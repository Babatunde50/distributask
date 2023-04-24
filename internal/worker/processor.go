package worker

import (
	"context"
	"fmt"

	"github.com/Babatunde50/distributask/internal/database"
	"github.com/hibiken/asynq"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
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
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				fmt.Println("Process task failed...")
			}),
		},
	)

	return &RedisTaskProcessor{
		server: server,
		db:     db,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendTask, processor.ProcessTaskSendTask)

	return processor.server.Start(mux)
}
