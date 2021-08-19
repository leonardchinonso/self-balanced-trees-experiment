package main

import (
	"fmt"
	"github.com/jmcvetta/randutil"
)

func main() {
	choices := make([]randutil.Choice, 0, 2)
	choices = append(choices, randutil.Choice{Weight: 1, Item: "a"})
	choices = append(choices, randutil.Choice{Weight: 1, Item: "b"})
	fmt.Println(choices)

	res, err := randutil.WeightedChoice(choices)
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
