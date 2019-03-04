# moosefs-client

[![Build Status](https://travis-ci.com/Hacky-DH/moosefs-client.svg?branch=master)](https://travis-ci.com/Hacky-DH/moosefs-client)

go client for [moosefs](https://github.com/moosefs/moosefs) version 3.0.103

Most code logic is from source of moosefs, and also take examples from [wfxiang08](https://github.com/wfxiang08/go-mfsclient) and [xiexiao](https://github.com/xiexiao/gopark/tree/master/moosefs).

# Usage

install with
```
go get github.com/Hacky-DH/moosefs-client
```
simple usage
```go
c, err := mfs.NewClient()
if err != nil {
	glog.Fatal(err)
}
defer c.Close()
err = c.WriteFile(local_file, mfs_file)
if err != nil {
	glog.Fatal(err)
}
```
