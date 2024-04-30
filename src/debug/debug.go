package debug

import (
	"fmt"
	"gorilla/object"
)

var PRINTEVALUATION = false

func PrintEvaluationStart(indent, blockName string) {
	if !PRINTEVALUATION {
		return
	}

	fmt.Print(indent)
	// for _, i := range a {
	// 	fmt.Print(i)
	// }
	fmt.Println(blockName)
}

func PrintEvaluationEnd(indent, blockName string, obj object.Object) {
	if !PRINTEVALUATION {
		return
	}

	fmt.Print(indent)
	fmt.Print(blockName, "(")

	if obj != nil {
		fmt.Print(obj.Type(), obj.Inspect())
	} else {
		fmt.Print("null")
	}
	fmt.Println(")")
}
