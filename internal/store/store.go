package store

import (
	"sync"

	"github.com/robinbaeckman/go-hotels/internal/hotel"
	pg "github.com/robinbaeckman/go-hotels/internal/store/postgres"
)

type PostgresStore struct {
	q pg.Querier
}

type memoryStore struct {
	mu     sync.RWMutex
	hotels []hotel.Hotel
}

func NewMemoryStore() hotel.Store {
	return &memoryStore{}
}

func NewPostgresStore(q pg.Querier) hotel.Store {
	return &PostgresStore{q: q}
}
