package benode

import "errors"

const (
	BSTR  BType = iota
	BINT  BType = iota
	BLIST BType = iota
	BDICT BType = iota
)

type (
	BType int
	BData any

	BObject struct {
		Type BType
		Data BData
	}
)

var (
	ErrTyp = errors.New("type not match")
)

func (o *BObject) List() ([]*BObject, error) {
	if o.Type != BLIST {
		return nil, ErrTyp
	}
	return o.Data.([]*BObject), nil
}

func (o *BObject) Int() (int64, error) {
	if o.Type != BINT {
		return 0, ErrTyp
	}
	return o.Data.(int64), nil
}

func (o *BObject) Dict() (map[*BObject]*BObject, error) {
	if o.Type != BDICT {
		return nil, ErrTyp
	}
	return o.Data.(map[*BObject]*BObject), nil
}

func (o *BObject) String() (string, error) {
	if o.Type != BSTR {
		return "", ErrTyp
	}
	return o.Data.(string), nil
}
