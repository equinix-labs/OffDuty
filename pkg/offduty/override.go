package offduty

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/relvacode/iso8601"
	"k8s.io/klog/v2"
)

var dayNameToInt = map[string]int{
	"Sunday":    0,
	"Monday":    1,
	"Tuesday":   2,
	"Wednesday": 3,
	"Thursday":  4,
	"Friday":    5,
	"Saturday":  6,
}

// Similar mapping to pagerduty.Override.
type Override struct {
	User     pagerduty.APIObject
	Schedule string
	Start    string
	End      string
}

// overlap is merely a start and end time.
type overlap struct {
	Start time.Time
	End   time.Time
}


// timeOverlap measures the amount of time that B overlaps with A.
func timeOverlap(sa time.Time, ea time.Time, sb time.Time, eb time.Time) (*overlap, error) {
	klog.Infof("finding overlap: a=%s to %s && b=%s to %s", sa, ea, sb, eb)

	// Does B start after the end of A?
	if sb.After(ea) {
		klog.Infof("no overlap, sb %s is after %s", sb, ea)
		return nil, nil
	}

	if sa == eb {
		klog.Infof("no overlap, start matches end: %s", sb)
		return nil, nil
	}

	// Does B end before the start of A?
	if sa.After(eb) {
		klog.Infof("no overlap, sa %s is after %s", sa, eb)
		return nil, nil
	}

	// At this point, we know there is some overlap!
	// If B starts after A .. the start is when B starts
	start := sb
	if start.Before(sa) {
		klog.Infof("adjusting start from %s to %s", sb, sa)
		start = sa
	}

	end := eb
	if end.After(ea) {
		klog.Infof("adjusting end from %s to %s", eb, ea)
		end = ea
	}

	klog.Infof("final overlap: %s-%s", start, end)
	if start == end {
		return nil, fmt.Errorf("calculated an invalid overlap: %s = %s", start, end)
	}

	return &overlap{Start: start, End: end}, nil
}

func parseHourMin(s string) (time.Duration, error) {
	if !strings.Contains(s, ":") {
		return time.Duration(0), fmt.Errorf("hh:mm value %q has no colon", s)
	}

	d := time.Duration(0)
	parts := strings.Split(s, ":")

	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Duration(0), fmt.Errorf("unparseable hour %q: %w", h, err)
	}
	d += (time.Duration(h) * time.Hour)

	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Duration(0), fmt.Errorf("unparseable minutes %q: %w", m, err)
	}
	d += (time.Duration(m) * time.Minute)

	klog.Infof("parsed %s as %v", s, d)
	return d, nil
}

func dailyTimeOverlaps(days []string, dStart time.Duration, dEnd time.Duration, begin time.Time, end time.Time) ([]*overlap, error) {
	if len(days) == 0 {
		days = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	}

	klog.Infof("finding daily overlaps for %s between %s/%s and %s/%s", days, dStart, dEnd, begin, end)

	matchDate := time.Date(begin.Year(), begin.Month(), begin.Day(), 0, 0, 0, 0, begin.Location())
	matchBegin := matchDate.Add(dStart)
	matchEnd := matchDate.Add(dEnd)
	klog.Infof("begin=%s matchBegin=%s matchEnd=%s", begin, matchBegin, matchEnd)

	includeDay := map[int]bool{}
	for _, d := range days {
		dn, ok := dayNameToInt[d]
		if !ok {
			return nil, fmt.Errorf("could not find day number for %q", d)
		}
		includeDay[dn] = true
	}

	var overlaps []*overlap

	// Increment by day
	for {
		if matchBegin.After(end) || matchEnd.After(end) {
			klog.Infof("breaking - time is after end: %s", end)
			break
		}

		if !includeDay[int(matchBegin.Weekday())] {
			klog.Infof("skipping %s - not in %s", matchBegin, days)
			matchBegin = matchBegin.Add(24 * time.Hour)
			matchEnd = matchEnd.Add(24 * time.Hour)
			continue
		}

		overlap, err := timeOverlap(begin, end, matchBegin, matchEnd)
		if err != nil {
			return nil, fmt.Errorf("time overlap: %w", err)
		}

		if overlap != nil {
			klog.Infof("found overlap: %+v", overlap)
			overlaps = append(overlaps, overlap)
		}

		klog.Infof("finding daily time overlaps for %s-%s between %s and %s ...", dStart, dEnd, begin, end)

		matchBegin = matchBegin.Add(24 * time.Hour)
		matchEnd = matchEnd.Add(24 * time.Hour)
	}

	return overlaps, nil
}

func CalculateOverrides(r Rule, sm map[string]*pagerduty.Schedule) ([]Override, error) {
	umap := map[string]bool{}
	for _, u := range r.Users {
		umap[u] = true
	}

	var overrides []Override

	// TODO: move this logic elsewhere
	if r.StartTime == "" {
		r.StartTime = "00:00"
	}

	if r.EndTime == "" {
		r.EndTime = "24:00"
	}
	for nick, s := range sm {
		uInfo := map[string]pagerduty.APIObject{}
		for _, u := range s.Users {
			uInfo[u.Summary] = u
		}

		klog.Infof("calculating %q overrides for %s ...", r.Name, nick, s.FinalSchedule)
		for _, rs := range s.FinalSchedule.RenderedScheduleEntries {
			if !umap[rs.User.Summary] && !umap[rs.User.ID] {
				klog.Infof("Skipping %q (not in %s)", rs.User.Summary, r.Users)
				continue
			}

			klog.Infof("%s is oncall from %s to %s ...", rs.User.Summary, rs.Start, rs.End)

			start, err := iso8601.ParseString(rs.Start)
			if err != nil {
				return nil, fmt.Errorf("iso8601 parse for %q: %w", rs.Start, err)
			}

			end, err := iso8601.ParseString(rs.End)
			if err != nil {
				return nil, fmt.Errorf("iso8601 parse for %q: %w", rs.End, err)
			}

			dStart, err := parseHourMin(r.StartTime)
			if err != nil {
				return nil, fmt.Errorf("unable to parse %q: %w", rs.End, err)
			}

			dEnd, err := parseHourMin(r.EndTime)
			if err != nil {
				return nil, fmt.Errorf("unable to parse %q: %w", rs.End, err)
			}

			klog.Infof("Calculating overlaps for %s on %s within %s to %s between %s and %s...",
				rs.User.Summary, r.Days, dStart, dEnd, start, end)
			os, err := dailyTimeOverlaps(r.Days, dStart, dEnd, start, end)
			if err != nil {
				return nil, fmt.Errorf("daily overlaps: %w", err)
			}

			for _, o := range os {
				klog.Infof("OVERRIDE on %s[%s]: %q gets %s to %s", nick, s.ID, r.Override[nick], o.Start, o.End)

				_, ok := uInfo[r.Override[nick]]
				if !ok {
					return nil, fmt.Errorf("%q is not a member of the %s[%s] schedule", r.Override[nick], s.Name, s.ID)
				}

				overrides = append(overrides, Override{
					User:     uInfo[r.Override[nick]],
					Schedule: s.ID,
					Start:    o.Start.Format(time.RFC3339),
					End:      o.End.Format(time.RFC3339),
				})
			}
		}
	}
	return overrides, nil
}
