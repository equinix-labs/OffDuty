package offduty

import (
	"context"
	"fmt"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"k8s.io/klog/v2"
)

var LookAhead = 60 * 24 * time.Hour

// ListSchedules returns a map of schedule names to schedule ID's.
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

// LoadSchedules returns a map of nicknames to schedule objects.
func LoadSchedules(ctx context.Context, c *pagerduty.Client, sm map[string]string, nicknames map[string]string, tz string) (map[string]*pagerduty.Schedule, error) {
	schedules := map[string]*pagerduty.Schedule{}
	until := time.Now().Add(LookAhead)
	opts := pagerduty.GetScheduleOptions{
		TimeZone: tz,
		Since:    time.Now().Format(time.RFC3339),
		Until:    until.Format(time.RFC3339),
	}

	for nick, full := range nicknames {
		sid := sm[full]
		if sid == "" {
			return schedules, fmt.Errorf("no schedule schedule named %q: %+v", full, sm)
		}

		klog.Infof("Getting schedule %q (nick=%s, id=%s, opts=%+v) ...", full, nick, sid, opts)
		s, err := c.GetScheduleWithContext(ctx, sid, opts)
		if err != nil {
			return schedules, fmt.Errorf("get schedule %q: %w", sid, err)
		}

		schedules[nick] = s
	}

	return schedules, nil
}
