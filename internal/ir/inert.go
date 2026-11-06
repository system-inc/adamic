package ir

import (
	"reflect"
	"strings"
)

// Inert reports whether an expression's own operation runs no program code and reaches nothing
// outside the program's own values: no call, no callback, no generator's body, no getter, no other
// task run while it waits, and no host (a file, the process, the console). What its operands do is
// theirs, and a proof walks them on its own. The list is fixed, and anything not on it, an operation
// added later included, is not inert, so a proof that asks this before passing a node it has no case
// for fails closed on a node it has never seen (Compiler's and Adamic's ruling on #qk8rztp S2c-2b:
// worker/purity.go's and native/reuse.go's defaults).
//
// Off the list, and each such proof's own case to judge: a call (Call, CallClosure, and the runtime
// loops that call a callback, ArrayMap, ArrayVisit, ArrayReduce, ArrayFrom, ArraySort, MapForEach,
// ParallelMap), a generator's step (GeneratorNext), a pattern's replacement callback (RegExpCall), a
// run-time key read, which calls whichever getter its key names (DynamicKeyedProperty), an encoding,
// which internal/flow takes to call back (JSONEncode), a suspension (Await, Yield), and every host
// operation (FileStatus, NodeFSFile, NodeHostCall, ProcessCall, ProgramArguments, ReadDirectory,
// ReadTextFile, WriteTextFile).
func Inert(expression Expression) bool {
	if expression == nil {
		return true
	}
	return inertExpressions[reflect.TypeOf(expression).Name()]
}

var inertExpressions = func() map[string]bool {
	result := map[string]bool{}
	for _, name := range strings.Fields(`ArrayConcat ArrayFill ArrayIndex ArrayIsArray ArrayJoin ArrayLiteral ArrayOfScalars ArrayPop
		ArrayPush ArrayReverse ArraySearch ArraySlice ArraySplice Binary BooleanConstant BooleanToString Box CharCodeAt
		CheckedCast ClosureSelf Coalesce CodePoints CollectionIterator Comma Concat Conditional Defined DynamicProperty
		Effects ErrorToString GeneratorDone HasAccessor HasOwn HasProperty InstanceOf IsCollectionIterator IsGenerator
		IsNull IsUndefined IterableOf JSONDecode JSONNull JSONStringify JoinEnter JoinLeave Length LibraryGlobal Logical
		LogicalAssignment MakeClosure MakeError MapClear MapDelete MapEntries MapGet MapHas MapKeys MapNew MapSet MapSize
		MapValues MathCall MaybeOf MaybeToString Narrow NodeBufferCall Null NumberCall NumberConstant NumberFormat
		NumberToString ObjectCall ObjectKeys ObjectLiteral OwnFunction PhantomMember PlaceholderUse PromiseValue Property
		Read RecordCall RecordCoalesce RecordLiteral RecordStorage RegExpGroup RegExpNew RegExpProperty SetAdd SetNew
		SetValues StringCall StringConstant StringFromCodes StringIndex StringIterator StringLength ToFixed Trim Truthy
		TypeOf TypedArrayFill TypedArrayNew TypedArraySet TypedArraySort TypedArraySubarray Unary Undefined UnionToString
		UnknownKind Unwrap Utf8At Utf8Length Void WeakOf WeakTarget`) {
		result[name] = true
	}
	return result
}()
