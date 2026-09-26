package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
	TestDBName string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("erro ao carregar .env: %w", err)
	}
	cfg := &Config{
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		TestDBName: os.Getenv("TEST_DB_NAME"),
	}

	if cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("variáveis de ambiente do banco incompletas")
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

// função DSN para ser usada em testes, conecta com o banco booklist_test
func (c *Config) TestDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.TestDBName)
}
