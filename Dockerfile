# Use Node.js LTS version as the base image
FROM node:20-alpine

# Set working directory
WORKDIR /app

# Copy package.json and package-lock.json files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy CA certificate if it exists
COPY ca-cert.pem ./ca-cert.pem

# Copy source code
COPY . .

# Expose the port defined in .env
EXPOSE 8080

# Start the application
CMD ["npm", "start"]
