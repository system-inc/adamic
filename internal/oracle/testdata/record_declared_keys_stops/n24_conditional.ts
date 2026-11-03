interface Opts { name: string; [key: string]: string | number; }
type Dict = { [key: string]: string | number };
function pick(): string { return "name"; }
const key: string = pick();
const opts: Opts = { name: "x" };
const other: Dict = {};
const wide: Dict = key.length > 0 ? opts : other;
wide[key] = 5;
console.log(typeof opts.name);
