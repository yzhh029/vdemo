package utils

import (
	"encoding/json"
	"fmt"
	"log"
)

func JsonPutLine(obj interface{}) {
	b, err := json.Marshal(obj)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	fmt.Println(string(b))
}

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}