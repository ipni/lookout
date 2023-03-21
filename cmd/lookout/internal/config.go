package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ipni/lookout"
	"github.com/ipni/lookout/check"
	"github.com/ipni/lookout/sample"
	"gopkg.in/yaml.v2"
)

type (
	CheckerType string
	SamplerType string
	Config      struct {
		Checkers map[string]struct {
			Type          CheckerType   `yaml:"type"`
			Timeout       time.Duration `yaml:"timeout"`
			IpniEndpoint  string        `yaml:"ipniEndpoint"`
			CascadeLabels []string      `yaml:"cascadeLabels"`
			Parallelism   int           `yaml:"parallelism"`
		} `yaml:"checkers"`
		Samplers map[string]struct {
			Type SamplerType `yaml:"type"`
		} `yaml:"samplers"`
		CheckInterval       time.Duration `yaml:"checkInterval"`
		CheckersParallelism int           `yaml:"checkersParallelism"`
		SamplersParallelism int           `yaml:"samplersParallelism"`
		MetricsListenAddr   string        `yaml:"metricsListenAddr"`
	}
)

const (
	ipniNonStreamingChecker CheckerType = "ipni-non-streaming"

	saturnOrchestratorTopCids SamplerType = "saturn-orch-top-cids"
	awesomeIpfsDatasets       SamplerType = "awesome-ipfs-datasets"
)

func NewConfig(p string) (*Config, error) {
	f, err := os.Open(filepath.Clean(p))
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) ToOptions() ([]lookout.Option, error) {
	var opts []lookout.Option
	var checkers []check.Checker
	for name, cc := range c.Checkers {
		copts := []check.Option{
			check.WithName(name),
			check.WithCascadeLabels(cc.CascadeLabels),
		}
		if cc.Timeout != 0 {
			copts = append(copts, check.WithCheckTimeout(cc.Timeout))
		}
		if cc.Parallelism != 0 {
			copts = append(copts, check.WithParallelism(cc.Parallelism))
		}
		if cc.IpniEndpoint != "" {
			copts = append(copts, check.WithIpniEndpoint(cc.IpniEndpoint))
		}

		switch cc.Type {
		case ipniNonStreamingChecker:
			checker, err := check.NewIpniNonStreamingChecker(copts...)
			if err != nil {
				return nil, err
			}
			checkers = append(checkers, checker)
		default:
			return nil, fmt.Errorf("unknown checker type: %s", cc.Type)
		}
	}
	opts = append(opts, lookout.WithCheckers(checkers...))

	var samplers []sample.Sampler
	for name, sc := range c.Samplers {
		nameOpt := sample.WithName(name)
		switch sc.Type {
		case saturnOrchestratorTopCids:
			s, err := sample.NewSaturnTopCidsSampler(nameOpt)
			if err != nil {
				return nil, err
			}
			samplers = append(samplers, s)
		case awesomeIpfsDatasets:
			s, err := sample.NewAwesomeIpfsDatasets(nameOpt)
			if err != nil {
				return nil, err
			}
			samplers = append(samplers, s)
		default:
			return nil, fmt.Errorf("unknown checker type: %s", sc.Type)
		}
	}
	opts = append(opts, lookout.WithSamplers(samplers...))

	if c.CheckInterval != 0 {
		opts = append(opts, lookout.WithCheckInterval(c.CheckInterval))
	}
	if c.CheckersParallelism > 0 {
		opts = append(opts, lookout.WithCheckersParallelism(c.CheckersParallelism))
	}
	if c.SamplersParallelism > 0 {
		opts = append(opts, lookout.WithSamplersParallelism(c.SamplersParallelism))
	}
	if c.MetricsListenAddr != "" {
		opts = append(opts, lookout.WithMetricsListenAddr(c.MetricsListenAddr))
	}
	return opts, nil
}
