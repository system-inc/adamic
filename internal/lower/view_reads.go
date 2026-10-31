package lower

import "github.com/system-inc/adamic/internal/ir"

// viewRead is what an IR expression does with an object operand it holds, for checked views. A
// value a view admitted stays a checked value under every type it's later seen as (Adamic's rule on
// #80cs7rz), so each expression that can hold one either checks it, reads none of its fields, or is
// NotYet where it may hold one. view_reads_test.go finds every expression type in internal/ir's
// source with an operand and fails on one viewReadOf doesn't class, so a new one can't be forgotten.
type viewRead uint8

const (
	// viewReadChecks reads each field it reads through a check: the view's, at a checked
	// receiver's field read, or the run time's against the declared type.
	viewReadChecks viewRead = iota + 1
	// viewReadsNoField reads no field of an object operand: it moves, compares or tests the
	// reference, or reads only what no view claims.
	viewReadsNoField
	// viewReadNotYet reads fields raw, so it's NotYet where its operand may hold a value from a
	// view's entry, failing closed on unknown reach: the widened read and the whole-object
	// readers (checkViewWidenedReads, checkViewCopies).
	viewReadNotYet
)

// viewReadOf is the one table: each IR expression with an operand, by its class, which the
// post-lowering passes read. Zero is an expression nobody has classed.
func viewReadOf(program *ir.Program, expression ir.Expression) viewRead {
	switch value := expression.(type) {
	case ir.Property:
		if program.CheckedField(value.ViewReceiverTypeID, value.Name) {
			return viewReadChecks
		}
		return viewReadNotYet
	case ir.ObjectLiteral:
		if value.Spread != nil {
			return viewReadNotYet
		}
		return viewReadsNoField
	case ir.ObjectCall:
		switch value.Method {
		case "assign":
			return viewReadNotYet
		case "values", "entries":
			// Unchecked, the operand is a literal's own const binding (enumerationProven), whose
			// fields no cast claimed.
			if value.Checked {
				return viewReadChecks
			}
		}
		return viewReadsNoField
	case ir.JSONEncode:
		return viewReadNotYet
	case ir.CheckedCast, ir.DynamicProperty:
		return viewReadChecks
	case ir.ArrayConcat, ir.ArrayFill, ir.ArrayFrom, ir.ArrayIndex, ir.ArrayIsArray, ir.ArrayJoin, ir.ArrayLiteral,
		ir.ArrayMap, ir.ArrayPop, ir.ArrayPush, ir.ArrayReduce, ir.ArrayReverse, ir.ArraySearch, ir.ArraySlice,
		ir.ArraySort, ir.ArraySplice, ir.ArrayVisit, ir.Await, ir.Binary, ir.BooleanToString, ir.Box, ir.Call,
		ir.CallClosure, ir.CharCodeAt, ir.Coalesce, ir.CodePoints, ir.CollectionIterator, ir.Comma, ir.Concat,
		ir.Conditional, ir.Defined, ir.Effects, ir.FileStatus, ir.HasAccessor, ir.HasOwn, ir.HasProperty,
		ir.InstanceOf, ir.IsNull, ir.IsUndefined, ir.JSONDecode, ir.JSONStringify, ir.Length, ir.Logical,
		ir.LogicalAssignment, ir.MakeError, ir.MapClear, ir.MapDelete, ir.MapEntries, ir.MapForEach, ir.MapGet,
		ir.MapHas, ir.MapKeys, ir.MapNew, ir.MapSet, ir.MapSize, ir.MapValues, ir.MathCall, ir.MaybeOf,
		ir.MaybeToString, ir.Narrow, ir.NodeBufferCall, ir.NodeFSFile, ir.NodeHostCall, ir.NumberCall,
		ir.NumberFormat, ir.NumberToString, ir.ObjectKeys, ir.ParallelMap, ir.PhantomMember, ir.PlaceholderUse, ir.ProcessCall,
		ir.PromiseValue, ir.ReadDirectory, ir.ReadTextFile, ir.RecordCall, ir.RecordCoalesce, ir.RecordLiteral,
		ir.RegExpCall, ir.RegExpGroup, ir.RegExpNew, ir.RegExpProperty, ir.SetAdd, ir.SetNew, ir.SetValues,
		ir.StringCall, ir.StringFromCodes, ir.StringIndex, ir.StringIterator, ir.StringLength, ir.ToFixed, ir.Trim,
		ir.Truthy, ir.TypeOf, ir.TypedArrayFill, ir.TypedArrayNew, ir.TypedArraySet, ir.TypedArraySort,
		ir.TypedArraySubarray, ir.Unary, ir.UnionToString, ir.Unwrap, ir.Utf8At, ir.Utf8Length, ir.Void, ir.WeakOf,
		ir.WeakTarget, ir.WriteTextFile:
		return viewReadsNoField
	}
	return 0
}
