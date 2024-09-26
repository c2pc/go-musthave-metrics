package metric

type Key string

func (k Key) String() string {
	return string(k)
}

type Value interface {
	float64 | int64
}

type Metric[T Value] interface {
	GetName() string
	PollStats()
	GetStats() map[Key]T
}
