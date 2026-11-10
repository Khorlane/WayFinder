package wtl

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"WayFinder/weg"
)

const liveDialTimeout = 5 * time.Second

// StreamLiveOutput copies server text to output without sending game input.
func StreamLiveOutput(address string, output io.Writer) error {
	conn, err := net.DialTimeout("tcp", address, liveDialTimeout)
	if err != nil {
		return fmt.Errorf("connect to MUD: %w", err)
	}
	defer conn.Close()

	if _, err := io.Copy(output, conn); err != nil {
		return fmt.Errorf("stream MUD output: %w", err)
	}
	return nil
}

// PlayLive relays commands and server output until the connection or window closes.
func PlayLive(ctx context.Context, address string, commands <-chan string, display func(weg.LiveText), connected func()) error {
	dialer := net.Dialer{Timeout: liveDialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("connect to MUD: %w", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stopClose := context.AfterFunc(ctx, func() { conn.Close() })
	defer stopClose()

	connected()
	writeErr := make(chan error, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case command, open := <-commands:
				if !open {
					return
				}
				if _, err := io.WriteString(conn, command+"\r\n"); err != nil {
					writeErr <- err
					conn.Close()
					return
				}
			}
		}
	}()

	_, err = io.Copy(weg.NewLiveGateway(display), conn)
	if ctx.Err() != nil {
		return nil
	}
	select {
	case err := <-writeErr:
		return fmt.Errorf("send MUD command: %w", err)
	default:
	}
	if err != nil {
		return fmt.Errorf("read MUD output: %w", err)
	}
	return nil
}
