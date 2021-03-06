// Parses the torfile file

package main

import (
	"crypto/sha1"
	"errors"
	"github.com/marksamman/bencode"
	"io"
)

type File struct {
	Name   string
	Length int64
}

type Torrent struct {
	Announce    string
	PieceLength int64
	Pieces      string
	File        File
	InfoHash	[20]byte
}

func createTorrent(reader io.Reader) (*Torrent, error) {
	dict, err := bencode.Decode(reader)
	if err != nil {
		return nil, err
	}

	info := dict["info"].(map[string]interface{})
	if _, ok := info["length"]; !ok {
		return nil, errors.New("multi-file torrents are not supported")
	}

	buf := bencode.Encode(info)
	infoHash := sha1.Sum(buf)
	
	file := File{
		Name:   info["name"].(string),
		Length: info["length"].(int64),
	}

	torrent := Torrent{
		Announce:    dict["announce"].(string),
		PieceLength: info["piece length"].(int64),
		Pieces:      info["pieces"].(string),
		File:        file,
		InfoHash:    infoHash,
	}

	return &torrent, nil
}

// Parses the reader for the torfile file into the torfile struct
func Parse(reader io.Reader) (*Torrent, error) {
	torrent, err := createTorrent(reader)
	if err != nil {
		return nil, err
	}

	return torrent, nil
}
