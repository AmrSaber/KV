package tests

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
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

	t.Run("Pipe get to set", func(t *testing.T) {
		sourceKey, distKey := "source-key", "dist-key"
		value := "some-value"

		// Set initial value
		RunKVSuccess(t, "set", sourceKey, value)

		readCmd := RunKVCommand(t, "get", sourceKey)
		writeCmd := RunKVCommand(t, "set", distKey)

		readerOutPipe, err := readCmd.StdoutPipe()
		if err != nil {
			t.Fatal("Failed to get stdout pipe for reader command:", err)
		}

		// Redirect write command's input to read command's output
		writeCmd.Stdin = readerOutPipe

		errGroup := &errgroup.Group{}

		errGroup.Go(func() error {
			err := readCmd.Start()
			if err != nil {
				return fmt.Errorf("Failed to start read command: %w", err)
			}

			err = readCmd.Wait()
			if err != nil {
				return fmt.Errorf("Failed to wait for read command: %w", err)
			}

			return nil
		})

		errGroup.Go(func() error {
			err := writeCmd.Start()
			if err != nil {
				return fmt.Errorf("Failed to start write command: %w", err)
			}

			err = writeCmd.Wait()
			if err != nil {
				return fmt.Errorf("Failed to wait for write command: %w", err)
			}

			return nil
		})

		if err := errGroup.Wait(); err != nil {
			t.Fatal(err)
		}

		distValue := RunKVSuccess(t, "get", distKey)
		if distValue != value {
			t.Fatalf("Expected %q, got %q", value, distValue)
		}
	})
}
