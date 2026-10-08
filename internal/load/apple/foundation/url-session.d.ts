// An addition to Foundation's generated session module, merged into its UrlSession: Swift's
// data(from:) as a promise to await, which the headers can't say (apple.c's adamic_apple_data makes
// the request and settles the promise as the task completes).

declare module 'apple/foundation/url-session' {
	import type { Url } from 'apple/foundation/url';

	/** What a session's data gives back: the body as UTF-8 text, and the HTTP status (0 where the response isn't HTTP's). */
	export interface TextResponse {
		readonly body: string;
		readonly status: number;
	}

	export interface UrlSession {
		/**
		 * -[NSURLSession dataTaskWithURL:completionHandler:], resumed: the body Apple fetched from the
		 * URL and its status. It rejects with Apple's description of the error, and where the body
		 * isn't UTF-8.
		 * @objc function adamic_apple_data 0.from:object -> promise
		 */
		data(options: { readonly from: Url }): Promise<TextResponse>;
	}
}
