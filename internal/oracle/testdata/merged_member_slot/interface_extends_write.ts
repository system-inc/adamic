class Base { b = 1 }
class Merged { a = 1 }
interface Merged extends Base {}
const m = new Merged();
m.b = 5;
console.log(String(m.b + 1));
