# Telegram Image to R2 Bot

A Telegram bot that automatically uploads images to Cloudflare R2 storage and returns public URLs.

## Features

- Accepts images sent as photos or documents
- Uploads files to Cloudflare R2
- Returns public URLs for uploaded files
- Owner-only access control

## Environment Variables

- `TELEGRAM_BOT_TOKEN`: Your Telegram bot token from BotFather
- `OWNER_ID`: Telegram user ID of the bot owner, you can get own from @myidbot
- `CF_ACCOUNT_ID`: Cloudflare account ID
- `CF_ACCESS_KEY_ID`: Cloudflare R2 access key ID
- `CF_ACCESS_KEY_SECRET`: Cloudflare R2 access key secret
- `CF_BUCKET_NAME`: Name of your R2 bucket

## Building and Running

### Using Docker

```bash
docker build -t telegram-r2-bot .
