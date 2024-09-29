package api

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	db "simple_bank/db/sqlc"

	"simple_bank/pkg"

	"simple_bank/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testQueries *db.Queries
	testDB      *pgxpool.Pool
)

func newTestServer(t *testing.T, store db.Store) *Server {
	cfg := &config.Config{
		TokenSymmetricKey:   pkg.RandomString(32),
		AccessTokenDuration: time.Minute,
	}
	server, err := NewServer(cfg, store)
	require.NoError(t, err)
	return server
}

func TestMain(m *testing.M) {
	cfg, err := config.LoadConfig("../")
	if err != nil {
		log.Fatal(err)
	}

	testDB, err = pgxpool.New(context.Background(), cfg.DBSource)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	defer testDB.Close()

	testQueries = db.New(testDB)
	os.Exit(m.Run())
}
