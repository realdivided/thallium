package main

import "time"

type Target struct {
	Name         string
	LocalAddress string
	Address      string
	Legacy       bool
	Period       duration
}

type DbStatusAdapter struct {
	Name     string
	Address  string
	Latency  int
	Online   int
	Max      int
	Time     int64
	Favicon  string
	Protocol int
	Version  string
	Motd     string
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

