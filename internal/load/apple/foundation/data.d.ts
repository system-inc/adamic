// Foundation's data: a seed binding, until the generator writes these (#qxe07rq).

declare module 'apple/foundation/data' {
	/**
	 * NSData
	 * @objc class NSData
	 */
	export class Data {
		private constructor();

		/**
		 * -[NSData length]
		 * @objc get length -> unsigned
		 */
		readonly length: number;

		/**
		 * [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding]: the bytes as UTF-8
		 * text, or undefined where they aren't UTF-8.
		 * @objc alloc NSString initWithData:encoding: this:object const(4):unsigned -> string?
		 */
		utf8Text(): string | undefined;
	}
}
