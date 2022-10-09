package raft

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/grafana/dskit/backoff"
	"github.com/lni/dragonboat/v4"
	"github.com/lni/dragonboat/v4/client"
	"github.com/lni/dragonboat/v4/config"
	"github.com/lni/dragonboat/v4/statemachine"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/Blinkuu/qms/internal/core/domain"
	"github.com/Blinkuu/qms/internal/core/ports"
	stor "github.com/Blinkuu/qms/internal/core/storage"
	"github.com/Blinkuu/qms/internal/core/storage/alloc/quota"
	"github.com/Blinkuu/qms/pkg/dto"
	"github.com/Blinkuu/qms/pkg/log"
	badgerlog "github.com/Blinkuu/qms/pkg/log/badger"
)

const (
	appliedEntryIndexKey string = "__applied_entry_index__"
)

type Storage struct {
	cfg        Config
	logger     log.Logger
	memberlist ports.MemberlistService
	nh         *dragonboat.NodeHost
	storages   map[uint64]*storage
	sessions   map[uint64]*client.Session

	shutdown     chan struct{}
	shutdownOnce sync.Once
}

func NewStorage(cfg Config, logger log.Logger, memberlist ports.MemberlistService) (*Storage, error) {
	if cfg.ReplicaIDOverride != "" {
		replicaID, err := strconv.ParseUint(trimBeforeSubstr(cfg.ReplicaIDOverride, "-"), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse replica_id_override: %w", err)
		}

		cfg.ReplicaID = replicaID + 1
	}

	if cfg.HostFromHostname {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("failed to get hostname: %w", err)
		}

		cfg.Host = fmt.Sprintf("%s.%s.default.svc.cluster.local", hostname, trimAfterSubstr(hostname, "-"))
	}

	var initialMembers map[uint64]string
	joined := false
	raftAddr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	logger.Info("trying to join raft cluster")
	alreadyMember, err := join(context.Background(), logger, memberlist, cfg.ReplicaID, raftAddr)
	if err == nil {
		logger.Info("successfully joined raft cluster")
		if !alreadyMember {
			joined = true
		}
		goto createNodeHost
	}

	if cfg.ReplicaID != 1 {
		if raftAndDataDirsExist(cfg.Dir, cfg.ReplicaID) {
			goto createNodeHost
		}

		return nil, fmt.Errorf("failed to join raft cluster: %w", err)
	}

	logger.Info("bootstrapping new raft cluster", "replicaID", cfg.ReplicaID)
	initialMembers = map[uint64]string{
		cfg.ReplicaID: raftAddr,
	}

createNodeHost:
	// raftDir: dir/raft_node_nodeId
	// dataDir: dir/data_node_nodeId
	logger.Info("creating raft data and dirs", "dir", cfg.Dir)
	raftDir, dataDir, err := createRaftAndDataDirs(cfg.Dir, cfg.ReplicaID)
	if err != nil {
		return nil, fmt.Errorf("failed to create raft and data dirs: %w", err)
	}

	logger.Infof("raftAddr=%s", raftAddr)
	nodeHostCfg := newNodeHostConfig(cfg.DeploymentID, raftDir, raftAddr)
	nh, err := dragonboat.NewNodeHost(nodeHostCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create new node host: %w", err)
	}

	var (
		storages = make(map[uint64]*storage)
		sessions = make(map[uint64]*client.Session)
	)

	for shardID := uint64(1); shardID <= cfg.Shards; shardID++ {
		shardDir := filepath.Join(dataDir, strconv.Itoa(int(shardID))) //clusterDataPath: base/data_node_nodeId/shardID

		st, err := newStorage(shardDir, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create new local storage: %w", err)
		}

		stateMachine := newStateMachine(st)
		if err != nil {
			return nil, fmt.Errorf("failed to create new state machine for shardID=%d: %w", shardID, err)
		}

		raftCfg := newRaftConfig(cfg.ReplicaID, shardID)
		logger.Infof("initialMembers=%+v", initialMembers)
		err = nh.StartOnDiskReplica(initialMembers, joined, func(_ uint64, _ uint64) statemachine.IOnDiskStateMachine { return stateMachine }, raftCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to start on disk replica for shardID=%d: %w", shardID, err)
		}

		storages[shardID] = st
		sessions[shardID] = nh.GetNoOPSession(shardID)
	}

	return &Storage{
		cfg:          cfg,
		logger:       logger,
		memberlist:   memberlist,
		nh:           nh,
		storages:     storages,
		sessions:     sessions,
		shutdown:     make(chan struct{}),
		shutdownOnce: sync.Once{},
	}, nil
}

func (s *Storage) Run(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdown:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if !s.isLeader() {
				continue
			}
		}
	}
}

func (s *Storage) Alloc(ctx context.Context, namespace, resource string, tokens, version int64) (int64, int64, bool, error) {
	id := strings.Join([]string{namespace, resource}, "_")
	shardID := shardIDFromString(id, s.cfg.Shards)

	allocCmd := NewAllocCommand(namespace, resource, tokens, version)
	result, err := allocCmd.RaftInvoke(ctx, s.nh, shardID, s.sessions[shardID])
	if err != nil {
		return 0, 0, false, fmt.Errorf("failed to raft invoke: %w", err)
	}

	typedResult := result.(AllocCommandResult)
	if typedResult.Err != "" {
		if stor.IsInvalidVersionError(typedResult.Err) {
			return 0, 0, false, stor.ErrInvalidVersion
		}

		return 0, 0, false, errors.New(typedResult.Err)
	}

	return typedResult.RemainingTokens, typedResult.CurrentVersion, typedResult.OK, nil
}

func (s *Storage) Free(ctx context.Context, namespace, resource string, tokens, version int64) (int64, int64, bool, error) {
	id := strings.Join([]string{namespace, resource}, "_")
	shardID := shardIDFromString(id, s.cfg.Shards)

	freeCmd := NewFreeCommand(namespace, resource, tokens, version)
	result, err := freeCmd.RaftInvoke(ctx, s.nh, shardID, s.sessions[shardID])
	if err != nil {
		return 0, 0, false, fmt.Errorf("failed to raft invoke: %w", err)
	}

	typedResult := result.(FreeCommandResult)
	if typedResult.Err != "" {
		if stor.IsInvalidVersionError(typedResult.Err) {
			return 0, 0, false, stor.ErrInvalidVersion
		}

		return 0, 0, false, errors.New(typedResult.Err)
	}

	return typedResult.RemainingTokens, typedResult.CurrentVersion, typedResult.OK, nil
}

func (s *Storage) RegisterQuota(ctx context.Context, namespace, resource string, cfg quota.Config) error {
	id := strings.Join([]string{namespace, resource}, "_")
	shardID := shardIDFromString(id, s.cfg.Shards)

	registerQuotaCmd := NewRegisterQuotaCommand(namespace, resource, cfg)
	_, err := registerQuotaCmd.RaftInvoke(ctx, s.nh, shardID, s.sessions[shardID])
	if err != nil {
		return fmt.Errorf("failed to raft invoke: %w", err)
	}

	return nil
}

func (s *Storage) AddRaftReplica(ctx context.Context, replicaID uint64, raftAddr string) (bool, error) {
	info := s.nh.GetNodeHostInfo(dragonboat.NodeHostInfoOption{SkipLogInfo: true})

	if len(info.ShardInfoList) > 0 {
		nodes := info.ShardInfoList[0].Nodes
		for _, addr := range nodes {
			if raftAddr == addr {
				s.logger.Info("replica is already part of the raft cluster", "replicaID", replicaID, "raftAddr", raftAddr)
				return true, nil
			}
		}
	}

	for shardID := uint64(1); shardID <= s.cfg.Shards; shardID++ {
		ms, err := s.nh.SyncGetShardMembership(ctx, shardID)
		if err != nil {
			s.logger.Info("failed to get shard membership", "raftAddr", raftAddr, "replicaID", replicaID, "shardID", shardID, "err", err)
			return false, fmt.Errorf("failed to get shard membership for replicaID=%d and shardID=%d: %w", replicaID, shardID, err)
		}

		err = s.nh.SyncRequestAddReplica(ctx, shardID, replicaID, raftAddr, ms.ConfigChangeID)
		if err != nil {
			s.logger.Info("failed to request add replica", "raftAddr", raftAddr, "replicaID", replicaID, "shardID", shardID, "err", err)
			return false, fmt.Errorf("failed to request add replica for replicaID=%d and shardID=%d: %w", replicaID, shardID, err)
		}
	}

	return false, nil
}

func (s *Storage) RemoveRaftReplica(ctx context.Context, replicaID uint64) error {
	// TODO: Handle already non-existent replica

	for shardID := uint64(1); shardID <= s.cfg.Shards; shardID++ {
		ms, err := s.nh.SyncGetShardMembership(ctx, shardID)
		if err != nil {
			return fmt.Errorf("failed to get shard membership for replicaID=%d and shardID=%d: %w", replicaID, shardID, err)
		}

		err = s.nh.SyncRequestDeleteReplica(ctx, shardID, replicaID, ms.ConfigChangeID)
		if err != nil {
			return fmt.Errorf("failed to request delete replica for replicaID=%d and shardID=%d: %w", replicaID, shardID, err)
		}
	}

	return nil
}

func (s *Storage) AwaitHealthy(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if s.AllShardsHealthy() {
			return nil
		}
	}
}

func (s *Storage) AllShardsHealthy() bool {
	for shardID := uint64(1); shardID <= s.cfg.Shards; shardID++ {
		if !s.ShardHealthy(shardID) {
			return false
		}
	}

	return true
}

func (s *Storage) ShardHealthy(shardID uint64) bool {
	_, _, valid, err := s.nh.GetLeaderID(shardID)
	if err != nil {
		return false
	}

	return valid
}

func (s *Storage) Shutdown(_ context.Context) error {
	var err error
	s.shutdownOnce.Do(func() {
		close(s.shutdown)
		err = s.nh.StopReplica(s.cfg.ShardID, s.cfg.ReplicaID)
		s.nh.Close()
	})

	return err
}

func (s *Storage) isLeader() bool {
	leaderID, _, valid, err := s.nh.GetLeaderID(s.cfg.ShardID)
	if err != nil {
		return false
	}

	if !valid {
		return false
	}

	if leaderID != s.cfg.ReplicaID {
		return false
	}

	return true
}

type item struct {
	Allocated int64
	Capacity  int64
	Version   int64
}

type storage struct {
	dir string
	db  *badger.DB
}

func newStorage(dir string, logger log.Logger) (*storage, error) {
	opts := badger.DefaultOptions(dir)
	opts.Logger = badgerlog.NewLogger(logger)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger: %w", err)
	}

	return &storage{
		dir: dir,
		db:  db,
	}, nil
}

func (s *storage) alloc(namespace, resource string, tokens, version int64, entryIdx uint64) (int64, int64, bool, error) {
	if s.db.IsClosed() {
		return 0, 0, false, errors.New("badger db is closed")
	}

	id := strings.Join([]string{namespace, resource}, "_")

	txn := s.db.NewTransaction(true)
	defer txn.Discard()

	it, err := get[item](txn, id)
	if err != nil {
		return 0, 0, false, fmt.Errorf("failed to get: %w", err)
	}

	if version != 0 && it.Version != version {
		return 0, 0, false, stor.ErrInvalidVersion
	}

	newAllocated := it.Allocated + tokens
	if newAllocated > it.Capacity {
		return it.Capacity - it.Allocated, it.Version, false, nil
	}

	it.Allocated = newAllocated
	it.Version += 1
	if err := set[item](txn, id, it); err != nil {
		return 0, 0, false, fmt.Errorf("failed to set item: %w", err)
	}

	if err := set[uint64](txn, appliedEntryIndexKey, entryIdx); err != nil {
		return 0, 0, false, fmt.Errorf("failed to set entry index: %w", err)
	}

	if err := txn.Commit(); err != nil {
		return 0, 0, false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return it.Capacity - it.Allocated, it.Version, true, nil
}

func (s *storage) free(namespace, resource string, tokens, version int64, entryIdx uint64) (int64, int64, bool, error) {
	if s.db.IsClosed() {
		return 0, 0, false, errors.New("badger db is closed")
	}

	id := strings.Join([]string{namespace, resource}, "_")

	txn := s.db.NewTransaction(true)
	defer txn.Discard()

	it, err := get[item](txn, id)
	if err != nil {
		return 0, 0, false, fmt.Errorf("failed to get: %w", err)
	}

	if version != 0 && it.Version != version {
		return 0, 0, false, stor.ErrInvalidVersion
	}

	newAllocated := it.Allocated - tokens
	if newAllocated < 0 {
		return it.Capacity - it.Allocated, it.Version, false, nil
	}

	it.Allocated = newAllocated
	it.Version += 1
	if err := set[item](txn, id, it); err != nil {
		return 0, 0, false, fmt.Errorf("failed to set item: %w", err)
	}

	if err := set[uint64](txn, appliedEntryIndexKey, entryIdx); err != nil {
		return 0, 0, false, fmt.Errorf("failed to set entry index: %w", err)
	}

	if err := txn.Commit(); err != nil {
		return 0, 0, false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return it.Capacity - it.Allocated, it.Version, true, nil
}

func (s *storage) registerQuota(namespace, resource string, cfg quota.Config, entryIdx uint64) error {
	if s.db.IsClosed() {
		return errors.New("badger db is closed")
	}

	id := strings.Join([]string{namespace, resource}, "_")

	txn := s.db.NewTransaction(true)
	defer txn.Discard()

	_, err := get[item](txn, id)
	if !errors.Is(err, badger.ErrKeyNotFound) {
		return nil
	}

	if err := set[item](txn, id, item{Allocated: 0, Capacity: cfg.Capacity, Version: 1}); err != nil {
		return fmt.Errorf("failed to set item :%w", err)
	}

	if err := set[uint64](txn, appliedEntryIndexKey, entryIdx); err != nil {
		return fmt.Errorf("failed to set entry index: %w", err)
	}

	if err := txn.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *storage) lastAppliedIndex() (uint64, error) {
	if s.db.IsClosed() {
		return 0, errors.New("badger db is closed")
	}

	txn := s.db.NewTransaction(false)
	defer txn.Discard()

	val, err := get[uint64](txn, appliedEntryIndexKey)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return 0, nil
		}

		return 0, fmt.Errorf("failed to get applied entry index key: %w", err)
	}

	return val, nil
}

func (s *storage) snapshot(w io.Writer, stopChan <-chan struct{}) error {
	select {
	case <-stopChan:
		return nil
	default:
	}

	_, err := s.db.Backup(w, 0)
	if err != nil {
		return fmt.Errorf("failed to backup badger db: %w", err)
	}

	return nil
}

func (s *storage) loadSnapshot(r io.Reader, stopChan <-chan struct{}) error {
	select {
	case <-stopChan:
		return nil
	default:
	}

	return s.db.Load(r, 256)
}

func (s *storage) close() error {
	return s.db.Close()
}

func get[T any](txn *badger.Txn, key string) (T, error) {
	item, err := txn.Get([]byte(key))
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to get key: %w", err)
	}

	var result T
	err = item.Value(func(val []byte) error {
		err := binary.Read(bytes.NewReader(val), binary.BigEndian, &result)
		if err != nil {
			return fmt.Errorf("failed to read bytes: %w", err)
		}

		return nil
	})
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to read item value: %w", err)
	}

	return result, nil
}

func set[T any](txn *badger.Txn, key string, value T) error {
	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.BigEndian, value); err != nil {
		return fmt.Errorf("failed to write to buffer: %w", err)
	}

	if err := txn.Set([]byte(key), buf.Bytes()); err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}

	return nil
}

func trimBeforeSubstr(s string, substr string) string {
	if idx := strings.LastIndex(s, substr); idx != -1 {
		return s[idx+1:]
	}
	return s
}

func trimAfterSubstr(s string, substr string) string {
	if idx := strings.LastIndex(s, substr); idx != -1 {
		return s[:idx]
	}
	return s
}

func shardIDFromString(key string, shards uint64) uint64 {
	return uint64(crc32.ChecksumIEEE([]byte(key))%uint32(shards)) + 1
}

func raftAndDataDirsExist(path string, replicaID uint64) bool {
	raftPath := filepath.Join(path, fmt.Sprintf("raft_node_%d", replicaID))
	dataPath := filepath.Join(path, fmt.Sprintf("data_node_%d", replicaID))

	if _, err := os.Stat(raftPath); os.IsExist(err) {
		return false
	}

	if _, err := os.Stat(dataPath); os.IsExist(err) {
		return false
	}

	return true
}

func createRaftAndDataDirs(path string, replicaID uint64) (string, string, error) {
	raftPath := filepath.Join(path, fmt.Sprintf("raft_node_%d", replicaID))
	dataPath := filepath.Join(path, fmt.Sprintf("data_node_%d", replicaID))
	if err := os.MkdirAll(raftPath, os.ModePerm); err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(dataPath, os.ModePerm); err != nil {
		return "", "", err
	}
	return raftPath, dataPath, nil
}

func newNodeHostConfig(deploymentId uint64, raftDir, raftAddr string) config.NodeHostConfig {
	return config.NodeHostConfig{
		DeploymentID:     deploymentId,
		WALDir:           raftDir,
		NodeHostDir:      raftDir,
		RTTMillisecond:   200,
		RaftAddress:      raftAddr,
		MutualTLS:        false,
		MaxSendQueueSize: 128 * 1024 * 1024,
		EnableMetrics:    false,
	}
}

func newRaftConfig(replicaID, shardID uint64) config.Config {
	return config.Config{
		ReplicaID:               replicaID,
		ShardID:                 shardID,
		CheckQuorum:             false,
		ElectionRTT:             20,
		HeartbeatRTT:            2,
		SnapshotEntries:         25 * 10000 * 10,
		CompactionOverhead:      25 * 10000,
		OrderedConfigChange:     true,
		MaxInMemLogSize:         256 * 1024 * 1024,
		SnapshotCompressionType: config.NoCompression,
		EntryCompressionType:    config.Snappy,
		DisableAutoCompactions:  false,
		IsNonVoting:             false,
		IsWitness:               false,
		Quiesce:                 false,
	}
}

func join(ctx context.Context, logger log.Logger, memberlist ports.MemberlistService, replicaID uint64, raftAddr string) (bool, error) {
	members, err := memberlist.Members(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get members: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return false, fmt.Errorf("failed to get hostname: %w", err)
	}

	var filteredMembers []domain.Instance
	for _, member := range members {
		if member.Hostname != hostname {
			filteredMembers = append(filteredMembers, member)
		}
	}

	if len(filteredMembers) < 1 {
		return false, errors.New("no members to join")
	}

	cli := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   1 * time.Second,
	}

	cfg := backoff.Config{
		MinBackoff: 0,
		MaxBackoff: 2 * time.Second,
		MaxRetries: 10,
	}

	bo := backoff.New(ctx, cfg)
	for bo.Ongoing() {
		for _, member := range filteredMembers {
			addr := net.JoinHostPort(member.Host, strconv.Itoa(member.HTTPPort))
			url := fmt.Sprintf("http://%s/api/v1/internal/raft/join", addr)
			body := dto.JoinRequestBody{ReplicaID: replicaID, RaftAddr: raftAddr}
			var bodyBuffer bytes.Buffer
			if err := json.NewEncoder(&bodyBuffer).Encode(body); err != nil {
				return false, fmt.Errorf("failed to encode alloc request body: %w", err)
			}

			r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &bodyBuffer)
			if err != nil {
				return false, fmt.Errorf("failed to create new request with context: %w", err)
			}

			logger.Info("trying to join cluster", "addr", addr)
			res, err := cli.Do(r)
			if err != nil {
				logger.Warn("failed to do request", "err", err)
				continue
			}
			defer func() {
				if err := res.Body.Close(); err != nil {
					logger.Warn("failed to close response body: %w", err)
				}
			}()

			if res.StatusCode != http.StatusOK {
				logger.Warn("invalid http status code", "statusCode", res.StatusCode)
				continue
			}

			resBody := dto.ResponseBody[dto.JoinResponseBody]{}
			if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
				logger.Warn("failed to decode response body", "err", err)
				continue
			}

			if resBody.Status != dto.StatusOK {
				logger.Warn("invalid status code", "statusCode", resBody.Status)
				continue
			}

			return resBody.Result.AlreadyMember, nil
		}

		bo.Wait()
	}

	return false, errors.New("all attempts failed")
}
