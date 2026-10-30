interface Named { readonly name: string; }
interface Judge { isNamed(value: object): value is Named; }
const judge: Judge = { isNamed: (value: object): value is Named => "name" in value };
function show(j: Judge, value: object): void { if (j.isNamed(value)) { console.log(value.name); } else { console.log("no"); } }
show(judge, { name: "n" });
