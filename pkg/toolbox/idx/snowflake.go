package idx

import (
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"kratos-admin/pkg/toolbox/errorx"
)

const (
	epoch          int64 = 1735660800000 // 2025-01-01 00:00:00 UTC+8
	timestampBits  uint8 = 41
	dataCenterBits uint8 = 5
	nodeBits       uint8 = 5
	sequenceBits   uint8 = 12

	maxDataCenterID int64 = -1 ^ (-1 << dataCenterBits)
	maxNodeID       int64 = -1 ^ (-1 << nodeBits)
	sequenceMask    int64 = -1 ^ (-1 << sequenceBits)
)

var (
	defaultSnowflake, _ = NewSnowflake(0, 0)
)

type Snowflake struct {
	mu            sync.Mutex
	lastTimestamp int64
	DataCenterID  int64
	NodeID        int64
	sequence      int64
}

func NewSnowflake(dataCenterID, nodeID int64) (*Snowflake, error) {
	if dataCenterID == 0 {
		// 从环境变量中加载
		dataCenterIdStr := os.Getenv("SNOWFLAKE_DATA_CENTER_ID")
		if dataCenterIdStr != "" {
			dataCenterID, _ = strconv.ParseInt(dataCenterIdStr, 10, 64)
		}
	}
	if dataCenterID == 0 {
		// 随机生成
		dataCenterID = rand.Int63n(maxDataCenterID) // #nosec G404
	}
	if dataCenterID < 0 || dataCenterID > maxDataCenterID {
		return nil, errorx.New("invalid data center ID")
	}

	if nodeID == 0 {
		// 从环境变量中加载
		nodeIdStr := os.Getenv("SNOWFLAKE_NODE_ID")
		if nodeIdStr != "" {
			nodeID, _ = strconv.ParseInt(nodeIdStr, 10, 64)
		}
	}
	if nodeID == 0 {
		// 随机生成
		nodeID = rand.Int63n(maxNodeID) // #nosec G404
	}
	if nodeID < 0 || nodeID > maxNodeID {
		return nil, errorx.New("invalid node ID")
	}

	return &Snowflake{
		DataCenterID: dataCenterID,
		NodeID:       nodeID,
	}, nil
}

func (s *Snowflake) NextId() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()

	if now < s.lastTimestamp {
		// 时钟回拨，等待
		now = s.waitNextMillis(s.lastTimestamp)
		s.sequence = 0
	} else if now == s.lastTimestamp {
		// 同一毫秒，序列号递增
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			// 当前毫秒序列号用完，等待下一个毫秒
			now = s.waitNextMillis(now)
		}
	} else {
		s.sequence = 0
	}

	s.lastTimestamp = now

	id := (now-epoch)<<(dataCenterBits+nodeBits+sequenceBits) |
		(s.DataCenterID << (nodeBits + sequenceBits)) |
		(s.NodeID << sequenceBits) |
		s.sequence

	return id
}

func (s *Snowflake) waitNextMillis(last int64) int64 {
	now := time.Now().UnixMilli()
	for now <= last {
		time.Sleep(100 * time.Microsecond)
		now = time.Now().UnixMilli()
	}
	return now
}

func SnowId() int64 {
	return defaultSnowflake.NextId()
}
