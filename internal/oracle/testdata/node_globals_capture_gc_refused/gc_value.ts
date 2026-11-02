// global.gc's call with its result used as a value. The checker lets the call through only past a
// non-null assertion, which .a refuses, so this is TypeScript.
import 'node:process';
try {
	const result = global.gc!();
	console.log(String(result));
} catch (error) {
	console.log('caught');
}
