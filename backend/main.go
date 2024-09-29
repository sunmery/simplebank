package main

import (
	"context"
	"fmt"

	"simple_bank/config"

	"simple_bank/api"

	"github.com/jackc/pgx/v5/pgxpool"
	db "simple_bank/db/sqlc"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	conn, newDBErr := pgxpool.New(context.Background(), cfg.DBSource)
	if newDBErr != nil {
		panic(fmt.Sprintf("Unable to connect to database: %v", err))
	}

	store := db.NewStore(conn)
	server, newServerErr := api.NewServer(cfg, store)
	if newServerErr != nil {
		panic(fmt.Sprintf("Unable to create server: %v", err))
	}

	err = server.Start(cfg.ServerAddress)
	if err != nil {
		panic(fmt.Sprintf("Unable to start server: %v", err))
	}
}
