package main

import (
	"fmt"
	"time"
)

func barissta() {
	fmt.Println("Barista Starting to make coffee")
	fmt.Println("Barista finished making coffee")
}
func main() {
	//
	fmt.Println("Coffee shop opens!")
	go barissta() //this makes the function concurrent
	time.Sleep(3 * time.Second)
	fmt.Println("Coffee shop closes!")

}

//the issue with the sequencial execution is that the barista has to wait for the customer to finish before starting to make coffee. This can lead to inefficiencies, especially if there are multiple customers waiting. By using concurrency, we can allow the barista to start making coffee while the customer is still placing their order, improving overall efficiency and reducing wait times.
//so basically if one step fails the program halts. To solve this we use goroutines to make the tasks concurrent.
