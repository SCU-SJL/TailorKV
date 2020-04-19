package main

import (
	"fmt"
)

func main() {
	fmt.Print(^int(^uint(0) >> 1))
}
