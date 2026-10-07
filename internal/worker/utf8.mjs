// Workers UTF-8 runtime. TextEncoder writes lone surrogates as U+FFFD.
// The last text is cached because byte walks ask about the same string repeatedly.
export function createUtf8(panic) {
	const encoder = new TextEncoder();
	let utf8Text = '';
	let utf8Encoded = new Uint8Array(0);
	function encoded(text) {
		if (text !== utf8Text) {
			utf8Text = text;
			utf8Encoded = encoder.encode(text);
		}
		return utf8Encoded;
	}
	function utf8Length(text) {
		return encoded(text).length;
	}
	function utf8At(text, index) {
		const bytes = encoded(text);
		if (!(index >= 0 && index < bytes.length && index === Math.trunc(index))) {
			panic(`RangeError: utf8At index ${index} is not a byte of a text of ${bytes.length} bytes`);
		}
		return bytes[index];
	}
	return { utf8Length, utf8At };
}
