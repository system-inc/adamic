package ir

import "reflect"

// Await remains an ordinary expression until normalization exposes its suspension point.
type Await struct {
	Site  int // one-based source site, resolved by lowering only for diagnostics
	Value Expression
	Of    Type
}

func (a Await) Type() Type { return a.Of }

// PromiseValue creates a settled library promise. Reject payloads are Errors.
type PromiseValue struct {
	Value  Expression
	Reject bool
}

func (PromiseValue) Type() Type { return Promise }

func (p *Program) HasAsync() bool {
	for _, function := range p.Functions {
		if function.Async {
			return true
		}
	}
	return false
}

func (f Function) BodyReturns() Type {
	if f.Async {
		return f.AsyncReturns
	}
	if f.Generator {
		// A generator's body returns no value its consumer reads: return v is not yet built.
		return 0
	}
	return f.Returns
}

// Suspends is a function whose body runs on a counted frame and can stop part way and go on later:
// an async function at an await, a generator at a yield. Each reader of Async decides whether it
// means this or async alone.
func (f Function) Suspends() bool { return f.Async || f.Generator }

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
	return []*GeneratedType{{"adamic_generated_frame_0", asyncFrameIdentity}, {"adamic_async_promise", asyncPromiseIdentity}, {"adamic_async_reaction", asyncReactionIdentity}}
}
func (t *GeneratedType) RuntimeBreaksCycles() bool {
	return t != nil && (t.identity == asyncFrameIdentity || t.identity == asyncPromiseIdentity || t.identity == asyncReactionIdentity)
}

// HasPromises selects the scheduler and conservative async ownership. Retained
// host watchers need that same loop even without an explicit Promise expression.
func (p *Program) HasPromises() bool {
	if p.HasAsync() {
		return true
	}
	for _, local := range p.Locals {
		if local.Type == Promise {
			return true
		}
	}
	for _, function := range p.Functions {
		if function.Returns == Promise || containsPromise(reflect.ValueOf(function.Body)) {
			return true
		}
	}
	return containsPromise(reflect.ValueOf(p.Main))
}

// Promise expressions and loop-backed host operations can occur without a Promise local or return signature,
// for example under typeof. Walk the ordinary IR operands, not checker metadata.
func containsPromise(value reflect.Value) bool {
	if !value.IsValid() {
		return false
	}
	if value.CanInterface() {
		if call, ok := value.Interface().(NodeFSFile); ok && call.RetainsListener() {
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
