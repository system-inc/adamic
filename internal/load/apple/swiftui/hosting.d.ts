// A SwiftUI view shown in AppKit, through swiftui.swift (internal/native/apple), written by hand
// until the generator writes these (#7xv3pcs). Programs use apple/swiftui/host, written in Adamic
// over this.

declare module 'apple/swiftui/hosting' {
	import type { View } from 'apple/appkit/view';
	import type { SwiftUIView } from 'apple/swiftui/native';

	/**
	 * NSHostingView<AnyView>, and the root it shows
	 * @objc class AdamicSwiftUIHost
	 */
	export class SwiftUIHost {
		/**
		 * init(root:)
		 * @objc init initWithRoot: 0:object
		 */
		constructor(root: SwiftUIView);

		/**
		 * the hosting view, for a window's contentView
		 * @objc get view -> object
		 */
		readonly view: View;

		/**
		 * shows a new root: what a view's body gives when its state changes
		 * @objc method render: 0:object -> void
		 */
		render(root: SwiftUIView): void;

		/**
		 * the width the view wants, laid out now
		 * @objc get fittingWidth -> double
		 */
		readonly fittingWidth: number;

		/**
		 * the height the view wants, laid out now
		 * @objc get fittingHeight -> double
		 */
		readonly fittingHeight: number;

		/**
		 * the rows SwiftUI's lists show, laid out now, section headers included (a List on macOS is an
		 * NSTableView): what a test reads where a list's fitting size can't tell what was rendered
		 * @objc get listRows -> integer
		 */
		readonly listRows: number;

		/**
		 * runs the action of the last button made with this title, the block its Button holds, as a
		 * press would; true when there was one. For tests: SwiftUI's own accessibility tree is empty
		 * until an assistive client connects.
		 * @objc method press: 0:string -> boolean
		 */
		press(title: string): boolean;
	}
}
