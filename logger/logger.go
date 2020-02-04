package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// 用户目录
var home, _ = os.UserHomeDir()

var logPath = home + "/xzp/error.log"

// WriteLog 写单行日志到文件
func WriteLog(msg string) {
	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	defer f.Close()
	if err != nil {
		log.Println("[ERR] 无法记录错误:", err)
		return
	}

	f.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC850), msg))
}
