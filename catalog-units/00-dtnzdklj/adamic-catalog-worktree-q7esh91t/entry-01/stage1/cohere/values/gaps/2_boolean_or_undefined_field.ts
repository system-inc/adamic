// Gap 2: a field of type boolean | undefined is refused, in a class or an object type; number,
// string and array fields that may be undefined lower.
class Holder {
	flag: boolean | undefined = undefined;
}
const holder = new Holder();
holder.flag = true;
console.log(`${holder.flag === true}`);
