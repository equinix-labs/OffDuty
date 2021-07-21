package main

import (
	"context"
	"fmt"
	"os"

	"github.com/PagerDuty/go-pagerduty"
	"k8s.io/klog/v2"
)

func scheduleMap(ctx context.Context, c *pagerduty.Client) (map[string]string, error) {
	opts := pagerduty.ListSchedulesOptions{}
	m := map[string]string{}

	for {
		klog.Infof("Listing schedules (offset=%d) ...", opts.Offset)
		resp, err := c.ListSchedulesWithContext(ctx, opts)
		if err != nil {
			return m, fmt.Errorf("list schedules: %w", err)
		}

		for _, p := range resp.Schedules {
			m[p.Name] = p.ID
		}

		if !resp.APIListObject.More {
			break
		}

		opts.Offset += uint(len(resp.Schedules))
	}
	return m, nil
}

func main() {
	token := os.Getenv("PAGERDUTY_TOKEN")
	client := pagerduty.NewClient(token)
	ctx := context.Background()

	m, err := scheduleMap(ctx, client)
	if err != nil {
		klog.Fatalf("schedule map failed: %v", err)
	}

	s, err := client.GetSchedule(m["Provisioning Primary"], pagerduty.GetScheduleOptions{})
	if err != nil {
		klog.Fatalf("get schedule: %v", err)
	}
	klog.Infof("get: %+v", s)
}
