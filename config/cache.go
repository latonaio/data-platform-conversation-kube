package config

import "os"

type REDIS struct {
	Address string
	Port    string
}

func newREDIS() *REDIS {
	return &REDIS{
		Address: os.Getenv("REDIS_HOST"),
		Port:    os.Getenv("REDIS_PORT"),
	}
}
