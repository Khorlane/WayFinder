package wtl

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"WayFinder/weg"
	"WayFinder/wmr"
	"WayFinder/wne"
)

const liveDialTimeout = 5 * time.Second

// StreamLiveOutput copies server text to output without sending game input.
func StreamLiveOutput(address string, output io.Writer) error {
	conn, err := net.DialTimeout("tcp", address, liveDialTimeout)
	if err != nil {
		return fmt.Errorf("connect to MUD: %w", err)
	}
	defer conn.Close()

	if err := receiveTelnet(conn, output, nil, nil); err != nil {
		return fmt.Errorf("stream MUD output: %w", err)
	}
	return nil
}

// PlayLive relays commands and server output until the connection or window closes.
func PlayLive(ctx context.Context, address string, commands <-chan string, display func(weg.LiveText), showMap func(wmr.Layout, string, error), connected func()) error {
	trace, closeTrace := openGMCPLog()
	defer closeTrace()
	gateway := weg.NewLiveGateway(func(update weg.LiveText) {
		if update.GMCPStatus != "" {
			trace.Print(update.GMCPStatus)
		}
		display(update)
	})
	configureLiveMap(gateway, showMap)
	dialer := net.Dialer{Timeout: liveDialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		trace.Printf("Connect failed: %v", err)
		return fmt.Errorf("connect to MUD: %w", err)
	}
	defer conn.Close()
	trace.Printf("Connected to %s", conn.RemoteAddr())
	trace.Print(weg.GMCPWaitingStatus)
	defer trace.Print("Live session ended")

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
				gateway.CommandSent(command)
				wireCommand := strings.ReplaceAll(command, "\xff", "\xff\xff") + "\r\n"
				if _, err := io.WriteString(conn, wireCommand); err != nil {
					writeErr <- err
					conn.Close()
					return
				}
			}
		}
	}()

	err = receiveTelnet(conn, gateway, gateway, trace)
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

func openGMCPLog() (*log.Logger, func()) {
	const name = "WayFinder-gmcp.log"
	path := name
	if exe, err := os.Executable(); err == nil {
		path = filepath.Join(filepath.Dir(exe), name)
	}
	trace := log.New(os.Stderr, fmt.Sprintf("[pid %d] ", os.Getpid()), log.LstdFlags|log.Lmicroseconds)
	flags := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	file, err := os.OpenFile(path, flags, 0o600)
	if err != nil && path != name {
		file, err = os.OpenFile(name, flags, 0o600)
	}
	if err != nil {
		trace.Printf("Could not open GMCP log: %v", err)
		return trace, func() {}
	}
	trace.SetOutput(file)
	return trace, func() { file.Close() }
}

func configureLiveMap(gateway *weg.LiveGateway, showMap func(wmr.Layout, string, error)) {
	path := "Rooms"
	if exe, err := os.Executable(); err == nil {
		besideExe := filepath.Join(filepath.Dir(exe), path)
		if _, err := os.Stat(besideExe); !os.IsNotExist(err) {
			path = besideExe
		}
	}
	world, err := LoadWorld(path)
	if err != nil {
		showMap(wmr.Layout{}, "", fmt.Errorf("Could not load Rooms/ for the map (%w)", err))
		return
	}
	if len(world.exits) == 0 {
		showMap(wmr.Layout{}, "", fmt.Errorf("No room files found in %s", path))
		return
	}
	discovery := NewDiscoveryState()
	worldView := wneWorldAdapter{w: world}
	navigation := wne.NewLiveNavigation(worldView, wneDiscoveryAdapter{d: discovery})
	mapper := wmr.NewMapper()
	mapper.BindTopology(world)
	gateway.OnRoomEntry(func(room weg.RoomEntry) {
		roomID := room.ID
		if roomID == "" {
			var err error
			roomID, err = navigation.ResolveRoom(room.Name, room.Exits, room.Direction)
			if err != nil {
				navigation.ClearCurrentRoom()
				showMap(mapper.SnapshotLayout(discovery, ""), room.Name, err)
				return
			}
		}
		arrival := navigation.ObserveRoom(roomID)
		id := fromWNERoomID(arrival.Room)
		var err error
		if arrival.Known {
			err = mapper.PlaceArrival(id, fromWNERoomID(arrival.From), arrival.Direction)
		} else {
			err = fmt.Errorf("Room %s is missing from Rooms/", id)
		}
		showMap(mapper.SnapshotLayout(discovery, id), room.Name, err)
	}, func(name string) bool { return len(worldView.RoomsNamed(name)) > 0 })
}
