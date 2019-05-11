package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/olorin/nagiosplugin"
	"google.golang.org/api/iterator"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	googlepb "github.com/golang/protobuf/ptypes/timestamp"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

const (
	checkIntervalInMinutes = 5
)

func main() {
	filter := flag.String("filter", "", "time series filter")
	project := flag.String("project", "", "name of the google pubsub project containing the monitored resource")
	threshold := flag.Int64("threshold", 0, "alert when result in greater than or equal to this threashold")

	flag.Parse()

	check := nagiosplugin.NewCheck()
	defer check.Finish()
	check.AddResult(nagiosplugin.OK, "Check succeeded")

	ctx := context.Background()
	client, err := monitoring.NewMetricClient(ctx)

	if err != nil {
		check.AddResult(nagiosplugin.UNKNOWN, "Failed to perform check")
		log.Fatalf("Failed to create client: %v", err)
	}

	intervalStartTime := &googlepb.Timestamp{Seconds: time.Now().Add(-time.Minute * checkIntervalInMinutes).Unix()}
	intervalEndTime := &googlepb.Timestamp{Seconds: time.Now().Unix()}

	request := &monitoringpb.ListTimeSeriesRequest{
		Name:   fmt.Sprintf("projects/%s", *project),
		Filter: *filter,
		Interval: &monitoringpb.TimeInterval{
			StartTime: intervalStartTime,
			EndTime:   intervalEndTime,
		},
		Aggregation: &monitoringpb.Aggregation{
			AlignmentPeriod:    &duration.Duration{Seconds: 60 * checkIntervalInMinutes},
			PerSeriesAligner:   monitoringpb.Aggregation_ALIGN_MEAN,
			CrossSeriesReducer: monitoringpb.Aggregation_REDUCE_MEAN,
		},
	}

	it := client.ListTimeSeries(ctx, request)
	handleResult(it, *threshold, check)
}

func handleResult(it *monitoring.TimeSeriesIterator, threshold int64, check *nagiosplugin.Check) {
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			check.AddResult(nagiosplugin.UNKNOWN, "Failed to perform check")
			break
		}

		if err != nil {
			check.AddResult(nagiosplugin.UNKNOWN, "Failed to perform check")
			log.Fatalf("Failed to fetch time series: %v", err)
		}

		if len(resp.Points) > 1 {
			check.AddResult(nagiosplugin.UNKNOWN, "Failed to perform check, too many points in result")
			log.Fatalf("Response contains more than 1 point, please refine filter and aggregation params so that only 1 point will return")
		}

		if resp.Points[0].GetValue().GetInt64Value() > threshold {
			check.AddResult(nagiosplugin.CRITICAL, "Failed to perform check")
			break
		}
	}
}
