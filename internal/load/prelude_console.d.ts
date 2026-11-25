// The console a program with no Node declarations has: the prelude's beside prelude.d.ts, served at
// /adamic-prelude/console.d.ts (source_fs.go). A Node program has prelude_node.d.ts's in its place.

declare const console: {
	log(message: string): void;
	error(message: string): void;
};
