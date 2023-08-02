package main

import (
	"database/sql"
	"log"

	"github.com/dravikant/simplebank/api"
	db "github.com/dravikant/simplebank/db/sqlc"
	"github.com/dravikant/simplebank/util"
	_ "github.com/lib/pq"
)

func main() {

	config, err := util.LoadConfig(".")

	if err != nil {
		log.Fatal("can not load config", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("can not connect to DB")
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("can not start server")
	}
}
