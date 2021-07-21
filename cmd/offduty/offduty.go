package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/tstromberg/offduty/pkg/offduty"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

var configPath = flag.String("config", "", "configuration file to load test cases from")

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	if *configPath == "" {
		klog.Exitf("--config is a required flag. See ./local-kubernetes.yaml, for example")
	}
	f, err := ioutil.ReadFile(*configPath)
	if err != nil {
		klog.Exitf("unable to read config: %v", err)
	}

	dc := &offduty.Config{}
	err = yaml.Unmarshal(f, &dc)
	if err != nil {
		klog.Exitf("unmarshal: %w", err)
	}

	klog.Infof("read config: %+v", dc)

	token := os.Getenv("PAGERDUTY_TOKEN")
	client := pagerduty.NewClient(token)
	ctx := context.Background()

	m, err := offduty.ScheduleMap(ctx, client)
	if err != nil {
		klog.Fatalf("schedule map failed: %v", err)
	}

	s, err := client.GetSchedule(m["Provisioning Primary"], pagerduty.GetScheduleOptions{})
	if err != nil {
		klog.Fatalf("get schedule: %v", err)
	}
	klog.Infof("get: %+v", s)
}
