package utils

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
)

func UUIDToOAPIPtr(id uuid.UUID) *types.UUID {
	t := types.UUID(id)
	return &t
}

func ToUUID(u pgtype.UUID) uuid.UUID {
	if !u.Valid {
		return uuid.UUID{}
	}
	return uuid.UUID(u.Bytes)
}

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}

func String(v string) *string      { return &v }
func Int(v int) *int               { return &v }
func Float32(v float32) *float32   { return &v }
func Strings(v []string) *[]string { return &v }
