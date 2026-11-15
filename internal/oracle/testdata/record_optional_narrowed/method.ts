// #70kdt66: a user method called through ?. on a narrowed record value a call removed (Compiler's road 3).
class Named {
	name: string;
	constructor(name: string) {
		this.name = name;
	}
	describe(): string {
		return "named " + this.name;
	}
}
const r: Record<string, Named> = {};
function clear(): void { delete r["k"]; }
r["k"] = new Named("a");
clear();
try { console.log(String(r["k"]?.describe())); } catch (error) { console.log("caught"); }
r["k"] = new Named("b");
console.log(String(r["k"]?.describe()));
