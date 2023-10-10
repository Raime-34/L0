package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func main() {

	clientID := "publisher"
	client, conErr := connect2Stan(clientID)
	if conErr != nil {
		log.Fatalln("Не удалось подключиться к stan-серверу")
	}

	for {
		orderSctr := getRandOrder()
		order, err := json.Marshal(orderSctr)
		if err != nil {
			fmt.Println("Ошибка при генерации заказа")
			continue
		}
		pubErr := client.Publish(subject, order)
		if pubErr != nil {
			fmt.Println("Проблема с отправкой данных...")
			client, _ = connect2Stan(clientID)
		} else {
			fmt.Printf("Данные отправлены... id %v\n", orderSctr.Order_uid)
		}
		time.Sleep(5 * time.Second)
	}

}
