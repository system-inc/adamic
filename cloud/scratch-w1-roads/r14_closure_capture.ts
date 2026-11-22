const seen = new WeakSet<object>();
const mark = (key: object): boolean => {
	seen.add(key);
	return seen.has(key);
};
const key = {};
console.log(`${mark(key)} ${seen.has(key)}`);
