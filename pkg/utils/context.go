package utils

import (
	"context"
	"time"
)

func GetContextWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
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
