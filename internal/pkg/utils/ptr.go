package utils

import "github.com/google/uuid"

func UUIDToPtr(id uuid.UUID) *uuid.UUID {
	return &id
}

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}

func String(v string) *string      { return &v }
func Int(v int) *int               { return &v }
func Float32(v float32) *float32   { return &v }
func Strings(v []string) *[]string { return &v }
