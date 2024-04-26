package logger

import "testing"

var repeatTimes = 2

func TestLogger(t *testing.T) {
	for i := 0; i < repeatTimes; i++ {
		KDebug("This is a debug message.")
		KInfo("This is an info message.")
		KWarning("This is a warning message.")
		KError("This is an error message.")
	}
}
