// AppKit's windows: a seed binding, until the generator writes these (#qxe07rq).

declare module 'apple/appkit/window' {
	import type { Rectangle } from 'apple/foundation/geometry';
	import type { View } from 'apple/appkit/view';

	/** NSWindowStyleMask */
	export type WindowStyle = 'Borderless' | 'Titled' | 'Closable' | 'Miniaturizable' | 'Resizable' | 'FullSizeContentView';

	/** NSBackingStoreType */
	export type BackingStore = 'Retained' | 'Nonretained' | 'Buffered';

	/**
	 * NSWindow
	 * @objc class NSWindow
	 */
	export class Window {
		/**
		 * -[NSWindow initWithContentRect:styleMask:backing:defer:]
		 * @objc init initWithContentRect:styleMask:backing:defer: 0:rectangle 1.styleMask:options(Borderless=0,Titled=1,Closable=2,Miniaturizable=4,Resizable=8,FullSizeContentView=32768) 1.backing:enum(Retained=0,Nonretained=1,Buffered=2) 1.defer?:boolean=no
		 */
		constructor(contentRectangle: Rectangle, options: { readonly styleMask: readonly WindowStyle[]; readonly backing: BackingStore; readonly defer?: boolean });

		/**
		 * -[NSWindow title], -[NSWindow setTitle:]
		 * @objc get title -> string
		 * @objc set setTitle: string
		 */
		title: string;

		/**
		 * -[NSWindow contentView], -[NSWindow setContentView:]
		 * @objc get contentView -> object?
		 * @objc set setContentView: object
		 */
		contentView: View | undefined;

		/**
		 * -[NSWindow isVisible]
		 * @objc get isVisible -> boolean
		 */
		readonly isVisible: boolean;

		/**
		 * -[NSWindow center]
		 * @objc method center -> void
		 */
		center(): void;

		/**
		 * -[NSWindow makeKeyAndOrderFront:]
		 * @objc method makeKeyAndOrderFront: nil:object? -> void
		 */
		makeKeyAndOrderFront(): void;

		/**
		 * -[NSWindow orderOut:]
		 * @objc method orderOut: nil:object? -> void
		 */
		orderOut(): void;

		/**
		 * -[NSWindow close]
		 * @objc method close -> void
		 */
		close(): void;
	}
}
