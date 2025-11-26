package tests

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestConcurrency(t *testing.T) {
	cleanup := SetupTestDB(t)
	defer cleanup()

	t.Run("concurrent calls succeed", func(t *testing.T) {
		var wg sync.WaitGroup
		runnersCount := 10

		for id := range runnersCount {
			wg.Go(func() {
				key := fmt.Sprintf("a-%d", id)
				for i := range 11 {
					RunKVSuccess(t, "set", key, fmt.Sprint(i))
					time.Sleep(5 * time.Millisecond)
				}
			})
		}

		wg.Wait()

		for id := range runnersCount {
			key := fmt.Sprintf("a-%d", id)
			value := RunKVSuccess(t, "get", key)
			if value != "10" {
				t.Errorf("Runner %d: expected %s, got %s", id, "10", value)
			}
		}
	})
}
