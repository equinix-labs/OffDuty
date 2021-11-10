package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/equinix-labs/OffDuty/pkg/offduty"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

var configPath = flag.String("config", "", "configuration file to load test cases from")
var dryRun = flag.Bool("dry-run", false, "pretend to make changes, but do not make them")

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

	cfg := &offduty.Config{}
	err = yaml.Unmarshal(f, &cfg)
	if err != nil {
		klog.Exitf("unmarshal: %w", err)
	}

	if *dryRun {
		cfg.Options.DryRun = true
	}

	klog.Infof("read config: %+v", cfg)

	token := os.Getenv("PAGERDUTY_TOKEN")
	if token == "" {
		klog.Exitf("$PAGERDUTY_TOKEN environment variable is unset. Get one from https://support.pagerduty.com/docs/generating-api-keys")
	}

	client := pagerduty.NewClient(token)
	ctx := context.Background()

	os, err := offduty.Process(ctx, client, cfg)
	if err != nil {
		klog.Fatalf("failed processing: %v", err)
	}

	klog.Infof("Successfully processed %d rules, resulting in %d overrides", len(cfg.Rules), len(os))
}
