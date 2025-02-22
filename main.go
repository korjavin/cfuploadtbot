package main

import (
	"bytes"
	"io"

	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Get environment variables
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	owner, err := strconv.ParseInt(os.Getenv("OWNER_ID"), 10, 64)
	if err != nil {
		log.Fatal("Invalid OWNER_ID")
	}

	// Cloudflare R2 credentials
	accountId := os.Getenv("CF_ACCOUNT_ID")
	accessKeyId := os.Getenv("CF_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("CF_ACCESS_KEY_SECRET")
	bucketName := os.Getenv("CF_BUCKET_NAME")

	// Initialize Telegram bot
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize S3 client for Cloudflare R2
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.NewFromConfig(cfg)

	// Set up updates channel
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Printf("Bot started")

	// Handle updates
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Check if message is from owner
		if update.Message.From.ID != owner {
			continue
		}

		// Handle photo messages
		if update.Message.Photo != nil || update.Message.Document != nil {
			var fileID string
			var fileName string

			if update.Message.Photo != nil {
				// Get the largest photo size
				photos := update.Message.Photo
				photo := photos[len(photos)-1]
				fileID = photo.FileID
				fileName = fmt.Sprintf("%d.jpg", time.Now().Unix())
			} else if update.Message.Document != nil {
				fileID = update.Message.Document.FileID
				fileName = update.Message.Document.FileName
			}

			// Get file URL from Telegram
			file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
			if err != nil {
				log.Printf("Error getting file: %v", err)
				continue
			}

			// Download file
			resp, err := http.Get(file.Link(botToken))
			if err != nil {
				log.Printf("Error downloading file: %v", err)
				continue
			}

			// Read the entire response body into a buffer
			fileData, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				log.Printf("Error reading file: %v", err)
				continue
			}

			// Upload to R2 using the buffer
			contentType := resp.Header.Get("Content-Type")
			_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
				Bucket:      &bucketName,
				Key:         &fileName,
				Body:        bytes.NewReader(fileData),
				ContentType: &contentType,
			})
			if err != nil {
				log.Printf("Error uploading to R2: %v", err)
				continue
			}

			// Generate public URL
			publicURL := fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", accountId, fileName)

			// Send response
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, publicURL)
			bot.Send(msg)
		}
	}
}
