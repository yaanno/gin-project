package utils

import (
	"context"
	"time"
)

func GetContextWithTimeout(duration ...time.Duration) (context.Context, context.CancelFunc) {
	defaultTimeout := 10 * time.Second

	if len(duration) > 0 {
		defaultTimeout = duration[0]
	}

	return context.WithTimeout(context.Background(), defaultTimeout)
}

func GetContextWithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func GetContextWithDeadline() (context.Context, context.CancelFunc) {
	return context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
}

func GetContextWithValue[T any](key string, value T) context.Context {
	return context.WithValue(context.Background(), key, value)
}

func GetContextWithLogger[T any](key string, value T) context.Context {
	return context.WithValue(context.Background(), key, value)
}

func GetContextWithDB[T any](key string, value T) context.Context {
	return context.WithValue(context.Background(), key, value)
}
