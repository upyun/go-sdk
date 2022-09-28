package upyun

import (
	"sync"
	"time"
)

type Recorder interface {
	Set(path string, breakpoint *BreakPointConfig)

	Get(path string) *BreakPointConfig

	Delete(path string)

	TimedClearance()
}

type MemoryRecorder struct {
	resumeRecorder sync.Map
}

func (recorder *MemoryRecorder) Get(path string) *BreakPointConfig {
	if value, ok := recorder.resumeRecorder.Load(path); ok {
		if breakPoint, ok := value.(*BreakPointConfig); ok {
			return breakPoint
		}
	}
	return nil
}

func (recorder *MemoryRecorder) Set(path string, breakpoint *BreakPointConfig) {
	recorder.resumeRecorder.Store(path, breakpoint)
}

func (recorder *MemoryRecorder) Delete(path string) {
	recorder.resumeRecorder.Delete(path)
}

func (recorder *MemoryRecorder) TimedClearance() {
	recorder.resumeRecorder.Range(func(key, value interface{}) bool {
		breakPoint, ok := value.(*BreakPointConfig)
		if !ok {
			return false
		}
		// breakPoint survival time was more than 24h
		if breakPoint.LastTime.Add(24 * time.Hour).Before(time.Now()) {
			recorder.resumeRecorder.Delete(key)
		}
		return true
	})
}
