package testlogger

/*
 * testlogger_test.go
 * Tests for testlogger.go
 * By J. Stuart McMurray
 * Created 20260118
 * Last Modified 20260118
 */

import "testing"

/* This all gets tricky, as failing tests get complicated. */

// Does it work in the happy case?
func TestTestLogBuffer(t *testing.T) {
	have := "kittens"
	tl, tb := New()
	tl.Printf("%s", have)
	tb.TestStartsWith(t, have)
	tb.TestEmpty(t)
}
