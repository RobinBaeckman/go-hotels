package utils_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/robinbaeckman/go-hotels/internal/pkg/utils"
)

func TestUUIDToPtr(t *testing.T) {
	id := uuid.New()
	ptr := utils.UUIDToPtr(id)

	if ptr == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *ptr != id {
		t.Errorf("expected %v, got %v", id, *ptr)
	}
}

func TestPtr_Generic(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		x := 42
		ptr := utils.Ptr(x)
		if ptr == nil || *ptr != x {
			t.Errorf("expected %v, got %v", x, ptr)
		}
	})

	t.Run("string", func(t *testing.T) {
		s := "hello"
		ptr := utils.Ptr(s)
		if ptr == nil || *ptr != s {
			t.Errorf("expected %q, got %v", s, ptr)
		}
	})
}

func TestString(t *testing.T) {
	s := "test"
	ptr := utils.String(s)
	if ptr == nil || *ptr != s {
		t.Errorf("expected %q, got %v", s, ptr)
	}
}

func TestInt(t *testing.T) {
	x := 99
	ptr := utils.Int(x)
	if ptr == nil || *ptr != x {
		t.Errorf("expected %d, got %v", x, ptr)
	}
}

func TestFloat32(t *testing.T) {
	f := float32(3.14)
	ptr := utils.Float32(f)
	if ptr == nil || *ptr != f {
		t.Errorf("expected %f, got %v", f, ptr)
	}
}

func TestStrings(t *testing.T) {
	list := []string{"one", "two"}
	ptr := utils.Strings(list)
	if ptr == nil || len(*ptr) != len(list) {
		t.Fatalf("expected pointer to slice of len %d", len(list))
	}
	for i := range list {
		if (*ptr)[i] != list[i] {
			t.Errorf("expected %s, got %s", list[i], (*ptr)[i])
		}
	}
}
