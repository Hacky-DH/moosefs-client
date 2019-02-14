package mfscli

import (
	"flag"
	"testing"
)

func init() {
	flag.Set("logtostderr", "true")
	flag.Set("v", "10")
	flag.Set("H", "127.0.0.1")
	flag.Set("P", "")
	flag.Parse()
}

func TestWrite(t *testing.T) {
	t.Skip()
	c, err := NewCLient()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	f, err := c.OpenOrCreate("testwfile")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("hello mfs"), 0)
	if err != nil {
		t.Fatal(err)
	}
	c.Unlink("testwfile")
}
