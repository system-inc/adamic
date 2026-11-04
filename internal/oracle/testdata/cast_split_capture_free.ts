// The .ts twin of cast_split_capture_free.a: the same capture-free split casts are proven in .ts too.
function show(pieces: readonly string[]): string {
	let shown = '';
	for (const piece of pieces) {
		shown += `[${piece.toUpperCase()} ${piece.length}]`;
	}
	return `${pieces.length}${shown}`;
}
for (const text of ['a\r\nb', 'a\nb', 'a\rb', '', 'one\n\ntwo\r\n']) {
	const lines = text.split(/\r\n?|\n/) as string[];
	console.log(`${JSON.stringify(lines)} ${show(lines)}`);
}
for (const text of ['  alpha beta\tgamma  ', 'single', '']) {
	console.log(show(text.trim().split(/\s+/) as string[]));
}
console.log(show('a(b(c'.split(/[(]/) as string[]));
console.log(show('a(b(c'.split(/\(/) as string[]));
console.log(show('xaybz'.split(/(?:a)|b/) as string[]));
console.log(show('xabcbab'.split(/(?<=a)b/) as string[]));
console.log(show('xbcab'.split(/(?<!a)b/) as string[]));
console.log(show('a b c d'.split(/\s/, 2) as string[]));
console.log(show('a;b'.split((/;/)) as string[]));
const viewed = 'p,q,r'.split(/,/) as readonly string[];
console.log(show(viewed));
