// The Adamic 0.1 prelude: what a program can use that es2024 does not declare.
// The full 0.1 library (docs/0.1.md) replaces es2024 with Adamic's own; until then, this.

declare const console: {
	log(message: string): void;
	error(message: string): void;
};

declare module 'adamic' {
	export function panic(message: string): never;

	// Input, opened in 0.2 (docs/0.1.md). A failure is a value: 0.2 has no exceptions yet.
	export function readTextFile(path: string): { readonly kind: 'Ok'; readonly text: string } | { readonly kind: 'Error'; readonly message: string };
	export function programArguments(): readonly string[];
	export function writeTextFile(path: string, text: string): { readonly kind: 'Ok' } | { readonly kind: 'Error'; readonly message: string };
	// The file system a walk needs: a directory's names in Node's order, and what a path names.
	export function readDirectory(path: string): { readonly kind: 'Ok'; readonly names: readonly string[] } | { readonly kind: 'Error'; readonly message: string };
	export function fileStatus(path: string): { readonly kind: 'Ok'; readonly type: 'file' | 'directory' | 'other'; readonly size: number; readonly symbolicLink: boolean } | { readonly kind: 'Error'; readonly message: string };

	export interface WeakBrand {
		readonly adamicWeak?: never;
	}
	export type Weak<Target extends object> = (Target & WeakBrand) | undefined;
}
