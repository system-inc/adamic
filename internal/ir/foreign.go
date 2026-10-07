package ir

// Foreign is a function whose body lives outside Adamic: an Objective-C message or a C function,
// declared in an apple/ binding file (internal/load/apple, docs/apple.md). Its IR body only panics,
// so a backend that can't make the call (JavaScript) says so out loud; the native backend replaces
// the body with the call itself (internal/native/foreign.go).
//
// Every reference parameter is borrowed, and a reference result comes back owned, as a fresh value.
type Foreign struct {
	Kind ForeignKind

	// Class is the Objective-C class a Construct or a ClassMessage is sent to.
	Class string

	// Selector is the message's selector, or the C function's symbol.
	Selector string

	// Arguments are the native arguments, in the order the selector or the C function takes them.
	// For an InstanceMessage the receiver isn't one of them: it's the function's first parameter.
	Arguments []ForeignArgument

	// Returns is the native result, and Retained says it comes back with a count the caller owns, as
	// alloc, new and copy give it; any other object result is retained on its way into Adamic.
	Returns  NativeType
	Retained bool
}

// ForeignKind is how a Foreign function is called.
type ForeignKind int

const (
	// Construct is new Class(...): alloc sent to the class, then Selector (an init) to what it made.
	Construct ForeignKind = iota + 1

	// ClassMessage is Selector sent to the class itself.
	ClassMessage

	// InstanceMessage is Selector sent to the function's first parameter.
	InstanceMessage

	// CFunction is a C function, Selector its symbol.
	CFunction
)

// ForeignArgument is one native argument and where its value comes from.
type ForeignArgument struct {
	Type NativeType

	// Parameter is the function's parameter that holds the value, or -1 when the source leaves the
	// field out and Default is passed instead.
	Parameter int

	// Default is what's passed when the source leaves an optional field out, as Type reads it.
	Default string
}

// NativeType is how a value crosses into native code and back.
type NativeType struct {
	Kind NativeKind

	// Nullable says nil is a value (T | null in Adamic): only an Object may be.
	Nullable bool

	// Names and Values are an enumeration's or an option set's members: the string literal Adamic
	// writes, and the integer the native side takes.
	Names  []string
	Values []int64

	// Parameters are a block's: what Apple hands the closure when it calls it.
	Parameters []NativeType
}

// NativeKind is a native value's representation.
type NativeKind int

const (
	NativeVoid NativeKind = iota + 1

	// NativeDouble is a double or a CGFloat, an Adamic number.
	NativeDouble

	// NativeInteger is a long or an NSInteger, and NativeUnsigned an unsigned long or an NSUInteger, an
	// Adamic number truncated toward zero on the way in.
	NativeInteger
	NativeUnsigned

	// NativeBoolean is a BOOL.
	NativeBoolean

	// NativeString is an NSString, an Adamic string.
	NativeString

	// NativeObject is any Objective-C object, held in Adamic by one strong reference (a foreign box).
	NativeObject

	// NativeRectangle is a CGRect, an Adamic { x, y, width, height }.
	NativeRectangle

	// NativeEnumeration is an integer named by one string literal, and NativeOptions an integer made
	// of the bits an array of string literals names.
	NativeEnumeration
	NativeOptions

	// NativeAction is a closure an Objective-C control calls when it acts: two native arguments, the
	// target and its action selector, made from one Adamic () => void.
	NativeAction

	// NativeBlock is a closure as an Objective-C block that returns nothing: Apple calls it with
	// Parameters, on whatever thread it likes, and the closure runs on the main thread.
	NativeBlock
)
