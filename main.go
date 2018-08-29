package main

import (
	"time"
	"fmt"
)

func main()  {
	fmt.Printf("time is %v \n", time.Now().Unix())
	time.Sleep(time.Second)
	fmt.Printf("time is %v", time.Now().Unix())
}



