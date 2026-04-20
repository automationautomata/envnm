package metrics

import "envmn/internal/repository/cache"

type RepositoryCacheMetricsName string

func ProvideRepositoryCacheMetrics(name RepositoryCacheMetricsName) (cache.RepositoryMetrics, error) {
	return NewRepositoryCacheMetrics(string(name))
}
