// The Adamic 0.1 prelude: what a program can use that es2024 does not declare.
// The full 0.1 library (docs/0.1.md) replaces es2024 with Adamic's own; until then, this.

declare const console: {
	log(message: string): void;
	error(message: string): void;
};

// The supported process surface. Environment and stream observations are read-only.
// Nonterminal streams have no isTTY value, as on Node.
declare const process: {
	exitCode: number | undefined;
	readonly exit: (code?: number) => never;
	readonly stdout: { readonly isTTY: true | undefined };
	readonly stderr: { readonly isTTY: true | undefined };
	readonly env: { readonly [name: string]: string | undefined };
};

declare module 'adamic' {
	export function parallelMap<T, R>(items: readonly T[], work: (item: T, index: number) => R): R[];
	// External checker library. Native build requires --tsgo <archive>; JavaScript is refused.
	// Paths and positions follow bridge/tsgo/tsgo.h. Release each handle exactly once.
	export function tsgoProgram(tsconfig: string, files: readonly string[]): number;
	export function tsgoQuery(program: number, file: string, bytePosition: number): { readonly nodeKind: number; readonly symbolName: string; readonly type: string };
	// Constrained union parts, framed as documented in tsgo.h. Exact byte span and kind.
	export function tsgoTypeParts(program: number, file: string, byteStart: number, byteEnd: number, nodeKind: string): string;
	// Length-framed checker facts. Questions and schema: bridge/tsgo/facts.md.
	export function tsgoInspect(program: number, file: string, byteStart: number, byteEnd: number, nodeKind: string, question: string): string;
	export function tsgoRelease(program: number): void;

	export function panic(message: string): never;

	// Input, opened in 0.2 (docs/0.1.md). A failure is a value: 0.2 has no exceptions yet.
	export function readTextFile(path: string): { readonly kind: 'Ok'; readonly text: string } | { readonly kind: 'Error'; readonly message: string };
	export function programArguments(): readonly string[];
	export function writeTextFile(path: string, text: string): { readonly kind: 'Ok' } | { readonly kind: 'Error'; readonly message: string };
	// The file system a walk needs: a directory's names in Node's order, and what a path names.
	export function readDirectory(path: string): { readonly kind: 'Ok'; readonly names: readonly string[] } | { readonly kind: 'Error'; readonly message: string };
	export function realPath(path: string): { readonly kind: 'Ok'; readonly path: string } | { readonly kind: 'Error'; readonly message: string };
	export function fileStatus(path: string): { readonly kind: 'Ok'; readonly type: 'file' | 'directory' | 'other'; readonly size: number; readonly symbolicLink: boolean } | { readonly kind: 'Error'; readonly message: string };

	// A string's UTF-8, read in place: how many bytes it is, and the byte at an index, 0 to 255. A lone
	// surrogate is U+FFFD's bytes, as TextEncoder writes it. utf8At panics where the index isn't a byte.
	export function utf8Length(text: string): number;
	export function utf8At(text: string, index: number): number;

	export interface WeakBrand {
		readonly adamicWeak?: never;
	}
	export type Weak<Target extends object> = (Target & WeakBrand) | undefined;
}

// ES2025 Set operations. Stage 0 accepts concrete library Sets with matching representations.
interface Set<T> {
	union<U>(other: ReadonlySetLike<U>): Set<T | U>;
	intersection<U>(other: ReadonlySetLike<U>): Set<T & U>;
	difference<U>(other: ReadonlySetLike<U>): Set<T>;
	symmetricDifference<U>(other: ReadonlySetLike<U>): Set<T | U>;
	isSubsetOf<U>(other: ReadonlySetLike<U>): boolean;
	isSupersetOf<U>(other: ReadonlySetLike<U>): boolean;
	isDisjointFrom<U>(other: ReadonlySetLike<U>): boolean;
}
interface ReadonlySet<T> {
	union<U>(other: ReadonlySetLike<U>): Set<T | U>;
	intersection<U>(other: ReadonlySetLike<U>): Set<T & U>;
	difference<U>(other: ReadonlySetLike<U>): Set<T>;
	symmetricDifference<U>(other: ReadonlySetLike<U>): Set<T | U>;
	isSubsetOf<U>(other: ReadonlySetLike<U>): boolean;
	isSupersetOf<U>(other: ReadonlySetLike<U>): boolean;
	isDisjointFrom<U>(other: ReadonlySetLike<U>): boolean;
}

// Built-in collection iterators always provide done, including false on a yielded result.
// Requiring it here avoids an optional boolean slot while preserving the ECMAScript result union.
type AdamicCollectionIteratorResult<T> =
	| { readonly done: false; readonly value: T }
	| { readonly done: true; readonly value: undefined };
interface MapIterator<T> {
	next(): AdamicCollectionIteratorResult<T>;
}
interface SetIterator<T> {
	next(): AdamicCollectionIteratorResult<T>;
}

interface ReadonlySetLike<T> {
	readonly size: number;
	readonly has: (value: T) => boolean;
	readonly keys: () => Iterator<T>;
}

// stringify can return undefined. The bundled TypeScript declaration's plain string result
// is not a proven type, so this overload keeps callers honest about the missing result.
interface JSON {
	stringify(value?: unknown, replacer?: unknown, space?: unknown): string | undefined;
}
