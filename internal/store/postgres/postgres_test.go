package postgres

import (
	"context"
	"testing"
)

func TestConnect_InvalidURL(t *testing.T) {
	ctx := context.Background()

	pool, err := Connect(ctx, "not-a-valid-url")
	if err == nil {
		t.Fatal("expected error for invalid database URL, got nil")
	}
	if pool != nil {
		t.Errorf("expected nil pool, got %#v", pool)
	}
}
