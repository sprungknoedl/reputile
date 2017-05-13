package model

import (
	"fmt"
	"net"
	"time"
)

type Entry struct {
	Source      string
	Domain      string
	IP          net.IP
	Last        time.Time
	Category    string
	Description string

	Err error
}

func SendError(err error) *Entry {
	return &Entry{Err: err}
}

func (e Entry) Key() string {
	return fmt.Sprintf("%s|%s|%s", e.Source, e.Domain, e.IP)
}
