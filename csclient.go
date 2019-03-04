package mfscli

/*
MIT License

Copyright (c) 2019 DHacky
*/

import (
	"fmt"
	"github.com/golang/glog"
	"hash/crc32"
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
	Version    uint32
	CSItems    CSItemMap
}

func NewCSClient(t *CSItem) (c *CSClient, err error) {
	c = new(CSClient)
	addr := t.addr()
	var conn net.Conn
	for i := 0; i < TCP_RETRY_TIMES; i++ {
		conn, err = net.DialTimeout("tcp", addr, TCP_CONNECT_TIMEOUT)
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
	c.conn.SetDeadline(time.Now().Add(TCP_RW_TIMEOUT))
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
	c.conn.SetDeadline(time.Now().Add(TCP_RW_TIMEOUT))
	n, err = io.ReadFull(c.conn, buf)
	if err != nil {
		c.Close()
	}
	return
}

func (d *CSData) Write(buf []byte, off uint64) (n uint32, err error) {
	if len(d.CSItems) == 0 {
		err = fmt.Errorf("no chunkserver found")
		return
	}
	for _, cs := range d.CSItems {
		var c *CSClient
		c, err = _cspool.Get(cs)
		if err != nil {
			return
		}
		css := []interface{}{
			d.ProtocolId,
			d.ChunkId,
			d.Version,
		}
		for _, _cs := range d.CSItems {
			css = append(css, _cs.Ip)
			css = append(css, _cs.Port)
		}
		msg := PackCmd(CLTOCS_WRITE, css...)
		if err = c.Send(msg); err != nil {
			err = fmt.Errorf("send write to cs error %v", err)
			return
		}
		var wid uint32 = 1
		pos := uint16((off & MFSCHUNKMASK) >> MFSBLOCKBITS)
		from := uint16(off & MFSBLOCKMASK)
		size := uint32(len(buf))
		for size > 0 {
			sz := MFSBLOCKSIZE - uint32(from)
			if sz > size {
				sz = size
			}
			glog.V(20).Infof("csclient write block buf[%d:%d] wid %d pos %d from %d",
				n, sz, wid, pos, from)
			err = d.WriteBlock(c, wid, pos, from, buf[n:n+sz])
			if err != nil {
				return
			}
			size -= sz
			n += sz
			pos += 1
			from = 0
			wid += 1
		}
		msg = PackCmd(CLTOCS_WRITE_FINISH, d.ChunkId, d.Version)
		if err = c.Send(msg); err != nil {
			err = fmt.Errorf("send write finish to cs error %v", err)
			return
		}
		// just write to one cs
		return
	}
	return
}

func (d *CSData) WriteBlock(c *CSClient, wid uint32, blockNum, off uint16,
	buf []byte) (err error) {
	crc := crc32.ChecksumIEEE(buf)
	msg := PackCmd(CLTOCS_WRITE_DATA, d.ChunkId, wid, blockNum, off,
		len(buf), crc, buf)
	if err = c.Send(msg); err != nil {
		err = fmt.Errorf("send data to cs error %v", err)
		return
	}
	rbuf := make([]byte, 21)
	var rcmd, size uint32 = ANTOAN_NOP, 4
	for rcmd == ANTOAN_NOP && size == 4 {
		n, e := c.Recv(rbuf)
		if e != nil {
			err = fmt.Errorf("recv from cs error %v", e)
			return
		}
		if n < 21 {
			err = fmt.Errorf("recv from cs size is too short")
			return
		}
		UnPack(rbuf, &rcmd, &size)
	}
	if rcmd != CSTOCL_WRITE_STATUS {
		err = fmt.Errorf("recv from cs bad command %d", rcmd)
		return
	}
	var cid uint64
	var wrid uint32
	var status uint8
	UnPack(rbuf[8:], &cid, &wrid, &status)
	if status != 0 {
		err = fmt.Errorf("write block %s", MFSStrerror(status))
		return
	}
	if cid != d.ChunkId || wid != wrid {
		err = fmt.Errorf("recv from cs bad cid %d wid %d", cid, wrid)
		return
	}
	return
}
func (d *CSData) Read(buf []byte, off uint64) (n uint32, err error) {
	if len(d.CSItems) == 0 {
		err = fmt.Errorf("no chunkserver found")
		return
	}
	for _, cs := range d.CSItems {
		var c *CSClient
		c, err = _cspool.Get(cs)
		if err != nil {
			return
		}
		msg := PackCmd(CLTOCS_READ, d.ProtocolId, d.ChunkId, d.Version,
			uint32(off), uint32(len(buf)))
		if err = c.Send(msg); err != nil {
			err = fmt.Errorf("send read to cs error %v", err)
			continue
		}
		from := uint16(off & MFSBLOCKMASK)
		size := uint32(len(buf))
		var rs uint32
		for size > 0 {
			sz := MFSBLOCKSIZE - uint32(from)
			if sz > size {
				sz = size
			}
			rs, err = d.ReadBlock(c, buf[n:n+sz], off)
			if err != nil {
				break
			}
			if rs != sz {
				continue
			}
			size -= sz
			n += sz
			from = 0
			off += uint64(sz)
		}
		if err == nil {
			return
		}
	}
	return
}

func (d *CSData) ReadBlock(c *CSClient, buf []byte, off uint64) (n uint32, err error) {
	read := func(sz uint32) (rbuf []byte, err error) {
		rbuf = make([]byte, sz)
		if _, err = c.Recv(rbuf); err != nil {
			err = fmt.Errorf("read block recv from cs error %v", err)
			return
		}
		return
	}
	rbuf, err := read(8)
	if err != nil {
		return
	}
	var cmd, sz uint32
	UnPack(rbuf, &cmd, &sz)
	if cmd == CSTOCL_READ_STATUS {
		if sz != 9 {
			err = fmt.Errorf("read block status wrong sizei %d!=9", sz)
			return
		}
		rbuf, err = read(sz)
		if err != nil {
			return
		}
		var cid uint64
		var status uint8
		UnPack(rbuf, &cid, &status)
		if status != 0 {
			err = fmt.Errorf("read block status error %s", MFSStrerror(status))
			return
		}
		if cid != d.ChunkId {
			err = fmt.Errorf("read block status wrong cid %d!=%d", cid, d.ChunkId)
			return
		}
		glog.V(10).Infof("read block status ok")
	} else if cmd == CSTOCL_READ_DATA {
		if sz < 20 {
			err = fmt.Errorf("read block data wrong size %d<20", sz)
			return
		}
		rbuf, err = read(20)
		if err != nil {
			return
		}
		var cid uint64
		var rpos, roff uint16
		var rsz, crc uint32
		UnPack(rbuf, &cid, &rpos, &roff, &rsz, &crc)
		if cid != d.ChunkId {
			err = fmt.Errorf("read block data wrong cid %d!=%d", cid, d.ChunkId)
			return
		}
		if rsz != uint32(len(buf)) {
			err = fmt.Errorf("read block data wrong size %d!=%d", rsz, len(buf))
			return
		}
		if sz != 20+rsz {
			err = fmt.Errorf("read block data wrong size %d!=20+%d", sz, rsz)
			return
		}
		pos := uint16((off & MFSCHUNKMASK) >> MFSBLOCKBITS)
		if pos != rpos {
			err = fmt.Errorf("read block data wrong pos %d!=%d", pos, rpos)
			return
		}
		offset := uint16(off & MFSBLOCKMASK)
		if offset != roff {
			err = fmt.Errorf("read block data wrong off %d!=%d", offset, roff)
			return
		}
		rbuf, err = read(rsz)
		if err != nil {
			return
		}
		ccrc := crc32.ChecksumIEEE(rbuf)
		if ccrc != crc {
			err = fmt.Errorf("read block data wrong crc %d!=%d", crc, ccrc)
			return
		}
		copy(buf, rbuf)
		n = rsz
	} else {
		err = fmt.Errorf("read block unknown rcmd %d", cmd)
	}
	return
}
