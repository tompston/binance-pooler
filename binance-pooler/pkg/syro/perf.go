package syro

import (
	"runtime"
	"time"
)

type MemStats struct {
	Time       time.Time `json:"time" bson:"time"`
	Alloc      uint64    `json:"alloc" bson:"alloc"`
	TotalAlloc uint64    `json:"total_alloc" bson:"total_alloc"`
	Sys        uint64    `json:"sys" bson:"sys"`
	NumGC      uint32    `json:"num_gc" bson:"num_gc"`
}

func NewMemStats() MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MemStats{
		Time:       time.Now().UTC(),
		Alloc:      m.Alloc / 1024,
		TotalAlloc: m.TotalAlloc / 1024,
		Sys:        m.Sys / 1024,
		NumGC:      m.NumGC,
	}
}

func (m MemStats) Store() error {
	return nil
}
