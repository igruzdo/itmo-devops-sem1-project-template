package main

import (
	"fmt"
)

type Account struct {
	name string
	surname string
}

func main() {
	transactions := [5]int{2, 3, 5, 10, 22};

	for idx, item := range transactions {
		fmt.Printf("This is %d with idx %d \n", item, idx);
	}

	part := transactions[1:5];

	fmt.Println(part);

	newUser := Account{"Ivan", "Petrov"};

	sayMaySame(newUser)
}

func sayMaySame(user Account) {
	fmt.Print(user.name)
}