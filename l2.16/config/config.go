package config

import "time"

type Config struct {
	MaxDepth         int
	MaxConcurrent    int
	Timeout          time.Duration
	UserAgent        string
	OutputDir        string
	RespectRobotsTxt bool
}