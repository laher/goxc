package main

import "fmt"

func Test() {
	fmt.Println("This should not build")
}

func main() {

	fmt.Println("This should not get built")
}
