package pokecache // same package = can access unexported types if needed

import (
	"fmt"
	"testing"
	"time"
)

// TestAddGet proves: whatever you Add(), you can Get() back immediately.
func TestAddGet(t *testing.T) {
	const interval = 5 * time.Second // long enough that reaper won't fire during the test

	cases := []struct {
		key string
		val []byte
	}{
		{key: "https://example.com", val: []byte("testdata")},
		{key: "https://example.com/path", val: []byte("moretestdata")},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			cache := NewCache(interval)
			cache.Add(c.key, c.val)     // store it
			val, ok := cache.Get(c.key) // retrieve it

			if !ok {
				t.Errorf("expected to find key") // test fails if key missing
				return
			}
			if string(val) != string(c.val) {
				t.Errorf("expected to find value") // test fails if value wrong
				return
			}
		})
	}
}

// TestReapLoop proves: entries ARE automatically deleted after the interval.
func TestReapLoop(t *testing.T) {
	const baseTime = 5 * time.Millisecond         // very short so the test doesn't take long
	const waitTime = baseTime + 5*time.Millisecond // wait a bit beyond the interval

	cache := NewCache(baseTime)
	cache.Add("https://example.com", []byte("testdata"))

	// Should exist right away
	_, ok := cache.Get("https://example.com")
	if !ok {
		t.Errorf("expected to find key")
		return
	}

	time.Sleep(waitTime) // let the reaper goroutine fire and clean up

	// Should be gone now
	_, ok = cache.Get("https://example.com")
	if ok {
		t.Errorf("expected to not find key") // test fails if key still exists
		return
	}
}
