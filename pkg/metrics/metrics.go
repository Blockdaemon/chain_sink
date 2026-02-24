package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	m "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

var G Meters = &Noop{}

const metricsNamespace = "chain_sink"

func Init(ctx context.Context) (m.MeterProvider, error) {
	res, err := resource.New(ctx,
		resource.WithDetectors(nil),
	)
	if err != nil {
		return nil, err
	}

	exporter, err := prometheus.New(
		prometheus.WithNamespace(metricsNamespace),
	)
	if err != nil {
		return nil, err
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(res))

	G, err = New(provider)
	if err != nil {
		return nil, err
	}

	otel.SetMeterProvider(provider)

	err = provider.ForceFlush(ctx)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func Handler() http.Handler {
	return promhttp.Handler()
}
