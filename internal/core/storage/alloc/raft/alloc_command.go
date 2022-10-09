package raft

import (
	"context"
	"fmt"

	"github.com/lni/dragonboat/v4"
	"github.com/lni/dragonboat/v4/client"
	"github.com/lni/dragonboat/v4/statemachine"
)

type AllocCommand struct {
	Namespace string
	Resource  string
	Tokens    int64
	SMResult  statemachine.Result
}

type AllocCommandResult struct {
	RemainingTokens int64
	OK              bool
}

func NewAllocCommand(namespace, resource string, tokens int64) *AllocCommand {
	return &AllocCommand{
		Namespace: namespace,
		Resource:  resource,
		Tokens:    tokens,
		SMResult:  statemachine.Result{},
	}
}

func (c *AllocCommand) Type() CommandType {
	return Alloc
}

func (c *AllocCommand) RaftInvoke(ctx context.Context, nh *dragonboat.NodeHost, _ uint64, session *client.Session) (any, error) {
	result, err := syncWrite[AllocCommandResult](ctx, nh, session, c)
	if err != nil {
		return nil, fmt.Errorf("failed to sync write: %w", err)
	}

	return result, nil
}

func (c *AllocCommand) LocalInvoke(storage *storage, entryIdx uint64) error {
	remainingTokens, ok, err := storage.alloc(c.Namespace, c.Resource, c.Tokens, entryIdx)
	if err != nil {
		return fmt.Errorf("failed to alloc: %w", err)
	}

	data := EncodeCommandResult(AllocCommandResult{RemainingTokens: remainingTokens, OK: ok})
	c.SMResult = statemachine.Result{
		Value: 1,
		Data:  data,
	}

	return nil
}

func (c *AllocCommand) Result() statemachine.Result {
	return c.SMResult
}
