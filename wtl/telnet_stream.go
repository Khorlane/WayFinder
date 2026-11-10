package wtl

import (
	"fmt"
	"io"
	"log"
	"net"

	"WayFinder/weg"
)

const (
	telnetSE   byte = 240
	telnetSB   byte = 250
	telnetWILL byte = 251
	telnetWONT byte = 252
	telnetDO   byte = 253
	telnetDONT byte = 254
	telnetIAC  byte = 255
	telnetGMCP byte = 201

	maxGMCPFrame = 16 * 1024
)

type telnetReadState uint8

const (
	readText telnetReadState = iota
	readCommand
	readOption
	readSubOption
	readSubData
	readSubCommand
)

type telnetStream struct {
	conn       net.Conn
	output     io.Writer
	gateway    *weg.LiveGateway
	trace      *log.Logger
	state      telnetReadState
	command    byte
	text       []byte
	gmcp       bool
	subData    []byte
	discardSub bool
}

func receiveTelnet(conn net.Conn, output io.Writer, gateway *weg.LiveGateway, trace *log.Logger) error {
	stream := telnetStream{conn: conn, output: output, gateway: gateway, trace: trace}
	var buffer [4096]byte
	firstRead := true
	for {
		n, err := conn.Read(buffer[:])
		if n > 0 {
			if firstRead {
				stream.tracef("First server read: %d bytes, starts with Telnet IAC: %t", n, buffer[0] == telnetIAC)
				firstRead = false
			}
			if consumeErr := stream.consume(buffer[:n]); consumeErr != nil {
				return consumeErr
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (s *telnetStream) consume(chunk []byte) error {
	for _, b := range chunk {
		switch s.state {
		case readText:
			if b == telnetIAC {
				s.state = readCommand
			} else {
				s.text = append(s.text, b)
			}
		case readCommand:
			s.state = readText
			switch b {
			case telnetIAC:
				s.text = append(s.text, b)
			case telnetWILL, telnetWONT, telnetDO, telnetDONT:
				s.command, s.state = b, readOption
			case telnetSB:
				s.state = readSubOption
			}
		case readOption:
			if err := s.flushText(); err != nil {
				return err
			}
			if err := s.negotiate(b); err != nil {
				return err
			}
			s.state = readText
		case readSubOption:
			s.subData = s.subData[:0]
			s.discardSub = b != telnetGMCP || !s.gmcp
			if b == telnetGMCP {
				s.tracef("RX Telnet %02X %02X %02X (GMCP subnegotiation, enabled: %t)", telnetIAC, telnetSB, b, s.gmcp)
			}
			s.state = readSubData
		case readSubData:
			if b == telnetIAC {
				s.state = readSubCommand
			} else {
				s.appendSubByte(b)
			}
		case readSubCommand:
			s.state = readSubData
			switch b {
			case telnetIAC:
				s.appendSubByte(b)
			case telnetSE:
				if !s.discardSub {
					if err := s.flushText(); err != nil {
						return err
					}
					room := s.gateway.ReceiveGMCP(s.subData)
					if room.ID == "" {
						s.tracef("Ignored GMCP payload (%d bytes, no valid HolyQuest.Room event)", len(s.subData))
					} else if room.Destinations == nil {
						s.tracef("RX HolyQuest.Room id=%q (no exit destinations supplied)", room.ID)
					} else {
						s.tracef("RX HolyQuest.Room id=%q exits=%d", room.ID, len(room.Destinations))
					}
				}
				s.state = readText
			default:
				s.discardSub = true
			}
		}
	}
	return s.flushText()
}

func (s *telnetStream) negotiate(option byte) error {
	s.tracef("RX Telnet %02X %02X %02X", telnetIAC, s.command, option)
	switch s.command {
	case telnetWILL:
		if option != telnetGMCP || s.gateway == nil {
			return s.reply(telnetDONT, option)
		}
		if !s.gmcp {
			if err := s.reply(telnetDO, option); err != nil {
				return err
			}
			s.gmcp = true
			s.gateway.SetGMCPEnabled(true)
		}
	case telnetWONT:
		if option == telnetGMCP && s.gmcp {
			if err := s.reply(telnetDONT, option); err != nil {
				return err
			}
			s.gmcp = false
			s.gateway.SetGMCPEnabled(false)
		}
	case telnetDO:
		return s.reply(telnetWONT, option)
	}
	return nil
}

func (s *telnetStream) reply(command, option byte) error {
	if _, err := s.conn.Write([]byte{telnetIAC, command, option}); err != nil {
		return fmt.Errorf("negotiate Telnet option %d: %w", option, err)
	}
	s.tracef("TX Telnet %02X %02X %02X", telnetIAC, command, option)
	return nil
}

func (s *telnetStream) tracef(format string, args ...any) {
	if s.trace != nil {
		s.trace.Printf(format, args...)
	}
}

func (s *telnetStream) appendSubByte(b byte) {
	if s.discardSub {
		return
	}
	if len(s.subData) == maxGMCPFrame {
		s.discardSub = true
		return
	}
	s.subData = append(s.subData, b)
}

func (s *telnetStream) flushText() error {
	if len(s.text) == 0 {
		return nil
	}
	n, err := s.output.Write(s.text)
	if err == nil && n != len(s.text) {
		err = io.ErrShortWrite
	}
	s.text = s.text[:0]
	return err
}
