interface Maybe { test?: (text: string) => boolean }
const box: { v: object } = { v: /a/ };
const m = <Maybe>box.v;
console.log(String(m.test === undefined));
