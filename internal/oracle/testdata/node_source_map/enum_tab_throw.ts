import process from 'node:process';
process.removeAllListeners('uncaughtException');
function printTopFrame(error: Error): void {
	console.error(String(error.stack).split('\n')[1]?.trim());
}
process.on('uncaughtException', printTopFrame);
enum Direction { North, South }
function sourceMapThrow(): never {
	throw new Error('source map position');
}
sourceMapThrow();
