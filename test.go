package main

import (
	"fmt"

	_ "github.com/lib/pq"
)

func main() {

	db := connect2Psql()
	res, err := db.Query("select count (*) from orders")
	var n int
	flag := res.Next()
	err2 := res.Scan(&n)
	fmt.Println(n, err, flag, err2)

}
