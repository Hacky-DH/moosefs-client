package mfscli

import (
	"fmt"
	"github.com/golang/glog"
	"net"
	"time"
)

// chunk server client
type CSClient struct {
	conn net.Conn
	addr *CSItem
	Version
}

type CSItem struct {
	Ip        uint32
	Port      uint16
	Version   uint32
	LabelMask uint32
}

type CSItemMap map[uint32]*CSItem

type CSData struct {
	ProtocolId uint8
	Length     uint64
	ChunkId    uint64
	Version    uint64
	CSItems    CSItemMap
}

func NewCSClient(t *CSItem) (c *CSClient, err error) {
	c = new(CSClient)
	addr := fmt.Sprintf("%d.%d.%d.%d:%d", t.Ip>>24, 0xff&(t.Ip>>16),
		0xff&(t.Ip>>8), 0xff&t.Ip, t.Port)
	var conn net.Conn
	for i := 0; i < 3; i++ {
		conn, err = net.DialTimeout("tcp", addr, time.Minute)
		if err == nil {
			c.conn = conn
			break
		}
		glog.V(8).Infof("connect chunk master error: %v retry #%d", err, i+1)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil {
		return
	}
	c.Version = GetVersion(t.Version)
	glog.V(8).Infof("connect chunk master %s successfully", addr)
	return
}
