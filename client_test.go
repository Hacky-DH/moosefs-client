package mfscli

import (
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	c := NewClient("127.0.0.1")
	if err := c.Connect(); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
	c.Close()
}
