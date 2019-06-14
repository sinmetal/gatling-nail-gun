package main

import (
	"context"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/option"
)

func CreateSpannerClient(ctx context.Context, db string, o ...option.ClientOption) (*spanner.Client, error) {
	dataClient, err := spanner.NewClient(ctx, db, o...)
	if err != nil {
		return nil, err
	}

	return dataClient, nil
}
