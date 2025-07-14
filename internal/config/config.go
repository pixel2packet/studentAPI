package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Addr string `yaml:"addr" env-required:"true"`
}

// env-default:"production"
// Defines a struct named `HTTPServer`, with a single field `Addr` (like `:8080` for the port).
type Config struct {
	Env         string `yaml:"env" env:"ENV" env-required:"true"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	// StoragePath string `yaml:"storage_path" env-default:"./storage.db"`
	// In above case error will be thorwn since path is not set "Config path is not set"
	// If no tag used "env-req/default" than Uses zero-value (e.g., empty string, 0)
	HTTPServer `yaml:"http_server"`
}

// | Tag                   | Purpose                                     |
// | --------------------- | ------------------------------------------- |
// | `yaml:"..."`          | Maps Go fields to YAML keys.                |
// | `env:"..."`           | Loads from a specific environment variable. |
// | `env-default:"..."`   | Default value if env var is not set.        |
// | `env-required:"true"` | Crashes if env var is missing.              |

// create multiple config files like:
// config/local.yaml
// config/staging.yaml
// config/production.yaml
// Then load them based on a flag or ENV variable:
// env := os.Getenv("ENV")
// configPath := fmt.Sprintf("config/%s.yaml", env)

func MustLoad() *Config {
	var configPath string

	configPath = os.Getenv("CONFIG_PATH")
	// fmt.Println("CONFIG Path:", configPath)

	if configPath == "" { // -config-path xyz
		flags := flag.String("config", "", "path to the confiuguration file")
		flag.Parse()
		// fmt.Println("flag config:", *flags)

		configPath = *flags

		// If no `CONFIG_PATH` env var was set, it checks for a **command-line flag** like:
		// " go run main.go -config=config/local.yaml"
		if configPath == "" {
			log.Fatal("Config path is not set")
		}
		// This runs before ReadConfig, so it ensures the app doesn’t even try to read a missing config file.
	}

	// Checks if the file actually exists. If not → crash with an error.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("can not read config file: %s", err.Error())
	}

	// Returns the filled config to use in your main program.
	return &cfg

	// Example Config File (`config.yml`)
	// env: development
	// storage_path: ./data/db.json
	// http_server:
	//   addr: ":8080"
}
