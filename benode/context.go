package benode

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"reflect"
	"tutorial/bt_demo/utils"
)

var (
	P ParseContext = &NodeContextImpl{}
)

type ParseContext interface {
	Scan(*bufio.Reader) Benode
	MustScan(*bufio.Reader) Benode
	Marshal(any) Benode
	Unmarshal(*bufio.Reader, ...any)
	CalSHA(any) [utils.SHALEN]byte

	Err() error
	Clean()
}

type NodeContextImpl struct {
	err error
}

func newNodeContext() ParseContext {
	return &NodeContextImpl{}
}

func (impl *NodeContextImpl) readInt(rd *bufio.Reader) (res int64) {
	for {
		if impl.Err() != nil {
			return 0
		}
		next := impl.peekByte(rd)
		if next < '0' || next > '9' {
			break
		}
		res = res*10 + int64(next-'0')
		_ = impl.readByte(rd)
	}
	return res
}

func (impl *NodeContextImpl) peekByte(rd *bufio.Reader) byte {
	if impl.Err() != nil {
		return 0
	}
	b, err := peekByte(rd)
	if err != nil {
		impl.addErr(fmt.Errorf("%w in\n %v", err, utils.LogSource()))
	}
	return b
}

func (impl *NodeContextImpl) readByte(rd *bufio.Reader) byte {
	if impl.Err() != nil {
		return 0
	}
	var bs []byte
	bs, err := readSlice(rd, 1)
	if err != nil {
		impl.addErr(fmt.Errorf("%w in\n %v", err, utils.LogSource()))
	}
	return bs[0]
}

func (impl *NodeContextImpl) readSlice(rd *bufio.Reader, l int) []byte {
	if impl.Err() != nil {
		return nil
	}
	bs, err := readSlice(rd, l)
	if err != nil {
		impl.addErr(err)
	}
	return bs
}

func (impl *NodeContextImpl) ScanInt(rd *bufio.Reader) *IntNode {
	if impl.Err() != nil {
		return nil
	}
	next := impl.readByte(rd)
	if next != IntStartSign {
		impl.addErr(fmt.Errorf("invalid int node split sign"))
	}
	data := impl.readInt(rd)
	next = impl.readByte(rd)
	if next != 'e' {
		impl.addErr(fmt.Errorf("invalid int node split sign"))
	}
	return &IntNode{
		data: utils.Of(data),
	}
}

func (impl *NodeContextImpl) Equal(src byte, tgt byte) {

}

func (impl *NodeContextImpl) ScanString(rd *bufio.Reader) *StringNode {
	if impl.Err() != nil {
		return nil
	}
	l := impl.readInt(rd)
	next := impl.readByte(rd)
	if next != ':' {
		impl.addErr(fmt.Errorf("StringNode: invalid split sign"))
	}
	data := impl.readSlice(rd, int(l))
	return &StringNode{
		data: utils.Of(utils.Str(data)),
	}
}

func (impl *NodeContextImpl) ScanDict(rd *bufio.Reader) *DictNode {
	if impl.Err() != nil {
		return nil
	}
	next := impl.readByte(rd)
	if next != DictStartSign {
		impl.addErr(fmt.Errorf("DictNode: invalid start sign"))
	}
	data := make(map[Benode]Benode, 0)
	for {
		if impl.Err() != nil {
			break
		}
		next := impl.peekByte(rd)
		if next == EndSign {
			_ = impl.readByte(rd)
			break
		}
		keyNode := impl.MustScan(rd)
		valNode := impl.MustScan(rd)
		data[keyNode] = valNode
	}
	return &DictNode{
		data: data,
	}
}

func (impl *NodeContextImpl) ScanList(rd *bufio.Reader) *ListNode {
	if impl.Err() != nil {
		return nil
	}
	next := impl.readByte(rd)
	if next != ListStartSign {
		impl.addErr(fmt.Errorf("ListNode: invalid start sign"))
	}
	var data []Benode
	for {
		if impl.Err() != nil {
			break
		}
		next := impl.peekByte(rd)
		if next == EndSign {
			_ = impl.readByte(rd)
			break
		}
		nextNode := impl.MustScan(rd)
		data = append(data, nextNode)
	}
	return &ListNode{
		data: data,
	}
}

func (impl *NodeContextImpl) Err() error {
	return impl.err
}

func (impl *NodeContextImpl) addErr(err error) {
	if impl.err != nil {
		impl.err = fmt.Errorf("%v:%#w", err, impl.err)
	} else {
		impl.err = err
	}
}

func (impl *NodeContextImpl) Scan(rd *bufio.Reader) (res Benode) {
	if impl.Err() != nil {
		return nil
	}
	next := impl.peekByte(rd)

	switch next {
	case IntStartSign:
		res = impl.ScanInt(rd)
	case ListStartSign:
		res = impl.ScanList(rd)
	case DictStartSign:
		res = impl.ScanDict(rd)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		res = impl.ScanString(rd)
	default:
		return nil
	}
	return res
}

func (impl *NodeContextImpl) Clean() {
	impl.err = nil
}

func (impl *NodeContextImpl) MustScan(rd *bufio.Reader) (res Benode) {
	res = impl.Scan(rd)
	if res == nil {
		impl.addErr(fmt.Errorf("Context: not  benode"))
	}
	return res
}

func (impl *NodeContextImpl) Marshal(src any) (res Benode) {
	if impl.Err() != nil {
		return nil
	}
	res, err := newBenode(reflect.ValueOf(src))
	if err != nil {
		impl.addErr(fmt.Errorf("Marshal: %w", err))
	}

	return res
}

func (impl *NodeContextImpl) Unmarshal(rd *bufio.Reader, resList ...any) {
	for i := 0; i < len(resList); i++ {
		node := impl.MustScan(rd)
		if impl.Err() != nil {
			if errors.Is(impl.Err(), io.EOF) {
				impl.Clean()
			}
			break
		}
		if err := node.Decode(&resList[i]); err != nil {
			impl.addErr(fmt.Errorf("Decode: %w", err))
		}
	}
}

func (impl *NodeContextImpl) CalSHA(res any) [utils.SHALEN]byte {
	if impl.Err() != nil {
		return [utils.SHALEN]byte{}
	}
	var buf bytes.Buffer
	node := impl.Marshal(res)
	if err := node.Write(&buf); err != nil {
		impl.addErr(fmt.Errorf("CalSHA: %w", err))
	}
	return sha1.Sum(buf.Bytes())
}
