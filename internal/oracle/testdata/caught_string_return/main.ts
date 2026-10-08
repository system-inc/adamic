function message(value: unknown): string { try { throw value; } catch (e) { return e.message; } }
console.log(message({}));
console.log('after');
