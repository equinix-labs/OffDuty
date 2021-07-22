package offduty

import (
	"github.com/PagerDuty/go-pagerduty"
	"k8s.io/klog/v2"
)

// Similar mapping to pagerduty.Override
type Override struct {
	UserID string
	Start  string
	End    string
}

/*
func timeOverlap(sa time.Time, ea time.Time, sb time.Time, eb time.Time) (time.Time, time.Time) {

}
*/

func CalculateOverrides(r Rule, sm map[string]*pagerduty.Schedule) ([]Override, error) {
	umap := map[string]bool{}
	for _, u := range r.Users {
		umap[u] = true
	}

	for nick, s := range sm {
		klog.Infof("calculating %q overrides for %s ...", r.Name, nick, s.FinalSchedule)
		for _, rs := range s.FinalSchedule.RenderedScheduleEntries {
			if !umap[rs.User.Summary] && !umap[rs.User.ID] {
				klog.Infof("Skipping %q (not in %s)", rs.User.Summary, r.Users)
				continue
			}
		}
	}
	return nil, nil
}
