package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var HOME, _ = os.UserHomeDir()
var LOG_PATH = HOME + "/xzp/error.log"

func WriteLog(msg string) {
	f, err := os.OpenFile(LOG_PATH, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	defer f.Close()
	if err != nil {
		log.Println("[ERR] 无法记录错误:", err)
		return
	}

	f.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC850), msg))
}
