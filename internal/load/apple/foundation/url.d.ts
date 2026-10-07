// Foundation's URLs: a seed binding, until the generator writes these (#qxe07rq).

declare module 'apple/foundation/url' {
	/**
	 * NSURL
	 * @objc class NSURL
	 */
	export class Url {
		/**
		 * -[NSURL initWithString:], which gives nil for text that isn't a URL: then it panics.
		 * @objc init initWithString: 0:string
		 */
		constructor(text: string);

		/**
		 * -[NSURL absoluteString]
		 * @objc get absoluteString -> string?
		 */
		readonly absoluteString: string | undefined;
	}
}
