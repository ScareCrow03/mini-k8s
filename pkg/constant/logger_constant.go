// 本文件用于定义一些常量，可以被其他包引用；但是常量不一定只定义在这一个文件中，也可以定义在本目录的其他文件中、然后package都是constant即可
package constant

// 与日志有关的、用户可设置的常量
var (
	// 日志输出目标，支持2bit的位运算，如果不需要日志输出，那么置0即可
	LOG_OUTPUT_TARGET = OUTPUT2FILE | OUTPUT2STDOUT

	// 日志文件的相对路径；它默认被JOIN到当前user的主目录下，例如/home/user1/minik8s_log_dir/log.txt
	LOG_FILE_PATH_DEFAULT = "minik8s_log_dir/log.txt"

	// 当前的日志级别
	CURRENT_LOG_LEVEL = DEBUG_LEVEL
)

/* 与日志输出有关的、预设不应被更改的常量 */
const (
	OUTPUT2STDOUT = 1
	OUTPUT2FILE   = 2
	ERROR_LEVEL   = 0
	WARNING_LEVEL = 1
	INFO_LEVEL    = 2
	DEBUG_LEVEL   = 3
)
