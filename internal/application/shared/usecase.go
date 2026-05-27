package shared

import "context"

type UseCase[In any, Out any] interface {
	Execute(ctx context.Context, in In) (Out, error)
}
