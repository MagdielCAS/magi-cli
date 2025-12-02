package compose

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/MagdielCAS/magi-cli/pkg/llm"
	"github.com/MagdielCAS/magi-cli/pkg/shared"
	"github.com/pterm/pterm"
)

// --- Database Services ---

func generateMongoDBConfig() string {
	return `  mongo:
    image: mongo:latest
    container_name: mongo
    restart: unless-stopped
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: password
    volumes:
      - mongodb_data:/data/db
    ports:
      - "27017:27017"`
}

func generateMongoDBReplicaConfig() string {
	return `  mongo:
    image: mongo:latest
    container_name: mongo
    restart: unless-stopped
    environment:
      MONGO_REPLICA_SET_MODE: primary
      MONGO_REPLICA_SET_NAME: rs0
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: password
    volumes:
      - ./keyfile:/keyfile
      - mongodb_data:/data/db
    ports:
      - "27017:27017"
    command: mongod --replSet rs0 --keyFile /keyfile --bind_ip_all

  init-replica-set:
    image: mongo:latest
    container_name: init-replica-set
    command: >
      bash -c "until mongosh mongo:27017/admin -u root -p password --eval 'rs.initiate({_id:\"rs0\",members:[{_id:0,host:\"mongo:27017\"}]});'
      do sleep 5; done"
    depends_on:
      - mongo`
}

func generatePostgreSQLConfig() string {
	return `  postgres:
    image: postgres:alpine
    container_name: postgres
    restart: unless-stopped
    ports:
      - "5432:5432"
    environment:
      POSTGRES_DB: mydb
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_storage:/var/lib/postgresql/data
    healthcheck:
      test: ['CMD-SHELL', 'pg_isready -h localhost -U user -d mydb']
      interval: 5s
      timeout: 5s
      retries: 10`
}

func generateMySQLConfig() string {
	return `  mysql:
    image: mysql:8.0
    container_name: mysql
    restart: unless-stopped
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: mydb
      MYSQL_USER: user
      MYSQL_PASSWORD: password
    volumes:
      - mysql_data:/var/lib/mysql`
}

func generateMariaDBConfig() string {
	return `  mariadb:
    image: mariadb:latest
    container_name: mariadb
    restart: unless-stopped
    ports:
      - "3306:3306"
    environment:
      MARIADB_ROOT_PASSWORD: rootpassword
      MARIADB_DATABASE: mydb
      MARIADB_USER: user
      MARIADB_PASSWORD: password
    volumes:
      - mariadb_data:/var/lib/mysql`
}

func generateRedisConfig() string {
	return `  redis:
    image: redis:alpine
    container_name: redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data`
}

func generateMemcachedConfig() string {
	return `  memcached:
    image: memcached:alpine
    container_name: memcached
    restart: unless-stopped
    ports:
      - "11211:11211"`
}

func generateRabbitMQConfig() string {
	return `  rabbitmq:
    image: rabbitmq:3-management-alpine
    container_name: rabbitmq
    restart: unless-stopped
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      RABBITMQ_DEFAULT_USER: user
      RABBITMQ_DEFAULT_PASS: password
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq`
}

func generateElasticsearchConfig() string {
	return `  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.1
    container_name: elasticsearch
    restart: unless-stopped
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data`
}

func generateMinIOConfig() string {
	return `  minio:
    image: minio/minio
    container_name: minio
    restart: unless-stopped
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data`
}

// --- Web/Proxy Services ---

func generateNginxConfig() string {
	return `  nginx:
    image: nginx:alpine
    container_name: nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certbot/conf:/etc/letsencrypt
      - ./certbot/www:/var/www/certbot
    depends_on:
      - app`
}

func generateImgProxyConfig() string {
	key, _ := generateSaltKey(64)
	salt, _ := generateSaltKey(64)
	return fmt.Sprintf(`  imgproxy:
    image: darthsim/imgproxy:latest
    container_name: imgproxy
    restart: unless-stopped
    ports:
      - "8081:8080"
    environment:
      IMGPROXY_KEY: '%s'
      IMGPROXY_SALT: '%s'
      IMGPROXY_ENABLE_WEBP_DETECTION: 'true'`, key, salt)
}

// --- Advanced Services ---

func generateN8NConfig() string {
	return `  n8n:
    image: n8nio/n8n
    container_name: n8n
    restart: always
    ports:
      - "5678:5678"
    environment:
      - N8N_HOST=localhost
      - N8N_PORT=5678
      - N8N_PROTOCOL=http
      - NODE_ENV=production
      - WEBHOOK_URL=http://localhost:5678/
      - GENERIC_TIMEZONE=UTC
      - DB_TYPE=postgresdb
      - DB_POSTGRESDB_HOST=postgres
      - DB_POSTGRESDB_PORT=5432
      - DB_POSTGRESDB_DATABASE=n8n
      - DB_POSTGRESDB_USER=user
      - DB_POSTGRESDB_PASSWORD=password
    volumes:
      - n8n_data:/home/node/.n8n
    depends_on:
      - postgres`
}

// --- Dependency Checks ---

func checkN8NDependencies(services []string, autoconfirm bool) (string, error) {
	if hasService(services, "PostgreSQL") {
		return "", nil
	}

	if !autoconfirm {
		confirmed := promptForConfirmation("PostgreSQL is required for N8N. Add it?")
		if !confirmed {
			return "", fmt.Errorf("PostgreSQL required for N8N")
		}
	} else {
		pterm.Info.Println("Auto-adding PostgreSQL dependency for N8N...")
	}

	return generatePostgreSQLConfig(), nil
}

func addN8NTemplates() string {
	// N8N doesn't strictly need templates in this simple config, but we can add anchors if needed
	return ""
}

// --- Post-Creation Actions ---

func createMongoKeyfile(ctx context.Context, compose string, autoconfirm bool) error {
	if !autoconfirm {
		if !promptForConfirmation("Create MongoDB keyfile?") {
			return nil
		}
	}

	key, err := generateSaltKey(756) // MongoDB keyfiles are typically longer
	if err != nil {
		return err
	}

	err = os.WriteFile("keyfile", []byte(key), 0400) // Read-only for owner
	if err != nil {
		return fmt.Errorf("failed to create keyfile: %w", err)
	}
	pterm.Success.Println("Created MongoDB keyfile")
	return nil
}

func createNginxConfigFile(ctx context.Context, compose string, autoconfirm bool) error {
	// Basic Nginx config
	config := `events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        server_name localhost;

        location / {
            proxy_pass http://app:3000; # Adjust upstream as needed
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
    }
}`

	// Validate/Optimize with AI
	pterm.Info.Println("Validating Nginx configuration with AI...")
	validatedConfig, err := validateNginxConfig(ctx, config, compose)
	if err == nil {
		config = validatedConfig
	} else {
		pterm.Warning.Printf("AI validation failed: %v. Using default config.\n", err)
	}

	err = os.WriteFile("nginx.conf", []byte(config), 0644)
	if err != nil {
		return fmt.Errorf("failed to create nginx.conf: %w", err)
	}
	pterm.Success.Println("Created nginx.conf")
	return nil
}

func validateNginxConfig(ctx context.Context, config, composeContent string) (string, error) {
	runtime, err := shared.BuildRuntimeContext()
	if err != nil {
		return "", err
	}

	builder := llm.NewServiceBuilder(runtime)
	service, err := builder.Build()
	if err != nil {
		return "", err
	}

	prompt := []llm.ChatMessage{
		{
			Role: "system",
			Content: `Expert Nginx configuration validator. Analyze and fix 
            indentation errors, misconfigurations, and reverse proxy rules. 
            Based on the docker-compose context provided, identify frontend/API services 
            and add appropriate proxy rules if missing.
			Return ONLY the valid nginx.conf content. Do not include markdown code blocks.`,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Docker Compose:\n%s\n\nNginx Config:\n%s", composeContent, config),
		},
	}

	req := llm.ChatCompletionRequest{
		Messages:  prompt,
		MaxTokens: 2048,
	}

	result, err := service.ChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	// Remove code blocks if present
	result = strings.TrimPrefix(result, "```nginx")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSuffix(result, "```")

	return result, nil
}

// --- Helpers ---

func generateSaltKey(length int) (string, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

func promptForConfirmation(question string) bool {
	result, _ := pterm.DefaultInteractiveConfirm.Show(question)
	return result
}
