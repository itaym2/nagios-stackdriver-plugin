package main

import (
	"context"
	"fmt"
	"log"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

func main() {
	ctx := context.Background()
	client, err := monitoring.NewMetricClient(ctx)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	request := &monitoringpb.ListTimeSeriesRequest{}

	it := client.ListTimeSeries(ctx, request)

	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			log.Fatalf("Failed to fetch time series: %v", err)
		}

		fmt.Printf("%s", resp.Metadata)
	}
}
