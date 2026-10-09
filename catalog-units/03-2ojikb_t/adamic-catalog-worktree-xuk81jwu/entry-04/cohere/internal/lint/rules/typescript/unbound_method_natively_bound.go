package typescript

// nativelyBoundMembers is upstream's `nativelyBoundMembers`, frozen.
//
// Upstream builds this set at module load by REFLECTING over the running Node global object: for
// each of fourteen supported namespaces it takes every own property that is a function and does not
// begin with an underscore. A compiled binary has no such object to reflect over, so the set is
// captured once and written down.
//
// # Why the set exists at all, and how load bearing it is
//
// The rule also has a type-level test, and upstream's comment says this name list exists because
// that test is not sufficient: a declaration can come from `@types/node` rather than from the
// default library, in which case `IsSymbolFromDefaultLibrary` answers false for something that is
// in fact natively bound.
//
// Measured rather than taken on trust. The installed rule was patched to answer false from this
// lookup and its whole corpus re-run: exactly FIVE of two hundred and eleven verdicts moved, and
// three of those five are `console`, which is declared in `@types/node` rather than in the default
// library. The other two are `({ parseInt } = Number)` and `({ log } = console)`.
//
// # What is different about a frozen list
//
// Upstream reads whichever interpreter it happens to run on, and its own comment notes that a
// namespace like `Intl` may be absent depending on how Node was compiled. This reads a snapshot.
// The two agree on the runtime it was captured from and can drift on another, which is a real
// divergence and is recorded here rather than hidden.
//
// Captured with all fourteen namespaces present, 147 members. Generated rather than typed, and
// sorted within each namespace so a regeneration produces a reviewable diff.
var nativelyBoundMembers = map[string]bool{
	// Number, 6 members
	"Number.isFinite":      true,
	"Number.isInteger":     true,
	"Number.isNaN":         true,
	"Number.isSafeInteger": true,
	"Number.parseFloat":    true,
	"Number.parseInt":      true,
	// Object, 23 members
	"Object.assign":                    true,
	"Object.create":                    true,
	"Object.defineProperties":          true,
	"Object.defineProperty":            true,
	"Object.entries":                   true,
	"Object.freeze":                    true,
	"Object.fromEntries":               true,
	"Object.getOwnPropertyDescriptor":  true,
	"Object.getOwnPropertyDescriptors": true,
	"Object.getOwnPropertyNames":       true,
	"Object.getOwnPropertySymbols":     true,
	"Object.getPrototypeOf":            true,
	"Object.groupBy":                   true,
	"Object.hasOwn":                    true,
	"Object.is":                        true,
	"Object.isExtensible":              true,
	"Object.isFrozen":                  true,
	"Object.isSealed":                  true,
	"Object.keys":                      true,
	"Object.preventExtensions":         true,
	"Object.seal":                      true,
	"Object.setPrototypeOf":            true,
	"Object.values":                    true,
	// String, 3 members
	"String.fromCharCode":  true,
	"String.fromCodePoint": true,
	"String.raw":           true,
	// RegExp, 1 members
	"RegExp.escape": true,
	// Symbol, 2 members
	"Symbol.for":    true,
	"Symbol.keyFor": true,
	// Array, 4 members
	"Array.from":      true,
	"Array.fromAsync": true,
	"Array.isArray":   true,
	"Array.of":        true,
	// Proxy, 1 members
	"Proxy.revocable": true,
	// Date, 3 members
	"Date.UTC":   true,
	"Date.now":   true,
	"Date.parse": true,
	// Atomics, 14 members
	"Atomics.add":             true,
	"Atomics.and":             true,
	"Atomics.compareExchange": true,
	"Atomics.exchange":        true,
	"Atomics.isLockFree":      true,
	"Atomics.load":            true,
	"Atomics.notify":          true,
	"Atomics.or":              true,
	"Atomics.pause":           true,
	"Atomics.store":           true,
	"Atomics.sub":             true,
	"Atomics.wait":            true,
	"Atomics.waitAsync":       true,
	"Atomics.xor":             true,
	// Reflect, 13 members
	"Reflect.apply":                    true,
	"Reflect.construct":                true,
	"Reflect.defineProperty":           true,
	"Reflect.deleteProperty":           true,
	"Reflect.get":                      true,
	"Reflect.getOwnPropertyDescriptor": true,
	"Reflect.getPrototypeOf":           true,
	"Reflect.has":                      true,
	"Reflect.isExtensible":             true,
	"Reflect.ownKeys":                  true,
	"Reflect.preventExtensions":        true,
	"Reflect.set":                      true,
	"Reflect.setPrototypeOf":           true,
	// console, 25 members
	"console.Console":        true,
	"console.assert":         true,
	"console.clear":          true,
	"console.context":        true,
	"console.count":          true,
	"console.countReset":     true,
	"console.createTask":     true,
	"console.debug":          true,
	"console.dir":            true,
	"console.dirxml":         true,
	"console.error":          true,
	"console.group":          true,
	"console.groupCollapsed": true,
	"console.groupEnd":       true,
	"console.info":           true,
	"console.log":            true,
	"console.profile":        true,
	"console.profileEnd":     true,
	"console.table":          true,
	"console.time":           true,
	"console.timeEnd":        true,
	"console.timeLog":        true,
	"console.timeStamp":      true,
	"console.trace":          true,
	"console.warn":           true,
	// Math, 36 members
	"Math.abs":      true,
	"Math.acos":     true,
	"Math.acosh":    true,
	"Math.asin":     true,
	"Math.asinh":    true,
	"Math.atan":     true,
	"Math.atan2":    true,
	"Math.atanh":    true,
	"Math.cbrt":     true,
	"Math.ceil":     true,
	"Math.clz32":    true,
	"Math.cos":      true,
	"Math.cosh":     true,
	"Math.exp":      true,
	"Math.expm1":    true,
	"Math.f16round": true,
	"Math.floor":    true,
	"Math.fround":   true,
	"Math.hypot":    true,
	"Math.imul":     true,
	"Math.log":      true,
	"Math.log10":    true,
	"Math.log1p":    true,
	"Math.log2":     true,
	"Math.max":      true,
	"Math.min":      true,
	"Math.pow":      true,
	"Math.random":   true,
	"Math.round":    true,
	"Math.sign":     true,
	"Math.sin":      true,
	"Math.sinh":     true,
	"Math.sqrt":     true,
	"Math.tan":      true,
	"Math.tanh":     true,
	"Math.trunc":    true,
	// JSON, 4 members
	"JSON.isRawJSON": true,
	"JSON.parse":     true,
	"JSON.rawJSON":   true,
	"JSON.stringify": true,
	// Intl, 12 members
	"Intl.Collator":            true,
	"Intl.DateTimeFormat":      true,
	"Intl.DisplayNames":        true,
	"Intl.DurationFormat":      true,
	"Intl.ListFormat":          true,
	"Intl.Locale":              true,
	"Intl.NumberFormat":        true,
	"Intl.PluralRules":         true,
	"Intl.RelativeTimeFormat":  true,
	"Intl.Segmenter":           true,
	"Intl.getCanonicalLocales": true,
	"Intl.supportedValuesOf":   true,
}
