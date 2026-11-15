interface Maybe { test?: (text: string) => boolean }
const box: { v: object } = { v: /a/ };
const m = box.v as Maybe;
console.log(String(m.test === undefined));
