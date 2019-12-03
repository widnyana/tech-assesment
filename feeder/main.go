package main

import (
	"bytes"
	"encoding/json"
	"github.com/bxcodec/faker"
	"log"
	"net/http"
)

type FakeData struct {
	Author string `faker:"name" json:"author"`
	Body   string `faker:"paragraph" json:"body"`
}

func main() {
	client := &http.Client{}

	log.Println("wololo!")

	for i := 1; i <= 1000; i++ {
		data := FakeData{}
		err := faker.FakeData(&data)
		if err != nil {
			log.Printf("i: %d | error faking data: %s\n", i, err.Error())
			continue
		}

		payload, err := json.Marshal(data)
		if err != nil {
			log.Printf("i: %d | error marshalin data: %s\n", i, err.Error())
			continue
		}

		req, err := http.NewRequest("POST", "http://localhost:3000/news", bytes.NewBuffer(payload))
		if err != nil {
			log.Printf("i: %d | error crafting request: %s\n", i, err.Error())
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("i: %d | error sending data: %s\n", i, err)
			continue
		}

		log.Printf("i: %d | resp: %d", i, resp.StatusCode)
	}
}
