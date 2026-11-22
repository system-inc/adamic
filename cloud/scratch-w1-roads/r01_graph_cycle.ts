class Ring {
	next: Ring | undefined = undefined;
	marks = new WeakSet<object>();
}
const a = new Ring();
const b = new Ring();
a.next = b;
b.next = a;
a.marks.add(b);
b.marks.add(a);
console.log(`${a.marks.has(b)} ${b.marks.has(a)} ${a.marks.has(a)}`);
