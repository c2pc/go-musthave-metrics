package storage

type Value interface {
	float64 | int64
}

type Storage[T Value] interface {
	GetName() string
	Get(key string) (T, error)
	Set(key string, value T) error
}
