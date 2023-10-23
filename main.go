package main

import (
	"fmt"

	"go-mini-test/routes"
	"go-mini-test/sqlstore"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db := sqlstore.DbConn()

	sqlstore.DropTable(db)
	sqlstore.CreateTable(db)
	sqlstore.CreateDummy(db)

	fmt.Println("started on port 10000")

	routes.HandleRequest()
}
