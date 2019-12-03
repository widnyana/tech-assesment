package cetak

import (
	"fmt"
	"log"
	"strings"
)

func Printf(format string, v ...interface{}) {
	if !strings.Contains(format, "\n") {
		format = fmt.Sprintf("%s\n", format)
	}

	log.Printf(format, v...)
}