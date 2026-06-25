package id

import (
	"sync"
	"time"
)

// Snowflake 雪花算法 ID 生成器。
// 结构：1位符号位 + 41位时间戳 + 10位工作节点 + 12位序列号
// 工作节点 ID 有效范围：0-1023
type Snowflake struct {
	mutex   sync.Mutex
	stamp   int64
	worker  int64
	seq     int64
	lastSeq int64 // 上次时钟回拨时的序列号，用于检测
}

// NewSnowflake 创建雪花算法 ID 生成器。
// workerID 为工作节点 ID，有效范围 0-1023。
func NewSnowflake(workerID int64) *Snowflake {
	if workerID < 0 || workerID > 1023 {
		workerID = 0
	}
	return &Snowflake{
		worker: workerID,
	}
}

// NextID 生成下一个唯一 ID。
// 若检测到时钟回拨，会等待直到时钟追上上次生成时间。
func (s *Snowflake) NextID() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now().UnixNano() / 1000000

	// 时钟回拨检测与处理
	if now < s.stamp {
		// 等待时钟追上
		for now < s.stamp {
			now = time.Now().UnixNano() / 1000000
		}
	}

	if s.stamp == now {
		s.seq = (s.seq + 1) & 4095
		if s.seq == 0 {
			for now <= s.stamp {
				now = time.Now().UnixNano() / 1000000
			}
		}
	} else {
		s.seq = 0
	}

	s.stamp = now
	return (now << 22) | (s.worker << 12) | s.seq
}
