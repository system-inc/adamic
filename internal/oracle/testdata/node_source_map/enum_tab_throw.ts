process.removeAllListeners('uncaughtException');
process.on('uncaughtException', error => console.error(String(error.stack).split('\n')[1].trim()));
enum Direction { North, South }
function sourceMapThrow(): never {
	throw new Error('source map position');
}
sourceMapThrow();
