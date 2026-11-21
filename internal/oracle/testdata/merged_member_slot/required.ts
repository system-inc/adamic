class Merged { a = 1; }
interface Merged { g: string; }
const m = new Merged();
console.log(String(m.g));
