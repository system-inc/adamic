// AppKit's text fields: a seed binding, until the generator writes these (#qxe07rq).

declare module 'apple/appkit/text-field' {
	import { View } from 'apple/appkit/view';

	/**
	 * NSTextField
	 * @objc class NSTextField
	 */
	export class TextField extends View {
		private constructor();

		/**
		 * +[NSTextField labelWithString:]
		 * @objc static labelWithString: 0:string -> object
		 */
		static label(text: string): TextField;

		/**
		 * -[NSTextField stringValue], -[NSTextField setStringValue:]
		 * @objc get stringValue -> string
		 * @objc set setStringValue: string
		 */
		stringValue: string;
	}
}
