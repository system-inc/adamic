// Foundation's URL responses: a seed binding, until the generator writes these (#qxe07rq).

declare module 'apple/foundation/url-response' {
	/**
	 * NSURLResponse
	 * @objc class NSURLResponse
	 */
	export class UrlResponse {
		private constructor();

		/**
		 * -[NSURLResponse MIMEType]
		 * @objc get MIMEType -> string
		 */
		readonly mimeType: string;

		/**
		 * -[NSURLResponse expectedContentLength]
		 * @objc get expectedContentLength -> integer
		 */
		readonly expectedContentLength: number;
	}
}
