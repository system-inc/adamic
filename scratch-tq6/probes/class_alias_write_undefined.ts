// A class's boolean field written undefined through a plain alias whose field is optional, then read
// through the class. Node prints: false, undefined, undefined.
class Holder {
  valid = false;
  name = "holder";
}
type Wide = { valid?: boolean | undefined; name: string };
const holder = new Holder();
const wide: Wide = holder;
console.log(String(holder.valid));
wide.valid = undefined;
console.log(String(wide.valid));
console.log(String(holder.valid));
