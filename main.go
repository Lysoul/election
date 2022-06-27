package main

import (
	"database/sql"
	"log"

	"election/api"
	db "election/db/sqlc"
	"election/util"

	_ "github.com/lib/pq"
)

func main() {

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load configuartion: ", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create a server: ", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}

}
