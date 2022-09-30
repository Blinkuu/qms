package raft

//
//import (
//	"fmt"
//	"os"
//	"path/filepath"
//	"strconv"
//	"strings"
//
//	"github.com/lni/dragonboat/v4"
//	"github.com/lni/dragonboat/v4/client"
//	"github.com/lni/dragonboat/v4/config"
//)
//
//type Storage struct {
//	nh    *dragonboat.NodeHost
//	smMap map[uint64]*Store
//	csMap map[uint64]*client.Session
//}
//
//const clusterSize = 3
//
//func NewStorage(cfg Config) (*Storage, error) {
//	// join node initial members must be empty
//	var initialMembers map[uint64]string
//	if cfg.Join {
//		initialMembers = map[uint64]string{}
//	}
//
//	// raftDir: base/raft_node_nodeId
//	// dataDir: base/data_node_nodeId
//	raftDir, dataDir, err := initPath(baseDir, nodeId)
//	if err != nil {
//		return nil, err
//	}
//
//	listenAddr := fmt.Sprintf("0.0.0.0:%s", strings.Split(addr, ":")[1])
//
//	nhc := newNodeHostConfig(deploymentId, raftDir, addr, listenAddr)
//
//	nh, err := dragonboat.NewNodeHost(nhc)
//	if err != nil {
//		return nil, err
//	}
//
//	var (
//		csMap            = make(map[uint64]*client.Session)
//		smMap            = make(map[uint64]*Store)
//		clusterId uint64 = 0
//	)
//
//	for clusterId = 0; clusterId < clusterSize; clusterId++ {
//		rc := newRaftConfig(nodeId, clusterId)
//		//clusterDataPath: base/data_node_nodeId/clusterId
//		clusterDataPath := filepath.Join(dataDir, strconv.Itoa(int(clusterId)))
//		store, err := newStore(clusterDataPath, cfs)
//		if err != nil {
//			return nil, err
//		}
//
//		stateMachine, err := newRocksDBStateMachine(clusterId, uint64(nodeId), store)
//		if err != nil {
//			return nil, err
//		}
//
//		if err := nh.StartOnDiskCluster(initialMembers, join, func(_ uint64, _ uint64) sm.IOnDiskStateMachine {
//			return stateMachine
//		}, rc); err != nil {
//			return nil, err
//		}
//
//		csMap[clusterId] = nh.GetNoOPSession(clusterId)
//		smMap[clusterId] = store
//	}
//
//	return &Storage{
//		nh:    nh,
//		smMap: smMap,
//		csMap: csMap,
//	}, nil
//}
//
//func initPath(path string, nodeId uint64) (string, string, error) {
//	raftPath := filepath.Join(path, fmt.Sprintf("raft_node_%d", nodeId))
//	dataPath := filepath.Join(path, fmt.Sprintf("data_node_%d", nodeId))
//	if err := os.MkdirAll(raftPath, os.ModePerm); err != nil {
//		return "", "", err
//	}
//	if err := os.MkdirAll(dataPath, os.ModePerm); err != nil {
//		return "", "", err
//	}
//	return raftPath, dataPath, nil
//}
//
//func newNodeHostConfig(deploymentId uint64, raftDir string, raftAddr, listenAddr string) config.NodeHostConfig {
//	return config.NodeHostConfig{
//		DeploymentID:     deploymentId,
//		WALDir:           raftDir,
//		NodeHostDir:      raftDir,
//		RTTMillisecond:   200,
//		RaftAddress:      raftAddr,
//		ListenAddress:    listenAddr,
//		MutualTLS:        false,
//		MaxSendQueueSize: 128 * 1024 * 1024,
//		EnableMetrics:    false,
//	}
//}
//
//func newRaftConfig(replicaID, shardID uint64) config.Config {
//	return config.Config{
//		ReplicaID:               replicaID,
//		ShardID:                 shardID,
//		CheckQuorum:             false,
//		ElectionRTT:             20,
//		HeartbeatRTT:            2,
//		SnapshotEntries:         25 * 10000 * 10,
//		CompactionOverhead:      25 * 10000,
//		OrderedConfigChange:     true,
//		MaxInMemLogSize:         256 * 1024 * 1024,
//		SnapshotCompressionType: config.NoCompression,
//		EntryCompressionType:    config.Snappy,
//		DisableAutoCompactions:  false,
//		IsNonVoting:             false,
//		IsWitness:               false,
//		Quiesce:                 false,
//	}
//}
