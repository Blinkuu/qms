package raft

import (
	"context"
	"fmt"

	"github.com/lni/dragonboat/v4"
	"github.com/lni/dragonboat/v4/client"
	"github.com/lni/dragonboat/v4/statemachine"
)

type FreeCommand struct {
	Namespace string
	Resource  string
	Tokens    int64
	SMResult  statemachine.Result
}

type FreeCommandResult struct {
	RemainingTokens int64
	OK              bool
}

func NewFreeCommand(namespace, resource string, tokens int64) *FreeCommand {
	return &FreeCommand{
		Namespace: namespace,
		Resource:  resource,
		Tokens:    tokens,
		SMResult:  statemachine.Result{},
	}
}

func (c *FreeCommand) Type() CommandType {
	return Free
}

func (c *FreeCommand) RaftInvoke(ctx context.Context, nh *dragonboat.NodeHost, _ uint64, session *client.Session) (any, error) {
	result, err := syncWrite[FreeCommandResult](ctx, nh, session, c)
	if err != nil {
		return nil, fmt.Errorf("failed to sync write: %w", err)
	}

	return result, nil
}

func (c *FreeCommand) LocalInvoke(storage *storage, entryIdx uint64) error {
	remainingTokens, ok, err := storage.free(c.Namespace, c.Resource, c.Tokens, entryIdx)
	if err != nil {
		return fmt.Errorf("failed to free: %w", err)
	}

	data := EncodeCommandResult(FreeCommandResult{RemainingTokens: remainingTokens, OK: ok})
	c.SMResult = statemachine.Result{
		Value: 1,
		Data:  data,
	}

	return nil
}

func (c *FreeCommand) Result() statemachine.Result {
	return c.SMResult
}
