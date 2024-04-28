package grpcclientapp

import (
	"context"
	"log/slog"
	"time"

	"github.com/a-romash/grpc-calculator/agent/internal/clients/orchestrator/grpc"
	"github.com/a-romash/grpc-calculator/agent/internal/domain/models"
	"github.com/a-romash/grpc-calculator/agent/internal/service/agent"
)

type GRPCCApp struct {
	log         *slog.Logger
	orch_client *grpc.Client
	agent       *agent.Agent
	id          int
}

func New(
	log *slog.Logger,
	addr string, // address of orchestrator service
	retriesCount int,
	countCalcs int,
	durations map[string]time.Duration,
) *GRPCCApp {
	orch_client, err := grpc.New(context.Background(), log, addr, 5*time.Minute, retriesCount)
	if err != nil {
		panic(err)
	}

	agent := agent.New(countCalcs, durations)

	var id int
	id, err = orch_client.RegisterNewAgent(context.Background())
	if err != nil {
		panic(err)
	}

	return &GRPCCApp{
		log:         log,
		orch_client: orch_client,
		agent:       agent,
		id:          id,
	}
}

func (a *GRPCCApp) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *GRPCCApp) Run() error {
	// const op = "gtpccapp.Run"

	go func() {
		for {
			idExpression, tokens, err := a.orch_client.GetExpressionToEvaluate(context.Background(), a.id)
			if err != nil {
				a.log.Error(err.Error())
			}
			if len(tokens) > 0 {
				a.log.Info("STARTED EVALUATING")
				expression := models.Create("", tokens, idExpression)

				a.agent.SolveExpression(&expression)

				a.orch_client.GiveResultOfExpression(context.Background(), expression.IdExpression, float32(expression.Result), a.id)
			}
			time.Sleep(10 * time.Second)
		}
	}()

	return nil
}

func (a *GRPCCApp) Stop() error {
	return a.orch_client.RemoveAgent(context.Background(), a.id)
}
