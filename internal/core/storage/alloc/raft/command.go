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
	Alloc         CommandType = 1
	Free          CommandType = 2
	RegisterQuota CommandType = 3
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

// TODO: Implement View() for alloc service
//func syncRead(ctx context.Context, nh *dragonboat.NodeHost, clusterId uint64, cmd Command) ([]byte, error) {
//	result, err := nh.SyncRead(ctx, clusterId, EncodeCommand(cmd))
//	if err != nil {
//		return nil, fmt.Errorf("failed to sync read: %w", err)
//	}
//
//	return result.([]byte), err
//}
