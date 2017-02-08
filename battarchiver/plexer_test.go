package main

import (
	"testing"
)

func TestPlexer(t *testing.T) {
	px := newPlexer()
	px.add("1")
	px.add("2")
	px.add("3")
	testEx(t, "1", px)
	testEx(t, "2", px)
	testEx(t, "3", px)
	testEx(t, "1", px)
	px.get()
	px.remove("3") //remove last
	testEx(t, "1", px)
	px.add("3")
	px.remove("3") //remove after
	testEx(t, "2", px)
	px.get()
	px.add("3")
	px.get()
	px.remove("1") //remove before
	testEx(t, "3", px)

}
func testEx(t *testing.T, ex string, px *plexer) {
	got := px.get()
	if got != ex {
		t.Errorf("Expected %v got %v", ex, got)
	}
}
