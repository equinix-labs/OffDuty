package offduty

import (
	"context"
	"fmt"

	"github.com/PagerDuty/go-pagerduty"
	"k8s.io/klog/v2"
)

// ListSchedules returns a map of schedule names to schedule ID's
func ListSchedules(ctx context.Context, c *pagerduty.Client) (map[string]string, error) {
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

// LoadSchedules returns a map of nicknames to schedule objects
func LoadSchedules(ctx context.Context, c *pagerduty.Client, sm map[string]string, nicknames map[string]string) (map[string]*pagerduty.Schedule, error) {
	m := map[string]*pagerduty.Schedule{}
	return m, nil
}
