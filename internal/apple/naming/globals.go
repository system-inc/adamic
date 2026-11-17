package naming

import "strings"

// Global names audited from all 69 transitive lib.es2024.d.ts libraries in the
// pinned TypeScript submodule. This is a list of names, not copied library code.
// Dictionary is additionally reserved by the Apple binding naming policy.
var globalNames = strings.Fields(`AggregateError AggregateErrorConstructor Array ArrayBuffer ArrayBufferConstructor ArrayBufferLike ArrayBufferTypes ArrayBufferView ArrayConstructor ArrayIterator ArrayLike AsyncGenerator AsyncGeneratorFunction AsyncGeneratorFunctionConstructor AsyncIterable AsyncIterableIterator AsyncIterator AsyncIteratorObject Atomics Awaited BigInt BigInt64Array BigInt64ArrayConstructor BigIntConstructor BigIntToLocaleStringOptions BigUint64Array BigUint64ArrayConstructor Boolean BooleanConstructor BuiltinIteratorReturn CallableFunction Capitalize ClassAccessorDecoratorContext ClassAccessorDecoratorResult ClassAccessorDecoratorTarget ClassDecorator ClassDecoratorContext ClassFieldDecoratorContext ClassGetterDecoratorContext ClassMemberDecoratorContext ClassMethodDecoratorContext ClassSetterDecoratorContext ConcatArray ConstructorParameters DataView DataViewConstructor Date DateConstructor DecoratorContext DecoratorMetadata DecoratorMetadataObject Error ErrorConstructor ErrorOptions EvalError EvalErrorConstructor Exclude Extract FinalizationRegistry FinalizationRegistryConstructor FlatArray Float32Array Float32ArrayConstructor Float64Array Float64ArrayConstructor Function FunctionConstructor Generator GeneratorFunction GeneratorFunctionConstructor IArguments ImportAssertions ImportAttributes ImportCallOptions ImportMeta Infinity InstanceType Int16Array Int16ArrayConstructor Int32Array Int32ArrayConstructor Int8Array Int8ArrayConstructor Intl Iterable IterableIterator Iterator IteratorObject IteratorResult IteratorReturnResult IteratorYieldResult JSON Lowercase Map MapConstructor MapIterator Math MethodDecorator NaN NewableFunction NoInfer NonNullable Number NumberConstructor Object ObjectConstructor Omit OmitThisParameter ParameterDecorator Parameters Partial Pick Promise PromiseConstructor PromiseConstructorLike PromiseFulfilledResult PromiseLike PromiseRejectedResult PromiseSettledResult PromiseWithResolvers PropertyDecorator PropertyDescriptor PropertyDescriptorMap PropertyKey Proxy ProxyConstructor ProxyHandler RangeError RangeErrorConstructor Readonly ReadonlyArray ReadonlyMap ReadonlySet Record ReferenceError ReferenceErrorConstructor Reflect RegExp RegExpConstructor RegExpExecArray RegExpIndicesArray RegExpMatchArray RegExpStringIterator Required ReturnType Set SetConstructor SetIterator SharedArrayBuffer SharedArrayBufferConstructor String StringConstructor StringIterator Symbol SymbolConstructor SyntaxError SyntaxErrorConstructor TemplateStringsArray ThisParameterType ThisType TypeError TypeErrorConstructor TypedPropertyDescriptor URIError URIErrorConstructor Uint16Array Uint16ArrayConstructor Uint32Array Uint32ArrayConstructor Uint8Array Uint8ArrayConstructor Uint8ClampedArray Uint8ClampedArrayConstructor Uncapitalize Uppercase WeakKey WeakKeyTypes WeakMap WeakMapConstructor WeakRef WeakRefConstructor WeakSet WeakSetConstructor decodeURI decodeURIComponent encodeURI encodeURIComponent escape eval isFinite isNaN parseFloat parseInt unescape Dictionary`)

func typeName(spelling, framework string) string {
	name := normalize(spelling, true)
	for _, global := range globalNames {
		if name == normalize(global, true) {
			if framework == "" {
				// MapType predates the framework fact. Known namespaces supply its default.
				for prefix, owner := range map[string]string{"NS": "Foundation", "UI": "UIKit", "CG": "CoreGraphics", "CF": "CoreFoundation", "CA": "QuartzCore", "AV": "AVFoundation"} {
					if strings.HasPrefix(spelling, prefix) {
						framework = owner
						break
					}
				}
			}
			if framework != "" {
				return normalize(framework, true) + name
			}
		}
	}
	return name
}
