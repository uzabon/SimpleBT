package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func LogSource() string {
	pc, file, line, _ := runtime.Caller(1)
	return fmt.Sprintf("%v#%v-%v", filepath.Base(file), line, runtime.FuncForPC(pc).Name())
}
