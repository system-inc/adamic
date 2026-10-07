// swiftui.swift: SwiftUI as Objective-C classes, so Adamic reaches it through the same bridge as
// AppKit (docs/apple.md). SwiftUI is Swift only: its views are generic value types no C can name.
// Here each view an Adamic program makes is an AdamicSwiftUIView holding an AnyView, made by class
// methods (text:, button:action:, verticalStack:children:) and changed by modifiers (padding:,
// font:), and AdamicSwiftUIHost shows one in an NSView and swaps it when the program renders again.
//
// This is the first, hand-written cut of what the generator will write per API an app uses
// (#7xv3pcs); it's compiled with swiftc and linked only into programs that use it.

import AppKit
import SwiftUI

@objc(AdamicSwiftUIView) public final class AdamicSwiftUIView: NSObject {
	let view: AnyView

	init(_ view: AnyView) {
		self.view = view
	}

	@objc public static func text(_ text: String) -> AdamicSwiftUIView {
		AdamicSwiftUIView(AnyView(Text(verbatim: text)))
	}

	// The action is an Adamic closure as a block; a press runs it on the main thread, at once. The
	// last button made with each title is kept, for AdamicSwiftUIHost.press.
	static var actions: [String: () -> Void] = [:]

	// When the program finishes, the actions kept are let go of (apple.c's drain), so the closures
	// they hold come back to Adamic and are freed before its counts are taken.
	@objc public static func forgetActions() {
		actions = [:]
	}

	@objc public static func button(_ title: String, action: @escaping @convention(block) () -> Void) -> AdamicSwiftUIView {
		actions[title] = action
		return AdamicSwiftUIView(AnyView(Button(title, action: action)))
	}

	@objc public static func verticalStack(_ spacing: Double, children: [AdamicSwiftUIView]) -> AdamicSwiftUIView {
		AdamicSwiftUIView(AnyView(VStack(spacing: CGFloat(spacing)) {
			ForEach(children.indices, id: \.self) { index in children[index].view }
		}))
	}

	@objc public static func horizontalStack(_ spacing: Double, children: [AdamicSwiftUIView]) -> AdamicSwiftUIView {
		AdamicSwiftUIView(AnyView(HStack(spacing: CGFloat(spacing)) {
			ForEach(children.indices, id: \.self) { index in children[index].view }
		}))
	}

	@objc public static func spacer() -> AdamicSwiftUIView {
		AdamicSwiftUIView(AnyView(Spacer()))
	}

	@objc public func padding(_ amount: Double) -> AdamicSwiftUIView {
		AdamicSwiftUIView(AnyView(view.padding(CGFloat(amount))))
	}

	// The font's name is one of Font's text styles, as the binding's enumeration spells them.
	@objc public func font(_ style: Int) -> AdamicSwiftUIView {
		let styles: [Font] = [.largeTitle, .title, .headline, .body, .caption, .caption2]
		return AdamicSwiftUIView(AnyView(view.font(styles[max(0, min(style, styles.count - 1))])))
	}
}

@objc(AdamicSwiftUIHost) public final class AdamicSwiftUIHost: NSObject {
	let hosting: NSHostingView<AnyView>

	@objc public init(root: AdamicSwiftUIView) {
		hosting = NSHostingView(rootView: root.view)
	}

	@objc public var view: NSView {
		hosting
	}

	@objc public func render(_ root: AdamicSwiftUIView) {
		hosting.rootView = root.view
	}

	// The size the view wants, laid out now: what a test reads to see what was rendered.
	@objc public var fittingWidth: Double {
		hosting.layoutSubtreeIfNeeded()
		return Double(hosting.fittingSize.width)
	}

	@objc public var fittingHeight: Double {
		hosting.layoutSubtreeIfNeeded()
		return Double(hosting.fittingSize.height)
	}

	// Runs the action of the last button made with this title, the block its Button holds, as a press
	// would: how a test presses a button in process. SwiftUI's own accessibility tree is empty until
	// an assistive client connects, so a test can't reach the button through it. Whether there was
	// one comes back.
	@objc public func press(_ title: String) -> Bool {
		guard let action = AdamicSwiftUIView.actions[title] else { return false }
		action()
		return true
	}
}
