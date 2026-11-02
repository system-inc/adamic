// global.gc called where Node has none. The checker lets the call through only past a non-null
// assertion, which .a refuses, so this is TypeScript. Node evaluates the argument, then throws a
// TypeError the catch sees, and the program goes on.
import 'node:process';

let marks = 0;

function mark(): boolean {
	marks++;
	return marks > 0;
}

try {
	global.gc!(mark());
	console.log('collected');
} catch (error) {
	if (error instanceof Error) {
		console.log(error.name + ': ' + error.message);
	}
}
console.log(String(marks));

try {
	(global.gc)!();
} catch (error) {
	if (error instanceof Error) {
		console.log(error.name + ': ' + error.message);
	}
}
console.log('after');
