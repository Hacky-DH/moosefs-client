package mfscli

import (
	"flag"
	"math/rand"
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
	data := make([]byte, 0x00020000)
	rand.Read(data)
	_, err = f.Write(data, 0)
	if err != nil {
		c.Unlink("testwfile")
		t.Fatal(err)
	}
	c.Unlink("testwfile")
}

func TestReadData(t *testing.T) {
	t.Skip()
	c, err := NewCLient()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	f, err := c.Open("testrfile")
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 0x08000000)
	_, err = f.Read(data, 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDir(t *testing.T) {
	t.Skip()
	c, err := NewCLient()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	d := "testdir888"
	err = c.Mkdir(d)
	if err != nil {
		t.Error(err)
	}
	_, err = c.Readdir(d)
	if err != nil {
		t.Error(err)
	}
	err = c.Rmdir(d)
	if err != nil {
		t.Error(err)
	}
}
