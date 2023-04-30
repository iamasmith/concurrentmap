package concurrentmap

import (
	"sync"
	"testing"

	"github.com/google/uuid"
)

type TestStruct struct {
	Id  string
	Val int
}

func TestPutGetStrStr(t *testing.T) {
	cA := Wrap(map[string]string{})
	cA.Put("foo", "hello")
	v, found := cA.Get("foo")
	if !found {
		t.Errorf("Could not get key")
	}
	if v != "hello" {
		t.Errorf("Value mismatch")
	}
}

func TestPutGetStrTestStructr(t *testing.T) {
	cA := Wrap(map[string]TestStruct{})
	originalA := TestStruct{Id: "test", Val: 1}
	cA.Put("foo", originalA)
	originalB := TestStruct{Id: "another", Val: 2}
	cA.Put("bar", originalB)
	vA, found := cA.Get("foo")
	if !found {
		t.Errorf("Could not get key")
	}
	if vA != originalA {
		t.Errorf("Value mismatch")
	}
	vB, found := cA.Get("bar")
	if !found {
		t.Errorf("Could not get key")
	}
	if vB != originalB {
		t.Errorf("Value mismatch")
	}
}

func TestDelete(t *testing.T) {
	cA := Wrap(map[string]TestStruct{})
	original := TestStruct{Id: "test", Val: 1}
	cA.Put("foo", original)
	v, found := cA.Get("foo")
	if !found {
		t.Errorf("Could not get key")
	}
	if v != original {
		t.Errorf("Value mismatch")
	}
	cA.Delete("foo")
	if _, found := cA.Get("foo"); found {
		t.Errorf("Should not get key")
	}
}

// Example because the calling syntax is quite terse
func (c *ConcurrentMap[M, K, V]) keys() []K {
	return c.WithRLock(
		func(m M) any {
			r := make([]K, len(m))
			i := 0
			for k := range m {
				r[i] = k
				i++
			}
			return r
		},
	).([]K)
}

func TestKeys(t *testing.T) {
	cA := Wrap(map[string]int{})
	cA.Put("one", 1)
	cA.Put("two", 2)
	keys := cA.keys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys %#v", keys)
	}
}

func TestSaturate(t *testing.T) {
	m := Wrap(map[int]uuid.UUID{})
	// <10 sec YMMV
	const runCount = 50000
	const maxKey = 1000
	defer func() {
		if err := recover(); err != nil {
			t.Error(err)
		}
	}()
	current := 0
	l := sync.Mutex{}
	c := sync.WaitGroup{}
	s := sync.WaitGroup{}
	s.Add(1)
	shutdown := make(chan struct{})
	go func() {
		defer func() {
			if err := recover(); err != nil {
				t.Error(err)
			}
		}()
		defer s.Done()
		// Generate a LOT of updates until clients are created
		key := 0
		for current < runCount {
			uuid := uuid.New()
			m.Put(key, uuid)
			key++
			if key > maxKey {
				key = 0
			}
		}
		close(shutdown)
		// Continue until they are down to 0
		for current > 0 {
			uuid := uuid.New()
			m.Put(key, uuid)
			key++
			if key > maxKey {
				key = 0
			}
		}
	}()
	for client := 0; client < runCount; client++ {
		c.Add(1)
		go func(client int) {
			defer c.Done()
			l.Lock()
			current++
			l.Unlock()
			defer func() {
				l.Lock()
				current--
				l.Unlock()
			}()
			for running := true; running; {
				select {
				case <-shutdown:
					running = false
				default:
					for _, v := range m.keys() {
						u, _ := m.Get(v)
						t.Log(client, v, u)
					}
				}
			}
		}(client)
	}
	c.Wait()
	s.Wait()
}
