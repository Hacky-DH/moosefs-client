package mfscli

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

func write(w io.Writer, data ...interface{}) {
	for _, d := range data {
		binary.Write(w, binary.BigEndian, d)
	}
}

func read(r io.Reader, data ...interface{}) {
	for _, d := range data {
		binary.Read(r, binary.BigEndian, d)
	}
}

func Pack(data ...interface{}) []byte {
	buf := new(bytes.Buffer)
	for _, d := range data {
		switch v := d.(type) {
		case []byte:
			buf.Write(v)
		case string:
			buf.WriteString(v)
		case *string:
			buf.WriteString(*v)
		case int:
			write(buf, int32(v))
		case *int:
			write(buf, int32(*v))
		case uint:
			write(buf, uint32(v))
		case *uint:
			write(buf, uint32(*v))
		default:
			write(buf, v)
		}
	}
	return buf.Bytes()
}

func UnPack(in []byte, out ...interface{}) {
	reader := bytes.NewReader(in)
	read(reader, out...)
}

func PackCmd(cmd uint32, data ...interface{}) []byte {
	size := 0
	for _, d := range data {
		switch v := d.(type) {
		case string:
			size += len(v)
		case *string:
			size += len(*v)
		case int, *int, uint, *uint:
			size += 4
		default:
			size += binary.Size(d)
		}
	}
	args := make([]interface{}, 0)
	args = append(args, cmd, uint32(size))
	args = append(args, data...)
	return Pack(args...)
}

func UnPackCmd(in []byte, out ...interface{}) (cmd uint32, err error) {
	reader := bytes.NewReader(in)
	var size uint32
	read(reader, &cmd, &size)
	if int(size) != reader.Len() {
		msg := fmt.Sprintf("cmd size %d not match with %d", size, reader.Len())
		err = errors.New(msg)
		return
	}
	read(reader, out...)
	return
}
