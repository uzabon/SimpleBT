package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func LogSource() string {
	pc, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%v#%v-%v", filepath.Base(file), line, runtime.FuncForPC(pc).Name())
}
