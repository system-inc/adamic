/*
 * Road P (#p2n17c4, rulings.md:117 applied to the right of an &&): a presence test on the right of an &&
 * that is tested for truthiness alone, recursively: the test of ?:, an if, a chain of &&, an && nested on
 * either side of another in an if, the operand of !, and a while. Each runs over a class's required
 * method, which method presence folds, and an optional member, given and missing, which shape 1 reads off
 * its slot. Node prints, a line per flag pair: abcd falsef | ---- true- | abcd falsef, then
 * a--- falsef | ---- true- | a--- falsef, then ---- true- | ---- true- | ---- true- twice.
 */
class Counter {
	m(): number {
		return 1;
	}
}

interface Maybe {
	m?(): number;
}

const counter = new Counter();
const given: Maybe = { m() { return 2; } };
const missing: Maybe = {};

function optional(flag: boolean, other: boolean, h: Maybe): string {
	const a = flag && h.m ? 'a' : '-';
	let b = '-';
	if (flag && other && h.m) {
		b = 'b';
	}
	let c = '-';
	if (flag && (other && h.m)) {
		c = 'c';
	}
	let d = '-';
	if ((flag && h.m) && other) {
		d = 'd';
	}
	const e = String(!(flag && h.m));
	let f = '-';
	while (flag && h.m) {
		f = 'f';
		break;
	}
	return a + b + c + d + ' ' + e + f;
}

function method(flag: boolean, other: boolean): string {
	const a = flag && counter.m ? 'a' : '-';
	let b = '-';
	if (flag && other && counter.m) {
		b = 'b';
	}
	let c = '-';
	if (flag && (other && counter.m)) {
		c = 'c';
	}
	let d = '-';
	if ((flag && counter.m) && other) {
		d = 'd';
	}
	const e = String(!(flag && counter.m));
	let f = '-';
	while (flag && counter.m) {
		f = 'f';
		break;
	}
	return a + b + c + d + ' ' + e + f;
}

function line(flag: boolean, other: boolean): string {
	return `${optional(flag, other, given)} | ${optional(flag, other, missing)} | ${method(flag, other)}`;
}

console.log(line(true, true));
console.log(line(true, false));
console.log(line(false, true));
console.log(line(false, false));
