const show = (a?: unknown): string => typeof a + ":" + String(a);
const narrow: (a: number) => string = show;
console.log(narrow(2));
