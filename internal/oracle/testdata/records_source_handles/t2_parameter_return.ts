// Language's t2 on #ny0c3gb, its held value a record literal: Held through a parameter and a return typed
// object, then object to an all-optional Read. Node prints text, a string in a number field.
interface Held { a: string; [key: string]: string }
interface Read { a?: number }
function keep(x: object): object {
	return x;
}
const m: Held = { a: "text" };
m["b"] = "more";
const r: Read = keep(m);
console.log(String(r.a));
