interface Maybe { test?: (text: string) => boolean }
const box: { v: object } = { v: /a/ };
const m = box.v as Maybe;
const f = m.test; console.log(f === undefined ? 'missing' : 'function');
