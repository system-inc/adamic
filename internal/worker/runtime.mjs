// Generated Workers runtime. Its UTF-8 functions are the Worker's own; the Node oracle keeps an independent copy.
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
// The JavaScript backend imports decodeJson. Its Workers implementation is a separate unit, so until it
// lands a call panics, loudly, rather than decoding with code the oracle also runs.
export function decodeJson() { panic('decodeJson: not yet available on Workers'); }
// Keep the backend's encoder import linkable until the Workers JSON runtime unit lands.
export function encodeJson() { panic('encodeJson: not yet available on Workers'); }
