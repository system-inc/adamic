// An addition to Foundation's generated data module, merged into its Data: a convenience Apple
// spells as two messages to another class.

declare module 'apple/foundation/data' {
	export interface Data {
		/**
		 * [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding]: the bytes as UTF-8
		 * text, or undefined where they aren't UTF-8.
		 * @objc alloc NSString initWithData:encoding: this:object const(4):unsigned -> string?
		 */
		utf8Text(): string | undefined;
	}
}
