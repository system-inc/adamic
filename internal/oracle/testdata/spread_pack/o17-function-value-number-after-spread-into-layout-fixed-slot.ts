type A = number | string;
const show = (first?: A, ...rest: A[]): void => { console.log(typeof first + ":" + String(rest.length)); };
show("a", "b");
const lead = (...rest: number[]): void => { console.log(String(rest.length)); };
lead(...[], 7);
