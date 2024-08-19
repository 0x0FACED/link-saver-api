package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x0FACED/link-saver-api/config"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

type RedisLink struct {
	Username string
	Link     string
}

func New(cfg config.RedisConfig) *Redis {
	// TODO: add more options
	client := redis.NewClient(
		&redis.Options{
			Addr: cfg.Host + ":" + cfg.Port,
		},
	)

	return &Redis{
		client: client,
	}
}

func (r *Redis) SaveLink(ctx context.Context, username, url string) error {
	key := fmt.Sprintf("links:%s", username)
	_, err := r.client.HSet(ctx, key, url, 24*time.Hour).Result()
	if err != nil {
		log.Println("error happened: ", err)
		return err
	}
	return nil
}

func (r *Redis) GetLink(ctx context.Context, username, url string) (*RedisLink, error) {
	key := fmt.Sprintf("links:%s", username)
	res, err := r.client.HGet(ctx, key, url).Result()
	if err == redis.Nil {
		log.Println("key doesnt exist")
		return nil, err
	}

	if err != nil {
		log.Println("error happened: ", err)
		return nil, err
	}

	log.Println("Result: ", res)

	link := &RedisLink{
		Username: username,
		Link:     res,
	}
	return link, nil
}

func (r *Redis) GetLinks(ctx context.Context, username string) ([]*RedisLink, error) {
	key := fmt.Sprintf("links:%s", username)
	res, err := r.client.HGetAll(ctx, key).Result()
	if err == redis.Nil {
		log.Println("key doesnt exist")
		return nil, err
	}

	if err != nil {
		log.Println("error happened: ", err)
		return nil, err
	}

	log.Println("Result: ", res)

	links := make([]*RedisLink, 0, len(res))

	for k, v := range res {
		links = append(links, &RedisLink{Username: k, Link: v})
	}

	return links, nil
}

func (r *Redis) DeleteLink(ctx context.Context, username, url string) error {
	key := fmt.Sprintf("links:%s", username)

	res, err := r.client.HDel(ctx, key, url).Result()
	if err == redis.Nil {
		log.Println("key doesnt exist")
		return err
	}
	if err != nil {
		log.Println("error happened: ", err)
		return err
	}

	log.Println("deleted: ", res)

	return nil
}
