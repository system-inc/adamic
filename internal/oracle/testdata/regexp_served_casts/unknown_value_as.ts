interface Maybe { test?: (text: string) => boolean }
const u: unknown = /a/;
const m = u as Maybe;
const f = m.test; console.log(f === undefined ? 'missing' : 'function');
