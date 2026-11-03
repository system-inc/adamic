// A getter named after an Object.prototype member on one class, and a readonly data field of that
// name on another (readonly, as a getter-only property is), read through their union: the class with
// the field has no accessor of that name, so its read takes the field on both backends, as Node does
// (#76y1mth). Node prints before, getter, field.
class Reading {
	get valueOf(): string {
		return 'getter';
	}
}
class Stored {
	readonly valueOf: string = 'field';
}
function show(value: Reading | Stored): string {
	return value.valueOf;
}
console.log('before');
console.log(show(new Reading()));
console.log(show(new Stored()));
