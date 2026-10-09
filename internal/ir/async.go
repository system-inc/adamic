package ir

// AsyncProgram is the first, straight-line suspension lowering. Its state graph is separate from
// synchronous tree bodies: no synchronous borrow, region or freshness proof may span an await.
type AsyncProgram struct {
	Functions []AsyncFunction
	Entry     int
	Generated []*GeneratedType
}
type AsyncFunction struct {
	Name       string
	Parameters []int
	Locals     []int
	Returns    Type
	States     []AsyncState
}
type AsyncState struct {
	Body   []Statement
	Await  *AsyncWait
	Target int
	Result Expression
	Throw  Expression
}
type AsyncWait struct {
	Function  int
	Arguments []Expression
	Value     Expression
	Of        Type
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
