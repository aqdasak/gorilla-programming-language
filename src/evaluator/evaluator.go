package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/debug"
	"monkey/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

// func Eval(node ast.Node, indent string) object.Object {
func Eval(node ast.Node, env *object.Environment, opt_indent ...string) object.Object {
	var indent string = ""
	if len(opt_indent) == 1 {
		indent = opt_indent[0]
	}
	// defer untrace(trace("Eval"))

	switch node := node.(type) {
	// Statements
	case *ast.Program:
		debug.PrintEvaluation(indent, "ast.Program")

		v := evalProgram(node, env, indent+"    ")

		if v != nil {
			debug.PrintEvaluation(indent, "ast.Program(", v.Type(), v.Inspect(), ")")
		}

		return v

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env, indent)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	// Expressions
	case *ast.IntegerLiteral:
		v := &object.Integer{Value: node.Value}

		debug.PrintEvaluation(indent, "ast.IntegerLiteral(", v.Type(), v.Inspect(), ")")

		return v

	case *ast.Boolean:
		debug.PrintEvaluation(indent, "ast.Boolean")

		v := nativeBoolToBooleanObject(node.Value, indent+"    ")

		debug.PrintEvaluation(indent, "ast.Boolean(", v.Type(), v.Inspect(), ")")

		return v

	case *ast.PrefixExpression:
		debug.PrintEvaluation(indent, "ast.PrefixExpression")

		right := Eval(node.Right, env, indent+"  ")
		if isError(right) {
			return right
		}

		v := evalPrefixExpression(node.Operator, right, indent+"    ")

		debug.PrintEvaluation(indent, "ast.PrefixExpression(", v.Type(), v.Inspect(), ")")

		return v

	case *ast.InfixExpression:
		debug.PrintEvaluation(indent, "ast.InfixExpression")

		left := Eval(node.Left, env, indent+"    ")
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env, indent+"    ")
		if isError(right) {
			return right
		}

		v := evalInfixExpression(node.Operator, left, right, indent+"    ")

		debug.PrintEvaluation(indent, "ast.InfixExpression(", v.Type(), v.Inspect(), ")")

		return v

	case *ast.BlockStatement:
		debug.PrintEvaluation(indent, "ast.BlockStatement")

		v := evalBlockStatement(node, env, indent+"|  ")

		debug.PrintEvaluation(indent, "ast.BlockStatement(", v.Type(), v.Inspect(), ")")
		return v

	case *ast.IfExpression:
		debug.PrintEvaluation(indent, "ast.IfExpression")

		v := evalIfExpression(node, env, indent+"│  ")

		debug.PrintEvaluation(indent, "ast.IfExpression(", v.Type(), v.Inspect(), ")")
		return v

	case *ast.LetStatement:
		val := Eval(node.Value, env, indent)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)

	case *ast.ReturnStatement:
		debug.PrintEvaluation(indent, "ast.ReturnStatement")

		val := Eval(node.ReturnValue, env, indent+"    ")
		if isError(val) {
			return val
		}

		v := &object.ReturnValue{Value: val}

		debug.PrintEvaluation(indent, "ast.ReturnStatement(", v.Type(), v.Inspect(), ")")

		return v
	}

	debug.PrintEvaluation(indent, "nil")
	return nil
}

func evalProgram(program *ast.Program, env *object.Environment, indent string) object.Object {
	//defer untrace(trace("evalProgram"))

	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env, indent)

		switch result := result.(type) {
		case *object.ReturnValue:
			debug.PrintEvaluation("\n Returning out of program", result.Value.Type(), result.Value.Inspect())

			return result.Value
		case *object.Error:
			debug.PrintEvaluation("\n Returning out of program", result.Type(), result.Inspect())

			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment, indent string) object.Object {
	//defer untrace(trace("evalBlockStatement"))
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env, indent)

		// Here we explicitly don’t unwrap the return value and only check the Type() of each evaluation result. If it’s object.RETURN_VALUE_OBJ we simply return the *object.ReturnValue, without unwrapping its .Value, so it stops execution in a possible outer block statement and bubbles up to evalProgram, where it finally get’s unwrapped.
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {

				debug.PrintEvaluation(indent+"\n"+indent+"Returning out of block", result.Type(), result.Inspect())

				return result
			}
		}
	}

	return result
}

func nativeBoolToBooleanObject(input bool, indent string) *object.Boolean {
	//defer untrace(trace("nativeBoolToBooleanObject"))
	if input {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(operator string, right object.Object, indent string) object.Object {
	//defer untrace(trace("evalPrefixExpression"))

	switch operator {
	case "!":
		return evalBangOperatorExpression(right, indent+"  ")
	case "-":
		return evalMinusPrefixOperatorExpression(right, indent+"  ")
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object, indent string) object.Object {
	//defer untrace(trace("evalBangOperatorExpression"))

	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object, indent string) object.Object {
	//defer untrace(trace("evalMinusPrefixOperatorExpression"))

	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(
	operator string,
	left, right object.Object,
	indent string,
) object.Object {
	//defer untrace(trace("evalInfixExpression"))

	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right, indent+"  ")
	case operator == "==":
		// We’re using pointer comparison here to check for equality between booleans. That works because we're always using pointers to our objects and in the case of booleans we only ever use two: TRUE and FALSE
		return nativeBoolToBooleanObject(left == right, indent+"  ")
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right, indent+"  ")

	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(
	operator string,
	left, right object.Object,
	indent string,
) object.Object {
	//defer untrace(trace("evalIntegerInfixExpression"))

	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal, indent+"  ")
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal, indent+"  ")
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal, indent+"  ")
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal, indent+"  ")
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment, indent string) object.Object {
	//defer untrace(trace("evalIfExpression"))

	condition := Eval(ie.Condition, env, indent)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env, indent)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env, indent)
	} else {
		return NULL
	}
}

func isTruthy(obj object.Object) bool {
	//defer untrace(trace("isTruthy"))

	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: " + node.Value)
	}
	return val
}
