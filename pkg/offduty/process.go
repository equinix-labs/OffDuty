package offduty

import (
	"context"
	"fmt"

	"github.com/PagerDuty/go-pagerduty"
	"k8s.io/klog/v2"
)

func Process(ctx context.Context, client *pagerduty.Client, cfg *Config) ([]Override, error) {
	sm, err := ListSchedules(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("list schedules: %w", err)
	}

	overrides := []Override{}

	for _, r := range cfg.Rules {
		s, err := LoadSchedules(ctx, client, sm, r.Schedules)
		if err != nil {
			return nil, fmt.Errorf("get schedules: %w", err)
		}

		os, err := CalculateOverrides(r, s)
		if err != nil {
			return nil, fmt.Errorf("calculate: %w", err)
		}

		for _, o := range os {
			klog.Infof("override for %q: %+v", o)
			if !cfg.Options.DryRun {
				err := ApplyOverride(ctx, client, o)
				if err != nil {
					return nil, fmt.Errorf("apply override for %+v: %w", o, err)
				}
			}
			overrides = append(overrides, o)
		}
	}

	return overrides, nil
	/*
		s, err := client.GetSchedule(m["Provisioning Primary"], pagerduty.GetScheduleOptions{})
		if err != nil {
			klog.Fatalf("get schedule: %v", err)
		}
		klog.Infof("get: %+v", s)
	*/
}
