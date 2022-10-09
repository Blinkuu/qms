package raft

import (
	"fmt"
	"io"

	"github.com/lni/dragonboat/v4/statemachine"
)

type stateMachine struct {
	storage *storage
}

func newStateMachine(storage *storage) *stateMachine {
	return &stateMachine{
		storage: storage,
	}
}

func (m *stateMachine) Open(stopChan <-chan struct{}) (uint64, error) {
	select {
	case <-stopChan:
		return 0, statemachine.ErrOpenStopped
	default:
		return m.storage.lastAppliedIndex()
	}
}

func (m *stateMachine) Update(entries []statemachine.Entry) ([]statemachine.Entry, error) {
	var result []statemachine.Entry
	for _, e := range entries {
		r, err := m.processEntry(e)
		if err != nil {
			return nil, fmt.Errorf("failed to process entry: %w", err)
		}

		result = append(result, r)
	}

	return result, nil
}

func (m *stateMachine) Lookup(_ interface{}) (interface{}, error) {
	panic("implement me!")
}

func (m *stateMachine) Sync() error {
	return nil
}

func (m *stateMachine) PrepareSnapshot() (interface{}, error) {
	return nil, nil
}

func (m *stateMachine) SaveSnapshot(_ interface{}, writer io.Writer, stopChan <-chan struct{}) error {
	// TODO: Ensure all SaveSnapshot properties are satisfied
	return m.storage.snapshot(writer, stopChan)
}

func (m *stateMachine) RecoverFromSnapshot(reader io.Reader, stopChan <-chan struct{}) error {
	return m.storage.loadSnapshot(reader, stopChan)
}

func (m *stateMachine) Close() error {
	return m.storage.close()
}

func (m *stateMachine) processEntry(e statemachine.Entry) (statemachine.Entry, error) {
	decodedCmd, err := DecodeCommand(e.Cmd)
	if err != nil {
		return statemachine.Entry{}, fmt.Errorf("failed to decode command: %w", err)
	}

	if err := decodedCmd.LocalInvoke(m.storage, e.Index); err != nil {
		return e, err
	}

	e.Result = decodedCmd.Result()

	return e, nil
}
