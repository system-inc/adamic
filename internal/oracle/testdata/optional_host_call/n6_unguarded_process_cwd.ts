import 'node:process';

function check(host: NodeJS.Process | undefined): string {
	return host?.cwd() === undefined ? 'undefined' : 'present';
}
console.log(check(undefined));
