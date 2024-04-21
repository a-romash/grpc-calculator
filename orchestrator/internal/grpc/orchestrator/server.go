package orchgrpc

import (
	"context"
	"strconv"

	shuntingYard "github.com/a-romash/go-shunting-yard"
	"github.com/a-romash/grpc-calculator/protos/gen/go/orchestrator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type serverAPI struct {
	orchestrator.UnimplementedOrchestratorServer
	orch Orchestrator
}

type Orchestrator interface {
	Heartbeat(
		ctx context.Context,
		is_alive bool,
		id_agent int,
	) error
	GetExpressionToEvaluate(
		ctx context.Context,
		id_agent int,
	) (string, []shuntingYard.RPNToken, error)
	SaveResultOfExpression(
		ctx context.Context,
		id_expression string,
		result float32,
	) error
}

func Register(gRPCServer *grpc.Server, orch Orchestrator) {
	orchestrator.RegisterOrchestratorServer(gRPCServer, &serverAPI{orch: orch})
}

// Implementations of gRPC handlers
func (s *serverAPI) Heartbeat(
	ctx context.Context,
	in *orchestrator.IsAlive,
) (*emptypb.Empty, error) {
	if in.IdAgent == 0 {
		return nil, status.Error(codes.InvalidArgument, "id_agent is required")
	}
	if err := s.orch.Heartbeat(ctx, in.IsAlive, int(in.IdAgent)); err != nil {
		return nil, status.Error(codes.Internal, "some problems with heartbeat")
	}

	return &emptypb.Empty{}, nil
}

// Converting shuntingYard.RPNToken -> orchestrator.RPNToken (proto)
func tokensToPrototokens(tokens []shuntingYard.RPNToken) []*orchestrator.RPNToken {
	var prototokens []*orchestrator.RPNToken
	for _, el := range tokens {
		var token *orchestrator.RPNToken
		switch el.Value.(type) {
		case string:
			token = &orchestrator.RPNToken{
				Type:  int32(el.Type),
				Value: el.Value.(string),
			}
		case int:
			token = &orchestrator.RPNToken{
				Type:  int32(el.Type),
				Value: strconv.Itoa(el.Value.(int)),
			}
		}
		prototokens = append(prototokens, token)
	}

	return prototokens
}

func (s *serverAPI) GetExpressionToEvaluate(
	ctx context.Context,
	in *orchestrator.IdAgent,
) (*orchestrator.Expression, error) {
	if in.IdAgent == 0 {
		return nil, status.Error(codes.InvalidArgument, "id_agent is required")
	}

	id_expression, tokens, err := s.orch.GetExpressionToEvaluate(ctx, int(in.IdAgent))
	if err != nil {
		return nil, status.Error(codes.Internal, "some problems with getting expressions")
	}

	return &orchestrator.Expression{
		IdExpression:      id_expression,
		PostfixExpression: tokensToPrototokens(tokens),
	}, nil
}

func (s *serverAPI) GiveResultOfExpression(
	ctx context.Context,
	in *orchestrator.ResultOfExpression,
) (*emptypb.Empty, error) {
	if in.GetIdExpression() == "" {
		return nil, status.Error(codes.InvalidArgument, "id_expression is required")
	}

	if in.GetResult() == 0 {
		return nil, status.Error(codes.InvalidArgument, "result is required")
	}

	err := s.orch.SaveResultOfExpression(ctx, in.IdExpression, in.Result)
	if err != nil {
		return nil, status.Error(codes.Internal, "some problem with saving")
	}

	return &emptypb.Empty{}, nil
}
