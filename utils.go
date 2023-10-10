package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	_ "github.com/lib/pq"
	"github.com/nats-io/stan.go"
	"golang.org/x/text/encoding/charmap"
)

const (
	// Данные для подкючения к Psql
	host     = "localhost"
	port     = 5432
	user     = "poop"
	password = "hush2012"
	dbname   = "test"

	// Данные для подкючения к Nats-streaming
	clusterID = "test-cluster"
	natsURL   = "nats://localhost:4222"
	subject   = "obama"
)

// func getOrderById(w http.ResponseWriter, req *http.Request) {
// 	id := req.URL.Query().Get("order_uid")
// 	fmt.Printf("%v  %T\n", id, id)
// 	order, flag := c.Get(id)
// 	if !flag {
// 		fmt.Fprintf(w, "Не найден заказ с id: %v", id)
// 		return
// 	}

// 	fmt.Fprint(w, order)
// }

func getOrders(db *sql.DB) []Order {
	rows, qErr := db.Query("SELECT * FROM orders")

	if qErr != nil {
		decoder := charmap.Windows1251.NewDecoder()
		aaaa, bbb := decoder.Bytes([]byte(qErr.Error()))
		fmt.Println(string(aaaa), bbb, qErr)
	}

	var orders []Order

	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.Order_uid,
			&order.Track_number,
			&order.Entry,
			&order.Locale,
			&order.Internal_signature,
			&order.Customer_id,
			&order.Delivery_service,
			&order.Shardkey,
			&order.Sm_id,
			&order.Date_created,
			&order.Oof_shard,
			&order.Payment.Request_id,
			&order.Payment.Currency,
			&order.Payment.Provider,
			&order.Payment.Amount,
			&order.Payment.Payment_dt,
			&order.Payment.Bank,
			&order.Payment.Delivery_cost,
			&order.Payment.Goods_total,
			&order.Payment.Custome_fee,
		)
		order.Payment.Transaction = order.Order_uid
		if err != nil {
			fmt.Println(err)
		}

		deliveryQ := fmt.Sprintf("SELECT deliverymans.name, deliverymans.phone, deliverymans.zip, deliverymans.city, deliverymans.address, deliverymans.region, deliverymans.email FROM delivery JOIN deliverymans ON deliverymans.deliveryman_id = delivery.deliveryman_id WHERE delivery.order_uid = '%v' LIMIT 1", order.Order_uid)
		rows, errDeliveryQ := db.Query(deliveryQ)
		if errDeliveryQ != nil {
			fmt.Println(errDeliveryQ)
		}

		rows.Next()
		rows.Scan(
			&order.Delivery.Name,
			&order.Delivery.Phone,
			&order.Delivery.Zip,
			&order.Delivery.City,
			&order.Delivery.Address,
			&order.Delivery.Region,
			&order.Delivery.Email,
		)

		itemsQ := fmt.Sprintf("SELECT products.* FROM ordered_products JOIN products ON products.chrt_id = ordered_products.chrt_id WHERE ordered_products.order_uid = '%v'", order.Order_uid)
		rows, errItemsQ := db.Query(itemsQ)
		if errItemsQ != nil {
			fmt.Println(errDeliveryQ)
		}

		for rows.Next() {
			var item Item

			rows.Scan(
				&item.Chrt_id,
				&item.Track_number,
				&item.Price,
				&item.Rid,
				&item.Name,
				&item.Sale,
				&item.Size,
				&item.Total_price,
				&item.Nm_id,
				&item.Brand,
				&item.Status,
			)

			order.Items = append(order.Items, item)
		}

		//fmt.Println(order, err)
		orders = append(orders, order)
	}

	return orders
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func insertOrder(db *sql.DB, order Order) {

	deliveman := order.Delivery
	items := order.Items

	var delivery_id string
	res, err := db.Query(
		fmt.Sprintf("INSERT INTO deliverymans (name, phone, zip, city, address, region, email) VALUES ('%v', '%v', '%v', '%v', '%v', '%v', '%v') RETURNING deliveryman_id", deliveman.Name, deliveman.Phone, deliveman.Zip, deliveman.City, deliveman.Address, deliveman.Region, deliveman.Email),
	)
	if err != nil {
		fmt.Println(err)
	}

	if res.Next() {
		err = res.Scan(&delivery_id)
	}

	_, err = db.Exec(
		fmt.Sprintf("INSERT INTO orders VALUES ('%v', '%v', '%v', '%v', '%v', '%v', '%v', '%v', %v, '%v', '%v', '%v', '%v', '%v', %v, %v, '%v', %v, %v, %v)", order.Order_uid, order.Track_number, order.Entry, order.Locale, order.Internal_signature, order.Customer_id, order.Delivery_service, order.Shardkey, order.Sm_id, order.Date_created.Format(time.RFC3339), order.Oof_shard, order.Payment.Request_id, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.Payment_dt, order.Payment.Bank, order.Payment.Delivery_cost, order.Payment.Goods_total, order.Payment.Custome_fee),
	)
	if err != nil {
		fmt.Println(err)
	}

	_, err = db.Exec(
		fmt.Sprintf("INSERT INTO delivery VALUES ('%v', %v)", order.Order_uid, delivery_id),
	)
	if err != nil {
		fmt.Println(err)
	}

	for _, item := range items {
		_, err = db.Exec(
			fmt.Sprintf("INSERT INTO products VALUES (%v, '%v', %v, '%v', '%v', %v, '%v', %v, %v, '%v', %v)", item.Chrt_id, item.Track_number, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.Total_price, item.Nm_id, item.Brand, item.Status),
		)
		if err != nil {
			fmt.Println(err)
		}

		_, err = db.Exec(
			fmt.Sprintf("INSERT INTO ordered_products VALUES ('%v', %v)", order.Order_uid, item.Chrt_id),
		)
		if err != nil {
			fmt.Println(err, order.Order_uid, item.Chrt_id)
		}
	}

}

func getRandOrder() (order Order) {

	// Order
	order.Order_uid = randSeq(19)
	order.Track_number = randSeq(14)
	order.Entry = randSeq(4)
	order.Locale = "en"
	order.Internal_signature = ""
	order.Customer_id = "test"
	order.Delivery_service = "meest"
	order.Shardkey = "9"
	order.Sm_id = 99
	order.Date_created = time.Now().UTC()
	order.Oof_shard = "1"

	//Payment
	order.Payment.Transaction = order.Order_uid
	order.Payment.Request_id = randSeq(10)
	order.Payment.Currency = "USD"
	order.Payment.Provider = "wbpay"
	order.Payment.Amount = rand.Intn(2000) + 1
	order.Payment.Payment_dt = time.Now().Unix()
	order.Payment.Bank = "alpha"
	order.Payment.Delivery_cost = 1500
	order.Payment.Goods_total = 300
	order.Payment.Custome_fee = 0

	// Delivery
	order.Delivery.Name = randSeq(11)
	order.Delivery.Phone = randSeq(11)
	order.Delivery.Zip = "2639809"
	order.Delivery.City = "Kiryat Mozkin"
	order.Delivery.Address = "Ploshad Mira 15"
	order.Delivery.Region = "Kraiot"
	order.Delivery.Email = "test@gmail.com"

	// Items
	var item Item
	item.Chrt_id = rand.Intn(9999999) + 1
	item.Track_number = order.Track_number
	item.Price = rand.Intn(2000) + 100
	item.Rid = randSeq(21)
	item.Name = randSeq(8)
	item.Sale = 0
	item.Size = "0"
	item.Total_price = item.Price
	item.Nm_id = rand.Intn(9999999) + 1
	item.Brand = randSeq(13)
	item.Status = 202
	order.Items = append(order.Items, item)

	return

}

func connect2Psql() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Подключение к серверу Posgres установлено")
	return db
}

func connect2Stan(clientID string) (client stan.Conn, conErr error) {

	fmt.Println("Подключение к stan-серверу")
	for {
		client, conErr = stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
		if conErr == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return
}
