package configs

import "github.com/spf13/viper"

var cfg *config

type config struct {
	API APIConfig
	DB  DBConf
	AI  AIConf
}

type APIConfig struct {
	Port string
}

type DBConf struct {
	Host     string
	Port     string
	User     string
	Pass     string
	Database string
}

type AIConf struct {
	OllamaHost   string
	DefaultModel string
}

func init() {
	viper.SetDefault("api.host", "localhost")
	viper.SetDefault("api.port", "9000")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("ai.ollama_host", "http://localhost:11434")
	viper.SetDefault("ai.default_model", "gemma3:1b")
}

func Load(path string) error {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(path) // pq o ideal é que esse arquivo com as variaveis sempre esteja ao lado do binário
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	cfg = new(config)

	cfg.API = APIConfig{
		Port: viper.GetString("api.port"),
	}

	cfg.DB = DBConf{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetString("database.port"),
		User:     viper.GetString("database.user"),
		Pass:     viper.GetString("database.pass"),
		Database: viper.GetString("database.name"),
	}

	cfg.AI = AIConf{
		OllamaHost:   viper.GetString("ai.ollama_host"),
		DefaultModel: viper.GetString("ai.default_model"),
	}

	return nil
}

func GetDB() DBConf {
	return cfg.DB
}

func GetServerPort() string {
	return cfg.API.Port
}

func GetAI() AIConf {
	return cfg.AI
}
