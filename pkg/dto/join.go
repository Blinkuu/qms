package dto

type JoinRequestBody struct {
	ReplicaID uint64 `json:"replica_id"`
	RaftAddr  string `json:"raft_addr"`
}

type JoinResponseBody struct {
	AlreadyMember bool `json:"already_member"`
}
