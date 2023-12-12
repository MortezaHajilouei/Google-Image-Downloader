package main

import (
	"fmt"
	"ggd/db"
	"ggd/downloader"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

func main() {

	if len(os.Args) < 7 {
		fmt.Print("add argument as: [query] [max-count] [db-host] [db-port] [db-user] [db-password] [db-name]")
		log.Fatal("args count is not correct")
	}

	query := os.Args[1]
	max := os.Args[2]
	dbHost := os.Args[3]
	dbPort := os.Args[4]
	dbUser := os.Args[5]
	dbPassword := os.Args[6]
	dbName := os.Args[7]

	maxInt, err := strconv.Atoi(max)
	if err != nil {
		log.Fatal(err)
	}

	database, err := db.Init(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	err = downloader.ProcessImages(query, maxInt, database.GetDb())
	if err != nil {
		log.Println(err)
	} else {
		log.Println("images saved in db")
	}
}
