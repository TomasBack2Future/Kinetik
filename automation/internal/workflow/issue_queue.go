package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/TomasBack2Future/Kinetik/automation/internal/types"
	"github.com/TomasBack2Future/Kinetik/automation/pkg/logger"
	"github.com/sirupsen/logrus"
)

// QueuedIssue represents an issue waiting to be processed
type QueuedIssue struct {
	Event      *types.IssueEvent
	QueuedAt   time.Time
	BranchName string
}

// IssueQueue manages sequential processing of issues
type IssueQueue struct {
	mu         sync.Mutex
	queue      []*QueuedIssue
	processing bool
	processor  func(context.Context, *QueuedIssue) error
}

func NewIssueQueue(processor func(context.Context, *QueuedIssue) error) *IssueQueue {
	return &IssueQueue{
		queue:     make([]*QueuedIssue, 0),
		processor: processor,
	}
}

// Enqueue adds an issue to the queue
func (q *IssueQueue) Enqueue(event *types.IssueEvent) string {
	q.mu.Lock()
	defer q.mu.Unlock()

	branchName := fmt.Sprintf("issue-%d-%d", event.Issue.Number, time.Now().Unix())

	queued := &QueuedIssue{
		Event:      event,
		QueuedAt:   time.Now(),
		BranchName: branchName,
	}

	q.queue = append(q.queue, queued)

	logger.WithFields(logrus.Fields{
		"issue_number": event.Issue.Number,
		"branch_name":  branchName,
		"queue_length": len(q.queue),
	}).Info("Issue added to queue")

	// Start processing if not already running
	if !q.processing {
		go q.processQueue()
	}

	return branchName
}

// processQueue processes issues one by one
func (q *IssueQueue) processQueue() {
	q.mu.Lock()
	if q.processing {
		q.mu.Unlock()
		return
	}
	q.processing = true
	q.mu.Unlock()

	defer func() {
		q.mu.Lock()
		q.processing = false
		q.mu.Unlock()
	}()

	for {
		q.mu.Lock()
		if len(q.queue) == 0 {
			q.mu.Unlock()
			logger.Info("Queue empty, stopping processor")
			return
		}

		// Get next issue
		issue := q.queue[0]
		q.queue = q.queue[1:]
		q.mu.Unlock()

		logger.WithFields(logrus.Fields{
			"issue_number": issue.Event.Issue.Number,
			"branch_name":  issue.BranchName,
			"queued_at":    issue.QueuedAt,
			"wait_time":    time.Since(issue.QueuedAt).String(),
		}).Info("Processing queued issue")

		// Process the issue
		ctx := context.Background()
		if err := q.processor(ctx, issue); err != nil {
			logger.WithFields(logrus.Fields{
				"issue_number": issue.Event.Issue.Number,
				"error":        err.Error(),
			}).Error("Failed to process queued issue")
		}

		logger.WithFields(logrus.Fields{
			"issue_number": issue.Event.Issue.Number,
			"branch_name":  issue.BranchName,
		}).Info("Completed processing queued issue")
	}
}

// GetQueueLength returns current queue size
func (q *IssueQueue) GetQueueLength() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.queue)
}

// IsProcessing returns whether queue is currently processing
func (q *IssueQueue) IsProcessing() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.processing
}
