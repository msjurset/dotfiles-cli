package color

import (
	"fmt"
	"os"
)

const (
	reset   = "\033[0m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	red     = "\033[31m"
	cyan    = "\033[36m"
	dim     = "\033[2m"
	bold    = "\033[1m"
)

var enabled = true

func init() {
	if os.Getenv("NO_COLOR") != "" {
		enabled = false
	}
}

func wrap(code, s string) string {
	if !enabled {
		return s
	}
	return code + s + reset
}

func Green(s string) string  { return wrap(green, s) }
func Yellow(s string) string { return wrap(yellow, s) }
func Red(s string) string    { return wrap(red, s) }
func Cyan(s string) string   { return wrap(cyan, s) }
func Dim(s string) string    { return wrap(dim, s) }
func Bold(s string) string   { return wrap(bold, s) }

func Greenf(format string, a ...any) string  { return Green(fmt.Sprintf(format, a...)) }
func Yellowf(format string, a ...any) string { return Yellow(fmt.Sprintf(format, a...)) }
func Redf(format string, a ...any) string    { return Red(fmt.Sprintf(format, a...)) }
func Cyanf(format string, a ...any) string   { return Cyan(fmt.Sprintf(format, a...)) }
func Dimf(format string, a ...any) string    { return Dim(fmt.Sprintf(format, a...)) }
