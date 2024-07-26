package config

import (
	"fmt"
	"os"
	"strconv"
)

type Conf struct {
	RMQ     *RMQ
	REDIS   *REDIS
	SERVER  *SERVER
	REQUEST *REQUEST
	DB      *Database
}

func NewConf() *Conf {
	return &Conf{
		RMQ:     newRMQ(),
		REDIS:   newREDIS(),
		SERVER:  newSERVER(),
		REQUEST: newREQUEST(),
		DB:      newDatabase(),
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		val = fallback
	}
	return val
}

func getEnvInt(key string, fallback int) int {
	rawVal := os.Getenv(key)
	val, err := strconv.Atoi(rawVal)
	if err != nil {
		fmt.Fprintf(os.Stderr, "environment %s required number type: %+v", key, err)
		val = fallback
	}
	return val
}
