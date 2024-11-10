package gapi

import (
	db "simple_bank/db/sqlc"
	"simple_bank/worker"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"simple_bank/pkg"

	"simple_bank/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testQueries *db.Queries
	testDB      *pgxpool.Pool
)

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	cfg := &config.Config{
		TokenSymmetricKey:   pkg.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(cfg, store, taskDistributor)
	require.NoError(t, err)
	return server
}
