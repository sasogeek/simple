package main

import (
	"time"
	"fmt"
)

func main() {
	start := time.Now()
	my_integer := 1
	my_float := 1.5
	my_string := "Hello Simple!"
	my_array := []any{1, 2, 3, }
	my_dict := map[any]any{"name": "Simple", "age": 1}
	fmt.Println(my_integer)
	fmt.Println(my_float)
	fmt.Println(my_string)
	fmt.Println(my_array)
	fmt.Println(my_dict)
	fmt.Println()
	if float64(my_integer) > float64(my_float) {
		fmt.Println("Integer is bigger")
	} else {
		fmt.Println("Float is bigger")
	}
	fmt.Println()
	fmt.Println("Counting from 1 to 10:")
	for i := range 10 {
		fmt.Println(int(i) + int(1))
	}
	fmt.Println()
	x := 10
	fmt.Println("Counting down from 10")
	for int(x) > int(0) {
		fmt.Println(x)
		x = int(x) - int(1)
	}
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	elapsed := time.Since(start).Seconds()
	fmt.Printf("Elapsed time: %.8f seconds\n", elapsed)
}
