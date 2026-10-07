// Foundation's errors: a seed binding, until the generator writes these (#qxe07rq).

declare module 'apple/foundation/error' {
	/**
	 * NSError, named so it doesn't hide the language's own Error where both are used.
	 * @objc class NSError
	 */
	export class FoundationError {
		private constructor();

		/**
		 * -[NSError localizedDescription]
		 * @objc get localizedDescription -> string
		 */
		readonly localizedDescription: string;

		/**
		 * -[NSError code]
		 * @objc get code -> integer
		 */
		readonly code: number;
	}
}
