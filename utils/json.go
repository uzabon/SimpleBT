package utils

import (
	"unsafe"
)

func Bytes2Str(src []byte) string {
	return unsafe.String(unsafe.SliceData(src), len(src))
}

func Str2Bytes(src string) []byte {
	return unsafe.Slice(unsafe.StringData(src), len(src))
}
