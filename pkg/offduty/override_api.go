package offduty

import (
	"context"
	"fmt"

	"github.com/PagerDuty/go-pagerduty"
	"k8s.io/klog/v2"
)

func ApplyOverride(ctx context.Context, client *pagerduty.Client, o Override) error {
	klog.Infof("Applying override: %+v ...", o)
	resp, err := client.CreateOverrideWithContext(ctx, o.Schedule, pagerduty.Override{User: o.User, Start: o.Start, End: o.End})
	if err != nil {
		return fmt.Errorf("failed to create override: %w", err)
	}
	klog.Infof("Override successfully applied: %+v", resp)
	return nil
}
