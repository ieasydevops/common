package options

import "time"

type ServerRunOptions struct {
	Name      string
	AliveTime time.Duration
}

func NewServerRunOptions() *ServerRunOptions {
	s := ServerRunOptions{
		Name:"mock cmd server ",
		AliveTime:  1 * time.Second,
	}
	return &s
}

