package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Each worker owns its artifacts. Only immutable compiler/runtime binaries are
// shared. The caller consumes futures in corpus order, retaining serial counts,
// the attempted-test limit, and the complete ordered audit log.
func (e *engine) parallelAttempts(files []string, limit int, classifyOnly bool) ([]chan result, error) {
	if e.jobs <= 1 || classifyOnly {
		return nil, nil
	}
	type task struct {
		index int
		test  classified
	}
	var tasks []task
	futures := make([]chan result, len(files))
	for i, file := range files {
		source, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		relative, err := filepath.Rel(filepath.Join(e.test262, "test"), file)
		if err != nil {
			return nil, err
		}
		test := classify(filepath.ToSlash(relative), string(source), e.adapt)
		if test.Skip != "" || (limit > 0 && len(tasks) >= limit) {
			continue
		}
		futures[i] = make(chan result, 1)
		tasks = append(tasks, task{i, test})
	}
	workers := make([]engine, e.jobs)
	for i := range workers {
		workers[i] = *e
		workers[i].work = filepath.Join(e.work, fmt.Sprintf("worker-%d", i))
		if err := os.MkdirAll(workers[i].work, 0o755); err != nil {
			return nil, err
		}
	}
	queue := make(chan task, len(tasks))
	for _, task := range tasks {
		queue <- task
	}
	close(queue)
	for i := range workers {
		worker := workers[i]
		go func() {
			for task := range queue {
				futures[task.index] <- worker.attempt(task.test)
			}
		}()
	}
	return futures, nil
}
