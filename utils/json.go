package utils

import (
	"unsafe"

	"github.com/bytedance/sonic"
)

func Str(src []byte) string {
	return *(*string)(unsafe.Pointer(&src))
}

func Bytes(src string) []byte {
	return *(*[]byte)(unsafe.Pointer(&struct {
		string
		Len int
	}{src, len(src)}))
}

func ToJson(tgt any) []byte {
	res, _ := sonic.Marshal(tgt)
	return res
}

func ToAny[T any](tgt []byte) (res T) {
	_ = sonic.Unmarshal(tgt, &res)
	return res
}
