package storage

type MetricValue interface {
	float64 | int64
}

type Storage[T MetricValue] interface {
	GetName() string
	Get(key string) (T, error)
	Set(key string, value T) error
}
