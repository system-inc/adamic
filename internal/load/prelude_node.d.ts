// The console and process a program with Node's declarations has, beside prelude.d.ts, served at
// /adamic-prelude/node.d.ts (source_fs.go) in prelude_console.d.ts's place. Its console is Node's own
// Console, so the standalone console's narrower signatures never collide with @types/node's var.
//
// recordreplay is what tsc's sys.ts reads and Node's process doesn't have (#b7phwab). Adamic on #b5w3ycg,
// Nov 7 2026: it reads as undefined on every target, as the absent member lowering folds
// (hostBindingAbsentMembers), and every write or install of it stays refused. Optional and unknown, so it
// promises nothing. It lives here, never in prelude.d.ts, because a NodeJS namespace there would let a
// program with no Node types name NodeJS.Process (TS2503 is its promise,
// stage3/drivers/parser/native-nodejs-process-type.a).

declare var console: Console;

declare namespace NodeJS {
	interface Process {
		recordreplay?: unknown;
	}
}
