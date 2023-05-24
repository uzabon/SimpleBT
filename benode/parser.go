package benode

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"tutorial/bt_demo/utils"
)

type Benode interface {
	Write(io.Writer) error
	EncodeValue(srcVal reflect.Value) (err error)
	Encode(any) error
	Decode(any) error
	DecodeValue(resVal reflect.Value) (err error)
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

	BenodeTag = "benode"
)

func CalSHA(node Benode) ([utils.SHALEN]byte, error) {
	var buf bytes.Buffer
	if err := node.Write(&buf); err != nil {
		return [utils.SHALEN]byte{}, err
	}
	return sha1.Sum(buf.Bytes()), nil
}

func Unmarshal[T any](rd *bufio.Reader, res T) error {
	impl := NewNodeContext()
	node := impl.Scan(rd)
	if impl.Err() != nil {
		if errors.Is(impl.Err(), io.EOF) {
			impl.Clean()
		}
		return impl.Err()
	}
	if err := node.Decode(&res); err != nil {
		return err
	}
	return nil
}

func Marshal[T any](src T) (res Benode, err error) {
	return marshalValue(reflect.ValueOf(src))
}

func marshalValue(srcVal reflect.Value) (res Benode, err error) {
	srcVal, _ = unwarpPtr(srcVal)

	switch srcVal.Kind() {
	case reflect.Map, reflect.Struct:
		res = &DictNode{}
	case reflect.Slice, reflect.Array:
		res = &ListNode{}
	case reflect.Int, reflect.Int64:
		res = &IntNode{}
	case reflect.String, reflect.Float64, reflect.Float32:
		res = &StringNode{}
	default:
		return nil, fmt.Errorf("%w: get %v", bTypErr, srcVal.Type())
	}
	if err := res.EncodeValue(srcVal); err != nil {
		return nil, err
	}
	return res, nil
}

type DictNode struct {
	data map[Benode]Benode
}

func (e *DictNode) Write(wd io.Writer) (err error) {
	if _, err = wd.Write([]byte{DictStartSign}); err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	for k, v := range e.data {
		if err = k.Write(wd); err != nil {
			return err
		}
		if err = v.Write(wd); err != nil {
			return err
		}
	}
	if _, err := wd.Write([]byte{EndSign}); err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	return nil
}

func (e *DictNode) Decode(res any) (err error) {
	resVal := reflect.ValueOf(res)
	return e.DecodeValue(resVal)
}
func (e *DictNode) DecodeValue(resVal reflect.Value) (err error) {
	resVal, _ = unwarpPtr(resVal)
	resTyp := resVal.Type()
	var newVal reflect.Value

	switch resTyp.Kind() {
	case reflect.Map:
		newVal = reflect.MakeMapWithSize(resTyp, len(e.data))
		for k, v := range e.data {
			kVal := reflect.New(resTyp.Key()).Elem()
			vVal := reflect.New(resTyp.Elem()).Elem()

			if err = k.DecodeValue(kVal); err != nil {
				return err
			}
			if err = v.DecodeValue(vVal); err != nil {
				return err
			}

			newVal.SetMapIndex(kVal, vVal)
		}
	case reflect.Interface:
		newVal = reflect.MakeMapWithSize(reflect.MapOf(resTyp, resTyp), len(e.data))
		for k, v := range e.data {
			kVal := reflect.New(resTyp).Elem()
			vVal := reflect.New(resTyp).Elem()

			if err = k.DecodeValue(kVal); err != nil {
				return err
			}
			if err = v.DecodeValue(vVal); err != nil {
				return err
			}

			newVal.SetMapIndex(kVal, vVal)
		}
	case reflect.Struct:
		strMap := make(map[string]int, resTyp.NumField())
		for i := 0; i < resTyp.NumField(); i++ {
			elemTag := resTyp.Field(i).Tag.Get(BenodeTag)
			strMap[elemTag] = i
		}

		newVal = reflect.New(resTyp).Elem()
		for k, v := range e.data {
			var kData string
			if err = k.Decode(&kData); err != nil {
				return err
			}
			if idx, ok := strMap[kData]; ok {
				vVal := newVal.Field(idx)
				if err = v.DecodeValue(vVal); err != nil {
					return err
				}
				newVal.Field(idx).Set(vVal)
			}
		}
	default:
		return fmt.Errorf("DictNode parse %v: %w", resTyp, bTypErr)
	}
	resVal.Set(newVal)
	return nil
}

func (e *DictNode) EncodeValue(srcVal reflect.Value) (err error) {
	srcVal, _ = unwarpPtr(srcVal)
	srcTyp := srcVal.Type()

	switch srcTyp.Kind() {
	case reflect.Map:
		keysVal := srcVal.MapKeys()
		e.data = make(map[Benode]Benode, len(keysVal))
		var knode, vnode Benode
		for i := 0; i < len(keysVal); i++ {
			if knode, err = marshalValue(keysVal[i]); err != nil {
				return err
			}
			if vnode, err = marshalValue(srcVal.MapIndex(keysVal[i])); err != nil {
				return err
			}
			e.data[knode] = vnode
		}
	case reflect.Struct:
		e.data = make(map[Benode]Benode, srcTyp.NumField())
		var knode, vnode Benode
		for i := 0; i < srcTyp.NumField(); i++ {
			field := srcTyp.Field(i)
			if knode, err = marshalValue(reflect.ValueOf(field.Tag.Get(BenodeTag))); err != nil {
				return err
			}
			if vnode, err = marshalValue(srcVal.Field(i)); err != nil {
				return err
			}
			e.data[knode] = vnode
		}
	default:
		return fmt.Errorf("%w: DictNode get %v", bTypErr, srcTyp)
	}
	return nil
}

func (e *DictNode) Encode(src any) (err error) {
	return e.DecodeValue(reflect.ValueOf(src))
}

type ListNode struct {
	data []Benode
}

func (e *ListNode) Write(wd io.Writer) (err error) {
	if _, err := wd.Write([]byte{ListStartSign}); err != nil {
		return fmt.Errorf("ListNode %w: %v", bIOErr, err)
	}
	for _, v := range e.data {
		if err = v.Write(wd); err != nil {
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
		return fmt.Errorf("ListNode parse %v: %w", resTyp, bTypErr)
	}
	resVal.Set(newVal)
	return nil
}

func (e *ListNode) Encode(src any) (err error) {
	return e.DecodeValue(reflect.ValueOf(src))
}

// ListNode can only decode to array | slice
func (e *ListNode) Decode(res any) (err error) {
	resVal := reflect.ValueOf(res)
	return e.DecodeValue(resVal)
}

func (e *ListNode) EncodeValue(srcVal reflect.Value) (err error) {
	srcVal, _ = unwarpPtr(srcVal)
	srcTyp := srcVal.Type()

	switch srcTyp.Kind() {
	case reflect.Array, reflect.Slice:
		e.data = make([]Benode, srcVal.Len())
		for i := 0; i < srcVal.Len(); i++ {
			elem, err := marshalValue(srcVal.Index(i))
			if err != nil {
				return nil
			}
			e.data[i] = elem
		}
	default:
		return fmt.Errorf("ListNode get %v: %w", srcTyp, bTypErr)
	}
	return nil
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

func (e *IntNode) Write(wd io.Writer) error {
	if e.data == nil {
		return bDataErr
	}
	if _, err := wd.Write([]byte(fmt.Sprintf("i%de", *e.data))); err != nil {
		return fmt.Errorf("IntNode %w: %v", bIOErr, err)
	}
	return nil
}

func (e *IntNode) Encode(src any) (err error) {
	return e.DecodeValue(reflect.ValueOf(src))
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
		return fmt.Errorf("%w: IntNode parse %v", bTypErr, resTyp)
	}
	resVal.Set(newVal)
	return nil
}

func (e *IntNode) EncodeValue(srcVal reflect.Value) (err error) {
	srcVal, _ = unwarpPtr(srcVal)
	srcTyp := srcVal.Type()

	switch srcTyp.Kind() {
	case reflect.Int, reflect.Int64:
		e.data = utils.Of(srcVal.Int())
	default:
		return fmt.Errorf("IntNode get %v: %w", srcTyp, bTypErr)
	}
	return nil
}

type StringNode struct {
	data *string
}

func (e *StringNode) Write(wd io.Writer) error {
	if e.data == nil {
		return bDataErr
	}
	if _, err := wd.Write([]byte(fmt.Sprintf("%d:%s", len(*e.data), *e.data))); err != nil {
		return fmt.Errorf("StringNode %w: %v", bIOErr, err)
	}
	return nil
}

func (e *StringNode) Encode(src any) (err error) {
	return e.DecodeValue(reflect.ValueOf(src))
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
			return fmt.Errorf("ParseInt %w: %v", bDataErr, err)
		}
		newVal = reflect.ValueOf(newData)
	case reflect.Interface:
		newVal = reflect.ValueOf(*e.data)
	default:
		return fmt.Errorf("%w: StringNode parse %v", bTypErr, resTyp)
	}
	resVal.Set(newVal)
	return nil
}

func (e *StringNode) EncodeValue(srcVal reflect.Value) (err error) {
	srcVal, _ = unwarpPtr(srcVal)
	srcTyp := srcVal.Type()

	switch srcTyp.Kind() {
	case reflect.String:
		e.data = utils.Of(srcVal.String())
	case reflect.Float32, reflect.Float64:
		e.data = utils.Of(strconv.FormatFloat(srcVal.Float(), 'e', -1, 64))
	default:
		return fmt.Errorf("%w: StringNode get %v", bTypErr, srcTyp)
	}
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
	if n, err = io.ReadFull(rd, b); err != nil {
		return nil, fmt.Errorf("readSlice: %w %v", bIOErr, utils.LogSource())
	}
	if n < l {
		return nil, fmt.Errorf("readSlice: EOF read:%v \n %v", n, utils.LogSource())
	}
	return b, nil
}

func peekByte(rd *bufio.Reader) (byte, error) {
	var b []byte
	var err error
	if b, err = rd.Peek(1); err != nil {
		return 0, fmt.Errorf("peakByte: %w %v", bIOErr, utils.LogSource())
	}
	return b[0], nil
}
