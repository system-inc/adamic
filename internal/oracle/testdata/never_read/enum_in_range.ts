// An open numeric enum parameter a switch exhausts, its default reading it (#p6vawef, Language's third
// premise for the own-function trust): the read is enum_never.go's, which traps a value no member is,
// and run with each member it agrees with Node.
enum K { A, B }
function stray(value: unknown): string { return "stray " + String(value); }
function name(k: K): string {
	switch (k) {
		case K.A: return "a";
		case K.B: return "b";
		default: return stray(k);
	}
}
console.log(name(K.A) + " " + name(K.B));
