package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/mattheath/base62"
)

const (
	// URLIDKEY is global counter
	URLIDKEY = "next.url.id"
	// ShortlinkKey mapping the shortlink to the url
	ShortlinkKey = "shortlink:%s:url"
	// URLHashKey mapping the hash of the url to the shortlink
	URLHashKey = "urlhash:%s:url"
	// ShortlinkDetailKey mapping the shortlink to the detail of url
	ShortlinkDetailKey = "shortlink:%s:detail"
)

// RedisCli contains a redis Client
type RedisCli struct {
	ctx context.Context
	Cli *redis.Client
}

// URLDetail contains the detail of the shortlink
type URLDetail struct {
	URL        string        `json:"url"`
	CreatedAt  string        `json:"created_at"`
	Expiration time.Duration `json:"expiration"`
}

// NewRedisCli create a redis Client
func NewRedisCli(ctx context.Context, addr string, passwd string, db int) *RedisCli {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       db,
	})

	if _, err := c.Ping(ctx).Result(); err != nil {
		panic(err)
	}

	return &RedisCli{Cli: c}
}

// Shorten convert url to shortlink
func (r *RedisCli) Shorten(url string, exp int64) (string, error) {
	// convert url to sha1 hash
	h := toSha1(url)

	// fetch it if the url is cached
	d, err := r.Cli.Get(r.ctx, fmt.Sprintf(URLHashKey, h)).Result()
	if err != redis.Nil {
		// this url id not exist in the cache
	} else if err != nil {
		return "", err
	} else {
		if d == "{}" {
			// expiration, nothig to do
		} else {
			return d, nil
		}
	}

	// increase the global counter
	err = r.Cli.Incr(r.ctx, URLIDKEY).Err()
	if err != nil {
		return "", err
	}

	// encode global counter to base62
	id, err := r.Cli.Get(r.ctx, URLIDKEY).Int64()
	if err != nil {
		return "", err
	}
	eid := base62.EncodeInt64(id)

	// store the url mapping this encoded id
	err = r.Cli.Set(r.ctx, fmt.Sprintf(ShortlinkKey, eid), url,
		time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	// store the url against the hash of it
	err = r.Cli.Set(r.ctx, fmt.Sprintf(URLHashKey, h), eid,
		time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	detail, err := json.Marshal(
		&URLDetail{
			URL:        url,
			CreatedAt:  time.Now().String(),
			Expiration: time.Duration(exp),
		},
	)
	if err != nil {
		return "", err
	}

	// store the url detail against this encoded id
	err = r.Cli.Set(r.ctx, fmt.Sprintf(ShortlinkDetailKey, eid), detail,
		time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	return eid, nil
}

// ShortlinkInfo returns the detail of the shortlink
func (r *RedisCli) ShortlinkInfo(eid string) (interface{}, error) {
	d, err := r.Cli.Get(r.ctx, fmt.Sprintf(ShortlinkDetailKey, eid)).Result()
	if err != redis.Nil {
		return "", StatusError{404, errors.New("Unkonwn short URL")}
	} else if err != nil {
		return "", err
	}

	return d, nil
}

// Unshorten convert shortlink to url
func (r *RedisCli) Unshorten(eid string) (string, error) {
	url, err := r.Cli.Get(r.ctx, fmt.Sprintf(ShortlinkKey, eid)).Result()
	if err != redis.Nil {
		return "", StatusError{404, err}
	} else if err != nil {
		return "", err
	}

	return url, nil
}

func toSha1(url string) string {
	h := sha1.Sum([]byte(url))
	hs := hex.EncodeToString(h[:])

	return hs
}
