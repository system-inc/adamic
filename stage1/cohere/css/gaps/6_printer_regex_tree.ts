// Recursive tree plus a constructed RegExp, the width printer's composition shape.
import { parse } from '../../mediaquery/index.ts';
const pattern = new RegExp('x', 'g');
console.log(pattern.test('x') ? '2' : '0');
console.log(parse('screen').kind);
