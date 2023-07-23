package internal

import (
	"log"
	"runtime"
	"strings"
)

func SetupLogger() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
}

func PrintErr(id string, err error) {
	log.Println(getFuncName(2), id, err)
}

func getFuncName(depth int) string {
	pc, _, _, _ := runtime.Caller(depth)
	funcPath := runtime.FuncForPC(pc).Name()
	funcName := strings.Split(funcPath, "/")[3]
	return funcName
}
