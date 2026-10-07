package ir

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
		if function.Returns == Promise {
			return true
		}
	}
	return false
}
