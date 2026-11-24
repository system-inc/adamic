type A = number | string;
const strs: string[] = ["a", "b"];
const show = (...rest: A[]): void => { for (const value of rest) console.log(typeof value + ":" + String(value)); };
show(...strs);
