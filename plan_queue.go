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

type PlanQueueService struct {
	queueName string
	targetURL string
	tasks     *cloudtasks.Client
}

func NewPlanQueueService(host string, tasks *cloudtasks.Client) (*PlanQueueService, error) {
	qn := os.Getenv("PLAN_QUEUE_NAME")
	if len(qn) < 1 {
		return nil, errors.New("required PLAN_QUEUE_NAME variable")
	}

	return &PlanQueueService{
		tasks:     tasks,
		queueName: qn,
		targetURL: fmt.Sprintf("https://%s/plan/", host),
	}, nil
}

type PlanQueueTask struct {
	SQL    string `json:"sql"`
	Param  string `json:"param"`
	LastID string `json:"lastID"`
}

func (s *PlanQueueService) AddTask(ctx context.Context, body *PlanQueueTask) error {
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
