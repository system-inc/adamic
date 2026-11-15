// #70kdt66: ?. reads a class field of a narrowed record value a call removed (Compiler's road 1).
class Box {
	size: number;
	constructor(size: number) {
		this.size = size;
	}
}
const r: Record<string, Box> = {};
function clear(): void { delete r["k"]; }
r["k"] = new Box(3);
clear();
try { console.log(String(r["k"]?.size)); } catch (error) { console.log("caught"); }
r["k"] = new Box(4);
console.log(String(r["k"]?.size));
