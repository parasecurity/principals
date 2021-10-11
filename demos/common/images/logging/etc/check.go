package main

import (
	"fmt"
	"time"
	"strings"
)

func main(){
	var pre strings.Builder
	pre.WriteString("Tarxidiamou ")
	pre.WriteString("kounouniountai")
	msg := "edw var: %d, %s"
	i := 42
	foo := "bar"
	now := time.Now().UnixNano()
	pre.WriteString(msg)
	fmt.Printf("%d ", now/1000)
	fmt.Printf(pre.String(), i, foo)
	// fmt.Printf(pre.String(), string(now), msg, i, foo)

}
