package mfscli

import (
	"flag"
	"testing"
	"time"
)

func init() {
	flag.Set("logtostderr", "true")
	flag.Set("v", "10")
	flag.Parse()
}

func TestConnect(t *testing.T) {
	t.Skip()
	c := NewClient("127.0.0.1")
	if err := c.Connect(); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
	c.Close()
}
