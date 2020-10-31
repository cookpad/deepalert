package inspector

import (
	"context"
	"fmt"

	"github.com/deepalert/deepalert"
	"github.com/deepalert/deepalert/internal/errors"
	"github.com/google/uuid"
)

// StartTest emulates inspector.Start, but
func StartTest(args Arguments, attr deepalert.Attribute) (*deepalert.TaskResult, error) {
	if args.Handler == nil {
		return nil, fmt.Errorf("Handler is not set in emitter.Argument")
	}
	if args.Author == "" {
		return nil, fmt.Errorf("Author is not set in emitter.Argument")
	}

	reportID := uuid.New().String()
	ctx := context.WithValue(context.Background(), contextKey, &reportID)

	result, err := args.Handler(ctx, attr)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to run Handler")
	}

	return result, nil
}
