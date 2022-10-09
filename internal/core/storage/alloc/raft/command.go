package raft

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/lni/dragonboat/v4"
	"github.com/lni/dragonboat/v4/client"
	"github.com/lni/dragonboat/v4/statemachine"
)

type CommandType byte

const (
	View          CommandType = 1
	Alloc         CommandType = 2
	Free          CommandType = 3
	RegisterQuota CommandType = 4
)

type Command interface {
	Type() CommandType
	RaftInvoke(ctx context.Context, nh *dragonboat.NodeHost, shardID uint64, session *client.Session) (result any, err error)
	LocalInvoke(storage *storage, entryIdx uint64) error
	Result() statemachine.Result
}

func EncodeCommand(cmd Command) []byte {
	var buf bytes.Buffer
	buf.WriteByte(byte(cmd.Type()))
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(cmd); err != nil {
		panic(fmt.Errorf("failed to encode command: %w", err))
	}

	return buf.Bytes()
}

func DecodeCommand(data []byte) (Command, error) {
	buf := bytes.NewBuffer(data[1:])
	decoder := gob.NewDecoder(buf)

	switch CommandType(data[0]) {
	case View:
		cmd := &ViewCommand{}
		if err := decoder.Decode(cmd); err != nil {
			panic(fmt.Errorf("failed to decode view command: %w", err))
		}

		return cmd, nil
	case Alloc:
		cmd := &AllocCommand{}
		if err := decoder.Decode(cmd); err != nil {
			panic(fmt.Errorf("failed to decode alloc command: %w", err))
		}

		return cmd, nil
	case Free:
		cmd := &FreeCommand{}
		if err := decoder.Decode(cmd); err != nil {
			panic(fmt.Errorf("failed to decode free command: %w", err))
		}

		return cmd, nil
	case RegisterQuota:
		cmd := &RegisterQuotaCommand{}
		if err := decoder.Decode(cmd); err != nil {
			panic(fmt.Errorf("failed to decode free command: %w", err))
		}

		return cmd, nil
	default:
		return nil, fmt.Errorf("unknown command: type=%b", CommandType(data[0]))
	}
}

func EncodeCommandResult(v any) []byte {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		panic(fmt.Errorf("failed to encode command SMResult: %w", err))
	}

	return buf.Bytes()
}

func DecodeCommandResult[T any](data []byte) T {
	var result T
	buf := bytes.NewBuffer(data)
	if err := gob.NewDecoder(buf).Decode(&result); err != nil {
		panic(fmt.Errorf("failed to encode command SMResult: %w", err))
	}

	return result
}

// TODO: Implement optimistic concurrency control (versioning)
func syncWrite[T any](ctx context.Context, nh *dragonboat.NodeHost, session *client.Session, cmd Command) (T, error) {
	result, err := nh.SyncPropose(ctx, session, EncodeCommand(cmd))
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to sync propose: %w", err)
	}

	return DecodeCommandResult[T](result.Data), nil
}

func syncRead[T any](ctx context.Context, nh *dragonboat.NodeHost, shardID uint64, cmd Command) (T, error) {
	result, err := nh.SyncRead(ctx, shardID, EncodeCommand(cmd))
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to sync read: %w", err)
	}

	return DecodeCommandResult[T](result.([]byte)), nil
}
