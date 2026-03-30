package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var repositoryMetricsLabels = []string{"table", "field", "hit"}

type repositoryMetrics struct {
	counter *prometheus.CounterVec
}

func NewRepositoryMetrics(name string) (*repositoryMetrics, error) {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
		},
		repositoryMetricsLabels,
	)
	if err := prometheus.Register(counter); err != nil {
		return nil, err
	}
	return &repositoryMetrics{
		counter: counter,
	}, nil
}

func (metrics *repositoryMetrics) Inc(table, field string, hit bool) {
	metrics.counter.WithLabelValues(table, field, strconv.FormatBool(hit)).Inc()
}
