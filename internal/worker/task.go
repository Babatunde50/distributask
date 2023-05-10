package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const TaskSendTask = "task:send_task"

type PayloadSendTask struct {
	TaskID int `json:"task_id"`
	UserID int `json:"user_id"`
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

	task := asynq.NewTask(TaskSendTask, jsonPayload, opts[0])

	_, err = distributor.client.Enqueue(task, opts[1])
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendTask(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendTask
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	// get task
	gottenTask, err := processor.db.GetTask(payload.TaskID, payload.UserID)

	if err != nil {
		return fmt.Errorf("failed to get task from the db %v", err)
	}

	gottenTask.Status = "in_progress"

	fmt.Println(gottenTask.UserId, "gottenTask.UserId")

	err = processor.db.UpdateTask(gottenTask)

	if err != nil {
		return fmt.Errorf("failed to update task from the db %v", err)
	}

	switch gottenTask.Type {
	case "image_processing":
		err := imageHandler(processor.db, gottenTask)
		if err != nil {
			return err
		}
	}

	return nil
}
