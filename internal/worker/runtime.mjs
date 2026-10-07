// Generated Workers runtime. The UTF-8 implementation is shared with the Node oracle.
export class AdamicPanic extends Error {
	constructor(message) {
		super(message);
		this.name = 'AdamicPanic';
	}
}
export function panic(message) {
	throw new AdamicPanic(message);
}
/* UTF8_RUNTIME */
export const { utf8Length, utf8At } = createUtf8(panic);
export function readTextFile() { panic('readTextFile: not available on Workers'); }
export function writeTextFile() { panic('writeTextFile: not available on Workers'); }
export function readDirectory() { panic('readDirectory: not available on Workers'); }
export function fileStatus() { panic('fileStatus: not available on Workers'); }
export function programArguments() { panic('programArguments: not available on Workers'); }
