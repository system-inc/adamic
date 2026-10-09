package javascript

// Preserve the inserted key and its call ABI while comparing module generic
// closures by declaration. Ordinary keys retain the host collection fast path.
const genericIdentityRuntime = `
const adamicGeneric = value => value instanceof AdamicClosure && value.identity !== 0;
const adamicEqual = (left, right) => left === right || (adamicGeneric(left) && adamicGeneric(right) && left.identity === right.identity);
const adamicSearch = (array, method, ...args) => { const value = args[0]; return (adamicGeneric(value) ? array.map(item => adamicEqual(item, value) ? value : item) : array)[method](...args); };
class AdamicMap extends Map {
 key(value) { if (adamicGeneric(value)) for (const key of super.keys()) if (adamicEqual(key, value)) return key; return value; }
 has(value) { return super.has(this.key(value)); }
 get(value) { return super.get(this.key(value)); }
 set(key, value) { return super.set(this.key(key), value); }
 delete(value) { return super.delete(this.key(value)); }
}
class AdamicSet extends Set {
 key(value) { if (adamicGeneric(value)) for (const key of super.values()) if (adamicEqual(key, value)) return key; return value; }
 has(value) { return super.has(this.key(value)); }
 add(value) { return super.add(this.key(value)); }
 delete(value) { return super.delete(this.key(value)); }
}
`
