package expressionparser

import (
	"strings"

	shuntingYard "github.com/a-romash/go-shunting-yard"
)

// Парсим наше выражение и заодно проверяем на валидность
func ParseExpression(expression string) ([]*shuntingYard.RPNToken, error) {
	// parse input expression to infix notation
	infixTokens, err := shuntingYard.Scan(expression)
	if err != nil {
		return nil, err
	}

	// convert infix notation to postfix notation(RPN)
	postfixTokens, err := shuntingYard.Parse(infixTokens)
	if err != nil {
		return nil, err
	}

	// Костыль :) (если вы это видите - значит я не исправил этот костыль или просто не увидел большой смысл)
	// Здесь мы проверяем на валидность наше выражение (да, по факту мы его считаем, но чисто чтобы проверить на валидность)
	_, err = shuntingYard.Evaluate(postfixTokens)
	if err != nil {
		return nil, err
	}

	return postfixTokens, nil
}

// Создаём ключ имподентности для выражения
func CreateImpodenceKey(expression string) string {
	// Да, костыль, да, можно было по другому, но как есть
	expression = strings.ReplaceAll(expression, "+", "_p_")
	expression = strings.ReplaceAll(expression, "-", "_min_")
	expression = strings.ReplaceAll(expression, "*", "_mul_")
	expression = strings.ReplaceAll(expression, "/", "_d_")
	expression = strings.ReplaceAll(expression, "(", "_o_")
	expression = strings.ReplaceAll(expression, ")", "_c_")
	return expression
}

// type Node struct {
// 	Left     *Node
// 	Right    *Node
// 	Operator string
// 	Value    float64
// }

// // ExpressionParser парсер нашего выражения
// func ExpressionParser(s string) (*Node, error) {
// 	var (
// 		tokens    = strings.Fields(s)
// 		stack     []*Node
// 		operators []string
// 	)

// 	for _, token := range tokens {
// 		switch token {
// 		case "+", "-", "*", "/":
// 			// если токен - оператор, то
// 			// проходимся по циклу операторов для того, чтобы распределить приоритет операторов
// 			for len(operators) > 0 && precedence(operators[len(operators)-1]) >= precedence(token) {
// 				popOperator(&stack, &operators)
// 			}
// 			operators = append(operators, token)
// 		//если значение - не оператор, то это число
// 		default:
// 			value, err := strconv.ParseFloat(token, 64)
// 			if err != nil {
// 				return nil, err
// 			}
// 			stack = append(stack, &Node{Value: value})
// 		}
// 	}
// 	// закидываем оставшиеся операторы в нод стек
// 	for len(operators) > 0 {
// 		popOperator(&stack, &operators)
// 	}

// 	if len(stack) != 1 {
// 		return nil, errors.New("err")
// 	}

// 	return stack[0], nil
// }

// // precedence для установления порядка действий
// func precedence(op string) int {
// 	switch op {
// 	case "+", "-":
// 		return 1
// 	case "*", "/":
// 		return 2
// 	default:
// 		return 0
// 	}
// }

// // popOperator заносит все в в нод стек
// // и убирает за собой взятые значения :)
// func popOperator(stack *[]*Node, operators *[]string) {
// 	operator := (*operators)[len(*operators)-1]
// 	*operators = (*operators)[:len(*operators)-1]

// 	right := (*stack)[len(*stack)-1]
// 	*stack = (*stack)[:len(*stack)-1]

// 	left := (*stack)[len(*stack)-1]
// 	*stack = (*stack)[:len(*stack)-1]

// 	node := &Node{Right: right, Left: left, Operator: operator}
// 	*stack = append(*stack, node)
// }

// // EvaluatePostOrder записывает в мапу субвыражения по порядку выполнения
// func EvaluatePostOrder(node *Node, subExpressions *map[int]string, counter *int) error {
// 	if node == nil {
// 		return nil
// 	}
// 	// если нод левого выражения не пуст, то и его "парсим"
// 	if node.Left != nil {
// 		err := EvaluatePostOrder(node.Left, subExpressions, counter)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	// тут так же, только с правым
// 	if node.Right != nil {
// 		err := EvaluatePostOrder(node.Right, subExpressions, counter)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	// если оба уже пустые, то добавляем само значение
// 	if node.Left == nil && node.Right == nil {
// 		(*subExpressions)[*counter] = fmt.Sprintf("%.2f", node.Value)
// 		*counter++
// 	}

// 	// если оператор есть
// 	if node.Operator != "" {
// 		// то индексируем субвыражения
// 		lastIndex := *counter - 1
// 		secondLastIndex := lastIndex - 1
// 		subExpression := fmt.Sprintf("%s %s %s", (*subExpressions)[secondLastIndex], node.Operator, (*subExpressions)[lastIndex])
// 		// сохраняем в мапу наше субвыражение :)
// 		(*subExpressions)[*counter] = subExpression
// 		*counter++
// 	}
// 	// по итогу, в уже определенной мапе будут записаны субвыражения
// 	// по своему месту в очереди. знаю что тут много чего наверное неправильно
// 	// но хотя бы че то есть :)
// 	return nil
// }

// func ValidatedPostOrder(s string) (map[int]string, error) {
// 	node, err := ExpressionParser(s)
// 	if err != nil {
// 		return nil, err
// 	}
// 	subExps := make(map[int]string)
// 	var counter int
// 	err = EvaluatePostOrder(node, &subExps, &counter)
// 	if err != nil {
// 		return nil, err
// 	}
// 	for key, val := range subExps {
// 		if len(val) == 4 {
// 			delete(subExps, key)
// 		}
// 	}
// 	return subExps, nil
// }
