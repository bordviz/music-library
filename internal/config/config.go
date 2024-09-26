package config

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env            string `yaml:"env" env-required:"true"`
	MigrationsPath string `yaml:"migrations_path" env-required:"true"`
	Database       `yaml:"database" env-required:"true"`
	HTTPServer     `yaml:"http_server" env-required:"true"`
	LibraryServer  `yaml:"library_server" env-required:"true"`
}

type Database struct {
	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	User     string        `yaml:"user"`
	Name     string        `yaml:"name"`
	Password string        `yaml:"password"`
	Timeout  time.Duration `yaml:"timeout" env-required:"true"`
	Attempts int           `yaml:"attempts" env-required:"true"`
	Delay    time.Duration `yaml:"delay" env-required:"true"`
}

type HTTPServer struct {
	Port        int           `yaml:"port" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-reqired:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-required:"true"`
}

type LibraryServer struct {
	Protocol string `yaml:"protocol" env-required:"true"`
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
}

func MustLoad() *Config {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println(".env file not found")
		os.Exit(1)
	}

	configPath := strings.TrimSpace(os.Getenv("CONFIG_PATH"))
	dbUser, dbPassword, dbName, dbHost, dbPort := strings.TrimSpace(os.Getenv("POSTGRES_USER")),
		strings.TrimSpace(os.Getenv("POSTGRES_PASSWORD")),
		strings.TrimSpace(os.Getenv("POSTGRES_DB")),
		strings.TrimSpace(os.Getenv("POSTGRES_HOST")),
		strings.TrimSpace(os.Getenv("POSTGRES_PORT"))

	if slices.Contains([]string{configPath, dbUser, dbPassword, dbName, dbHost, dbPort}, "") {
		fmt.Println("missing required environment variables in .env file")
		os.Exit(1)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("config file %s not found", configPath)
		os.Exit(1)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		fmt.Printf("error reading config file: %s", err.Error())
		os.Exit(1)
	}

	cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.Host = dbUser, dbPassword, dbName, dbHost
	var err error
	cfg.Database.Port, err = strconv.Atoi(dbPort)
	if err != nil {
		fmt.Printf("invalid POSTGRES_PORT value: %s", dbPort)
		os.Exit(1)
	}

	return &cfg

}
