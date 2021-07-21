package offduty

import "github.com/PagerDuty/go-pagerduty"

// Similar mapping to pagerduty.Override
type Override struct {
	UserID string
	Start  string
	End    string
}

func CalculateOverrides(r Rule, s map[string]*pagerduty.Schedule) ([]Override, error) {
	return nil, nil
}
