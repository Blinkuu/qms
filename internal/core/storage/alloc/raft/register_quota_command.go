package raft

import (
	"context"
	"fmt"

	"github.com/lni/dragonboat/v4"
	"github.com/lni/dragonboat/v4/client"
	"github.com/lni/dragonboat/v4/statemachine"

	"github.com/Blinkuu/qms/internal/core/storage/alloc/quota"
)

type RegisterQuotaCommand struct {
	Namespace string
	Resource  string
	Cfg       quota.Config
	SMResult  statemachine.Result
}

type RegisterQuotaCommandResult struct{}

func NewRegisterQuotaCommand(namespace, resource string, cfg quota.Config) *RegisterQuotaCommand {
	return &RegisterQuotaCommand{
		Namespace: namespace,
		Resource:  resource,
		Cfg:       cfg,
		SMResult:  statemachine.Result{},
	}
}

func (c *RegisterQuotaCommand) Type() CommandType {
	return RegisterQuota
}

func (c *RegisterQuotaCommand) RaftInvoke(ctx context.Context, nh *dragonboat.NodeHost, _ uint64, session *client.Session) (any, error) {
	result, err := syncWrite[RegisterQuotaCommandResult](ctx, nh, session, c)
	if err != nil {
		return nil, fmt.Errorf("failed to sync write: %w", err)
	}

	return result, nil
}

func (c *RegisterQuotaCommand) LocalInvoke(storage *storage, entryIdx uint64) error {
	err := storage.registerQuota(c.Namespace, c.Resource, c.Cfg, entryIdx)
	if err != nil {
		return fmt.Errorf("failed to register quota: %w", err)
	}

	data := EncodeCommandResult(RegisterQuotaCommandResult{})
	c.SMResult = statemachine.Result{
		Value: 1,
		Data:  data,
	}

	return nil
}

func (c *RegisterQuotaCommand) Result() statemachine.Result {
	return c.SMResult
}
