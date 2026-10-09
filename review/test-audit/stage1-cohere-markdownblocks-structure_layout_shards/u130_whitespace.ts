import { DocumentArena } from '../document.ts';
import { printWhitespace } from '../whitespace.ts';
const a = new DocumentArena();
const token = { present: true, type: 'word', value: 'x', kind: 'non-cjk', cj: false, leading: false, trailing: false };
console.log(a.node(printWhitespace(a, { value: ' ', proseWrap: 'always', link: false, previous: token, next: token, afterNext: token, ancestorKinds: [], ancestorSetext: [], samples: [] })).kind);
