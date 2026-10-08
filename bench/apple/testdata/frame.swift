// frame.swift: frame.a's loop in Swift, the comparison (bench/apple/run.go): the same list,
// rendered and laid out again each frame.

import AppKit
import SwiftUI

struct Supplement {
	let name: String
	let dose: String
	let symbol: String
	let morning: Bool
}

let supplements = [
	Supplement(name: "Creatine", dose: "5 g", symbol: "bolt.fill", morning: true),
	Supplement(name: "Vitamin D", dose: "2000 IU", symbol: "sun.max.fill", morning: true),
	Supplement(name: "Magnesium", dose: "400 mg", symbol: "moon.fill", morning: false),
]

struct SupplementRow: View {
	let supplement: Supplement

	var body: some View {
		HStack(spacing: 10) {
			Image(systemName: supplement.symbol).foregroundStyle(Color.accentColor)
			Text(verbatim: supplement.name)
			Spacer()
			Text(verbatim: supplement.dose).foregroundStyle(Color.secondary)
		}
	}
}

func body(_ frame: Int) -> AnyView {
	let morning = supplements.filter { $0.morning }
	let evening = supplements.filter { !$0.morning }
	return AnyView(NavigationStack {
		List {
			Section("Morning") { ForEach(morning.indices, id: \.self) { index in SupplementRow(supplement: morning[index]) } }
			Section("Evening \(frame)") { ForEach(evening.indices, id: \.self) { index in SupplementRow(supplement: evening[index]) } }
		}
		.navigationTitle("Stack")
	})
}

func rows(_ hosting: NSView) -> Int {
	hosting.layoutSubtreeIfNeeded()
	func tables(_ view: NSView) -> [NSTableView] {
		(view as? NSTableView).map { [$0] } ?? view.subviews.flatMap(tables)
	}
	return tables(hosting).reduce(0) { $0 + $1.numberOfRows }
}

let frames = CommandLine.arguments.count > 1 ? Int(CommandLine.arguments[1]) ?? 0 : 0
let hosting = NSHostingView(rootView: body(0))
let window = NSWindow(contentRect: NSRect(x: 0, y: 0, width: 360, height: 480), styleMask: [.titled], backing: .buffered, defer: true)
window.isReleasedWhenClosed = false
window.contentView = hosting
var total = rows(hosting)
if frames > 0 {
	for frame in 1...frames {
		hosting.rootView = body(frame)
		total += rows(hosting)
	}
}
print("frames \(frames): \(total)")
window.close()
