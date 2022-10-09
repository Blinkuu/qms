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
	Version   int64
	SMResult  statemachine.Result
}

type AllocCommandResult struct {
	RemainingTokens int64
	CurrentVersion  int64
	OK              bool
	Err             string
}

func NewAllocCommand(namespace, resource string, tokens, version int64) *AllocCommand {
	return &AllocCommand{
		Namespace: namespace,
		Resource:  resource,
		Tokens:    tokens,
		Version:   version,
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
	remainingTokens, currentVersion, ok, err := storage.alloc(c.Namespace, c.Resource, c.Tokens, c.Version, entryIdx)
	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	data := EncodeCommandResult(AllocCommandResult{RemainingTokens: remainingTokens, CurrentVersion: currentVersion, OK: ok, Err: errStr})
	c.SMResult = statemachine.Result{
		Value: 1,
		Data:  data,
	}

	return nil
}

func (c *AllocCommand) Result() statemachine.Result {
	return c.SMResult
}
