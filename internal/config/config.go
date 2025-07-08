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
	Env         string `yaml:"env" env:"ENV" env-required:"true" env-default:"production"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

func MustLoad() *Config {
	var configPath string

	configPath = os.Getenv("CONFIG_PATH")

	if configPath == "" { // -config-path xyz
		flags := flag.String("config", "", "path to the confiuguration file")
		flag.Parse()

		configPath = *flags

		// If no `CONFIG_PATH` env var was set, it checks for a **command-line flag** like:
		// " go run main.go -config=./config.yml "
		if configPath == "" {
			log.Fatal("Config path is not set")
		}
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
