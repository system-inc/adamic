// The Adamic 0.1 prelude: what a program can use that es2024 does not declare.
// The full 0.1 library (docs/0.1.md) replaces es2024 with Adamic's own; until then, this.

declare const console: {
	log(message: string): void;
	error(message: string): void;
};

declare module 'adamic' {
	// External checker library. Native build requires --tsgo <archive>; JavaScript is refused.
	// Paths and positions follow bridge/tsgo/tsgo.h. Release each handle exactly once.
	export function tsgoProgram(tsconfig: string, files: readonly string[]): number;
	export function tsgoQuery(program: number, file: string, bytePosition: number): { readonly nodeKind: number; readonly symbolName: string; readonly type: string };
	export function tsgoRelease(program: number): void;

	export function panic(message: string): never;

	// Input, opened in 0.2 (docs/0.1.md). A failure is a value: 0.2 has no exceptions yet.
	export function readTextFile(path: string): { readonly kind: 'Ok'; readonly text: string } | { readonly kind: 'Error'; readonly message: string };
	export function programArguments(): readonly string[];
	export function writeTextFile(path: string, text: string): { readonly kind: 'Ok' } | { readonly kind: 'Error'; readonly message: string };
	// The file system a walk needs: a directory's names in Node's order, and what a path names.
	export function readDirectory(path: string): { readonly kind: 'Ok'; readonly names: readonly string[] } | { readonly kind: 'Error'; readonly message: string };
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
