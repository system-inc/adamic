package ir

import "reflect"

// Await remains an ordinary expression until normalization exposes its suspension point.
type Await struct {
	Site  int // one-based source site, resolved by lowering only for diagnostics
	Value Expression
	Of    Type
}

func (await Await) Type() Type { return await.Of }

// PromiseValue creates a settled library promise. Reject payloads are Errors.
type PromiseValue struct {
	Value  Expression
	Reject bool
}

func (PromiseValue) Type() Type { return Promise }

func (program *Program) HasAsync() bool {
	for _, function := range program.Functions {
		if function.Async {
			return true
		}
	}
	return false
}

func (function Function) BodyReturns() Type {
	if function.Async {
		return function.AsyncReturns
	}
	if function.Generator {
		// A generator's body returns no value its consumer reads: return v is not yet built.
		return 0
	}
	return function.Returns
}

// Suspends is a function whose body runs on a counted frame and can stop part way and go on later:
// an async function at an await, a generator at a yield. Each reader of Async decides whether it
// means this or async alone.
func (function Function) Suspends() bool { return function.Async || function.Generator }

// GeneratedType has an unexported provenance identity. Source names, fields and brands cannot
// construct it. Only these canonical protocol identities are accepted by the cycle finder.
type runtimeIdentity struct{ kind int }
type GeneratedType struct {
	Name     string
	identity *runtimeIdentity
}

var asyncFrameIdentity = &runtimeIdentity{1}
var asyncPromiseIdentity = &runtimeIdentity{2}
var asyncReactionIdentity = &runtimeIdentity{3}

func AsyncGeneratedTypes() []*GeneratedType {
	return []*GeneratedType{
		{"adamic_generated_frame_0", asyncFrameIdentity},
		{"adamic_async_promise", asyncPromiseIdentity},
		{"adamic_async_reaction", asyncReactionIdentity},
	}
}
func (generatedType *GeneratedType) RuntimeBreaksCycles() bool {
	return generatedType != nil && (generatedType.identity == asyncFrameIdentity ||
		generatedType.identity == asyncPromiseIdentity ||
		generatedType.identity == asyncReactionIdentity)
}

// HasPromises selects the scheduler and conservative async ownership. Retained
// host watchers need that same loop even without an explicit Promise expression.
func (program *Program) HasPromises() bool {
	if program.HasAsync() {
		return true
	}
	for _, local := range program.Locals {
		if local.Type == Promise {
			return true
		}
	}
	for _, function := range program.Functions {
		if function.Returns == Promise || containsPromise(reflect.ValueOf(function.Body)) {
			return true
		}
	}
	return containsPromise(reflect.ValueOf(program.Main))
}

// Promise expressions and loop-backed host operations can occur without a Promise local or return signature,
// for example under typeof. Walk the ordinary IR operands, not checker metadata.
func containsPromise(value reflect.Value) bool {
	if !value.IsValid() {
		return false
	}
	if value.CanInterface() {
		if call, ok := value.Interface().(NodeFSFile); ok && call.Operation == "host_watch_file" {
			return true
		}
		if expression, ok := value.Interface().(Expression); ok && expression.Type() == Promise {
			return true
		}
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		return !value.IsNil() && containsPromise(value.Elem())
	case reflect.Struct:
		if value.Type().PkgPath() != reflect.TypeFor[PromiseValue]().PkgPath() {
			return false
		}
		for index := 0; index < value.NumField(); index++ {
			if containsPromise(value.Field(index)) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			if containsPromise(value.Index(index)) {
				return true
			}
		}
	}
	return false
}
