package utils

import (
	"unsafe"
)

func Str(src []byte) string {
	return unsafe.String(unsafe.SliceData(src), len(src))
}

func Bytes(src string) []byte {
	return unsafe.Slice(unsafe.StringData(src), len(src))
}
