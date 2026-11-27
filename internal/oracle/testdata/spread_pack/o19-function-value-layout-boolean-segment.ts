type B = boolean | string;
const flags: boolean[] = [true, false];
const show = (...rest: B[]): void => { for (const value of rest) console.log(typeof value + ":" + String(value)); };
show("a", "b");
const lead = (first: B, ...rest: boolean[]): void => { console.log(typeof first + ":" + String(first) + ":" + String(rest.length)); };
lead(true, ...flags, false);
