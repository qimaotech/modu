package main

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

func TestReadInitInput_CancelWithoutInput(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	completed := make(chan error, 1)
	go func() { _, err := readInitInput(ctx, reader); completed <- err }()
	select {
	case err := <-completed:
		t.Fatalf("returned without input or cancellation: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	cancel()
	select {
	case err := <-completed:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("input prompt ignored cancellation")
	}
}
