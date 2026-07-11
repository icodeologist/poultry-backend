package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

func ConnectTODB() {
	ctx := context.Background()
	connString := "postgres://poultry_admin:idontknow@localhost:5432/poultry_db?sslmode=disable"

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("Unable to create connection pool : %v\n", err)
	}
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("Unable to ping databse : %v\n", err)
	}

	fmt.Println("Successfully  connected to database")

}
