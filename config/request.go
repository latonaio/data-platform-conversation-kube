package config

import (
	"fmt"
	"os"
)

func newREQUEST() *REQUEST {
	return &REQUEST{
		host: os.Getenv("NESTJS_DATA_CONNECTION_REQUEST_CONTROL_MANAGER_HOST"),
		port: os.Getenv("NESTJS_DATA_CONNECTION_REQUEST_CONTROL_MANAGER_PORT"),
	}
}

type REQUEST struct {
	host string
	port string
}

func (c *REQUEST) RequestURL() string {
	if c.port == "" {
		return fmt.Sprintf("http://%s/api/request", c.host)
	}
	return fmt.Sprintf("http://%s:%s/api/request", c.host, c.port)
}
