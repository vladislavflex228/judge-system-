package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbHost     string
	DbPort     string
	DbUser     string
	DbPassword string
	DbName     string
	ServerPort string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load() //Загружает все пары вида ключ - значение в OC , чтобы мы в getEnv смогли подгрузить их

	cfg := &Config{
		DbHost:     getEnv("DB_HOST", "localhost"),
		DbPort:     getEnv("DB_PORT", "5432"),
		DbUser:     getEnv("DB_USER", "postgres"),
		DbPassword: getEnv("DB_PASSWORD", "your_password_here"),
		DbName:     getEnv("DB_NAME", "judge_db"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}

	return cfg, nil

}

func getEnv(key, fallback string) string {
	if value, exist := os.LookupEnv(key); exist { //LookupEnv ос смотрит в окружение и пытается достать переменную с нужным ключом
		return value
	}

	return fallback
}

// Процесс и его окружение:
// Когда программа на Go запускается,операционная система выделяет для неё отдельный процесс.
// У каждого процесса в его адресном пространстве в ОЗУ есть изолированная таблица(массив строк),
// которая называется Environment Block(блок окружения).
// Как раз в этот Блок окружения я подгружаю переменные из файла .env
// После этого я могу смотреть в Блок окружения и доставать оттуда значения под определенными ключами.
