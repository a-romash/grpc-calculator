package agent

import (
	"calculator/internal/model"
	"calculator/pkg/config"
	"log"
	"sync"
	"time"

	shuntingYard "github.com/a-romash/go-shunting-yard"
)

type Agent struct {
	mu          sync.Mutex
	Calculators []*Calculator
	Queue       chan *model.ExpressionPart
}

func NewAgent(countCalculators int) *Agent {
	queue := make(chan *model.ExpressionPart)
	miniCalcs := make([]*Calculator, countCalculators)

	for i := 0; i < countCalculators; i++ {
		miniCalcs[i] = NewCalculator(i)
	}

	a := &Agent{
		Calculators: miniCalcs,
		Queue:       queue,
	}

	go a.distributeTasks()

	return a
}

func (a *Agent) AddTask(exp *model.ExpressionPart) {
	a.mu.Lock()
	a.Queue <- exp
	a.mu.Unlock()
}

func (a *Agent) distributeTasks() {
	i := 0
	countOfCalculators := len(a.Calculators)
	for task := range a.Queue {
		// Try to send the task to each worker in turn
		for {
			// Use a select statement with a default case to avoid blocking
			if a.Calculators[i].AddTask(task) {
				log.Printf("task send to Calculator[%b]", a.Calculators[i].id)
				break
			}

			i++

			if i < 0 || i >= countOfCalculators {
				i = 0
			}
		}
	}
}

func (a *Agent) SolveExpression(exp *model.Expression) {
	stack := make([]*shuntingYard.RPNToken, 0)

	for _, token := range exp.PostfixExpression {
		if token.Type == model.Operand {
			stack = append(stack, token)
			continue
		}

		if len(stack) < 2 {
			continue
		}
		num2 := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		num1 := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		duration := GetOperationDuration(token.Value.(string))

		exprPart := model.NewExpressionPart(num1, num2, token, exp.IdExpression, duration)
		a.AddTask(exprPart)

		stack = append(stack, <-exprPart.Result)
		close(exprPart.Result)
	}

	setResultsToExpression(exp, stack[0].Value.(float64))
}

func setResultsToExpression(exp *model.Expression, result float64) {
	exp.Result = result
	exp.Status = model.Solved
	timeCompleted := time.Now()
	exp.SolvedAt = timeCompleted
}

func GetOperationDuration(operation string) int {
	return config.Config.AgentResolveTime[operation]
}
