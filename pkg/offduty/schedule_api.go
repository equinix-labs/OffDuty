package offduty

import (
	"context"
	"fmt"

	"github.com/PagerDuty/go-pagerduty"
	"k8s.io/klog/v2"
)

// ScheduleMap returns a map of schedule names to schedule ID's
func ScheduleMap(ctx context.Context, c *pagerduty.Client) (map[string]string, error) {
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
