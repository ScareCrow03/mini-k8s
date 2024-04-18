package logger

import (
	"fmt"
	"mini-k8s/pkg/constant"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// 即使这几个变量不会被其他包引用，但是c/cpp的开发习惯，还是将这些变量名全部大写
const (
	ERROR_STR   = "[ERROR]"
	WARNING_STR = "[WARNING]"
	INFO_STR    = "[INFO]"
	DEBUG_STR   = "[DEBUG]"
)

var (
	curLogLevel int
	logFilePath string
)

// 用于承载日志输出的对象
var logFileFD *os.File

func init() { // go语言的包级别函数，当前模块被加载时执行一次
	fmt.Print("logger init\n")
	logFileFD = nil
	curLogLevel = constant.CURRENT_LOG_LEVEL

	// 获取当前用户的主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Failed to get the user home directory:", err)
		return
	}

	// 将日志文件的相对路径添加到主目录的路径中
	logFilePath = filepath.Join(homeDir, constant.LOG_FILE_PATH)

	// 创建日志文件的目录
	err = os.MkdirAll(filepath.Dir(logFilePath), 0755)
	if err != nil {
		fmt.Println("Failed to create log directory:", err)
		return
	}

	if curLogLevel&constant.OUTPUT2FILE != 0 {
		// 只有指定了需要日志输出时，才尝试能否打开日志文件
		logFileFD, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Failed to open log file:", err)

			curLogLevel &= ^constant.OUTPUT2FILE // 关闭文件输出，保持这个量与文件打开情况一致
		}
	}

}

// 这个函数并不是go语言的包级别函数，只是我们期望在退出时能够释放资源（当然可能不释放也没事）
func Close() {
	if logFileFD != nil {
		logFileFD.Close()
	}
}

// 基本的输出函数；接收报错等级的label串（也可以传递其他的任何东西、作为一个前缀），以及格式化字符串和参数
func KLog(prefixStr string, format string, a ...interface{}) {
	msg := fmt.Sprintf("%s [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), prefixStr, fmt.Sprintf(format, a...)) // 注意格式化时间串的逻辑，必须选择这个时间串（或它的一部分）来表示yyyy-MM-dd HH:mm:ss之类的逻辑，因为它是Go的创始人选择的时间串

	if curLogLevel&constant.OUTPUT2STDOUT != 0 {
		// 将日志输出到stdout
		fmt.Print(msg)
	}

	if curLogLevel&constant.OUTPUT2STDOUT != 0 && logFileFD != nil {
		// 将日志输出到文件
		fmt.Fprint(logFileFD, msg)
	}
}

// 包装好的Error级别的输出函数，当curLogLevel超过0时，就会输出Error级别的日志；以下几个函数同理，可以更自定义地设置日志输出的样式、信息等
func KError(format string, a ...interface{}) {
	if curLogLevel >= constant.ERROR_LEVEL { // 一旦出错，一并输出调用位置在哪个文件的哪一行
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "???"
			line = 0
		}
		prefix := fmt.Sprintf("%s:%d", file, line)
		KLog(ERROR_STR+prefix, format, a...)
	}
}

func KWarning(format string, a ...interface{}) {
	if curLogLevel >= constant.WARNING_LEVEL {
		KLog(WARNING_STR, format, a...)
	}
}

func KInfo(format string, a ...interface{}) {
	if curLogLevel >= constant.INFO_LEVEL {
		KLog(INFO_STR, format, a...)
	}
}

func KDebug(format string, a ...interface{}) {
	if curLogLevel >= constant.DEBUG_LEVEL {
		KLog(DEBUG_STR, format, a...)
	}
}
