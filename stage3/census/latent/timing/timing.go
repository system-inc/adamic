// Package timing records sequential latent unit costs separately from findings.
package timing

import (
	"encoding/json"
	"io"
	"syscall"
	"time"
)

type UnitTime struct {
	File        string  `json:"file"`
	Where       string  `json:"where"`
	WallSeconds float64 `json:"wall_seconds"`
	CPUSeconds  float64 `json:"cpu_seconds"`
}

type UnitTimer struct{ encoder *json.Encoder }

func NewUnitTimer(output io.Writer) *UnitTimer { return &UnitTimer{json.NewEncoder(output)} }

func processCPUSeconds() float64 {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		panic(err)
	}
	return float64(usage.Utime.Sec) + float64(usage.Utime.Usec)/1e6 + float64(usage.Stime.Sec) + float64(usage.Stime.Usec)/1e6
}

// Start includes registration, lowering and panic recovery, but excludes writing
// the time record. A nil timer leaves the deterministic census path untimed.
func (timer *UnitTimer) Start(file, where string) func() {
	if timer == nil {
		return func() {}
	}
	wall, cpu := time.Now(), processCPUSeconds()
	return func() {
		record := UnitTime{file, where, time.Since(wall).Seconds(), processCPUSeconds() - cpu}
		if err := timer.encoder.Encode(record); err != nil {
			panic(err)
		}
	}
}
