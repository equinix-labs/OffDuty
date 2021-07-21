package offduty

type Rule struct {
	Name      string            `yaml:"name"`
	Schedules map[string]string `yaml:"schedules"`
	Users     []string          `yaml:"users"`
	Days      []string          `yaml:"days"`
	StartTime string            `yaml:"start_time"`
	Override  map[string]string `yaml:"override"`
	Fallbacks []string          `yaml:"fallbacks"`
}

type Config struct {
	Options Options `yaml:"options"`
	Rules   []Rule  `yaml:"rules"`
}

type Options struct {
	LookAheadDays int  `yaml:"look_ahead_days"`
	DryRun        bool `yaml:"dry_run"`
}
