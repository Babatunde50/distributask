package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const TaskSendTask = "task:send_task"

type PayloadSendTask struct {
	Id int `json:"id"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendTask(
	ctx context.Context,
	payload *PayloadSendTask,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	fmt.Println("being called ...", payload)

	task := asynq.NewTask(TaskSendTask, jsonPayload, opts...)

	// info, err := distributor.client.EnqueueContext(ctx, task)
	info, err := distributor.client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	fmt.Println(info, "__info__")

	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendTask(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendTask
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	//TODO: Do your stuffs

	fmt.Println(payload, "__processing__task__")

	return nil
}
