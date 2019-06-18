package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2beta3"
	"github.com/morikuni/failure"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2beta3"
)

type FireQueueService struct {
	queueName string
	targetURL string
	tasks     *cloudtasks.Client
}

func NewFireQueueService(host string, tasks *cloudtasks.Client) (*FireQueueService, error) {
	qn := os.Getenv("FIRE_QUEUE_NAME")
	if len(qn) < 1 {
		return nil, errors.New("required FIRE_QUEUE_NAME variable")
	}

	return &FireQueueService{
		tasks:     tasks,
		queueName: qn,
		targetURL: fmt.Sprintf("https://%s/fire/", host),
	}, nil
}

type FireQueueTask struct {
	SQL           string `json:"sql"`
	SchemaVersion int64  `json:"schemaVersion"`
	StartID       string `json:"startID"`
	LastID        string `json:"lastID"`
	Limit         int    `json:"limit"`
}

func (s *FireQueueService) AddTask(ctx context.Context, body *FireQueueTask) error {
	message, err := json.Marshal(body)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("failed json.Marshal. body=%+v\n", body))
	}

	req := &taskspb.CreateTaskRequest{
		Parent: s.queueName,
		Task: &taskspb.Task{
			PayloadType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        s.targetURL,
					Body:       []byte(message),
					AuthorizationHeader: &taskspb.HttpRequest_OidcToken{
						OidcToken: &taskspb.OidcToken{
							ServiceAccountEmail: ServiceAccountEmail,
						},
					},
				},
			},
		},
	}

	_, err = s.tasks.CreateTask(ctx, req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("failed cloudtasks.CreateTask. body=%+v\n", body))
	}

	return nil
}
