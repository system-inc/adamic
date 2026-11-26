// Compiler's row on #xfpsqtn: one assertNever call reached by two paths, a switch and a chain of ifs, so
// the trap is the read's, judged each time it runs, and not the call site's.
enum Side {
	Left,
	Right,
}
function assertNever(value: never): never {
	throw new Error("Illegal value: " + value);
}
function sideName(side: Side, bySwitch: boolean): string {
	if (bySwitch) {
		switch (side) {
			case Side.Left:
				return "left";
			case Side.Right:
				return "right";
		}
	} else {
		if (side === Side.Left) return "l";
		if (side === Side.Right) return "r";
	}
	return assertNever(side);
}
function outside(n: number): Side {
	return n as Side;
}
console.log(sideName(Side.Left, true) + " " + sideName(outside(1), true) + " " + sideName(Side.Left, false) + " " + sideName(outside(1), false));
