package workflow

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/TomasBack2Future/Kinetik/automation/internal/types"
	"github.com/TomasBack2Future/Kinetik/automation/pkg/logger"
)

func init() {
	logger.Init("error") // Initialize logger for tests
}

func TestIssueQueue_Enqueue(t *testing.T) {
	processed := make([]int, 0)
	var mu sync.Mutex

	processor := func(ctx context.Context, queued *QueuedIssue) error {
		mu.Lock()
		processed = append(processed, queued.Event.Issue.Number)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond) // Simulate work
		return nil
	}

	queue := NewIssueQueue(processor)

	// Enqueue multiple issues
	event1 := &types.IssueEvent{Issue: types.Issue{Number: 1}}
	event2 := &types.IssueEvent{Issue: types.Issue{Number: 2}}
	event3 := &types.IssueEvent{Issue: types.Issue{Number: 3}}

	queue.Enqueue(event1)
	queue.Enqueue(event2)
	queue.Enqueue(event3)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(processed) != 3 {
		t.Errorf("Expected 3 issues processed, got %d", len(processed))
	}

	// Check order
	if processed[0] != 1 || processed[1] != 2 || processed[2] != 3 {
		t.Errorf("Issues not processed in order: %v", processed)
	}
}

func TestIssueQueue_GetQueueLength(t *testing.T) {
	processor := func(ctx context.Context, queued *QueuedIssue) error {
		time.Sleep(50 * time.Millisecond) // Slow processing
		return nil
	}

	queue := NewIssueQueue(processor)

	event := &types.IssueEvent{Issue: types.Issue{Number: 1}}
	queue.Enqueue(event)
	queue.Enqueue(event)
	queue.Enqueue(event)

	// Check queue length before processing completes
	time.Sleep(10 * time.Millisecond)
	length := queue.GetQueueLength()

	if length < 0 || length > 3 {
		t.Errorf("Unexpected queue length: %d", length)
	}
}

func TestIssueQueue_IsProcessing(t *testing.T) {
	processor := func(ctx context.Context, queued *QueuedIssue) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	queue := NewIssueQueue(processor)

	if queue.IsProcessing() {
		t.Error("Queue should not be processing initially")
	}

	event := &types.IssueEvent{Issue: types.Issue{Number: 1}}
	queue.Enqueue(event)

	time.Sleep(10 * time.Millisecond)

	if !queue.IsProcessing() {
		t.Error("Queue should be processing")
	}

	time.Sleep(100 * time.Millisecond)

	if queue.IsProcessing() {
		t.Error("Queue should have finished processing")
	}
}
