type B = boolean | string;
const flags: boolean[] = [true, false];
const show = (...rest: B[]): void => { let text = ""; for (const value of rest) text += typeof value + ":" + String(value) + ","; console.log(text); };
const all = (...rest: boolean[]): void => { console.log(String(rest.length)); };
let pick: (...rest: boolean[]) => void = all;
pick(...flags);
pick = show;
pick(...flags, false);
