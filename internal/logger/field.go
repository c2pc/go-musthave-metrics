package logger

import "go.uber.org/zap"

type Field struct {
	Key   string
	Value interface{}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

func convertFields(fields ...Field) []zap.Field {
	formatedFields := make([]zap.Field, len(fields))

	for i, field := range fields {
		formatedFields[i] = zap.Any(field.Key, field.Value)
	}

	return formatedFields
}
