package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/0x0FACED/link-saver-api/config"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}
type RedisLink struct {
	ID   string
	Link string
}

func New(cfg config.RedisConfig) *Redis {
	client := redis.NewClient(
		&redis.Options{
			Addr: cfg.Host + ":" + cfg.Port,
		},
	)
	return &Redis{
		client: client,
	}
}

func (r *Redis) SaveLink(ctx context.Context, userId int64, url, originalURL string, urlID int32) error {
	key := fmt.Sprintf("links:%d:%s", userId, originalURL)
	value := fmt.Sprintf("%d:%s", urlID, url)

	err := r.client.SetEx(ctx, key, value, 24*time.Hour).Err()
	if err != nil {
		return err
	}

	globalKey := fmt.Sprintf("links:%d:urls", userId)

	_, err = r.client.SAdd(ctx, globalKey, originalURL).Result()
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) GetOriginalURL(ctx context.Context, userId int64, generatedLink string) (string, error) {
	globalKey := fmt.Sprintf("links:%d:urls", userId)

	urls, err := r.client.SMembers(ctx, globalKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}

	for _, originalURL := range urls {
		key := fmt.Sprintf("links:%d:%s", userId, originalURL)

		value, err := r.client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return "", err
		}

		parts := strings.SplitN(value, ":", 2)
		if len(parts) == 2 && parts[1] == generatedLink {
			return originalURL, nil
		}
	}

	return "", nil
}

func (r *Redis) GetLink(ctx context.Context, userId int64, originalURL string) (*RedisLink, error) {
	key := fmt.Sprintf("links:%d:%s", userId, originalURL)

	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, err
		}
		return nil, err
	}

	parts := strings.SplitN(value, ":", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid value format in Redis")
	}

	return &RedisLink{
		ID:   parts[0],
		Link: parts[1],
	}, nil
}

func (r *Redis) GetLinks(ctx context.Context, userId int64) ([]*RedisLink, error) {
	globalKey := fmt.Sprintf("links:%d:urls", userId)

	originalURLs, err := r.client.SMembers(ctx, globalKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, err
		}
		return nil, err
	}

	links := make([]*RedisLink, 0, len(originalURLs))

	for _, originalURL := range originalURLs {
		key := fmt.Sprintf("links:%d:%s", userId, originalURL)
		value, err := r.client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}

		parts := strings.SplitN(value, ":", 2)
		if len(parts) < 2 {
			continue
		}

		links = append(links, &RedisLink{
			ID:   parts[0],
			Link: parts[1],
		})
	}

	return links, nil
}

func (r *Redis) DeleteLink(ctx context.Context, userId int64, originalURL string) error {
	key := fmt.Sprintf("links:%d:%s", userId, originalURL)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	globalKey := fmt.Sprintf("links:%d:urls", userId)
	_, err = r.client.SRem(ctx, globalKey, originalURL).Result()
	if err != nil {
		return err
	}

	return nil
}
