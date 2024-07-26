package config

import (
	"fmt"
	"os"
)

func newSERVER() *SERVER {
	return &SERVER{
		host: os.Getenv("SERVER_HOST"),
		port: os.Getenv("SERVER_PORT"),
	}
}

type SERVER struct {
	host string
	port string
}

func (c *SERVER) ServerURL() string {
	return fmt.Sprintf("%s:%s", c.host, c.port)
}
