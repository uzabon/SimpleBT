package benode

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"playground/utils"
)

type Encoder struct{}

var (
	ErrInvalidCode = errors.New("invalid encoding")
)

func (*Encoder) encodeStr(r io.Reader) (string, error) {
	rd, ok := r.(*bufio.Reader)
	if !ok {
		rd = bufio.NewReader(r)
	}
	data, err := encodeDecimal(rd)
	if err != nil {
		return "", fmt.Errorf("encodeDecimal:%w", err)
	}
	b, err := rd.ReadByte()
	if err != nil {
		return "", fmt.Errorf("readByte:%w", err)
	}
	if b != ':' {
		return "", ErrInvalidCode
	}
	var bs []byte
	l, err := io.ReadAtLeast(rd, bs, int(data))
	if err != nil {
		return "", fmt.Errorf("readAtLeast:%w", err)
	}
	if l != int(data) {
		return "", fmt.Errorf("readAtLeast:read %d bytes, expect %d", l, data)
	}
	return utils.Bytes2Str(bs), nil
}

func (*Encoder) encodeInt(r io.Reader) (int64, error) {
	rd, ok := r.(*bufio.Reader)
	if !ok {
		rd = bufio.NewReader(r)
	}
	b, err := rd.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("readByte:%w", err)
	}
	if b != 'i' {
		return 0, ErrInvalidCode
	}
	data, err := encodeDecimal(rd)
	if err != nil {
		return 0, fmt.Errorf("encodeDecimal:%w", err)
	}
	b, err = rd.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("readByte:%w", err)
	}
	if b != 'e' {
		return 0, ErrInvalidCode
	}
	return data, nil
}

func encodeDecimal(r *bufio.Reader) (res int64, err error) {
	res = 0
	sign := int64(1)
	start := true
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		if start {
			start = false
			if b == '-' {
				sign = -1
				continue
			}
		}

		if b <= '0' || b > '9' {
			break
		}
		res = res*10 + int64(b) - '0'
	}
	return res * sign, nil
}
