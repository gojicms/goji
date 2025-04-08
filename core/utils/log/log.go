package log

import (
	"fmt"
	"os"
)

const (
	LogVerbose LogLevel = 1 << iota
	LogInfo
	LogWarn
	LogError
	LogDebug
)

const (
	RCSuccess              = 0
	RCDatabase             = 0xDB // For issues related to databases
	RCAdminConfig          = 0x4D
	RCUnknownError         = 1 // Everything else that's likely on our side
	RCServicesConfig       = 0x534111CE5
	RCInvalidAppInvocation = 0xDEADBEEF // This is on you!
)

type LogLevel int

var Level = LogInfo

const color_yellow = "\033[33m"
const color_red = "\033[31m"
const color_green = "\033[32m"
const color_blue = "\033[34m"
const color_magenta = "\033[35m"
const color_cyan = "\033[36m"
const color_white = "\033[37m"
const color_black = "\033[40m"

const format_bold = "\033[1m"

const reset = "\033[0m"

func Debug(group string, message string, args ...interface{}) {
	if Level&LogDebug != 0 {
		str := fmt.Sprintf(message, args...)
		fmt.Printf("%s%s[%s] %s%s\n", color_magenta, format_bold, group, str, reset)
	}
}

func Log(group string, message string, args ...interface{}) {
	if Level&LogVerbose != 0 {
		str := fmt.Sprintf(message, args...)
		fmt.Printf("[%s] %s\n", group, str)
	}
}

func Success(group string, message string, args ...interface{}) {
	str := fmt.Sprintf(message, args...)
	fmt.Printf("%s%s[%s] %s%s\n", color_green, format_bold, group, str, reset)
}

func Info(group string, message string, args ...interface{}) {
	if Level&LogInfo != 0 {
		str := fmt.Sprintf(message, args...)
		fmt.Printf("%s[%s] %s%s\n", color_blue, group, str, reset)
	}
}

func Warn(group string, message string, args ...interface{}) {
	if Level&LogWarn != 0 {
		str := fmt.Sprintf(message, args...)
		fmt.Printf("%s%s[%s] %s%s\n", color_yellow, format_bold, group, str, reset)
	}
}

func Error(group string, message string, args ...interface{}) {
	if Level&LogError != 0 {
		str := fmt.Sprintf(message, args...)
		fmt.Printf("%s%s[%s] %s%s\n", color_red, format_bold, group, str, reset)
	}
}

func Fatal(returnCode int, group string, message string, args ...interface{}) {
	str := fmt.Sprintf(message, args...)
	fmt.Printf("%s%s A Fatal Error Occured %s\n", color_red, format_bold, reset)
	fmt.Printf("%s%s [%s] %s%s\n", color_red, format_bold, group, str, reset)
	os.Exit(returnCode)
}
