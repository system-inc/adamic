interface Options { outFile?: string; [name: string]: string | number | undefined; }
function key(): string { return 'outFile'; }
const name: string = key();
const options: Options = {};
const source: { x: string | number | undefined } = { x: 3 };
({ x: options[name] } = source);
console.log(typeof options.outFile);
console.log(String(options.outFile));
// The day an ElementAccessExpression target in a destructuring assignment lowers, this program and
// r13p, its plain write, must reach one verdict through l.slotType, and a mutant skipping the
// target's slot judgment must turn this one red (Adamic's condition on #9cwq1ts).
