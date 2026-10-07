// The production Go path-state implementation below is copied without edits.
package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	state := correctnessNoProcessExitAfterOutputState{}
	for _, step := range os.Args[1:] {
		if step == "reset" {
			state = nil
		} else if strings.HasPrefix(step, "w") {
			state = state.with(step[1:])
		} else {
			id, err := strconv.Atoi(step[1:])
			if err != nil {
				panic(err)
			}
			state = state.entering(id)
		}
		fmt.Println(state.key())
	}
}

type correctnessNoProcessExitAfterOutputState []string

func correctnessNoProcessExitAfterOutputChainHolds(outer string, inner string) bool {
	for _, id := range strings.Split(strings.Trim(inner, ","), ",") {
		if id != "" && !strings.Contains(outer, ","+id+",") {
			return false
		}
	}
	return true
}

func (state correctnessNoProcessExitAfterOutputState) with(chain string) correctnessNoProcessExitAfterOutputState {
	next := correctnessNoProcessExitAfterOutputState{}
	for _, held := range state {
		if correctnessNoProcessExitAfterOutputChainHolds(chain, held) {
			// Something already held is forgotten only where this one is too.
			return state
		}
		if !correctnessNoProcessExitAfterOutputChainHolds(held, chain) {
			next = append(next, held)
		}
	}
	next = append(next, chain)
	sort.Strings(next)
	return next
}

func (state correctnessNoProcessExitAfterOutputState) entering(try int) correctnessNoProcessExitAfterOutputState {
	member := "," + strconv.Itoa(try) + ","
	next := correctnessNoProcessExitAfterOutputState{}
	for _, held := range state {
		if !strings.Contains(held, member) {
			next = append(next, held)
		}
	}
	return next
}

func (state correctnessNoProcessExitAfterOutputState) key() string {
	return strings.Join(state, "|")
}
