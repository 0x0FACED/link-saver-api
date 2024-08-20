package redis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/0x0FACED/link-saver-api/config"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}
type RedisLink struct {
	Description string
	Link        string
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

func (r *Redis) SaveLink(ctx context.Context, username, description, url string) error {
	key := fmt.Sprintf("links:%s", username)
	field := fmt.Sprintf("%s:%s", description, url)

	_, err := r.client.HSet(ctx, key, field, 24*time.Hour).Result()
	if err != nil {
		log.Println("Error saving link to Redis: ", err)
		return err
	}
	return nil
}

func (r *Redis) GetLink(ctx context.Context, username, description string) (*RedisLink, error) {
	key := fmt.Sprintf("links:%s", username)

	linksMap, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Println("Key doesn't exist in Redis")
			return nil, err
		}
		log.Println("Error retrieving link from Redis: ", err)
		return nil, err
	}

	for field, url := range linksMap {
		if strings.HasPrefix(field, description+":") {
			return &RedisLink{
				Description: description,
				Link:        url,
			}, nil
		}
	}

	return nil, redis.Nil
}

func (r *Redis) GetLinks(ctx context.Context, username string) ([]*RedisLink, error) {
	key := fmt.Sprintf("links:%s", username)

	linksMap, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Println("Key doesn't exist in Redis")
			return nil, err
		}
		log.Println("Error retrieving links from Redis: ", err)
		return nil, err
	}

	links := make([]*RedisLink, 0, len(linksMap))
	for field, url := range linksMap {
		parts := strings.SplitN(field, ":", 2)
		if len(parts) == 2 {
			links = append(links, &RedisLink{
				Description: parts[0],
				Link:        url,
			})
		}
	}

	return links, nil
}
func (r *Redis) DeleteLink(ctx context.Context, username, description string) error {
	key := fmt.Sprintf("links:%s", username)

	linksMap, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Println("Key doesn't exist in Redis")
			return err
		}
		log.Println("Error retrieving links from Redis: ", err)
		return err
	}

	for field := range linksMap {
		if strings.HasPrefix(field, description+":") {
			_, err := r.client.HDel(ctx, key, field).Result()
			if err != nil {
				log.Println("Error deleting link from Redis: ", err)
				return err
			}
			log.Println("Deleted link: ", field)
			break
		}
	}

	return nil
}
