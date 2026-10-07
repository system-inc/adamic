const inside = /\B/u.exec('1\u{1F30D}aac');
console.log(inside === null ? 'null' : `${inside.index}`);
const lookahead = /(?!\W)/u.exec('\u{1F30D}');
console.log(lookahead === null ? 'null' : `${lookahead.index}`);
console.log(`${/[\q{e}]/iv.test('E')}`);
console.log(`${new RegExp('(?i:x|[^a-z])').test('B')}`);
console.log(`${new RegExp('(?i:a)|\\P{Ll}', 'v').test('Σ')}`);
const empty = /[\q{ab|a|}]/iv.exec('A');
console.log(empty === null ? 'null' : `${empty[0]}|`);
