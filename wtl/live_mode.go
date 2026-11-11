package wtl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
func PlayLive(ctx context.Context, address string, commands <-chan string, display func(weg.LiveText), showMap func(wmr.Layout, string, error, wmr.LayoutEditing), connected func()) error {
	trace, closeTrace := openGMCPLog()
	defer closeTrace()
	capture, closeCapture := openMapperCapture(trace)
	defer closeCapture()
	gateway := weg.NewLiveGateway(func(update weg.LiveText) {
		if update.GMCPStatus != "" {
			trace.Print(update.GMCPStatus)
		}
		display(update)
	})
	var store *liveMapStore
	if path, err := liveMapPath(address); err != nil {
		trace.Printf("Could not locate saved map: %v", err)
	} else {
		store = &liveMapStore{path: path}
	}
	configureLiveMap(gateway, showMap, trace, capture, store)
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
	trace := log.New(os.Stderr, fmt.Sprintf("[pid %d] ", os.Getpid()), log.LstdFlags|log.Lmicroseconds)
	file, err := openLiveDiagnostic("WayFinder-gmcp.log")
	if err != nil {
		trace.Printf("Could not open GMCP log: %v", err)
		return trace, func() {}
	}
	trace.SetOutput(file)
	return trace, func() { file.Close() }
}

func openLiveDiagnostic(name string) (*os.File, error) {
	path := name
	if exe, err := os.Executable(); err == nil {
		path = filepath.Join(filepath.Dir(exe), name)
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	file, err := os.OpenFile(path, flags, 0o600)
	if err != nil && path != name {
		file, err = os.OpenFile(name, flags, 0o600)
	}
	return file, err
}

func openMapperCapture(trace *log.Logger) (func(wmr.RebuildCapture) error, func()) {
	var mu sync.Mutex
	var file *os.File
	var writeErr error
	closed := false
	open := sync.OnceValue(func() *json.Encoder {
		file, writeErr = openLiveDiagnostic("WayFinder-mapper.jsonl")
		if writeErr != nil {
			trace.Printf("Could not open mapper captures: %v", writeErr)
			return nil
		}
		return json.NewEncoder(file)
	})
	return func(capture wmr.RebuildCapture) error {
			mu.Lock()
			defer mu.Unlock()
			if closed {
				return fmt.Errorf("the game session has ended")
			}
			if encoder := open(); encoder != nil && writeErr == nil {
				entry := struct {
					Time time.Time `json:"time"`
					PID  int       `json:"pid"`
					wmr.RebuildCapture
				}{time.Now(), os.Getpid(), capture}
				if writeErr = encoder.Encode(entry); writeErr != nil {
					trace.Printf("Could not write mapper capture attempt=%d: %v", capture.Attempt, writeErr)
				}
			}
			return writeErr
		}, func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if file != nil {
				file.Close()
			}
		}
}

func configureLiveMap(gateway *weg.LiveGateway, showMap func(wmr.Layout, string, error, wmr.LayoutEditing), trace *log.Logger, capture func(wmr.RebuildCapture) error, store *liveMapStore) {
	var state savedLiveMap
	var loadErr error
	if store != nil {
		state, loadErr = store.load()
	}
	discovery, navigation, mapper, restoreErr := state.restore()
	loadErr = errors.Join(loadErr, restoreErr)
	if loadErr != nil {
		loadErr = fmt.Errorf("could not restore saved map: %w", loadErr)
		trace.Print(loadErr)
		discovery, navigation, mapper, _ = (savedLiveMap{}).restore()
		store = nil
	}
	mapper.SetIssueReporter(trace.Printf)
	if capture != nil {
		mapper.SetRebuildCapture(func(entry wmr.RebuildCapture) { _ = capture(entry) })
	}
	// Room reports and submissions from the Win32 thread share the live mapper.
	var mapMu sync.Mutex
	var pending *wmr.Correction
	var pendingRoom wne.RoomID
	var viewVersion uint64
	saveProgress := func() error {
		if store != nil {
			if err := store.save(discovery, navigation, mapper); err != nil {
				trace.Printf("Could not save map: %v", err)
				return fmt.Errorf("could not save map: %w", err)
			}
		}
		return nil
	}
	recordView := func(layout wmr.Layout, title string, arrival wne.RoomArrival, err error) {
		if capture == nil {
			return
		}
		entry := wmr.RebuildCapture{
			Stage:          "view",
			From:           fromWNERoomID(arrival.From),
			To:             fromWNERoomID(arrival.Room),
			Direction:      arrival.Direction,
			Current:        layout.Current,
			Title:          title,
			Rooms:          make(map[wmr.SolverRoomID][2]int, len(layout.Rooms)),
			Links:          layout.Links,
			ExitDirections: layout.ExitDirections,
		}
		for _, placed := range layout.Rooms {
			entry.Rooms[wmr.SolverRoomID(placed.ID)] = [2]int{placed.R, placed.C}
		}
		if err != nil {
			entry.Error = err.Error()
		}
		_ = capture(entry)
	}
	trackEdit := func(edit *wmr.Correction, room wne.RoomID, onApplied func(wmr.Layout, error)) {
		pending, pendingRoom = edit, room
		submit := edit.Submit
		edit.Submit = func(rooms []wmr.Room) (wmr.Layout, error) {
			mapMu.Lock()
			defer mapMu.Unlock()
			if pending != edit {
				return wmr.Layout{}, fmt.Errorf("the live map changed while this edit was open")
			}
			applied, err := submit(rooms)
			if err == nil {
				pending = nil
				onApplied(applied, saveProgress())
			}
			return applied, err
		}
		edit.Discard = func() {
			mapMu.Lock()
			defer mapMu.Unlock()
			if pending == edit {
				pending = nil
			}
		}
	}
	editingForView := func(layout wmr.Layout, title string, correction *wmr.Correction) wmr.LayoutEditing {
		viewVersion++
		version := viewVersion
		editing := wmr.LayoutEditing{Failure: correction}
		if len(layout.Rooms) != 0 {
			editing.Begin = func() *wmr.Correction {
				mapMu.Lock()
				defer mapMu.Unlock()
				if viewVersion != version || pending != nil {
					return nil
				}
				current := mapper.SnapshotLayout(discovery, layout.Current)
				edit := mapper.NewMapEdit(current, discovery, capture)
				id := wne.RoomID(current.Current)
				trackEdit(edit, id, func(applied wmr.Layout, saveErr error) {
					recordView(applied, title, wne.RoomArrival{Room: id, From: id}, saveErr)
				})
				return edit
			}
		}
		return editing
	}
	if layout := mapper.SnapshotLayout(discovery, ""); len(layout.Rooms) != 0 || loadErr != nil {
		showMap(layout, "", loadErr, editingForView(layout, "", nil))
	}

	loadFallback := sync.OnceValue(func() error {
		world, err := loadLiveWorld()
		if err == nil {
			navigation.SetFallbackWorld(wneWorldAdapter{w: world})
		}
		return err
	})
	fallbackNoticeShown := false
	gateway.OnRoomEntry(func(room weg.RoomEntry) {
		mapMu.Lock()
		var publish func()
		defer func() {
			mapMu.Unlock()
			if publish != nil {
				publish()
			}
		}()
		queueView := func(layout wmr.Layout, err error, correction *wmr.Correction) {
			editing := editingForView(layout, room.Name, correction)
			publish = func() { showMap(layout, room.Name, err, editing) }
		}
		roomID := room.ID
		exitsChanged := false
		if room.Destinations != nil {
			exitsChanged = navigation.RecordExits(roomID, room.Destinations)
			if exitsChanged {
				mapper.RefreshRoomExits(fromWNERoomID(roomID))
			}
		} else if roomID == "" || !navigation.HasRoom(roomID) {
			if err := loadFallback(); err != nil {
				if roomID != "" {
					navigation.ObserveRoom(roomID)
					err = errors.Join(err, saveProgress())
				}
				pending = nil
				queueView(mapper.SnapshotLayout(discovery, ""), err, nil)
				fallbackNoticeShown = true
				return
			}
		}
		if roomID == "" {
			var err error
			roomID, err = navigation.ResolveRoom(room.Name, room.Exits, room.Direction)
			if err != nil {
				navigation.ClearCurrentRoom()
				pending = nil
				queueView(mapper.SnapshotLayout(discovery, ""), err, nil)
				return
			}
		}
		if pending != nil && pendingRoom == roomID && !exitsChanged {
			return
		}
		pending = nil
		arrival := navigation.ObserveRoom(roomID)
		id := fromWNERoomID(arrival.Room)
		var err error
		if arrival.Known {
			err = mapper.PlaceArrival(id, fromWNERoomID(arrival.From), arrival.Direction)
			if err != nil {
				trace.Printf("Mapper placement failed room=%q from=%q direction=%q: %v", id, arrival.From, arrival.Direction, err)
			}
		} else {
			err = fmt.Errorf("Room %s is missing from Rooms/", id)
		}
		layout := mapper.SnapshotLayout(discovery, id)
		viewErr := errors.Join(err, saveProgress())
		if arrival.From != arrival.Room || exitsChanged {
			recordView(layout, room.Name, arrival, viewErr)
		}
		correction := mapper.NewCorrection(layout, discovery, err, capture)
		if correction != nil {
			trackEdit(correction, roomID, func(applied wmr.Layout, saveErr error) {
				recordView(applied, room.Name, arrival, saveErr)
			})
		}
		queueView(layout, viewErr, correction)
	}, func(name string) bool {
		mapMu.Lock()
		if err := loadFallback(); err != nil {
			if fallbackNoticeShown || pending != nil {
				mapMu.Unlock()
				return false
			}
			fallbackNoticeShown = true
			layout := mapper.SnapshotLayout(discovery, "")
			editing := editingForView(layout, "", nil)
			mapMu.Unlock()
			showMap(layout, "", err, editing)
			return false
		}
		known := len(navigation.RoomsNamed(name)) > 0
		mapMu.Unlock()
		return known
	})
}

func loadLiveWorld() (*World, error) {
	path := "Rooms"
	if exe, err := os.Executable(); err == nil {
		besideExe := filepath.Join(filepath.Dir(exe), path)
		if _, err := os.Stat(besideExe); !os.IsNotExist(err) {
			path = besideExe
		}
	}
	world, err := LoadWorld(path)
	if err != nil {
		return nil, fmt.Errorf("Could not load Rooms/ for legacy mapping (%w)", err)
	}
	if len(world.exits) == 0 {
		return nil, fmt.Errorf("No room files found in %s", path)
	}
	return world, nil
}
