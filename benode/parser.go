package benode

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type Benode interface {
	Encode(io.Writer) error
	Decode(any) error
	DecodeValue(resVal reflect.Value) (err error)
	// Decode()
}

var (
	_ Benode = (*StringNode)(nil)
	_ Benode = (*IntNode)(nil)
	_ Benode = (*ListNode)(nil)
	_ Benode = (*DictNode)(nil)

	bIOErr   = fmt.Errorf("read/write error")
	bTypErr  = fmt.Errorf("mismatch type")
	bDataErr = fmt.Errorf("invalid data")
)

const (
	DictStartSign = 'd'
	ListStartSign = 'l'
	IntStartSign  = 'i'
	EndSign       = 'e'
	SplitSign     = ':'
)

type DictNode struct {
	data map[Benode]Benode
}

func (e *DictNode) Encode(wd io.Writer) (err error) {
	if _, err = wd.Write([]byte{DictStartSign}); err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	for k, v := range e.data {
		if err = k.Encode(wd); err != nil {
			return err
		}
		if err = v.Encode(wd); err != nil {
			return err
		}
	}
	if _, err := wd.Write([]byte{EndSign}); err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	return nil
}

func (e *DictNode) Decode(res any) (err error) {
	panic("todo")
}
func (e *DictNode) DecodeValue(resVal reflect.Value) (err error) {
	panic("todo")
}

type ListNode struct {
	data []Benode
}

func (e *ListNode) Encode(wd io.Writer) (err error) {
	if _, err := wd.Write([]byte{ListStartSign}); err != nil {
		return fmt.Errorf("ListNode %w: %v", bIOErr, err)
	}
	for _, v := range e.data {
		if err = v.Encode(wd); err != nil {
			return err
		}
	}
	if _, err := wd.Write([]byte{EndSign}); err != nil {
		return fmt.Errorf("ListNode %w: %v", bIOErr, err)
	}
	return nil
}

func (e *ListNode) DecodeValue(resVal reflect.Value) (err error) {
	resVal, _ = unwarpPtr(resVal)
	resTyp := resVal.Type()
	var newVal reflect.Value

	switch resTyp.Kind() {
	case reflect.Slice:
		newVal = reflect.New(resTyp).Elem()
		for i := 0; i < len(e.data); i++ {
			elemVal := reflect.New(resTyp.Elem()).Elem()
			if err = e.data[i].DecodeValue(elemVal); err != nil {
				return err
			}
			newVal = reflect.Append(newVal, elemVal)
		}
	case reflect.Array:
		newVal = reflect.New(reflect.ArrayOf(len(e.data), resTyp.Elem())).Elem()
		for i := 0; i < len(e.data); i++ {
			elemVal := reflect.New(resTyp.Elem()).Elem()
			if err = e.data[i].DecodeValue(elemVal); err != nil {
				return err
			}
			newVal.Index(i).Set(elemVal)
		}
	case reflect.Interface:
		// to []any
		newVal = reflect.MakeSlice(reflect.SliceOf(resTyp), len(e.data), len(e.data))
		for i := 0; i < len(e.data); i++ {
			elemVal := reflect.New(resTyp).Elem()
			if err = e.data[i].DecodeValue(elemVal); err != nil {
				return err
			}
			newVal.Index(i).Set(elemVal)
		}
	default:
		return bTypErr
	}
	resVal.Set(newVal)
	return nil
}

// ListNode can only decode to array | slice
func (e *ListNode) Decode(res any) (err error) {
	resVal := reflect.ValueOf(res)
	return e.DecodeValue(resVal)
}

func unwarpPtr(resVal reflect.Value) (reflect.Value, int) {
	ptrCap := 0

	for {
		if resVal.Kind() != reflect.Pointer {
			break
		}
		if !resVal.Elem().CanSet() {
			resVal.Set(reflect.New(resVal.Type()).Elem())
		}
		ptrCap++
		resVal = resVal.Elem()
	}
	return resVal, ptrCap
}

func warpPtr(resVal reflect.Value, ptrCap int) reflect.Value {
	for i := 0; i < ptrCap; i++ {
		resVal = reflect.Indirect(resVal)
	}
	for i := ptrCap; i < 0; i++ {
		resVal = resVal.Elem()
	}
	return resVal
}

type IntNode struct {
	data *int64
}

func (e *IntNode) Encode(wd io.Writer) error {
	if e.data == nil {
		return bDataErr
	}
	if _, err := wd.Write([]byte(fmt.Sprintf("i%de", *e.data))); err != nil {
		return fmt.Errorf("IntNode %w: %v", bIOErr, err)
	}
	return nil
}

// IntNode only decode to ...int/ float /string
func (e *IntNode) Decode(res any) (err error) {
	if e.data == nil {
		return bDataErr
	}
	resVal := reflect.ValueOf(res)
	return e.DecodeValue(resVal)
}

func (e *IntNode) DecodeValue(resVal reflect.Value) (err error) {
	resVal, _ = unwarpPtr(resVal)
	resTyp := resVal.Type()
	var newVal reflect.Value

	switch resTyp.Kind() {
	case reflect.Int:
		newVal = reflect.ValueOf(int(*e.data))
	case reflect.Int64:
		newVal = reflect.ValueOf(*e.data)
	case reflect.Float32:
		newVal = reflect.ValueOf(float32(*e.data))
	case reflect.Float64:
		newVal = reflect.ValueOf(float64(*e.data))
	case reflect.String:
		newVal = reflect.ValueOf(strconv.FormatInt(*e.data, 10))
	case reflect.Interface:
		newVal = reflect.ValueOf(*e.data)
	default:
		return bTypErr
	}
	resVal.Set(newVal)
	return nil
}

type StringNode struct {
	data *string
}

func (e *StringNode) Encode(wd io.Writer) error {
	if e.data == nil {
		return bDataErr
	}
	if _, err := wd.Write([]byte(fmt.Sprintf("%d:%s", len(*e.data), *e.data))); err != nil {
		return fmt.Errorf("StringNode %w: %v", bIOErr, err)
	}
	return nil
}

func (e *StringNode) DecodeValue(resVal reflect.Value) (err error) {
	resVal, _ = unwarpPtr(resVal)
	resTyp := resVal.Type()
	var newVal reflect.Value

	switch resTyp.Kind() {
	case reflect.String:
		newVal = reflect.ValueOf(*e.data)
	case reflect.Float64:
		newData, err := strconv.ParseFloat(*e.data, 64)
		if err != nil {
			return fmt.Errorf("ParseFloat %w: %v", bDataErr, err)
		}
		newVal = reflect.ValueOf(newData)
	case reflect.Int64:
		newData, err := strconv.ParseInt(*e.data, 10, 64)
		if err != nil {
			return fmt.Errorf("ParseFloat %w: %v", bDataErr, err)
		}
		newVal = reflect.ValueOf(newData)
	case reflect.Interface:
		newVal = reflect.ValueOf(*e.data)
	default:
		return bTypErr
	}
	resVal.Set(newVal)
	return nil
}

// StringNode only decode to string, float64, int64
func (e *StringNode) Decode(res any) (err error) {
	if e.data == nil {
		return bDataErr
	}
	resVal := reflect.ValueOf(res)
	return e.DecodeValue(resVal)
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
