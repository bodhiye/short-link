package main

import (
	"context"
	"log"
	"os"
	"strconv"
)

type Env struct {
	S Storage
}

func getEnv(ctx context.Context) *Env {
	addr := os.Getenv("APP_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	passwd := os.Getenv("APP_REDIS_PASSWD")
	if passwd == "" {
		passwd = ""
	}

	dbS := os.Getenv("APP_REDIS_DB")
	if dbS == "" {
		dbS = "0"
	}

	db, err := strconv.Atoi(dbS)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("connect to redis (add: %s password: %s db: %s)", addr, passwd, db)

	r := NewRedisCli(ctx, addr, passwd, db)
	return &Env{S: r}
}
