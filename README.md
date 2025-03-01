# DigitalOcean Database Connection Tester

A simple Go web application to test connections to a DigitalOcean managed database and manage news articles.

## Setup

1. Clone this repository
2. Copy `.env.example` to `.env` and update with your database credentials
3. Download your DigitalOcean CA certificate:
   - Go to your database in the DigitalOcean control panel
   - Click on "Connection Details"
   - Download the CA certificate
   - Save it to the project directory as `ca-certificate.crt`
4. Build and run the application:

```bash
go build
./test-conn
```

## SSL/TLS Connection

This application supports secure connections to your DigitalOcean managed database:

1. Set `DB_SSL_MODE=require` in your `.env` file
2. Set `DB_CA_CERT_PATH` to the path of your downloaded CA certificate
3. Keep `DB_SKIP_VERIFY=false` for proper certificate verification

For testing purposes only, you can set `DB_SKIP_VERIFY=true` to bypass certificate validation.

## Features

- Test database connections using environment variables or custom parameters
- Browse news articles stored in your database
- Search for specific articles
- View article details
- Create, edit, and delete articles
- Upload images directly or use image URLs

## Article Management

The application provides a complete article management system:

- **Create Articles**: Add new articles with title, description, author, and optional images
- **Edit Articles**: Modify existing articles and update their images
- **Delete Articles**: Remove articles from the database
- **Image Support**: Upload images or use remote image URLs

## Database Migration

The application automatically sets up the database schema when first run. It creates:

- An `articles` table to store article content
- Appropriate columns for storing images directly in the database

## Project Structure

- `/db`: Database connection utilities and migrations
- `/models`: Data models for the application
- `/handlers`: HTTP request handlers
- `/templates`: HTML templates for the UI
- `/static`: Static assets like CSS files
