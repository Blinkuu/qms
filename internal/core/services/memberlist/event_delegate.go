package memberlist

import (
	"github.com/hashicorp/memberlist"

	"github.com/Blinkuu/qms/internal/core/domain/cloud"
)

// EventDelegate is a simpler delegate that is used only to receive
// notifications about members joining and leaving. The methods in this
// delegate may be called by multiple goroutines, but never concurrently.
// This allows you to reason about ordering.
type EventDelegate interface {
	// NotifyJoin is invoked when an instance is detected to have joined.
	NotifyJoin(*cloud.Instance)

	// NotifyLeave is invoked when an instance is detected to have left.
	NotifyLeave(*cloud.Instance)

	// NotifyUpdate is invoked when an instance is detected to have
	// updated, usually involving the metadata.
	NotifyUpdate(*cloud.Instance)
}

type eventDelegateAdapter struct {
	EventDelegate
}

func (a eventDelegateAdapter) NotifyJoin(node *memberlist.Node) {
	instance, err := nodeToInstance(node)
	if err != nil {
		panic(err)
	}

	a.EventDelegate.NotifyJoin(instance)
}

func (a eventDelegateAdapter) NotifyLeave(node *memberlist.Node) {
	instance, err := nodeToInstance(node)
	if err != nil {
		panic(err)
	}

	a.EventDelegate.NotifyLeave(instance)

}

func (a eventDelegateAdapter) NotifyUpdate(node *memberlist.Node) {
	instance, err := nodeToInstance(node)
	if err != nil {
		panic(err)
	}

	a.EventDelegate.NotifyUpdate(instance)

}
