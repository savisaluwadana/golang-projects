package main

import (
	"fmt"
	"time"
)

func makeDrink(barista string) {
	fmt.Printf("Barista %s Starting to make coffee\n", barista)
	time.Sleep(2 * time.Second)
	fmt.Printf("Barista %s finished making coffee\n", barista)
}
func main() {
	//
	fmt.Println("Coffee shop opens!")
	go makeDrink("Alice") //this makes the function concurrent
	go makeDrink("Bob")
	go makeDrink("Charlie")

	time.Sleep(4 * time.Second)
	fmt.Println("All drinks are probably ready!")
	fmt.Println("Coffee shop closes!")

}

//the issue with the sequencial execution is that the barista has to wait for the customer to finish before starting to make coffee. This can lead to inefficiencies, especially if there are multiple customers waiting. By using concurrency, we can allow the barista to start making coffee while the customer is still placing their order, improving overall efficiency and reducing wait times.
//so basically if one step fails the program halts. To solve this we use goroutines to make the tasks concurrent.
