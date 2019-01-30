package mfscli

import (
	"fmt"
	"github.com/golang/glog"
	"io"
	"net"
	"sync"
	"time"
)

// chunk server client
type CSClient struct {
	conn net.Conn
	addr *CSItem
	Version
}

type CSPool struct {
	pool map[string][]*CSClient
	sync.Mutex
}

type CSItem struct {
	Ip        uint32
	Port      uint16
	Version   uint32
	LabelMask uint32
}

func (t *CSItem) addr() string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", t.Ip>>24, 0xff&(t.Ip>>16),
		0xff&(t.Ip>>8), 0xff&t.Ip, t.Port)
}

type CSItemMap map[uint32]*CSItem

// reponse data from master
type CSData struct {
	ProtocolId uint8
	Length     uint64
	ChunkId    uint64
	Version    uint64
	CSItems    CSItemMap
}

func NewCSClient(t *CSItem) (c *CSClient, err error) {
	c = new(CSClient)
	addr := t.addr()
	var conn net.Conn
	for i := 0; i < 3; i++ {
		conn, err = net.DialTimeout("tcp", addr, time.Minute)
		if err == nil {
			c.conn = conn
			c.conn.SetDeadline(time.Now().Add(time.Minute))
			break
		}
		glog.V(8).Infof("connect chunk master error: %v retry #%d", err, i+1)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil {
		return
	}
	c.addr = t
	c.Version = GetVersion(t.Version)
	glog.V(8).Infof("connect chunk master %s successfully", addr)
	return
}

func NewCSPool() *CSPool {
	return &CSPool{
		pool: make(map[string][]*CSClient),
	}
}

var _cspool *CSPool

func init() {
	_cspool = NewCSPool()
}

func (p *CSPool) Close() {
	for _, cs := range p.pool {
		for _, c := range cs {
			c.Close()
		}
	}
	_cspool = NewCSPool()
}

func (p *CSPool) Get(t *CSItem) (c *CSClient, err error) {
	addr := t.addr()
	p.Lock()
	cs, ok := p.pool[addr]
	if ok && len(cs) > 0 {
		//take the last one
		p.pool[addr] = nil
		p.pool[addr] = cs[:len(cs)-1]
		p.Unlock()
		c = cs[len(cs)-1]
		return
	}
	p.Unlock()
	return NewCSClient(t)
}

func (p *CSPool) Put(c *CSClient) {
	if c == nil {
		return
	}
	p.Lock()
	defer p.Unlock()
	addr := c.conn.RemoteAddr().String()
	cs, ok := p.pool[addr]
	if !ok {
		cs = nil
	}
	p.pool[addr] = append(cs, c)
}

func (c *CSClient) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

func (c *CSClient) Send(msg []byte) error {
	if c.conn == nil {
		return fmt.Errorf("connection to chunkserver is lost")
	}
	startSend := 0
	for startSend < len(msg) {
		sent, err := c.conn.Write(msg[startSend:])
		if err != nil {
			c.Close()
			return err
		}
		startSend += sent
	}
	return nil
}

func (c *CSClient) Recv(buf []byte) (n int, err error) {
	if c.conn == nil {
		err = fmt.Errorf("connection to chunkserver is lost")
		return
	}
	n, err = io.ReadFull(c.conn, buf)
	if err != nil {
		c.Close()
	}
	return
}

func (d *CSData) Write(buf []byte, off uint32) (n int, err error) {
	//_cspool.Get()
	return
}
