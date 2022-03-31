package evaluator

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"scream/ast"
	"scream/object"
)

var (
	NULL    = &object.Null{}
	TRUE    = &object.Boolean{Value: true}
	FALSE   = &object.Boolean{Value: false}
	PRAGMAS = make(map[string]int)
	CTX     = context.Background()
)

var builtins = map[string]*object.Builtin{}

func Eval(node ast.Node, env *object.Environment) object.Object {
	return EvalContext(context.Background(), node, env)
}

func EvalContext(ctx context.Context, node ast.Node, env *object.Environment) object.Object {

	select {
	case <-ctx.Done():
		return &object.Error{Message: ctx.Err().Error()}
	default:
	}

	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.NullLiteral:
		return NULL
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.PostfixExpression:
		return evalPostfixExpression(env, node.Operator, node)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		res := evalInfixExpression(node.Operator, left, right, env)
		if isError(res) {
			fmt.Printf("Error: %s\n", res.Inspect())
			if PRAGMAS["strict"] == 1 {
				os.Exit(1)
			}
		}
		return (res)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.TernaryExpression:
		return evalTernaryExpression(node, env)
	case *ast.ForLoopExpression:
		return evalForLoopExpression(node, env)
	case *ast.ForeachStatement:
		return evalForeachExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
		return val
	case *ast.ConstStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.SetConst(node.Name.Value, val)
		return val
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		defaults := node.Defaults
		return &object.Function{Parameters: params, Env: env, Body: body, Defaults: defaults}
	case *ast.FunctionDefineLiteral:
		params := node.Parameters
		body := node.Body
		defaults := node.Defaults
		env.Set(node.TokenLiteral(), &object.Function{Parameters: params, Env: env, Body: body, Defaults: defaults})
		return NULL
	case *ast.ObjectCallExpression:
		res := evalObjectCallExpression(node, env)
		if isError(res) {
			fmt.Fprintf(os.Stderr, "Error calling object-method %s\n", res.Inspect())
			if PRAGMAS["strict"] == 1 {
				os.Exit(1)
			}
		}
		return res
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpression(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		res := applyFunction(env, function, args)
		if isError(res) {
			fmt.Fprintf(os.Stderr, "Error calling `%s` : %s\n", node.Function, res.Inspect())
			if PRAGMAS["strict"] == 1 {
				os.Exit(1)
			}
			return res
		}
		return res

	case *ast.ArrayLiteral:
		elements := evalExpression(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.RegexpLiteral:
		return &object.Regexp{Value: node.Value, Flags: node.Flags}
	case *ast.BacktickLiteral:
		return backTickOperation(node.Value)
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
	case *ast.AssignStatement:
		return evalAssignStatement(node, env)
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	case *ast.SwitchExpression:
		return evalSwitchStatement(node, env)
	}
	return nil
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range block.Statements {
		result = Eval(statement, env)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalPostfixExpression(env *object.Environment, operator string, node *ast.PostfixExpression) object.Object {
	switch operator {
	case "++":
		val, ok := env.Get(node.Token.Literal)
		if !ok {
			return newError("%s is unknown", node.Token.Literal)
		}

		switch arg := val.(type) {
		case *object.Integer:
			v := arg.Value
			env.Set(node.Token.Literal, &object.Integer{Value: v + 1})
			return arg
		default:
			return newError("%s is not an int", node.Token.Literal)

		}
	case "--":
		val, ok := env.Get(node.Token.Literal)
		if !ok {
			return newError("%s is unknown", node.Token.Literal)
		}

		switch arg := val.(type) {
		case *object.Integer:
			v := arg.Value
			env.Set(node.Token.Literal, &object.Integer{Value: v - 1})
			return arg
		default:
			return newError("%s is not an int", node.Token.Literal)
		}
	default:
		return newError("unknown operator: %s", operator)
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
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

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	switch obj := right.(type) {
	case *object.Integer:
		return &object.Integer{Value: -obj.Value}
	case *object.Float:
		return &object.Float{Value: -obj.Value}
	default:
		return newError("unknown operator: -%s", right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object, env *object.Environment) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalFloatIntegerInfixExpression(operator, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalIntegerFloatInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "&&":
		return nativeBoolToBooleanObject(objectToNativeBoolean(left) && objectToNativeBoolean(right))
	case operator == "||":
		return nativeBoolToBooleanObject(objectToNativeBoolean(left) || objectToNativeBoolean(right))
	case operator == "!~":
		return notMatches(left, right)
	case operator == "~=":
		return matches(left, right, env)

	case operator == "==":
		return nativeBoolToBooleanObject(left == right)

	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ:
		return evalBooleanInfixExpression(operator, left, right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func matches(left, right object.Object, env *object.Environment) object.Object {

	str := left.Inspect()

	if right.Type() != object.REGEXP_OBJ {
		return newError("regexp required for regexp-match, given %s", right.Type())
	}

	val := right.(*object.Regexp).Value
	if right.(*object.Regexp).Flags != "" {
		val = "(?" + right.(*object.Regexp).Flags + ")" + val
	}

	r, err := regexp.Compile(val)

	if err != nil {
		return newError("error compiling regexp '%s': %s", right.Inspect(), err)
	}

	res := r.FindStringSubmatch(str)

	// Do we have any captures?
	if len(res) > 1 {
		for i := 1; i < len(res); i++ {
			env.Set(fmt.Sprintf("$%d", i), &object.String{Value: res[i]})
		}
	}

	if len(res) > 0 {
		return TRUE
	}

	return FALSE
}

func notMatches(left, right object.Object) object.Object {
	str := left.Inspect()

	if right.Type() != object.REGEXP_OBJ {
		return newError("regexp required for regexp-match, given %s", right.Type())
	}

	val := right.(*object.Regexp).Value
	if right.(*object.Regexp).Flags != "" {
		val = "(?" + right.(*object.Regexp).Flags + ")" + val
	}

	r, err := regexp.Compile(val)

	if err != nil {
		return newError("error compiling regexp '%s': %s", right.Inspect(), err)
	}

	if r.MatchString(str) {
		return FALSE
	}

	return TRUE
}
func evalBooleanInfixExpression(operator string, left, right object.Object) object.Object {
	l := &object.String{Value: string(left.Inspect())}
	r := &object.String{Value: string(right.Inspect())}

	switch operator {
	case "<":
		return evalStringInfixExpression(operator, l, r)
	case "<=":
		return evalStringInfixExpression(operator, l, r)
	case ">":
		return evalStringInfixExpression(operator, l, r)
	case ">=":
		return evalStringInfixExpression(operator, l, r)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value
	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "+=":
		return &object.Integer{Value: leftVal + rightVal}
	case "%":
		return &object.Integer{Value: leftVal % rightVal}
	case "**":
		return &object.Integer{Value: int64(math.Pow(float64(leftVal), float64(rightVal)))}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "-=":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "*=":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "/=":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	case "..":
		len := int(rightVal-leftVal) + 1
		array := make([]object.Object, len)
		i := 0
		for i < len {
			array[i] = &object.Integer{Value: leftVal}
			leftVal++
			i++
		}
		return &object.Array{Elements: array}
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}
func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Float).Value
	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "+=":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "-=":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "*=":
		return &object.Float{Value: leftVal * rightVal}
	case "**":
		return &object.Float{Value: math.Pow(leftVal, rightVal)}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "/=":
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalFloatIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := float64(right.(*object.Integer).Value)
	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "+=":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "-=":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "*=":
		return &object.Float{Value: leftVal * rightVal}
	case "**":
		return &object.Float{Value: math.Pow(leftVal, rightVal)}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "/=":
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIntegerFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := float64(left.(*object.Integer).Value)
	rightVal := right.(*object.Float).Value
	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "+=":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "-=":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "*=":
		return &object.Float{Value: leftVal * rightVal}
	case "**":
		return &object.Float{Value: math.Pow(leftVal, rightVal)}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "/=":
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	l := left.(*object.String)
	r := right.(*object.String)

	switch operator {
	case "==":
		return nativeBoolToBooleanObject(l.Value == r.Value)
	case "!=":
		return nativeBoolToBooleanObject(l.Value != r.Value)
	case ">=":
		return nativeBoolToBooleanObject(l.Value >= r.Value)
	case ">":
		return nativeBoolToBooleanObject(l.Value > r.Value)
	case "<=":
		return nativeBoolToBooleanObject(l.Value <= r.Value)
	case "<":
		return nativeBoolToBooleanObject(l.Value < r.Value)
	case "+":
		return &object.String{Value: l.Value + r.Value}
	case "+=":
		return &object.String{Value: l.Value + r.Value}
	}

	return newError("unknown operator: %s %s %s",
		left.Type(), operator, right.Type())
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	var permit []string
	i := 1
	for i < 32 {
		permit = append(permit, fmt.Sprintf("$%d", i))
		i++
	}
	nEnv := object.NewTemporaryScope(env, permit)
	condition := Eval(ie.Condition, nEnv)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(ie.Consequence, nEnv)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, nEnv)
	} else {
		return NULL
	}
}

func evalTernaryExpression(te *ast.TernaryExpression, env *object.Environment) object.Object {

	condition := Eval(te.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(te.IfTrue, env)
	}
	return Eval(te.IfFalse, env)
}

func evalAssignStatement(a *ast.AssignStatement, env *object.Environment) (val object.Object) {
	evaluated := Eval(a.Value, env)
	if isError(evaluated) {
		return evaluated
	}

	switch a.Operator {
	case "+=":
		current, ok := env.Get(a.Name.String())
		if !ok {
			return newError("%s is unknown", a.Name.String())
		}

		res := evalInfixExpression("+=", current, evaluated, env)
		if isError(res) {
			fmt.Printf("Error handling += %s\n", res.Inspect())
			return res
		}

		env.Set(a.Name.String(), res)
		return res

	case "-=":

		current, ok := env.Get(a.Name.String())
		if !ok {
			return newError("%s is unknown", a.Name.String())
		}

		res := evalInfixExpression("-=", current, evaluated, env)
		if isError(res) {
			fmt.Printf("Error handling -= %s\n", res.Inspect())
			return res
		}

		env.Set(a.Name.String(), res)
		return res

	case "*=":
		current, ok := env.Get(a.Name.String())
		if !ok {
			return newError("%s is unknown", a.Name.String())
		}

		res := evalInfixExpression("*=", current, evaluated, env)
		if isError(res) {
			fmt.Printf("Error handling *= %s\n", res.Inspect())
			return res
		}

		env.Set(a.Name.String(), res)
		return res

	case "/=":

		current, ok := env.Get(a.Name.String())
		if !ok {
			return newError("%s is unknown", a.Name.String())
		}

		res := evalInfixExpression("/=", current, evaluated, env)
		if isError(res) {
			fmt.Printf("Error handling /= %s\n", res.Inspect())
			return res
		}

		env.Set(a.Name.String(), res)
		return res

	case "=":
		if PRAGMAS["strict"] == 1 {
			_, ok := env.Get(a.Name.String())
			if !ok {
				fmt.Printf("Setting unknown variable '%s' is a bug under strict-pragma!\n", a.Name.String())
				os.Exit(1)
			}
		}

		env.Set(a.Name.String(), evaluated)
	}
	return evaluated
}

func evalSwitchStatement(se *ast.SwitchExpression, env *object.Environment) object.Object {

	obj := Eval(se.Value, env)

	for _, opt := range se.Choices {

		if opt.Default {
			continue
		}

		for _, val := range opt.Expr {

			out := Eval(val, env)

			// Is it a literal match?
			if obj.Type() == out.Type() &&
				(obj.Inspect() == out.Inspect()) {

				blockOut := evalBlockStatement(opt.Block, env)
				return blockOut
			}

			if out.Type() == object.REGEXP_OBJ {

				m := matches(obj, out, env)
				if m == TRUE {

					out := evalBlockStatement(opt.Block, env)
					return out

				}
			}
		}
	}

	for _, opt := range se.Choices {

		// skip default
		if opt.Default {

			out := evalBlockStatement(opt.Block, env)
			return out
		}
	}

	return nil
}

func evalForLoopExpression(fle *ast.ForLoopExpression, env *object.Environment) object.Object {
	rt := &object.Boolean{Value: true}
	for {
		condition := Eval(fle.Condition, env)
		if isError(condition) {
			return condition
		}
		if isTruthy(condition) {
			rt := Eval(fle.Consequence, env)
			if !isError(rt) && (rt.Type() == object.RETURN_VALUE_OBJ || rt.Type() == object.ERROR_OBJ) {
				return rt
			}
		} else {
			break
		}
	}
	return rt
}

func evalForeachExpression(fle *ast.ForeachStatement, env *object.Environment) object.Object {

	val := Eval(fle.Value, env)

	helper, ok := val.(object.Iterable)
	if !ok {
		return newError("%s object doesn't implement the Iterable interface", val.Type())
	}

	var permit []string
	permit = append(permit, fle.Ident)
	if fle.Index != "" {
		permit = append(permit, fle.Index)
	}

	child := object.NewTemporaryScope(env, permit)

	helper.Reset()

	ret, idx, ok := helper.Next()

	for ok {

		child.Set(fle.Ident, ret)

		idxName := fle.Index
		if idxName != "" {
			child.Set(fle.Index, idx)
		}

		rt := Eval(fle.Body, child)

		if !isError(rt) && (rt.Type() == object.RETURN_VALUE_OBJ || rt.Type() == object.ERROR_OBJ) {
			return rt
		}
		ret, idx, ok = helper.Next()
	}

	return &object.Null{}
}

func isTruthy(obj object.Object) bool {
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

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		result = Eval(statement, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
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

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	fmt.Fprintf(os.Stderr, "identifier not found: %s\n", node.Value)
	if PRAGMAS["strict"] == 1 {
		os.Exit(1)
	}
	return newError("identifier not found: " + node.Value)
}

func evalExpression(exps []ast.Expression, env *object.Environment) []object.Object {
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

func splitCommand(input string) []string {
	r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
	res := r.FindAllString(input, -1)

	var result []string
	for _, e := range res {
		result = append(result, trimQuotes(e, '"'))
	}
	return (result)
}
func trimQuotes(in string, c byte) string {
	if len(in) >= 2 {
		if in[0] == c && in[len(in)-1] == c {
			return in[1 : len(in)-1]
		}
	}
	return in
}

func backTickOperation(command string) object.Object {

	toExec := splitCommand(command)
	cmd := exec.Command(toExec[0], toExec[1:]...)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()

	if err != nil && err != err.(*exec.ExitError) {
		fmt.Printf("Failed to run '%s' -> %s\n", command, err.Error())
		return NULL
	}

	stdout := &object.String{Value: outb.String()}
	stderr := &object.String{Value: errb.String()}
	stdoutKey := &object.String{Value: "stdout"}
	stdoutHash := object.HashPair{Key: stdoutKey, Value: stdout}

	stderrKey := &object.String{Value: "stderr"}
	stderrHash := object.HashPair{Key: stderrKey, Value: stderr}

	newHash := make(map[object.HashKey]object.HashPair)
	newHash[stdoutKey.HashKey()] = stdoutHash
	newHash[stderrKey.HashKey()] = stderrHash

	return &object.Hash{Pairs: newHash}
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
		return newError("index operator not support:%s", left.Type())

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

func evalStringIndexExpression(input, index object.Object) object.Object {
	str := input.(*object.String).Value
	idx := index.(*object.Integer).Value
	max := int64(len(str))
	if idx < 0 || idx > max {
		return NULL
	}

	chars := []rune(str)

	ret := chars[idx]

	return &object.String{Value: string(ret)}
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
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

func applyFunction(env *object.Environment, fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendEnv)
		return upwrapReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(env, args...)
	default:
		return newError("not a function: %s", fn.Type())
	}

}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for key, val := range fn.Defaults {
		env.Set(key, Eval(val, env))
	}
	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			env.Set(param.Value, args[paramIdx])
		}
	}
	return env
}

func upwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func RegisterBuiltin(name string, fun object.BuiltinFunction) {
	builtins[name] = &object.Builtin{Fn: fun}
}

func SetContext(ctx context.Context) {
	CTX = ctx
}

func evalObjectCallExpression(call *ast.ObjectCallExpression, env *object.Environment) object.Object {

	obj := Eval(call.Object, env)
	if method, ok := call.Call.(*ast.CallExpression); ok {

		args := evalExpression(call.Call.(*ast.CallExpression).Arguments, env)
		ret := obj.InvokeMethod(method.Function.String(), *env, args...)
		if ret != nil {
			return ret
		}

		attempts := []string{}
		attempts = append(attempts, strings.ToLower(string(obj.Type())))
		attempts = append(attempts, "object")

		for _, prefix := range attempts {
			name := prefix + "." + method.Function.String()

			if fn, ok := env.Get(name); ok {

				extendEnv := extendFunctionEnv(fn.(*object.Function), args)

				extendEnv.Set("self", obj)

				evaluated := Eval(fn.(*object.Function).Body, extendEnv)
				obj = upwrapReturnValue(evaluated)
				return obj
			}
		}

	}

	return newError("Failed to invoke method: %s", call.Call.(*ast.CallExpression).Function.String())
}

func objectToNativeBoolean(o object.Object) bool {
	if r, ok := o.(*object.ReturnValue); ok {
		o = r.Value
	}
	switch obj := o.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.String:
		return obj.Value != ""
	case *object.Regexp:
		return obj.Value != ""
	case *object.Null:
		return false
	case *object.Integer:
		if obj.Value == 0 {
			return false
		}
		return true
	case *object.Float:
		if obj.Value == 0.0 {
			return false
		}
		return true
	case *object.Array:
		if len(obj.Elements) == 0 {
			return false
		}
		return true
	case *object.Hash:
		if len(obj.Pairs) == 0 {
			return false
		}
		return true
	default:
		return true
	}
}
