package storage

type Value interface {
	float64 | int64
}

type Storage[T Value] interface {
	GetName() string
	Get(key string) (T, error)
	GetString(key string) (string, error)
	GetAll() (map[string]T, error)
	GetAllString() (map[string]string, error)
	Set(key string, value T) error
	SetString(key string, value string) error
}
