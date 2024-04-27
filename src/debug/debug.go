package debug

import "fmt"

var PRINTEVALUATION = false

func PrintEvaluation(a ...interface{}) {
	if !PRINTEVALUATION {
		return
	}

	for _, i := range a {
		fmt.Print(i)
	}
	fmt.Println()
}
