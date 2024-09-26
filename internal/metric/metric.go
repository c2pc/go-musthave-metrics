package metric

type Value interface {
	float64 | int64
}

type Metric[T Value] interface {
	GetName() string
	PollStats()
	GetStats() map[string]T
}
