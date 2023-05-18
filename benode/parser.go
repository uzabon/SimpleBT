package benode

import (
	"errors"
	"fmt"
	"io"

	"tutorial/bt_demo/utils"
)

type NodeType interface {
	StringNode

	Decode(io.Reader) error
	Encode(io.Writer) error
}

var (
	bErr chan error

	bIOErr error = errors.New("read/write error")
)

type NodeFactory struct{}

type StringNode struct {
	data *string
}

func (e *StringNode) Encode(wd io.Writer) {
	if e.data == nil {
	}
	if _, err := wd.Write([]byte(fmt.Sprintf("%d:%s", len(*e.data), *e.data))); err != nil {
		bErr <- fmt.Errorf("encode: %w", err)
	}
}

func (e *StringNode) Decode(rd io.Reader) {
	res, next := readInt(rd)
	if next != ':' {
		bErr <- fmt.Errorf("invalid string node split sign")
		return
	}
	b := readSlice(rd, res)
	e.data = utils.Of(string(b))
}

func readSlice(rd io.Reader, l int64) []byte {
	b := make([]byte, l)
	if _, err := rd.Read(b); err != nil {
		bErr <- fmt.Errorf("readSlice: %w", err)
	}
	return b
}

func readByte(rd io.Reader) byte {
	b := make([]byte, 1)
	if _, err := rd.Read(b); err != nil {
		bErr <- fmt.Errorf("readByte: %w", err)
	}
	return b[0]
}

// parseInt parses an integer from the reader
func readInt(rd io.Reader) (res int64, next byte) {
	for {
		next = readByte(rd)
		if next < '0' && next > '9' {
			break
		}
		res = res*10 + int64(next-'0')
	}
	return res, next
}
