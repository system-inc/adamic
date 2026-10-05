// The Adamic 0.1 prelude: what a program can use that es2024 does not declare.
// The full 0.1 library (docs/0.1.md) replaces es2024 with Adamic's own; until then, this.

declare const console: {
	log(message: string): void;
	error(message: string): void;
};

declare module 'adamic' {
	export function panic(message: string): never;
}
