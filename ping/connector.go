package ping

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	mcnet "github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/net/packet"
)

type Pinger interface {
	Ping() (*Status, error)
}

type Standard struct {
	Addr         string
	LocalAddress string
	Timeout      time.Duration

	inited             bool
	parsedLocalAddress *net.TCPAddr
	requestPacket      packet.Packet
}

func (s *Standard) Ping() (*Status, error) {
	if err := s.init(); err != nil {
		return nil, err
	}

	dialer := net.Dialer{
		Timeout: s.Timeout,
	}
	if s.LocalAddress != "" {
		dialer.LocalAddr = s.parsedLocalAddress
	}
	rawConn, err := dialer.Dial("tcp", s.Addr)
	if err != nil {
		return nil, err
	}
	_ = rawConn.SetDeadline(time.Now().Add(s.Timeout))
	conn := mcnet.WrapConn(rawConn)

	defer func(conn *mcnet.Conn) {
		_ = conn.Close()
	}(conn)

	// handshake
	err = conn.WritePacket(s.requestPacket)
	if err != nil {
		return nil, fmt.Errorf("sending handshake: %v", err)
	}

	// status request
	startTime := time.Now()
	err = conn.WritePacket(packet.Marshal(0))
	if err != nil {
		return nil, fmt.Errorf("sending status: %v", err)
	}

	// response
	var recv packet.Packet
	err = conn.ReadPacket(&recv)
	if err != nil {
		return nil, fmt.Errorf("receiving response: %v", err)
	}

	var jsonPayload packet.String
	if err = recv.Scan(&jsonPayload); err != nil {
		return nil, fmt.Errorf("scanning list: %v", err)
	}

	status := &Status{
		Latency: time.Since(startTime),
	}
	if err = json.Unmarshal([]byte(jsonPayload), status); err != nil {
		return nil, fmt.Errorf("unmarshal json fail: %v", err)
	}

	return status, nil
}

func (s *Standard) init() error {
	if s.inited {
		return nil
	}

	if !strings.ContainsRune(s.Addr, ':') {
		s.Addr += ":25565"
	}

	host, strPort, err := net.SplitHostPort(s.Addr)
	if err != nil {
		return fmt.Errorf("could not split host and port: %v", err)
	}

	port, err := strconv.ParseUint(strPort, 10, 16)
	if err != nil {
		return fmt.Errorf("port must be a number: %v", err)
	}

	if s.LocalAddress != "" {
		s.parsedLocalAddress = &net.TCPAddr{
			IP: net.ParseIP(s.LocalAddress),
		}
	}

	s.requestPacket = packet.Marshal(0x00,
		packet.VarInt(5),
		packet.String(host),
		packet.UnsignedShort(port),
		packet.Byte(1),
	)

	s.inited = true
	return nil
}
