// Spec: a shorthand or spread whose slot holds `undefined` or is a union
// Expect: refused a value of type R seen as W, whose readonly field t becomes writable
interface R {
	readonly t: number;
}

interface W {
	t: number;
}

function make(p: R): { p: W } | undefined {
	if (p.t > 0) {
		return { p };
	}
	return undefined;
}

const r: R = { t: 2 };
const x = make(r);
if (x) {
	x.p.t = 1;
}
console.log(String(r.t));
