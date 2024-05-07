package evaluator

import (
	"fmt"
	"gorilla/ast"
	"gorilla/debug"
	"gorilla/object"
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
		debug.PrintEvaluationStart(indent, "ast.Program")

		v := evalProgram(node, env, indent+"    ")

		if v != nil {
			debug.PrintEvaluationEnd(indent, "ast.Program", v)
		}

		return v

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env, indent)

	// Expressions
	case *ast.IntegerLiteral:
		v := &object.Integer{Value: node.Value}

		debug.PrintEvaluationEnd(indent, "ast.IntegerLiteral", v)

		return v

	case *ast.Boolean:
		debug.PrintEvaluationStart(indent, "ast.Boolean")

		v := nativeBoolToBooleanObject(node.Value)

		debug.PrintEvaluationEnd(indent, "ast.Boolean", v)

		return v

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}

		return &object.Array{Elements: elements}

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)

	case *ast.HashLiteral:
		return evalHashLiteral(node, env)

	case *ast.PrefixExpression:
		debug.PrintEvaluationStart(indent, "ast.PrefixExpression")

		right := Eval(node.Right, env, indent+"  ")
		if isError(right) {
			return right
		}

		v := evalPrefixExpression(node.Operator, right, indent+"    ")

		debug.PrintEvaluationEnd(indent, "ast.PrefixExpression", v)

		return v

	case *ast.InfixExpression:
		debug.PrintEvaluationStart(indent, "ast.InfixExpression")

		left := Eval(node.Left, env, indent+"    ")
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env, indent+"    ")
		if isError(right) {
			return right
		}

		v := evalInfixExpression(node.Operator, left, right, indent+"    ")

		debug.PrintEvaluationEnd(indent, "ast.InfixExpression", v)

		return v

	case *ast.BlockStatement:
		debug.PrintEvaluationStart(indent, "ast.BlockStatement")

		v := evalBlockStatement(node, env, indent+"|  ")

		debug.PrintEvaluationEnd(indent, "ast.BlockStatement", v)
		return v

	case *ast.IfExpression:
		debug.PrintEvaluationStart(indent, "ast.IfExpression")

		v := evalIfExpression(node, env, indent+"│  ")

		debug.PrintEvaluationEnd(indent, "ast.IfExpression", v)
		return v

	case *ast.WhileExpression:
		return evalWhileExpression(node, env)

	case *ast.LetStatement:
		val := Eval(node.Value, env, indent)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)

	case *ast.ReturnStatement:
		debug.PrintEvaluationStart(indent, "ast.ReturnStatement")

		val := Eval(node.ReturnValue, env, indent+"    ")
		if isError(val) {
			return val
		}

		v := &object.ReturnValue{Value: val}

		debug.PrintEvaluationEnd(indent, "ast.ReturnStatement", v)

		return v
	}

	debug.PrintEvaluationStart(indent, "nil")
	return nil
}

func evalProgram(program *ast.Program, env *object.Environment, indent string) object.Object {
	//defer untrace(trace("evalProgram"))

	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env, indent)

		switch result := result.(type) {
		case *object.ReturnValue:
			debug.PrintEvaluationEnd("", "\n Returning out of program", result.Value)

			return result.Value
		case *object.Error:
			debug.PrintEvaluationEnd("", "\n Returning out of program", result)

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

				debug.PrintEvaluationEnd(indent, "\n"+indent+"Returning out of block", result)

				return result
			}
		}
	}

	return result
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
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

	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)

	case operator == "==":
		// We’re using pointer comparison here to check for equality between booleans. That works because we're always using pointers to our objects and in the case of booleans we only ever use two: TRUE and FALSE
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)

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
	case "%":
		return &object.Integer{Value: leftVal % rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
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
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: " + node.Value)
}

func evalExpressions(
	exps []ast.Expression,
	env *object.Environment,
) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)

		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		return fn.Fn(args...)

	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func evalStringInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	if operator == "+" {
		leftVal := left.(*object.String).Value
		rightVal := right.(*object.String).Value

		return &object.String{Value: leftVal + rightVal}

	} else if operator == "==" {
		leftVal := left.(*object.String).Value
		rightVal := right.(*object.String).Value

		return &object.Boolean{Value: leftVal == rightVal}

	} else {
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	case left.Type() == object.STRING_OBJ:
		return evalStringIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)

	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

func evalHashLiteral(
	node *ast.HashLiteral,
	env *object.Environment,
) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalStringIndexExpression(str, index object.Object) object.Object {
	stringObject := str.(*object.String)

	idx := index.(*object.Integer).Value
	max := int64(len(stringObject.Value) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return &object.String{Value: string(stringObject.Value[idx])}

}

func evalWhileExpression(we *ast.WhileExpression, env *object.Environment) object.Object {
	condition := Eval(we.Condition, env)
	if isError(condition) {
		return condition
	}

	var result object.Object = NULL
	for {
		condition := Eval(we.Condition, env)

		if !isTruthy(condition) {
			if result == nil {
				return NULL
			}
			return result
		}

		result = Eval(we.Body, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result
		}

	}
}
