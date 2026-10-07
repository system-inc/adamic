import { programArguments } from 'adamic';
const count = programArguments().length;
try {
	console.log('x'.repeat(count));
} catch (error) {
	console.log(error instanceof Error ? error.message : '?');
}
