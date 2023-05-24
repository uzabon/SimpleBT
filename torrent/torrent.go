package torrent

import (
	"bufio"
	"io"
	"tutorial/bt_demo/benode"
	"tutorial/bt_demo/utils"
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
	Announce  string
	Name      string
	FileLen   int
	PieceLen  int
	PiecesSHA [][utils.SHALEN]byte
}

func ParseTorrentFile(rd io.Reader) (tf *TorrentFile, err error) {
	var rf *RawFile
	var node benode.Benode
	checkErr := func(fn func()) {
		if err != nil {
			return
		}
		fn()
	}

	checkErr(func() {
		err = benode.Unmarshal(bufio.NewReader(rd), &rf)
	})
	checkErr(func() {
		tf = &TorrentFile{
			Announce: rf.Announce,
			Name:     rf.Info.Name,
			PieceLen: rf.Info.PieceLen,
			FileLen:  rf.Info.Len,
		}
	})
	checkErr(func() {
		node, err = benode.Marshal(rf.Info)
	})
	checkErr(func() {
		tf.InfoSHA, err = benode.CalSHA(node)
	})

	pieces := utils.Bytes(rf.Info.Pieces)
	cnt := len(pieces) / utils.SHALEN
	tf.PiecesSHA = make([][utils.SHALEN]byte, cnt)
	for i := 0; i < cnt; i++ {
		copy(tf.PiecesSHA[i][:], pieces[i*utils.SHALEN:(i+1)*utils.SHALEN])
	}

	return tf, nil
}
