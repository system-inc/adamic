// A SwiftUI view shown in AppKit, through swiftui.swift (internal/native/apple): a seed binding,
// until the generator writes these (#7xv3pcs).

declare module 'apple/swiftui/host' {
	import type { View as AppKitView } from 'apple/appkit/view';
	import type { View } from 'apple/swiftui/views';

	/**
	 * NSHostingView<AnyView>, and the root it shows
	 * @objc class AdamicSwiftUIHost
	 */
	export class Host {
		/**
		 * init(root:)
		 * @objc init initWithRoot: 0:object
		 */
		constructor(root: View);

		/**
		 * the hosting view, for a window's contentView
		 * @objc get view -> object
		 */
		readonly view: AppKitView;

		/**
		 * shows a new root: what a view's body gives when its state changes
		 * @objc method render: 0:object -> void
		 */
		render(root: View): void;

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
		 * runs the action of the last button made with this title, the block its Button holds, as a
		 * press would; true when there was one. For tests: SwiftUI's own accessibility tree is empty
		 * until an assistive client connects.
		 * @objc method press: 0:string -> boolean
		 */
		press(title: string): boolean;
	}
}
