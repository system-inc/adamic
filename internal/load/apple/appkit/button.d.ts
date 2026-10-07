// AppKit's buttons: a seed binding, until the generator writes these (#qxe07rq).

declare module 'apple/appkit/button' {
	import { View } from 'apple/appkit/view';

	/**
	 * NSButton
	 * @objc class NSButton
	 */
	export class Button extends View {
		/**
		 * +[NSButton buttonWithTitle:target:action:]
		 * @objc static buttonWithTitle:target:action: 0:string 1:action -> object
		 */
		constructor(title: string, action: () => void);

		/**
		 * -[NSButton title], -[NSButton setTitle:]
		 * @objc get title -> string
		 * @objc set setTitle: string
		 */
		title: string;

		/**
		 * -[NSButton performClick:]
		 * @objc method performClick: nil:object? -> void
		 */
		performClick(): void;
	}
}
