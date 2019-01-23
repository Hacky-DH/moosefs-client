package mfscli

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/golang/glog"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Client struct {
	conn      net.Conn
	addr      string
	password  string
	uid       uint32
	gid       uint32
	sessionId uint32
	sync.Mutex
	Version
}

func NewClientPwd(addr, pwd string, heartbeat bool) (c *Client) {
	c = &Client{
		password: pwd,
		uid:      uint32(os.Getuid()),
		gid:      uint32(os.Getgid()),
	}
	ip := strings.Split(addr, ":")
	if len(ip) < 2 {
		//mfs client port
		c.addr = addr + ":9421"
	}
	if heartbeat {
		go c.heartbeat()
	}
	return
}

func NewClient(addr string) *Client {
	return NewClientPwd(addr, "", true)
}

func (c *Client) Connect() (err error) {
	if c.conn != nil {
		return
	}
	var conn net.Conn
	c.Lock()
	defer c.Unlock()
	for i := 0; i < 3; i++ {
		conn, err = net.DialTimeout("tcp", c.addr, time.Minute)
		if err == nil {
			c.conn = conn
			break
		}
		glog.V(8).Infof("connect mfs master error: %v retry #%d", err, i+1)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil {
		return
	}
	glog.V(8).Infof("connect mfs master %s successfully", c.addr)
	return
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

func (c *Client) heartbeat() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	nop := func() error {
		if err := c.Connect(); err != nil {
			return fmt.Errorf("connect error %s", err.Error())
		}
		c.Lock()
		defer c.Unlock()
		msg := PackCmd(ANTOAN_NOP, 0)
		if err := c.Send(msg); err != nil {
			return fmt.Errorf("connect error %s", err.Error())
		}
		glog.V(10).Info("sent heartbeat nop")
		return nil
	}
	for {
		select {
		case <-ticker.C:
			if err := nop(); err != nil {
				glog.Error(err)
				c.Close()
			}
		}
	}
}

func (c *Client) Send(msg []byte) error {
	if err := c.Connect(); err != nil {
		return fmt.Errorf("connect to mfs master error %s", err.Error())
	}
	c.Lock()
	defer c.Unlock()
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

func (c *Client) Recv(buf []byte) (n int, err error) {
	if err = c.Connect(); err != nil {
		fmt.Errorf("connect to mfs master error %s", err.Error())
		return
	}
	c.Lock()
	defer c.Unlock()
	n, err = io.ReadFull(c.conn, buf)
	if err != nil {
		c.Close()
	}
	return
}

func (c *Client) doCmd(cmd uint32, args ...interface{}) (r []byte, err error) {
	msg := PackCmd(cmd, args...)
	if err = c.Send(msg); err != nil {
		err = fmt.Errorf("send error %s", err.Error())
		return
	}
	buf := make([]byte, 8)
	var rcmd, size uint32 = ANTOAN_NOP, 4
	for rcmd == ANTOAN_NOP && size == 4 {
		_, err = c.Recv(buf)
		if err != nil {
			err = fmt.Errorf("cmd recv error %s", err.Error())
			return
		}
		read(bytes.NewBuffer(buf), &rcmd, &size)
	}
	if rcmd != cmd+1 {
		err = fmt.Errorf("mfs master cmd %d bad answer rcmd %d", cmd, rcmd)
		return
	}
	glog.V(10).Infof("command %d size %d", cmd, size)
	if size > 0 {
		buf = make([]byte, size)
		if _, err = c.Recv(buf); err != nil {
			err = fmt.Errorf("data recv error %s", err.Error())
			return
		}
		r = buf
	}
	return
}

func getStatus(buf []byte) (err error) {
	if len(buf) < 1 {
		err = fmt.Errorf("got wrong size %d<1 from mfsmaster", len(buf))
		return
	}
	var code uint8
	UnPack(buf, &code)
	if code != 0 {
		err = fmt.Errorf("got error from mfsmaster: %s", MFSStrerror(code))
		return
	}
	return
}

func (c *Client) CreateSession() (err error) {
	err = c.MasterVersion()
	if err != nil {
		return
	}
	var buf []byte
	if c.sessionId == 0 {
		pwFinal := make([]byte, 16)
		if len(c.password) > 0 {
			buf, err = c.doCmd(CLTOMA_FUSE_REGISTER, FUSE_REGISTER_BLOB_ACL,
				REGISTER_GETRANDOM)
			if err != nil {
				return
			}
			if len(buf) != 32 {
				err = fmt.Errorf("got wrong size %d!=32 from mfsmaster", len(buf))
				return
			}
			pwMd5 := md5.Sum([]byte(c.password))
			md := md5.New()
			md.Write(buf[:16])
			md.Write(pwMd5[:])
			md.Write(buf[16:])
			pwFinal = md.Sum(nil)
		}
		buf, err = c.doCmd(CLTOMA_FUSE_REGISTER, FUSE_REGISTER_BLOB_ACL,
			REGISTER_NEWSESSION, c.Version, 2, "/\000", 2, "/\000", pwFinal)
	} else {
		buf, err = c.doCmd(CLTOMA_FUSE_REGISTER, FUSE_REGISTER_BLOB_ACL,
			REGISTER_RECONNECT, c.sessionId, c.Version)
	}
	if err != nil {
		return
	}
	if len(buf) == 1 {
		err = getStatus(buf)
		if err != nil {
			return
		}
		if c.sessionId != 0 {
			glog.V(8).Infof("reuse session id %d", c.sessionId)
			return
		}
	}
	if len(buf) < 43 {
		err = fmt.Errorf("got wrong size %d<43 from mfsmaster", len(buf))
		return
	}
	var id uint32
	UnPack(buf[4:], &id)
	if 0 != c.sessionId {
		c.CloseSession()
	}
	c.sessionId = id
	glog.V(8).Infof("create new session id %d", id)
	return
}

func (c *Client) CloseSession() (err error) {
	if c.sessionId == 0 {
		return
	}
	buf, err := c.doCmd(CLTOMA_FUSE_REGISTER, FUSE_REGISTER_BLOB_ACL,
		REGISTER_CLOSESESSION, c.sessionId)
	if err != nil {
		return
	}
	err = getStatus(buf)
	if err != nil {
		return
	}
	glog.V(8).Infof("close session id %d", c.sessionId)
	c.sessionId = 0
	return
}

func (c *Client) RemoveSession(sessionId uint32) (err error) {
	buf, err := c.doCmd(CLTOMA_SESSION_COMMAND, uint8(0), sessionId)
	if err != nil {
		return
	}
	err = getStatus(buf)
	if err != nil {
		return
	}
	glog.V(8).Infof("remove session id %d", sessionId)
	return
}
