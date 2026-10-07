// AppKit views: a seed binding, until the generator writes these (#qxe07rq).

declare module 'apple/appkit/view' {
	import type { Rectangle } from 'apple/foundation/geometry';

	/**
	 * NSView
	 * @objc class NSView
	 */
	export class View {
		/**
		 * -[NSView initWithFrame:]
		 * @objc init initWithFrame: 0:rectangle
		 */
		constructor(frame: Rectangle);

		/**
		 * -[NSView addSubview:]
		 * @objc method addSubview: 0:object -> void
		 */
		addSubview(view: View): void;

		/**
		 * -[NSView setFrame:]
		 * @objc method setFrame: 0:rectangle -> void
		 */
		setFrame(frame: Rectangle): void;

		/**
		 * -[NSView isHidden]
		 * @objc get isHidden -> boolean
		 */
		readonly isHidden: boolean;
	}
}
