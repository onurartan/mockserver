package logger

import (
	"fmt"
	"strings"
	"time"
)

import "github.com/fatih/color"

type Config struct {
	ShowTimestamp bool
}

var LoggerConfig = Config{
	ShowTimestamp: true,
}

var (
	successStyle   = color.New(color.FgGreen, color.Bold)
	errorStyle     = color.New(color.FgRed, color.Bold)
	warnStyle      = color.New(color.FgYellow, color.Bold)
	infoStyle      = color.New(color.FgCyan)
	bannerStyle    = color.New(color.FgHiMagenta, color.Bold)
	messageStyle   = color.New(color.FgHiWhite)
	timestampStyle = color.New(color.FgHiBlack)
)

func printEmptyLines(count int) {
	if count <= 0 {
		return
	}
	fmt.Print(strings.Repeat("\n", count))
}

func printTimestamp() string {
	if LoggerConfig.ShowTimestamp {
		return timestampStyle.Sprintf("[%s] ", time.Now().Format("15:04:05"))
	}
	return ""
}

// Main log function
// prefix: log type (OK, ERROR, WARN, etc.)
// style: color and style
// msg: log message
// addEmptyLines: optional parameters → [0]=number of lines, [1]=line insertion position, [2]=starting space
func logWithType(prefix string, style *color.Color, msg string, addEmptyLines ...int) {
	n := 0        // number of blank lines
	space := 0    // leading space
	position := 1 // line insertion position (1=before, -1=after)

	if len(addEmptyLines) > 0 {
		n = addEmptyLines[0]
	}
	if len(addEmptyLines) > 1 {
		position = addEmptyLines[1]
	}

	if len(addEmptyLines) > 2 {
		space = addEmptyLines[2]
	}

	if position > 0 {
		printEmptyLines(n)
	}

	fmt.Print(strings.Repeat(" ", space))
	fmt.Print(printTimestamp())
	fmt.Print(style.Sprintf("[%s] ", prefix))
	fmt.Println(messageStyle.Sprint(msg))

	if position == -1 {
		printEmptyLines(n)
	}
}
