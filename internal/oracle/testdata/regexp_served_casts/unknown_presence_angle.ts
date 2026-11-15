interface Maybe { test?: (text: string) => boolean }
const u: unknown = /a/;
const m = <Maybe>u;
console.log(String(m.test === undefined));
