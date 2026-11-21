class Base { b = 1 }
class Merged { a = 1 }
interface Merged extends Base {}
console.log(String(new Merged().b + 1));
