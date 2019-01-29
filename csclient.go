package mfscli

import (
	"net"
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
