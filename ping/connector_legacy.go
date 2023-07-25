package ping

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Tnze/go-mc/chat"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
)

type Legacy struct {
	Addr         string
	LocalAddress string
	Timeout      time.Duration

	inited             bool
	parsedLocalAddress *net.TCPAddr
	requestPacket      []byte
	utf16Decoder       *encoding.Decoder
}

func (p *Legacy) Ping() (*Status, error) {
	if err := p.init(); err != nil {
		return nil, err
	}

	dialer := net.Dialer{
		Timeout: p.Timeout,
	}
	if p.LocalAddress != "" {
		dialer.LocalAddr = p.parsedLocalAddress
	}
	conn, err := dialer.Dial("tcp", p.Addr)
	if err != nil {
		return nil, err
	}

	defer func(conn net.Conn) {
		if err := conn.Close(); err != nil {
			log.Println("could not close connection", err)
		}
	}(conn)

	_ = conn.SetDeadline(time.Now().Add(p.Timeout))

	if _, err := conn.Write(p.requestPacket); err != nil {
		return nil, fmt.Errorf("write ping packet: %v", err)
	}

	startTime := time.Now()
	response := make([]byte, 3)
	if read, err := conn.Read(response); err != nil {
		return nil, fmt.Errorf("read response: %v", err)
	} else if read != 3 {
		return nil, fmt.Errorf("invalid server response")
	}

	length := int(binary.BigEndian.Uint16(response[1:3]))
	if length > 256 {
		return nil, fmt.Errorf("response is too big")
	}

	strReader := bufio.NewReaderSize(p.utf16Decoder.Reader(conn), 256)

	runes := make([]rune, 0, length)
	for i := 0; i < length; i++ {
		readRune, _, err := strReader.ReadRune()
		if err != nil {
			return nil, err
		}
		runes = append(runes, readRune)
	}

	respStr := string(runes)
	if !strings.HasPrefix(respStr, "§1") {
		return nil, fmt.Errorf("invalid response")
	}

	split := strings.Split(respStr, "\x00")
	if len(split) != 6 {
		return nil, fmt.Errorf("invalid response: %s", respStr)
	}

	protocol, _ := strconv.Atoi(split[1])
	online, _ := strconv.Atoi(split[4])
	max, _ := strconv.Atoi(split[5])

	return &Status{
		Latency: time.Since(startTime),
		Players: StatusPlayers{
			Online: online,
			Max:    max,
		},
		Version: StatusVersion{
			Protocol: protocol,
			Name:     split[2],
		},
		Description: chat.Message{
			Text: split[3],
		},
	}, nil
}

func (p *Legacy) buildLegacyRequestPacket(host string, port int) []byte {
	packet := []byte{
		0xFE, // Ping packet id
		0x01, // ping payload

		0xFA,       // CustomPayload packet id
		0x00, 0x0B, // length of MC|PingHost in UTF16BE
		0x00, 0x4D, 0x00, 0x43, 0x00, 0x7C, 0x00, 0x50, 0x00,
		0x69, 0x00, 0x6E, 0x00, 0x67, 0x00, 0x48, 0x00, 0x6F,
		0x00, 0x73, 0x00, 0x74, // MC|PingHost string
	}
	encodedHost, _ := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).
		NewEncoder().Bytes([]byte(host))
	encodedLenBytes := make([]byte, 4)
	binary.BigEndian.PutUint16(encodedLenBytes, uint16(7+len(encodedHost)))

	packet = append(packet, encodedLenBytes[:2]...) // payload length
	packet = append(packet,
		0x4A, // protocol version
	)

	binary.BigEndian.PutUint16(encodedLenBytes, uint16(len(host)))

	packet = append(packet, encodedLenBytes[:2]...) // hostname length
	packet = append(packet, encodedHost...)         // hostname
	binary.BigEndian.PutUint32(encodedLenBytes, uint32(port))
	packet = append(packet, encodedLenBytes...) // port

	return packet
}

func (p *Legacy) init() error {
	if p.inited {
		return nil
	}

	if !strings.ContainsRune(p.Addr, ':') {
		p.Addr += ":25565"
	}

	host, strPort, err := net.SplitHostPort(p.Addr)
	if err != nil {
		return fmt.Errorf("could not split host and port: %v", err)
	}

	port, err := strconv.ParseUint(strPort, 10, 16)
	if err != nil {
		return fmt.Errorf("port must be a number: %v", err)
	}

	if p.LocalAddress != "" {
		p.parsedLocalAddress = &net.TCPAddr{
			IP: net.ParseIP(p.LocalAddress),
		}
	}

	p.requestPacket = p.buildLegacyRequestPacket(host, int(port))

	p.utf16Decoder = unicode.
		UTF16(unicode.BigEndian, unicode.IgnoreBOM).
		NewDecoder()

	p.inited = true
	return nil
}
