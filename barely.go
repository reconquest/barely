package barely

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"unicode"
)

var (
	escapeSequenceRegexp = regexp.MustCompile("\x1b[^m]+m")
)

// StatusBar is the implementation of configurable status bar.
type StatusBar struct {
	sync.Locker

	format *template.Template
	last   string

	status interface{}

	width int
}

// NewStatusBar returns new StatusBar object, initialized with given template.
//
// Template will be used later on Render calls.
func NewStatusBar(format *template.Template) *StatusBar {
	return &StatusBar{
		format: format,
	}
}

func (bar *StatusBar) SetWidth(w int) {
	bar.Lock()
	defer bar.Unlock()
	bar.width = w
}

// Lock locks StatusBar object if locker object was set with SetLock method
// to prevent multi-threading race conditions.
//
// StatusBar will be locked in the Set and Render methods.
func (bar *StatusBar) Lock() {
	if bar.Locker != nil {
		bar.Locker.Lock()
	}
}

// Unlock unlocks previously locked StatusBar object.
func (bar *StatusBar) Unlock() {
	if bar.Locker != nil {
		bar.Locker.Unlock()
	}
}

// SetLock sets locker object, that will be used for Lock and Unlock methods.
func (bar *StatusBar) SetLock(lock sync.Locker) {
	bar.Locker = lock
}

// SetStatus sets data which will be used in the template execution, which is
// previously set through NewStatusBar function.
func (bar *StatusBar) SetStatus(data interface{}) {
	bar.Lock()
	defer bar.Unlock()

	bar.status = data
}

// Render renders specified template and writes it to the specified writer.
//
// Also, render result will be remembered and will be used to generate clear
// sequence which can be obtained from Clear method call.
func (bar *StatusBar) Render(writer io.Writer) error {
	bar.Lock()
	defer bar.Unlock()

	buffer := &bytes.Buffer{}

	if bar.status == nil {
		return nil
	}

	err := bar.format.Execute(buffer, bar.status)
	if err != nil {
		return fmt.Errorf(
			`error during rendering status bar: %s`,
			err,
		)
	}

	if str := buffer.String(); bar.width > 0 && graphicLength(str) > bar.width {
		buffer.Reset()
		buffer.WriteString(trimTo(str, bar.width))
	}

	fmt.Fprintf(buffer, "\r")

	bar.last = escapeSequenceRegexp.ReplaceAllLiteralString(
		buffer.String(),
		``,
	)

	_, err = io.Copy(writer, buffer)
	if err != nil {
		return fmt.Errorf(
			`can't write status bar: %s`,
			err,
		)
	}

	return nil
}

func graphicLength(str string) int {
	c := 0
	for _, r := range str {
		if unicode.IsGraphic(r) {
			c++
		}
	}

	return c
}

func trimTo(str string, l int) string {
	orig := []rune(str)
	if len(orig) < 4 {
		return str
	}

	return string(orig[:l-3]) + "…"
}

// Clear writes clear sequence in the specified writer, which is represented by
// whitespace sequence followed by "\r".
func (bar *StatusBar) Clear(writer io.Writer) {
	bar.Lock()
	defer bar.Unlock()

	fmt.Fprint(
		writer,
		strings.Repeat(" ", len(strings.TrimRight(bar.last, " \r\t")))+"\r",
	)

	bar.last = ""
}
