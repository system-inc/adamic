package ir

import "reflect"

// Await remains an ordinary expression until normalization exposes its suspension point.
type Await struct {
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
	return f.Returns
}

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

// Promise expressions can occur without a Promise local or return signature,
// for example under typeof. Walk the ordinary IR operands, not checker metadata.
func containsPromise(value reflect.Value) bool {
	if !value.IsValid() {
		return false
	}
	if value.CanInterface() {
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
