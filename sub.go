package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/stan.go"
	"github.com/patrickmn/go-cache"
	"github.com/rs/cors"
)

func main() {

	clientID := "subscriber"
	subject := "obama"
	validation := validator.New(validator.WithRequiredStructEnabled())

	client, conErr := connect2Stan(clientID)
	if conErr != nil {
		log.Fatalln("Не удалось подключиться к stan-серверу")
	}

	db := connect2Psql()
	c := cache.New(cache.NoExpiration, cache.NoExpiration)

	client.Subscribe(subject, func(msg *stan.Msg) {

		fmt.Println("Получены данные...")

		var order Order
		json.Unmarshal(msg.Data, &order)
		valErr := validation.Struct(order)

		if valErr != nil {
			fmt.Println("\tДанные не прошли валидацию")
			return
		}

		c.Set(order.Order_uid, string(msg.Data), cache.NoExpiration)
		insertOrder(db, order)

	})

	orders := getOrders(db)
	for _, order := range orders {
		c.Set(order.Order_uid, order, cache.NoExpiration)
		// fmt.Println(order.Order_uid)
	}

	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Разрешить доступ с любого домена (не рекомендуется для продакшена)
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"*"},
	})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("order_uid")
		fmt.Printf("%v  %T\n", id, id)
		order, flag := c.Get(id)
		if !flag {
			fmt.Fprintf(w, "Не найден заказ с id: %v", id)
			return
		}

		fmt.Fprint(w, order)
	})
	http.Handle("/", cors.Handler(handler))
	http.ListenAndServe(":8080", nil)

	select {}
}
