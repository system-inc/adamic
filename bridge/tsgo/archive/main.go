// Build with go build -buildmode=c-archive -o /path/tsgo.a ./bridge/tsgo/archive.
package main

/*
#include "tsgo.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"

	bridge "github.com/system-inc/adamic/bridge/tsgo/checker"
)

var programs = struct {
	sync.Mutex
	next uint64
	live map[uint64]*bridge.Program
}{live: make(map[uint64]*bridge.Program)}

func text(view C.tsgo_view) string { return C.GoStringN(view.data, C.int(view.length)) }
func buffer(value string) C.tsgo_buffer {
	if len(value) == 0 {
		return C.tsgo_buffer{}
	}
	data := C.malloc(C.size_t(len(value)))
	if data == nil {
		panic("out of memory allocating bridge output")
	}
	C.memcpy(data, unsafe.Pointer(unsafe.StringData(value)), C.size_t(len(value)))
	return C.tsgo_buffer{data: (*C.char)(data), length: C.size_t(len(value))}
}
func failed(error *C.tsgo_buffer, status int, value any) C.int {
	*error = buffer(fmt.Sprint(value))
	return C.int(status)
}
func recoverFailure(status *C.int, error *C.tsgo_buffer) {
	if value := recover(); value != nil {
		*status = failed(error, C.TSGO_CHECKER, value)
	}
}

//export tsgo_go_create
func tsgo_go_create(config C.tsgo_view, files *C.tsgo_view, count C.size_t, handle *C.tsgo_handle, error *C.tsgo_buffer) (status C.int) {
	defer recoverFailure(&status, error)
	programs.Lock()
	defer programs.Unlock()
	roots := make([]string, int(count))
	for index, file := range unsafe.Slice(files, int(count)) {
		roots[index] = text(file)
	}
	program, err := bridge.Open(text(config), roots)
	if err != nil {
		return failed(error, C.TSGO_CHECKER, err)
	}
	// Adamic numbers represent these IDs exactly; refuse rather than wrap or round.
	if programs.next == (1<<53)-1 {
		return failed(error, C.TSGO_HANDLE, "handle space exhausted")
	}
	startProfile()
	programs.next++
	programs.live[programs.next] = program
	*handle = C.tsgo_handle(programs.next)
	return C.TSGO_OK
}

//export tsgo_go_query
func tsgo_go_query(handle C.tsgo_handle, file C.tsgo_view, position C.uint64_t, result *C.tsgo_result, error *C.tsgo_buffer) (status C.int) {
	defer recoverFailure(&status, error)
	programs.Lock()
	defer programs.Unlock()
	program := programs.live[uint64(handle)]
	if program == nil {
		return failed(error, C.TSGO_HANDLE, "invalid or released checker handle")
	}
	answer, err := program.Query(text(file), uint64(position))
	if err != nil {
		return failed(error, C.TSGO_CHECKER, err)
	}
	result.kind = C.uint32_t(answer.Kind)
	result.symbol = buffer(answer.Symbol)
	result._type = buffer(answer.Type)
	return C.TSGO_OK
}

//export tsgo_go_type_parts
func tsgo_go_type_parts(handle C.tsgo_handle, file C.tsgo_view, start, end C.uint64_t, kind C.tsgo_view, parts, error *C.tsgo_buffer) (status C.int) {
	defer recoverFailure(&status, error)
	programs.Lock()
	defer programs.Unlock()
	started := queryStarted()
	defer queryFinished(started, true)
	program := programs.live[uint64(handle)]
	if program == nil {
		return failed(error, C.TSGO_HANDLE, "invalid or released checker handle")
	}
	answer, err := program.TypeParts(text(file), uint64(start), uint64(end), text(kind))
	if err != nil {
		return failed(error, C.TSGO_CHECKER, err)
	}
	*parts = buffer(answer)
	return C.TSGO_OK
}

//export tsgo_go_inspect
func tsgo_go_inspect(handle C.tsgo_handle, file C.tsgo_view, start, end C.uint64_t, kind, question C.tsgo_view, facts, error *C.tsgo_buffer) (status C.int) {
	defer recoverFailure(&status, error)
	programs.Lock()
	defer programs.Unlock()
	started := queryStarted()
	defer queryFinished(started, false)
	program := programs.live[uint64(handle)]
	if program == nil {
		return failed(error, C.TSGO_HANDLE, "invalid or released checker handle")
	}
	answer, err := program.Inspect(text(file), uint64(start), uint64(end), text(kind), text(question))
	if err != nil {
		return failed(error, C.TSGO_CHECKER, err)
	}
	*facts = buffer(answer)
	return C.TSGO_OK
}

//export tsgo_go_release
func tsgo_go_release(handle C.tsgo_handle, error *C.tsgo_buffer) (status C.int) {
	defer recoverFailure(&status, error)
	programs.Lock()
	defer programs.Unlock()
	if programs.live[uint64(handle)] == nil {
		return failed(error, C.TSGO_HANDLE, "invalid or released checker handle")
	}
	delete(programs.live, uint64(handle))
	if len(programs.live) == 0 {
		stopProfile()
	}
	return C.TSGO_OK
}
func main() {}
