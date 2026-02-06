// Package testlogger - Simple logger which logs to a buffer, for testing
package testlogger

/*
 * testlogger.go
 * Simple logger which logs to a buffer, for testing
 * By J. Stuart McMurray
 * Created 20260118
 * Last Modified 20260118
 */

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"
)

// TestLogBuffer is a buffer with a couple of test functions attached.
type TestLogBuffer struct {
	bytes.Buffer
}

// New returns a log.Logger which logs to the returned TestLogBuffer.
// Log lines will have no timestamps.
func New() (*log.Logger, *TestLogBuffer) {
	buf := new(TestLogBuffer)
	return log.New(buf, "", 0), buf
}

// TestStartsWith calls t.Errorf if b doesn't start with wantLines, to which
// newlines will be appended.
// The buffer will have a number of bytes read corresponding to the length
// of wantLines plus newlines, up to the size of the buffer.
func (b *TestLogBuffer) TestStartsWith(t *testing.T, wantLines ...string) {
	t.Helper()
	for _, want := range wantLines {
		/* Make sure we still have something buffered. */
		if 0 == b.Len() {
			t.Errorf("Log buffer empty, expected\n%q", want)
			continue
		}
		want += "\n"
		if got := string(b.Next(len(want))); got != want {
			t.Errorf(
				"Log incorrect\ngot:\n%q\nwant:\n%q",
				got,
				want,
			)
		}
	}
}

// TestEmpty calls t.Errorf and resets b if b isn't already empty.
func (b *TestLogBuffer) TestEmpty(t *testing.T) {
	t.Helper()
	defer b.Reset()
	if got := b.Bytes(); 0 != len(got) {
		var ls []string
		for l := range bytes.Lines(got) {
			ls = append(ls, fmt.Sprintf(
				"%q",
				bytes.TrimRight(l, "\n"),
			))
		}
		t.Errorf(
			"Log buffer not empty, contains\n%s",
			strings.Join(ls, "\n"),
		)
	}
}
