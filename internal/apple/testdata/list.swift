// list.swift: list.a's witness, the same supplements list written as plain SwiftUI, rendered and
// read the same way (witness_test.go). Nothing here is Adamic.

import AppKit
import SwiftUI

struct Supplement {
	let name: String
	let dose: String
	let symbol: String
}

var morning = [
	Supplement(name: "Creatine", dose: "5 g", symbol: "bolt.fill"),
	Supplement(name: "Vitamin D", dose: "2000 IU", symbol: "sun.max.fill"),
]
var evening = [Supplement(name: "Magnesium", dose: "400 mg", symbol: "moon.fill")]

func row(_ supplement: Supplement) -> some View {
	HStack(spacing: 8) {
		Image(systemName: supplement.symbol).foregroundStyle(Color.accentColor)
		Text(verbatim: supplement.name)
		Spacer()
		Text(verbatim: supplement.dose).foregroundStyle(Color.secondary)
	}
}

func body() -> AnyView {
	AnyView(NavigationStack {
		VStack(spacing: 0) {
			List {
				Section("Morning") { ForEach(morning.indices, id: \.self) { index in row(morning[index]) } }
				Section("Evening") { ForEach(evening.indices, id: \.self) { index in row(evening[index]) } }
			}
			.listStyle(.inset)
			Text(verbatim: "\(morning.count + evening.count) supplements").font(.caption).multilineTextAlignment(.center).padding(8)
		}
		.navigationTitle("Stack")
	})
}

func number(_ value: CGFloat) -> String {
	let double = Double(value)
	return double == double.rounded() ? String(Int(double)) : String(double)
}

let hosting = NSHostingView(rootView: body())
let window = NSWindow(contentRect: NSRect(x: 0, y: 0, width: 360, height: 480), styleMask: [.titled], backing: .buffered, defer: true)
window.isReleasedWhenClosed = false
window.contentView = hosting

func size() -> String {
	hosting.layoutSubtreeIfNeeded()
	return "\(number(hosting.fittingSize.width)) by \(number(hosting.fittingSize.height))"
}

// The rows the list shows, section headers included: a List on macOS is an NSTableView.
func rows() -> Int {
	hosting.layoutSubtreeIfNeeded()
	func tables(_ view: NSView) -> [NSTableView] {
		(view as? NSTableView).map { [$0] } ?? view.subviews.flatMap(tables)
	}
	return tables(hosting).reduce(0) { $0 + $1.numberOfRows }
}

print("rows: \(rows()), size: \(size())")
evening.append(Supplement(name: "Glycine", dose: "3 g", symbol: "leaf.fill"))
hosting.rootView = body()
print("supplements: \(morning.count + evening.count), rows: \(rows())")

// One row on its own: its size is its symbol's, name's and dose's.
let alone = NSHostingView(rootView: row(morning[1]))
alone.layoutSubtreeIfNeeded()
print("row: \(number(alone.fittingSize.width)) by \(number(alone.fittingSize.height))")
window.close()
