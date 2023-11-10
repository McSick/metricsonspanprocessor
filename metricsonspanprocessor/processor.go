package metricsonspanprocessor

import (
	"context"

	"github.com/mcsick/metricsonspanprocessor/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.uber.org/zap"
)

// Traces is a processor that can consume traces.
type Traces interface {
	component.Component
	consumer.Traces
}

// Metrics is a processor that can consume metrics.
type Metrics interface {
	component.Component
	consumer.Metrics
}

// Logs is a processor that can consume logs.
type Logs interface {
	component.Component
	consumer.Logs
}

var processorCapabilities = consumer.Capabilities{MutatesData: true}

type metricsonspanprocessor struct {
}

// NewFactory returns a new factory for the Filter processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, metadata.MetricsStability),
		processor.WithTraces(createTracesProcessor, metadata.TracesStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

type Config struct {
}

/* METRICS */

func createMetricsProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	fp, err := newM2PMetricProcessor(set.TelemetrySettings, cfg.(*Config))
	if err != nil {
		return nil, err
	}
	return processorhelper.NewMetricsProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		fp.processMetrics,
		processorhelper.WithCapabilities(processorCapabilities))
}

func newM2PMetricProcessor(set component.TelemetrySettings, cfg *Config) (*m2pMetricProcessor, error) {
	p := &m2pMetricProcessor{
		logger: set.Logger,
	}
	return p, nil
}

var (
	storedmetrics map[string]pmetric.ScopeMetricsSlice = make(map[string]pmetric.ScopeMetricsSlice)
)

type m2pMetricProcessor struct {
	logger *zap.Logger
}

func (mp *m2pMetricProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	mp.logger.Info("Made it to Metrics")
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		rm.ScopeMetrics()
		hostname, found := rm.Resource().Attributes().Get("service.name")
		if !found {
			mp.logger.Info("Error getting service name")
			continue
		}
		mp.logger.Info("Found Metrics for Service", zap.String("service", hostname.AsString()))
		storedmetrics[hostname.AsString()] = rm.ScopeMetrics()
	}
	return md, nil
}

/* TRACES */

type m2pSpanProcessor struct {
	logger *zap.Logger
}

func (sp *m2pSpanProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {

	return td, nil
}

func newM2PSpansProcessor(set component.TelemetrySettings, cfg *Config) (*m2pSpanProcessor, error) {
	p := &m2pSpanProcessor{
		logger: set.Logger,
	}
	return p, nil
}

func createTracesProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Traces,
) (processor.Traces, error) {
	fp, err := newM2PSpansProcessor(set.TelemetrySettings, cfg.(*Config))
	if err != nil {
		return nil, err
	}
	return processorhelper.NewTracesProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		fp.processTraces,
		processorhelper.WithCapabilities(processorCapabilities))
}
