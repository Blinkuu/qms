package raft

import (
	"context"
	"fmt"

	"github.com/lni/dragonboat/v4"
	"github.com/lni/dragonboat/v4/client"
	"github.com/lni/dragonboat/v4/statemachine"
)

type ViewCommand struct {
	Namespace string
	Resource  string
	SMResult  statemachine.Result
}

type ViewCommandResult struct {
	Allocated int64
	Capacity  int64
	Version   int64
	Err       string
}

func NewViewCommand(namespace, resource string) *ViewCommand {
	return &ViewCommand{
		Namespace: namespace,
		Resource:  resource,
		SMResult:  statemachine.Result{},
	}
}

func (c *ViewCommand) Type() CommandType {
	return View
}

func (c *ViewCommand) RaftInvoke(ctx context.Context, nh *dragonboat.NodeHost, shardID uint64, _ *client.Session) (any, error) {
	result, err := syncRead[ViewCommandResult](ctx, nh, shardID, c)
	if err != nil {
		return nil, fmt.Errorf("failed to sync write: %w", err)
	}

	return result, nil
}

func (c *ViewCommand) LocalInvoke(storage *storage, _ uint64) error {
	allocated, capacity, version, err := storage.view(c.Namespace, c.Resource)
	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	data := EncodeCommandResult(ViewCommandResult{Allocated: allocated, Capacity: capacity, Version: version, Err: errStr})
	c.SMResult = statemachine.Result{
		Value: 1,
		Data:  data,
	}

	return nil
}

func (c *ViewCommand) Result() statemachine.Result {
	return c.SMResult
}
