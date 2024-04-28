package agent

import (
	"log"
	"sync"
	"time"

	shuntingYard "github.com/a-romash/go-shunting-yard"
	"github.com/a-romash/grpc-calculator/agent/internal/domain/models"
)

type Agent struct {
	mu          sync.Mutex
	Calculators []*Calculator
	Queue       chan *models.ExpressionPart
	Durations   map[string]time.Duration
}

func New(countCalculators int, durations map[string]time.Duration) *Agent {
	queue := make(chan *models.ExpressionPart)
	miniCalcs := make([]*Calculator, countCalculators)

	for i := 0; i < countCalculators; i++ {
		miniCalcs[i] = NewCalculator(i)
	}

	a := &Agent{
		Calculators: miniCalcs,
		Queue:       queue,
		Durations:   durations,
	}

	go a.distributeTasks()

	return a
}

func (a *Agent) AddTask(exp *models.ExpressionPart) {
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

func (a *Agent) SolveExpression(exp *models.Expression) {
	stack := make([]*shuntingYard.RPNToken, 0)

	// fmt.Println(exp.PostfixExpression)

	for _, token := range exp.PostfixExpression {
		// fmt.Println(stack)
		if token.Type == models.Operand {
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

		duration := a.GetOperationDuration(token.Value.(string))

		exprPart := models.NewExpressionPart(num1, num2, token, exp.IdExpression, duration)
		a.AddTask(exprPart)

		stack = append(stack, <-exprPart.Result)
		close(exprPart.Result)
	}
	// fmt.Print("123")
	// result, _ := shuntingYard.Evaluate(exp.PostfixExpression)

	setResultsToExpression(exp, stack[0].Value.(float64))
}

func setResultsToExpression(exp *models.Expression, result float64) {
	exp.Result = result
	exp.Status = models.Solved
	timeCompleted := time.Now()
	exp.SolvedAt = timeCompleted
}

func (a *Agent) GetOperationDuration(operation string) time.Duration {
	return a.Durations[operation]
}
