package main

import (
	"encoding/binary"
	"fmt"
	"github.com/marksamman/bencode"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

type Tracker struct {
	Complete   int64
	Incomplete int64
	Interval   int64
	Peers      []Peer
}

type Peer struct {
	IP   net.IP
	Port uint16
}


func buildURL(t *Torrent) (string, error) {
	uri, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	// TODO: Remove hardcoded identifier
	peerID := "a3kdu7g38behf73ha79d"
	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{peerID},
		"port":       []string{strconv.Itoa(6881)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":		  []string{strconv.Itoa(int(t.File.Length))},
	}

	uri.RawQuery = params.Encode()
	return uri.String(), nil
}

func parsePeers(peersBin []byte) ([]Peer, error) {
	const peerSize = 6 // 4 for IP, 2 for port
	numPeers := len(peersBin) / peerSize
	if len(peersBin)%peerSize != 0 {
		err := fmt.Errorf("Received malformed peers")
		return nil, err
	}
	peers := make([]Peer, numPeers)
	for i := 0; i < numPeers; i++ {
		offset := i * peerSize
		peers[i].IP = net.IP(peersBin[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16(peersBin[offset+4 : offset+6])
	}
	return peers, nil
}

func createTracker(reader io.Reader) (*Tracker, error) {
	dict, err := bencode.Decode(reader)
	if err != nil {
		return nil, err
	}

	b := []byte(dict["peers"].(string))
	res, err := parsePeers(b)
	if err != nil {
		return nil, err
	}

	tracker := Tracker{
		Complete: dict["complete"].(int64),
		Incomplete: dict["incomplete"].(int64),
		Interval: dict["interval"].(int64),
		Peers: res,
	}

	return &tracker, nil
}

func GetTrackers(t* Torrent) (*Tracker, error) {
	uri, err := buildURL(t)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(uri)
	defer resp.Body.Close()

	tracker, err := createTracker(resp.Body)
	if err != nil {
		return nil, err
	}

	return tracker, err
}