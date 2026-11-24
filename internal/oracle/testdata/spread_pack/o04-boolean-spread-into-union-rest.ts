type A = boolean | string;
const flags: boolean[] = [true, false];
function show(...rest: A[]): void { for (const value of rest) console.log(typeof value + ":" + String(value)); }
show(...flags);
