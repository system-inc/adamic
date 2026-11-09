// A record literal handed to an immediately invoked arrow (#ny0c3gb, Language's forms): its parameter
// holds it as `{}`, and the body sees that as an all-optional Read. Node prints text, a string in a
// number field.
interface Held { a: string; [key: string]: string }
interface Read { a?: number }
const m: Held = { a: "text" };
m["b"] = "more";
((held: {}): void => {
	const r: Read = held;
	console.log(String(r.a));
})(m);
