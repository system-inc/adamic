interface TextRange { pos: number; end: number }
function move(range: TextRange, pos: number, end: number): void { range.pos = pos; range.end = end; }
const range: { readonly pos: 0; readonly end: 1 } = { pos: 0, end: 1 };
move(range, 0, 1);
console.log(range.pos.toString() + ' ' + range.end.toString());
