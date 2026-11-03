// A getter named after an Object.prototype member on one class, and a data field of that name on
// another, read through their union: the class with the field has no accessor of that name, so its
// read takes the field on both backends, as Node does (#76y1mth). Node prints before, getter, field.
class Reading {
	get valueOf(): string {
		return 'getter';
	}
}
class Stored {
	valueOf = 'field';
}
function show(value: Reading | Stored): string {
	return value.valueOf;
}
console.log('before');
console.log(show(new Reading()));
console.log(show(new Stored()));
