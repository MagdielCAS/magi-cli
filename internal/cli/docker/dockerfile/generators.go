package dockerfile

import "fmt"

// GenerateGoDockerfile generates a Dockerfile for a Go project
func GenerateGoDockerfile(port string) string {
	if port == "" {
		port = "8080"
	}
	return fmt.Sprintf(`FROM golang:alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .
CMD ["./main"]
EXPOSE %s`, port)
}

// GenerateNextNuxtDockerfile generates a Dockerfile for Next.js or Nuxt.js projects
func GenerateNextNuxtDockerfile(framework, port, buildPath string) string {
	if port == "" {
		port = "3000"
	}

	// Multi-stage build with optimization
	return fmt.Sprintf(`FROM node:20.7-bookworm-slim AS base
COPY .env /app/.env
WORKDIR /app

FROM base AS prod-deps
COPY package.json package-lock.json ./
RUN npm ci --only=production

FROM base AS build
COPY . .
RUN npm install && npm run build

FROM base
COPY --from=prod-deps /app/node_modules ./node_modules
COPY --from=build /app/%s ./%s
USER node
EXPOSE %s
CMD ["npm", "run", "start"]`, buildPath, buildPath, port)
}

// GenerateNodeDockerfile generates a basic Dockerfile for Node.js projects
func GenerateNodeDockerfile(port string) string {
	if port == "" {
		port = "3000"
	}
	return fmt.Sprintf(`FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
EXPOSE %s
CMD ["npm", "start"]`, port)
}
