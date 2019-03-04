package mfscli

/*
MIT License

Copyright (c) 2019 DHacky
*/

import (
	"crypto/md5"
	"flag"
	"math/rand"
	"os"
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
	c, err := NewClient()
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
	c, err := NewClient()
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
	c, err := NewClient()
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

func TestWriteReadFile(t *testing.T) {
	t.Skip()
	c, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	data := make([]byte, 0x04500000)
	rand.Read(data)
	lname := "/tmp/wrfilename889"
	rname := "/wrfilename889"
	wmd5 := md5.Sum(data)
	f, err := os.Create(lname)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	err = c.WriteFile(lname, rname)
	if err != nil {
		t.Fatal(err)
	}
	os.Remove(lname)
	//read
	err = c.ReadFile(rname, lname)
	if err != nil {
		c.Unlink(rname)
		t.Fatal(err)
	}
	c.Unlink(rname)
	f, err = os.Open(lname)
	if err != nil {
		t.Fatal(err)
	}
	data = make([]byte, 0x04500000)
	_, err = f.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if wmd5 != md5.Sum(data) {
		t.Fatal("md5 is not equal")
	}
	os.Remove(lname)
}
