package compose

import (
	"context"
)

// ServiceConfig defines the configuration for a Docker Compose service
type ServiceConfig struct {
	ConfigFunc          func() string                        // Generate service YAML
	AfterComposeCreated AfterComposeCreated                  // Post-creation actions
	BeforeServices      func() string                        // Add templates/anchors
	CheckOtherServices  func([]string, bool) (string, error) // Dependency resolution
}

// AfterComposeCreated is a function that runs after the compose file is created
type AfterComposeCreated func(ctx context.Context, compose string, autoconfirm bool) error

// ServiceConfigs is the registry of available services
var ServiceConfigs = map[string]ServiceConfig{
	"MongoDB": {
		ConfigFunc: generateMongoDBConfig,
	},
	"MongoDB with Replica Set": {
		ConfigFunc:          generateMongoDBReplicaConfig,
		AfterComposeCreated: createMongoKeyfile,
	},
	"PostgreSQL": {
		ConfigFunc: generatePostgreSQLConfig,
	},
	"Redis": {
		ConfigFunc: generateRedisConfig,
	},
	"Nginx": {
		ConfigFunc:          generateNginxConfig,
		AfterComposeCreated: createNginxConfigFile,
	},
	"N8N": {
		ConfigFunc:         generateN8NConfig,
		CheckOtherServices: checkN8NDependencies,
		BeforeServices:     addN8NTemplates,
	},
	"ImgProxy": {
		ConfigFunc: generateImgProxyConfig,
	},
	"MySQL": {
		ConfigFunc: generateMySQLConfig,
	},
	"MariaDB": {
		ConfigFunc: generateMariaDBConfig,
	},
	"Memcached": {
		ConfigFunc: generateMemcachedConfig,
	},
	"RabbitMQ": {
		ConfigFunc: generateRabbitMQConfig,
	},
	"Elasticsearch": {
		ConfigFunc: generateElasticsearchConfig,
	},
	"MinIO": {
		ConfigFunc: generateMinIOConfig,
	},
	// Add more services here as needed
}

// Helper function to check if a service is in the list
func hasService(services []string, serviceName string) bool {
	for _, s := range services {
		if s == serviceName {
			return true
		}
	}
	return false
}
