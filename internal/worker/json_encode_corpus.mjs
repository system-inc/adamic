import { writeFileSync } from 'node:fs';
let seed = 0x813efcb2;
const random = () => seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0;
const strings = ['', '\ud800', '\udc00', '\ud800\ud800\udc00\udc00', '🌍', 'a\0b\n"\\', '\u2028', 'héllo'];
const numbers = ['-0','1e400','-1e400','5e-324','9007199254740993','0.30000000000000004','1e21','-1e-7'];
const pick = values => values[random()%values.length];
const s = () => JSON.stringify(pick(strings));
const texts = [], expected = [];
for (let which = 0; which < 4; which++) for (let index = 0; index < 10000; index++) {
	let text;
	if (which === 0) {
		const items = [];
		for (let n = random()%8; n > 0; n--) items.push(`{"name":${s()},"score":${pick(numbers)}${index%3 ? ',"note":'+s() : ''}}`);
		text = `{"items":[${items.join(',')}],"enabled":${random()%2 === 0}${index%3 ? ',"title":'+s() : ''}}`;
	} else if (which === 1) {
		const values = []; for (let n = index === 99 ? 10000 : random()%20; n > 0; n--) values.push(pick(numbers));
		text = '['+values.join(',')+']';
	} else if (which === 2) text = index%2 ? `{"kind":"Text","text":${s()}${index%3 ? ',"note":'+s() : ''}}` : `{"kind":"Count","count":${pick(numbers)}}`;
	else text = `[${s()},${pick(numbers)},${random()%2 === 0}]`;
	texts.push(text);
	// All objects were built in declared order. This witness walks no schema.
	expected.push(JSON.stringify(JSON.parse(text)));
}
writeFileSync(process.argv[2], JSON.stringify(texts));
writeFileSync(process.argv[3], expected.join('\n')+'\n');
