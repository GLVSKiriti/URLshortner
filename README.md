# GoShort
- An Implementation of URL shortner with GO Lang

## Teckstacks
- Go
- Go Fiber
- Redis
- Docker

## Features
- Users can create custom shorten URL
- Rate limiter to restrict users to exploit the API
- Containerized the whole application using Docker and Docker-compose for easy set-up
- Suitable checks for non-redundancy in URL creation

## Project Setup

- Clone the repository using git clone [<repo_url>](https://github.com/GLVSKiriti/URLshortner.git)
- `Cd URLshortner` Run this command
- `go mod tidy` Run this command to install all required golang dependencies
- `docker compose up -d` so it starts the servers for Go Fiber and and Redis database at ports 3000 and 6379 respectively
    `Note:` Make sure your PC is installed docker and docker compose
- Now you can test this GoShort API with Postman in your system
