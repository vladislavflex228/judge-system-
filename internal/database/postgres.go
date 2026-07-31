package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreatePool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)

	if err != nil {
		return nil, fmt.Errorf("error in ParsingConfig : %w", err)
	}

	// Конфиг - это чертеж , неживой объект , описывающий наш будущий пул , после его создания требуется настроить его оптимальными настройками

	config.MaxConns = 25                       // Макс 25 (горячих + активных + холодных) соединений 	10m-25m оптимально
	config.MinConns = 5                        // Мин 5 горячих соединений (готовые но не используются)	2m-5m оптимально
	config.MaxConnLifetime = 30 * time.Minute  // Принудительно переводит любое соединение в старое для обновления 30m - 1h оптимально
	config.MaxConnIdleTime = 10 * time.Minute  // Макс время простоя горячих соединений								5m - 15m оптимально
	config.HealthCheckPeriod = 1 * time.Minute // Выявление мертвых соединений(Соединения , разорванные физически , о которых Go еще не знает)

	config.ConnConfig.ConnectTimeout = 5 * time.Second // Количество времени на установку соединения

	pool, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		return nil, fmt.Errorf("failed to creat pgxpool : %w", err)
	}

	timeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)

	defer cancel()

	if err := pool.Ping(timeCtx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres : %w", err)
	}

	return pool, nil

}
