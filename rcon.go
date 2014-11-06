package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"log"
	"fmt"
	"sync/atomic"
	"time"
)

const (
	SERVERDATA_AUTH = 3
	SERVERDATA_EXECCOMMAND = 2
	SERVERDATA_AUTH_RESPONSE = 0
	SERVERDATA_RESPONSE_VALUE = 2
)

const readBufferSize = 4110

// RCON Protocol specs can be found here: https://developer.valvesoftware.com/wiki/Source_RCON_Protocol
// Thanks to https://github.com/james4k/ (james4k) for some of the RCON functions

type RemoteConsole struct {
	conn      net.Conn
	readbuf   []byte
	readmu    sync.Mutex
	reqid     int32
	queuedbuf []byte
}

var (
	ErrAuthFailed          = errors.New("rcon: authentication failed")
	ErrInvalidAuthResponse = errors.New("rcon: invalid response type during auth")
	ErrUnexpectedFormat    = errors.New("rcon: unexpected response format")
	ErrResponseTooLong     = errors.New("rcon: response too long")
)

func (r *RemoteConsole) WriteData(data string, v ...interface{}) (requestId int, err error){
	buffer := fmt.Sprintf(data, v...)
	log.Printf("Sent(RCON): %s\n", buffer)
	return r.writeCmd(SERVERDATA_EXECCOMMAND, buffer)
}

func (r *RemoteConsole) Read() (response string, requestId int, err error) {
	var respType int
	var respBytes []byte
	respType, requestId, respBytes, err = r.readResponse(2 * time.Minute)
	if err != nil || respType != SERVERDATA_RESPONSE_VALUE {
		response = ""
		requestId = 0
	} else {
		response = string(respBytes)
	}
	return
}

func (r *RemoteConsole) Close() error {
	return r.conn.Close()
}

func newRequestId(id int32) int32 {
	if id&0x0fffffff != id {
		return int32((time.Now().UnixNano() / 100000) % 100000)
	}
	return id + 1
}

func (r *RemoteConsole) writeCmd(cmdType int32, str string) (int, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, 14+len(str)))
	reqid := atomic.LoadInt32(&r.reqid)
	reqid = newRequestId(reqid)
	atomic.StoreInt32(&r.reqid, reqid)

	binary.Write(buffer, binary.LittleEndian, int32(10+len(str)))
	binary.Write(buffer, binary.LittleEndian, int32(reqid))
	binary.Write(buffer, binary.LittleEndian, int32(cmdType))
	buffer.WriteString(str)
	binary.Write(buffer, binary.LittleEndian, byte(0))
	binary.Write(buffer, binary.LittleEndian, byte(0))

	r.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err := r.conn.Write(buffer.Bytes())
	return int(reqid), err
}

func (r *RemoteConsole) readResponse(timeout time.Duration) (int, int, []byte, error) {
	r.readmu.Lock()
	defer r.readmu.Unlock()

	r.conn.SetReadDeadline(time.Now().Add(timeout))
	var size int
	var err error
	if r.queuedbuf != nil {
		copy(r.readbuf, r.queuedbuf)
		size = len(r.queuedbuf)
		r.queuedbuf = nil
	} else {
		size, err = r.conn.Read(r.readbuf)
		if err != nil {
			return 0, 0, nil, err
		}
	}
	if size < 4 {
		// need the 4 byte packet size...
		s, err := r.conn.Read(r.readbuf[size:])
		if err != nil {
			return 0, 0, nil, err
		}
		size += s
	}

	var dataSize32 int32
	b := bytes.NewBuffer(r.readbuf[:size])
	binary.Read(b, binary.LittleEndian, &dataSize32)
	if dataSize32 < 10 {
		return 0, 0, nil, ErrUnexpectedFormat
	}

	totalSize := size
	dataSize := int(dataSize32)
	if dataSize > 4106 {
		return 0, 0, nil, ErrResponseTooLong
	}

	for dataSize+4 > totalSize {
		size, err := r.conn.Read(r.readbuf[totalSize:])
		if err != nil {
			return 0, 0, nil, err
		}
		totalSize += size
	}

	data := r.readbuf[4 : 4+dataSize]
	if totalSize > dataSize+4 {
		// start of the next buffer was at the end of this packet.
		// save it for the next read.
		r.queuedbuf = r.readbuf[4+dataSize : totalSize]
	}

	return r.readResponseData(data)
}

func (r *RemoteConsole) readResponseData(data []byte) (int, int, []byte, error) {
	var requestId, responseType int32
	var response []byte
	b := bytes.NewBuffer(data)
	binary.Read(b, binary.LittleEndian, &requestId)
	binary.Read(b, binary.LittleEndian, &responseType)
	response, err := b.ReadBytes(0x00)
	if err != nil && err != io.EOF {
		return 0, 0, nil, err
	}
	if err == nil {
		// if we didn't hit EOF, we have a null byte to remove
		response = response[:len(response)-1]
	}
	return int(responseType), int(requestId), response, nil
}