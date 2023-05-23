package benode

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

type NodeType interface {
	Encode(io.Writer, chan<- error)
	// Decode()
}

var (
	_ NodeType = (*StringNode)(nil)
	_ NodeType = (*IntNode)(nil)
	_ NodeType = (*ListNode)(nil)
	_ NodeType = (*DictNode)(nil)

	bIOErr error = errors.New("read/write error")
)

const (
	DictStartSign = 'd'
	ListStartSign = 'l'
	IntStartSign  = 'i'
	EndSign       = 'e'
	SplitSign     = ':'
)

type DictNode struct {
	data map[NodeType]NodeType
}

func (e *DictNode) Encode(wd io.Writer, ch chan<- error) {
	if _, err := wd.Write([]byte{DictStartSign}); err != nil {
		ch <- fmt.Errorf("encode: %w", err)
	}
	for k, v := range e.data {
		k.Encode(wd, ch)
		v.Encode(wd, ch)
	}
	if _, err := wd.Write([]byte{EndSign}); err != nil {
		ch <- fmt.Errorf("encode: %w", err)
	}
}

type ListNode struct {
	data []NodeType
}

func (e *ListNode) Encode(wd io.Writer, ch chan<- error) {
	if _, err := wd.Write([]byte{ListStartSign}); err != nil {
		ch <- fmt.Errorf("ListNode encode: %w", err)
	}
	for _, v := range e.data {
		v.Encode(wd, ch)
	}
	if _, err := wd.Write([]byte{EndSign}); err != nil {
		ch <- fmt.Errorf("ListNode encode: %w", err)
	}
}

func ToGenericList(n ...NodeType) []NodeType {
	return n
}

type IntNode struct {
	data *int64
}

func (e *IntNode) Encode(wd io.Writer, ch chan<- error) {
	if e.data == nil {
		return
	}
	if _, err := wd.Write([]byte(fmt.Sprintf("i%de", *e.data))); err != nil {
		ch <- fmt.Errorf("IntNode encode: %w", err)
	}
}

type StringNode struct {
	data *string
}

func (e *StringNode) Encode(wd io.Writer, ch chan<- error) {
	if e.data == nil {
		return
	}
	if _, err := wd.Write([]byte(fmt.Sprintf("%d:%s", len(*e.data), *e.data))); err != nil {
		ch <- fmt.Errorf("encode: %w", err)
	}
}

func readSlice(rd *bufio.Reader, l int) (b []byte, err error) {
	b = make([]byte, l)
	var n int
	if n, err = rd.Read(b); err != nil {
		return nil, fmt.Errorf("readSlice: %w", bIOErr)
	}
	if n < l {
		return nil, fmt.Errorf("readSlice: EOF")
	}
	return b, nil
}

func peekByte(rd *bufio.Reader) (byte, error) {
	var b []byte
	var err error
	if b, err = rd.Peek(1); err != nil {
		return 0, fmt.Errorf("peakByte: %w", bIOErr)
	}
	return b[0], nil
}
