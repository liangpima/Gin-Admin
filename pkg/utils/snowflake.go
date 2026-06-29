package utils

import (
	"sync"
	"time"
)

const (
	epoch          = 1288834974657
	machineBits    = 10
	sequenceBits   = 12
	maxMachineID   = (1 << machineBits) - 1
	maxSequence    = (1 << sequenceBits) - 1
	machineIDShift = sequenceBits
	timeShift      = machineBits + sequenceBits
)

type Snowflake struct {
	mu        sync.Mutex
	machineID int64
	sequence  int64
	lastTime  int64
}

var sf *Snowflake

func InitSnowflake(machineID int64) {
	if machineID < 0 || machineID > maxMachineID {
		machineID = 1
	}
	sf = &Snowflake{machineID: machineID}
}

func (s *Snowflake) NextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()

	if now < s.lastTime {
		now = s.lastTime
	}

	if now == s.lastTime {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			for now <= s.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastTime = now

	return ((now - epoch) << timeShift) | (s.machineID << machineIDShift) | s.sequence
}

func GenerateID() int64 {
	if sf == nil {
		InitSnowflake(1)
	}
	return sf.NextID()
}
