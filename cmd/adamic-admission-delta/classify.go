package main

import (
	"sync"
	"time"
)

// Each worker owns its results slot; aggregation stays in stable corpus order.
func classifyPrograms(tree, base, head string, programs []entry, workers int, limit time.Duration) {
	jobs := make(chan int)
	var group sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := range jobs {
				record := &programs[index]
				record.Base = execute(tree, limit, base, "admission-lower", record.Path)
				if base == head {
					record.Head = record.Base
				} else {
					record.Head = execute(tree, limit, head, "admission-lower", record.Path)
				}
				record.Base.Stdout, record.Head.Stdout = "", ""
				record.Class = classify(record.Base, record.Head)
			}
		}()
	}
	for index := range programs {
		jobs <- index
	}
	close(jobs)
	group.Wait()
}
