enum Kind { Circle, Square }
function fail(kind: unknown): never { throw new Error("unexpected kind " + String(kind)); }
function corners(kind: Kind): number {
	switch (kind) {
		case Kind.Circle: return 0;
		case Kind.Square: return 4;
	}
	return fail(kind);
}
console.log(String(corners(Kind.Circle)));
function other(): number { return 7; }
const kind: Kind = other();
console.log(String(corners(kind)));
