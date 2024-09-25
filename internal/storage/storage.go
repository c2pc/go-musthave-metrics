package storage

type MetricValue interface {
	String() string
}

type Storage[T MetricValue] interface {
	GetName() string
	Get(key string) (string, error)
	Set(key string, value T) error
}
