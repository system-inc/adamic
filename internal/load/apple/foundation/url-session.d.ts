// Foundation's URL sessions: a seed binding, until the generator writes these (#qxe07rq).

declare module 'apple/foundation/url-session' {
	import type { Data } from 'apple/foundation/data';
	import type { FoundationError } from 'apple/foundation/error';
	import type { Url } from 'apple/foundation/url';
	import type { UrlResponse } from 'apple/foundation/url-response';

	/**
	 * NSURLSessionDataTask
	 * @objc class NSURLSessionDataTask
	 */
	export class UrlSessionDataTask {
		private constructor();

		/**
		 * -[NSURLSessionTask resume]
		 * @objc method resume -> void
		 */
		resume(): void;
	}

	/**
	 * NSURLSession
	 * @objc class NSURLSession
	 */
	export class UrlSession {
		private constructor();

		/**
		 * +[NSURLSession sharedSession]
		 * @objc get sharedSession -> object
		 */
		static readonly shared: UrlSession;

		/**
		 * -[NSURLSession dataTaskWithURL:completionHandler:]: the completion runs on the main thread.
		 * @objc method dataTaskWithURL:completionHandler: 0:object 1:block(object?,object?,object?) -> object
		 */
		dataTask(url: Url, completion: (data: Data | undefined, response: UrlResponse | undefined, error: FoundationError | undefined) => void): UrlSessionDataTask;
	}
}
