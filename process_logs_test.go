package main

import (
	"testing"
)

// We use minute buckets. Want to make sure timestamps are properly translated into such buckets
func TestTimestamp2Bucket(t *testing.T) {
	in := []byte("2014-12-25T19:02:49.959156+00:00")
	want := int64(1419534169 / 60)
	got := timestamp2Bucket(in)
	if got != want {
		t.Errorf("TestTimestamp2Bucket(%v) == %v, want %v", in, got, want)
	}
}
