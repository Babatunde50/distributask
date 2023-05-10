package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Babatunde50/distributask/internal/database"
	"github.com/hibiken/asynq"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendTask(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	db     *database.DB
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, db *database.DB) TaskProcessor {

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 3,
				QueueDefault:  2,
				QueueLow:      1,
			},
			StrictPriority: true,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				fmt.Println("Process task failed...", err)
				// update task status to failed...
				var payload PayloadSendTask
				if err := json.Unmarshal(task.Payload(), &payload); err != nil {
					return
				}

				gottenTask, err := db.GetTask(payload.TaskID, payload.UserID)

				if err != nil {
					return
				}

				gottenTask.Status = "failed"

				gottenTask.RetryCount += 1

				db.UpdateTask(gottenTask)
			}),
			RetryDelayFunc: func(n int, e error, task *asynq.Task) time.Duration {
				return time.Duration(time.Duration.Seconds(20))
			},
		},
	)

	return &RedisTaskProcessor{
		server: server,
		db:     db,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	// mux.HandleFunc(TaskSendTask, processor.ProcessTaskSendTask)

	mux.Handle(TaskSendTask, loggingMiddleware(asynq.HandlerFunc(processor.ProcessTaskSendTask)))

	return processor.server.Start(mux)
}

// middleware...
func loggingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		start := time.Now()
		log.Printf("Start processing %q", t.Type())
		err := h.ProcessTask(ctx, t)
		if err != nil {
			return err
		}
		log.Printf("Finished processing %q: Elapsed Time = %v", t.Type(), time.Since(start))
		return nil
	})
}
