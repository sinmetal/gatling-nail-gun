package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/cloudtasks/apiv2beta3"
	"cloud.google.com/go/spanner"
	"github.com/sinmetal/gcpmetadata"
)

var ProjectID string
var ServiceAccountEmail string
var TasksClient *cloudtasks.Client
var SpannerClient *spanner.Client

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("Hello world received a request.")
	target := os.Getenv("TARGET")
	if target == "" {
		target = "World"
	}
	fmt.Fprintf(w, "Hello %s!\n", target)
}

func main() {
	log.Print("Hello world sample started.")

	http.HandleFunc("/setup/", handleSetupAPI)
	http.HandleFunc("/plan/", handlePlanAPI)
	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func init() {
	projectID, err := gcpmetadata.GetProjectID()
	if err != nil {
		log.Fatalf("failed ProjectID.err=%+v\n", err)
		os.Exit(1)
	}
	ProjectID = projectID
	log.Printf("ProjectID is %s\n", projectID)

	sa, err := gcpmetadata.GetServiceAccountEmail()
	if err != nil {
		log.Fatalf("failed get ServiceAccountEmail.err=%+v\n", err)
		os.Exit(1)
	}
	ServiceAccountEmail = sa

	{
		client, err := cloudtasks.NewClient(context.Background())
		if err != nil {
			log.Fatalf("failed cloudtasks.NewClient.err=%+v", err)
		}
		TasksClient = client
	}

	{
		db := os.Getenv("SPANNER_DATABASE")
		SpannerClient, err = CreateSpannerClient(context.Background(), db)
		if err != nil {
			log.Fatalf("failed CreateSpannerClient. DB=%s, err=%+v", db, err)
		}
	}
}
