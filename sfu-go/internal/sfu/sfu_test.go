package sfu

import (
	"testing"

	"babelgrid/sfu/internal/config"
)

func TestFirstNonEmpty(t *testing.T) {
	if firstNonEmpty("", "b") != "b" {
		t.Fatal("want b")
	}
	if firstNonEmpty("a", "b") != "a" {
		t.Fatal("want a")
	}
}

func TestEnqueueNeverBlocks(t *testing.T) {
	// no workers draining (Workers=0 clamped to 1, but jobs buffer is 256);
	// pushing far more than the buffer must not deadlock.
	m := &Manager{jobs: make(chan AudioFrame, 4), rooms: map[string]*room{}}
	for i := 0; i < 100; i++ {
		m.Enqueue(AudioFrame{Room: "r", PCM: []int16{1, 2, 3}})
	}
	// if we got here without blocking, the drop-on-full path works
}

func TestGetRoomStable(t *testing.T) {
	m := NewManager(config.Config{Workers: 1, Targets: []string{"es"}})
	a := m.getRoom("x")
	b := m.getRoom("x")
	if a != b {
		t.Fatal("getRoom should return the same room for the same id")
	}
	rooms, _ := m.Stats()
	if rooms != 1 {
		t.Fatalf("rooms=%d", rooms)
	}
}
