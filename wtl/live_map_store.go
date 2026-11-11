package wtl

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"WayFinder/wmr"
	"WayFinder/wne"
)

const savedMapVersion = 1

type savedLiveMap struct {
	Version int                                  `json:"version"`
	Visited []wne.RoomID                         `json:"visited"`
	Exits   map[wne.RoomID]map[string]wne.RoomID `json:"exits"`
	Layout  wmr.SavedLayout                      `json:"layout"`
}

type liveMapStore struct {
	path string
	last []byte
}

func liveMapPath(address string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("%x.json", sha256.Sum256([]byte(address)))
	return filepath.Join(dir, "WayFinder", "maps", name), nil
}

func (s *liveMapStore) load() (savedLiveMap, error) {
	var state savedLiveMap
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return state, err
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return savedLiveMap{}, err
	}
	if state.Version != savedMapVersion {
		return savedLiveMap{}, fmt.Errorf("unsupported saved map version %d", state.Version)
	}
	s.last = data
	return state, nil
}

func (state savedLiveMap) restore() (*DiscoveryState, *wne.LiveNavigation, *wmr.Mapper, error) {
	discovery := NewDiscoveryState()
	navigation := wne.NewLiveNavigation(nil, wneDiscoveryAdapter{d: discovery})
	mapper := wmr.NewMapper()
	mapper.BindTopology(wneTopologyAdapter{t: navigation})
	for _, id := range state.Visited {
		if id == "" {
			return nil, nil, nil, errors.New("saved discovery contains an empty room ID")
		}
		discovery.Discover(fromWNERoomID(id))
	}
	for id, exits := range state.Exits {
		if !discovery.IsDiscovered(fromWNERoomID(id)) {
			return nil, nil, nil, fmt.Errorf("saved exits refer to unvisited room %q", id)
		}
		for dir, destination := range exits {
			if dir == "" || destination == "" {
				return nil, nil, nil, fmt.Errorf("invalid saved exit for %q", id)
			}
		}
		navigation.RecordExits(id, exits)
	}
	if err := mapper.RestoreLayout(state.Layout, discovery); err != nil {
		return nil, nil, nil, err
	}
	return discovery, navigation, mapper, nil
}

func (s *liveMapStore) save(discovery *DiscoveryState, navigation *wne.LiveNavigation, mapper *wmr.Mapper) error {
	visited := discoveredRoomIDs(discovery)
	state := savedLiveMap{
		Version: savedMapVersion,
		Visited: visited,
		Exits:   make(map[wne.RoomID]map[string]wne.RoomID, len(visited)),
		Layout:  mapper.SaveLayout(),
	}
	for _, id := range visited {
		if navigation.HasRoom(id) {
			state.Exits[id] = navigation.ExitsFrom(id)
		}
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if bytes.Equal(data, s.last) {
		return nil
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, "map-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	// Check for updates from a different session before replacing the saved map.
	current, err := os.ReadFile(s.path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if !bytes.Equal(current, s.last) {
		return errors.New("the saved map changed in another session")
	}
	if err := os.Rename(file.Name(), s.path); err != nil {
		return err
	}
	s.last = data
	return nil
}
