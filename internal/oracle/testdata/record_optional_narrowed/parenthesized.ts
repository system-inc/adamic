// #70kdt66: a parenthesized narrowed record value a call removed, read through ?. in a template and an if (Emit's rows).
const r: Record<string, number[]> = {};
function clear(): void { delete r["k"]; }
r["k"] = [1];
clear();
try {
	console.log(`${(r["k"])?.length}`);
	if (r["k"]?.includes(1)) { console.log("yes"); } else { console.log("no"); }
} catch (error) { console.log("caught"); }
console.log("after");
