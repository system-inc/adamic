// Cycletools records lowering decisions or profiles one lowering operation.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func main() {
	profile := flag.String("profile", "", "CPU profile output for one input")
	flag.Parse()
	if flag.NArg() != 1 {
		panic("usage: cycletools [-profile cpu.pprof] input.a")
	}
	started := time.Now()
	program, err := load.Load([]string{flag.Arg(0)})
	loaded := time.Now()
	lowered := loaded
	if err == nil {
		if *profile != "" {
			output, failure := os.Create(*profile)
			if failure != nil {
				panic(failure)
			}
			defer output.Close()
			if failure := pprof.StartCPUProfile(output); failure != nil {
				panic(failure)
			}
		}
		_, err = lower.Lower(context.Background(), program)
		lowered = time.Now()
		if *profile != "" {
			pprof.StopCPUProfile()
		}
	}
	decision := "accepted"
	if err != nil {
		decision = fmt.Sprintf("%T: %s", err, err)
	}
	result := struct {
		Path         string  `json:"path"`
		Decision     string  `json:"decision"`
		LoadSeconds  float64 `json:"load_seconds"`
		LowerSeconds float64 `json:"lower_seconds"`
	}{flag.Arg(0), decision, loaded.Sub(started).Seconds(), lowered.Sub(loaded).Seconds()}
	if failure := json.NewEncoder(os.Stdout).Encode(result); failure != nil {
		panic(failure)
	}
}
