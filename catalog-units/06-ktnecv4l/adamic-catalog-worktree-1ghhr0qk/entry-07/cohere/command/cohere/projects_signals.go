package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

// projectChildren are the project runs a mixed-repository run has going, held so a signal to this run
// reaches them.
//
// Each child is a whole cohere run that ends its own engine on a signal (see runEngineBinaryUntil), but
// only a signal it receives. A terminal's interrupt reaches the whole foreground group, while a caller's
// `kill <pid>`, or a timeout's terminate, reaches this process alone. That used to end this process and
// leave its children checking, a Swift child's compiler among them, with nobody waiting for their verdict
// (#nqb3mjv). So this run catches the interrupt, terminate and hangup it can, sends each to every child,
// waits for them to end, and then ends as the signal would have ended it. A second signal kills them
// outright. A kill this process cannot catch still leaves them running: nothing can forward that.
type projectChildren struct {
	mutex   sync.Mutex
	running map[*exec.Cmd]bool
	// first is the signal that stopped this run, or nil while none has.
	first   os.Signal
	count   int
	signals chan os.Signal
}

// forwardSignalsToChildren starts catching the signals a run forwards. Call stop when the children have
// all ended.
func forwardSignalsToChildren() *projectChildren {
	children := &projectChildren{running: map[*exec.Cmd]bool{}, signals: make(chan os.Signal, 1)}
	signal.Notify(children.signals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for received := range children.signals {
			children.forward(received)
		}
	}()
	return children
}

// forward sends a signal to every running child: the signal itself the first time, a kill after that.
func (children *projectChildren) forward(received os.Signal) {
	children.mutex.Lock()
	defer children.mutex.Unlock()
	if children.first == nil {
		children.first = received
	}
	children.count++
	for child := range children.running {
		if children.count > 1 {
			_ = child.Process.Kill()
			continue
		}
		// A platform that cannot deliver the signal, Windows for all but a kill, ends the child instead,
		// since the point is that it does not outlive this run.
		if err := child.Process.Signal(received); err != nil && !errors.Is(err, os.ErrProcessDone) {
			_ = child.Process.Kill()
		}
	}
}

// run starts a child and waits for it, holding it while it runs. A child is never started once this run has
// been stopped, so a signal between two starts cannot leave the second one running.
func (children *projectChildren) run(child *exec.Cmd) error {
	children.mutex.Lock()
	if children.first != nil {
		children.mutex.Unlock()
		return fmt.Errorf("not started: this run was stopped by %v", children.first)
	}
	// Started under the lock, so a signal arriving now is forwarded to this child once it is held, never
	// between its start and its holding.
	if err := child.Start(); err != nil {
		children.mutex.Unlock()
		return err
	}
	children.running[child] = true
	children.mutex.Unlock()

	err := child.Wait()

	children.mutex.Lock()
	delete(children.running, child)
	children.mutex.Unlock()
	return err
}

// received is the number of the signal that stopped this run, or 0 when none did.
func (children *projectChildren) received() int {
	children.mutex.Lock()
	defer children.mutex.Unlock()
	if number, isNumbered := children.first.(syscall.Signal); isNumbered {
		return int(number)
	}
	return 0
}

// stop stops catching signals. After it returns no signal is delivered to the channel, so it is closed and
// the forwarding goroutine ends.
func (children *projectChildren) stop() {
	signal.Stop(children.signals)
	close(children.signals)
}
