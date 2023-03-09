package metrics

import (
	"context"
	"net/http"
	"sync"

	"github.com/ipni/lookout/check"
	"github.com/ipni/lookout/sample"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/metric"
)

type Metrics struct {
	exporter *prometheus.Exporter

	checkLatencyHistogram   instrument.Int64Histogram
	sampleSetSizeGauge      instrument.Int64ObservableGauge
	lookupSuccessRatioGauge instrument.Float64ObservableGauge

	observablesLock     sync.RWMutex
	sampleSetSizes      map[string]int64
	lookupSuccessRatios map[attribute.Set]float64
}

func New() *Metrics {
	return &Metrics{
		sampleSetSizes:      make(map[string]int64),
		lookupSuccessRatios: make(map[attribute.Set]float64),
	}
}

func (m *Metrics) Start() error {
	var err error
	if m.exporter, err = prometheus.New(
		prometheus.WithoutUnits(),
		prometheus.WithoutScopeInfo(),
		prometheus.WithoutTargetInfo()); err != nil {
		return err
	}
	provider := metric.NewMeterProvider(metric.WithReader(m.exporter))
	meter := provider.Meter("ipni/lookout")

	if m.checkLatencyHistogram, err = meter.Int64Histogram(
		"ipni/lookout/check_latency",
		instrument.WithUnit("ms"),
		instrument.WithDescription("The elapsed time per check in milliseconds."),
	); err != nil {
		return err
	}
	if m.sampleSetSizeGauge, err = meter.Int64ObservableCounter(
		"ipni/lookout/sample_set_size",
		instrument.WithUnit("1"),
		instrument.WithDescription("The sample set size returned by samplers."),
		instrument.WithInt64Callback(m.observeSampleSetSize),
	); err != nil {
		return err
	}
	if m.lookupSuccessRatioGauge, err = meter.Float64ObservableGauge(
		"ipni/lookout/lookup_success_ratio",
		instrument.WithUnit("%"),
		instrument.WithDescription("The lookup success ratio as a number between 0 and 1."),
		instrument.WithFloat64Callback(m.observeLookupSuccessRatio),
	); err != nil {
		return err
	}
	return nil
}

func (m *Metrics) observeSampleSetSize(_ context.Context, observer instrument.Int64Observer) error {
	m.observablesLock.RLock()
	defer m.observablesLock.RUnlock()
	for sampler, size := range m.sampleSetSizes {
		observer.Observe(size, attribute.String("sampler", sampler))
	}
	return nil
}

func (m *Metrics) observeLookupSuccessRatio(_ context.Context, observer instrument.Float64Observer) error {
	m.observablesLock.RLock()
	defer m.observablesLock.RUnlock()
	for attrs, ratio := range m.lookupSuccessRatios {
		observer.Observe(ratio, attrs.ToSlice()...)
	}
	return nil
}

func (m *Metrics) NotifySampleSet(_ context.Context, ss *sample.Set) {
	m.observablesLock.Lock()
	defer m.observablesLock.Unlock()
	m.sampleSetSizes[ss.Name] = int64(len(ss.Multihashes))
}

func (m *Metrics) NotifyCheckResults(ctx context.Context, results *check.Results) {
	checkerAttr := attribute.String("checker", results.CheckerName)
	sampleAttr := attribute.String("sample", results.SampleSetName)
	var success int
	for _, result := range results.Results {
		if result.StatusCode == http.StatusOK {
			success++
		}
		// TODO check error for context timeout or cancellation
		m.checkLatencyHistogram.Record(
			ctx,
			result.Elapsed.Milliseconds(),
			checkerAttr,
			sampleAttr,
			attribute.Int("status", result.StatusCode),
			attribute.Bool("error", result.Err != nil),
			attribute.String("timeout", result.Timeout.String()),
			attribute.Bool("streaming", result.Streaming),
		)
	}
	var ratio float64
	if total := len(results.Results); total > 0 {
		ratio = float64(success) / float64(total)
	}
	// Store ratio even if it is zero so that it can be used for alerting.
	// If it is zero, the chances are something is not right.
	m.observablesLock.Lock()
	defer m.observablesLock.Unlock()
	m.lookupSuccessRatios[attribute.NewSet(checkerAttr, sampleAttr)] = ratio
}

func (m *Metrics) Shutdown(ctx context.Context) error {
	var err error
	if m.exporter != nil {
		err = m.exporter.Shutdown(ctx)
	}
	return err
}
