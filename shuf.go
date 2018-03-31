package main

import (
	"math/rand"
	"fmt"
	"time"
	"strconv"
)

func main() {
	a := []int{1, 2, 3, 4, 5, 6, 7, 8}
	//rand.Seed(time.Now().UnixNano())
	//for i := len(a) - 1; i > 0; i-- { // Fisher–Yates shuffle
	//	j := rand.Intn(i + 1)
	//	a[i], a[j] = a[j], a[i]
	//}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	for i := len(a) - 1; i > 0; i-- {
		//fmt.Println(strconv.Itoa(a[i]))
		fmt.Println("a: " + strconv.FormatInt(0, 16))
	}
}
