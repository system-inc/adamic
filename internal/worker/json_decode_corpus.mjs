import { writeFileSync } from 'node:fs';
let seed = 0x4d187a39;
const random = () => seed = (Math.imul(seed, 1664525) + 1013904223) >>> 0;
const numbers = ['-0', '1e400', '-1e400', '5e-324', '9007199254740993', '0', '12.5', '1E+2'];
const strings = ['""', '"\\uD800"', '"\\uDC00"', '"\\uD83C\\uDF0D"', '"🌍"', '"a\\u0000b"', '"\\uD800\\n"'];
const pick = values => values[random() % values.length];
const malformed = ['', ' ', '[1,]', '{"x":1,}', '01', '-01', 'nullx', 'tru', 'fals', 'nul', '-', '1.', '1e+', '"\\x"', '"\\u12xz"', '"\\u000"', '"a\nb"', '{x:1}', '{"x" 1}', '\u00a01', '\ufeff1', '\r\n {"🌍":0 ?}', '"🌍" x'];
const texts = [];
for (let which = 0; which < 4; which++) {
  const fixed = ['{}', '[]', 'null', 'true', '1e400', '"\\ud800"', '[[]]', '[[1,-0,1e400]]',
    '{"required":1,"required":2,"enabled":false}', '{"required":2,"enabled":true,"extra":1}', '{"first":"old","first":"new","required":2,"enabled":false}',
    '{"kind":"Text","text":"\\ud800"}', '{"kind":"Count","count":-0}', '{"kind":"Count","count":"bad"}',
    '{"inner":{"name":"old","name":"new","score":1e400},"rows":[]}',
    ...[127,128,129].map(depth => '['.repeat(depth) + '0' + ']'.repeat(depth)),
    ...[127,128,129].map(depth => '{"required":0,"drop":' + '['.repeat(depth) + '0' + ']'.repeat(depth) + '}'),
    ...malformed];
  for (let index = 0; index < 10000; index++) {
    if (index < fixed.length) { texts.push(fixed[index]); continue; }
    const n = pick(numbers), s = pick(strings);
    let text;
    if (which === 0) text = `{"inner":{"name":${s},"score":${n},"flag":false},"rows":[{"name":${pick(strings)},"score":${pick(numbers)},"flag":true}],"extra":{"drop":[null,true]},"note":${pick(strings)}}`;
    else if (which === 1) text = `[[${n},${pick(numbers)}],[],[${pick(numbers)}]]`;
    else if (which === 2) text = random()%2 ? `{"kind":"Text","text":${s},"note":${pick(strings)},"extra":false}` : `{"kind":"Count","count":${n},"extra":[]}`;
    else text = `{"first":${s},"required":${n},"words":[${pick(strings)},${pick(strings)}],"enabled":false,"extra":{}}`;
    switch (random() % 8) {
      case 0: text = pick(malformed); break;
      case 1: text = text.slice(0, random() % text.length); break;
      case 2: text += ' trailing'; break;
      case 3: text = text.replace('"required":', '"ignored":').replace('"score":', '"ignored":').replace('"text":', '"ignored":'); break;
      case 4: text = text.replace('"required":', '"required":"bad","ignored":').replace('"name":', '"name":false,"ignored":'); break;
    }
    texts.push(text);
  }
}
writeFileSync(process.argv[2], JSON.stringify(texts));
