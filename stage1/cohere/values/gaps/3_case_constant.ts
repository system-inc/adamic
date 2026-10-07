// Gap 3: a case that names a constant is refused, for a number or a string; a literal lowers.
const newline = 0x0a;
function describe(code: number): string {
	switch (code) {
		case newline:
			return 'newline';
	}
	return 'other';
}
console.log(`${describe(10)} ${describe(32)}`);
