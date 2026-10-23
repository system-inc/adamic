// counter.swift: counter.a's witness, the same view written as plain SwiftUI, pressed and read the
// same way (witness_test.go). Nothing here is Adamic.

import AppKit
import SwiftUI

final class Count: ObservableObject {
	@Published var value = 0
}

struct Counter: View {
	@ObservedObject var count: Count

	var body: some View {
		VStack(spacing: 12) {
			Text(verbatim: "Count: \(count.value)").font(.title)
			button("Add one") { count.value += 1 }
		}
		.padding(20)
	}
}

// The last button made with each title keeps its action here, as the Adamic shim does, and pressed
// runs it: SwiftUI's accessibility tree is empty without an assistive client.
var actions: [String: () -> Void] = [:]

func button(_ title: String, action: @escaping () -> Void) -> Button<Text> {
	actions[title] = action
	return Button(title, action: action)
}

func number(_ value: CGFloat) -> String {
	let double = Double(value)
	return double == double.rounded() ? String(Int(double)) : String(double)
}

let count = Count()
let hosting = NSHostingView(rootView: Counter(count: count))
let window = NSWindow(contentRect: NSRect(x: 0, y: 0, width: 300, height: 200), styleMask: [.titled], backing: .buffered, defer: true)
window.isReleasedWhenClosed = false
window.contentView = hosting

func size() -> String {
	hosting.layoutSubtreeIfNeeded()
	return "\(number(hosting.fittingSize.width)) by \(number(hosting.fittingSize.height))"
}

func pressed(_ title: String) -> Bool {
	guard let action = actions[title] else { return false }
	action()
	return true
}

print("size: \(size())")
print("pressed: \(pressed("Add one")), count: \(count.value)")
print("pressed: \(pressed("Add one")), count: \(count.value)")
print("missing: \(pressed("Subtract one")), count: \(count.value)")
for _ in 0..<8 {
	_ = pressed("Add one")
}
print("count: \(count.value), size: \(size())")
window.close()
