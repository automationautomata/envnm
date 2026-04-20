package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var repositoryMetricsLabels = []string{"table", "field", "hit"}

type repositoryCacheMetrics struct {
	counter *prometheus.CounterVec
}

func NewRepositoryCacheMetrics(name string) (*repositoryCacheMetrics, error) {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
		},
		repositoryMetricsLabels,
	)
	if err := prometheus.Register(counter); err != nil {
		return nil, err
	}
	return &repositoryCacheMetrics{
		counter: counter,
	}, nil
}

func (metrics *repositoryCacheMetrics) Inc(table, field string, hit bool) {
	metrics.counter.WithLabelValues(table, field, strconv.FormatBool(hit)).Inc()
}
