package test

import (
	"bufio"
	"io"
	"tutorial/bt_demo/benode"
	"tutorial/bt_demo/utils"
)

var (
	Ctx benode.ParseContext = &benode.NodeContextImpl{}
)

type RawInfo struct {
	Name     string `benode:"name"`
	Len      int    `benode:"length"`
	PieceLen int    `benode:"piece length"`
	Pieces   string `benode:"pieces"`
}

type RawFile struct {
	Announce string   `benode:"announce"`
	Info     *RawInfo `benode:"info"`
}

type TorrentFile struct {
	InfoSHA   [utils.SHALEN]byte
	RawInfo   *RawFile
	PiecesSHA [][utils.SHALEN]byte
}

func ParseTorrentFile(rd io.Reader) (tf *TorrentFile, err error) {
	var rf *RawFile
	Ctx.Unmarshal(bufio.NewReader(rd), &rf)
	if Ctx.Err() != nil {
		return nil, Ctx.Err()
	}

	tf = &TorrentFile{
		RawInfo: rf,
	}
	panic("todo")
}
