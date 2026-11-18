// supplements.swift: supplements.a's witness, the same app written in Swift: SwiftUI, URLSession,
// JSONDecoder and Codable, each completion hopping to the main thread as the Adamic bridge does
// (witness_test.go). Nothing here is Adamic.

import AppKit
import SwiftUI

enum Timing: String, Codable {
	case morning = "Morning"
	case evening = "Evening"
}

struct Supplement: Codable {
	let name: String
	let dose: String
	let symbol: String
	let timing: Timing
}

final class Supplements: ObservableObject {
	@Published var value: [Supplement] = []
}

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

struct SupplementsView: View {
	@ObservedObject var supplements: Supplements

	func rows(_ timing: Timing) -> [Supplement] {
		supplements.value.filter { $0.timing == timing }
	}

	var body: some View {
		if supplements.value.isEmpty {
			NavigationStack { ProgressView().navigationTitle("Stack") }
		} else {
			NavigationStack {
				List {
					Section("Morning") { ForEach(rows(.morning).indices, id: \.self) { index in SupplementRow(supplement: rows(.morning)[index]) } }
					Section("Evening") { ForEach(rows(.evening).indices, id: \.self) { index in SupplementRow(supplement: rows(.evening)[index]) } }
				}
				.listStyle(.inset)
				.navigationTitle("Stack")
			}
		}
	}
}

let supplements = Supplements()
let hosting = NSHostingView(rootView: SupplementsView(supplements: supplements))
let window = NSWindow(contentRect: NSRect(x: 0, y: 0, width: 360, height: 480), styleMask: [.titled], backing: .buffered, defer: true)
window.isReleasedWhenClosed = false
window.contentView = hosting

// The rows the list shows, section headers included: a List on macOS is an NSTableView.
func rows() -> Int {
	hosting.layoutSubtreeIfNeeded()
	func tables(_ view: NSView) -> [NSTableView] {
		(view as? NSTableView).map { [$0] } ?? view.subviews.flatMap(tables)
	}
	return tables(hosting).reduce(0) { $0 + $1.numberOfRows }
}

print("loading: rows \(rows())")
let address = CommandLine.arguments.count > 1 ? CommandLine.arguments[1] : "http://127.0.0.1:1"
let main = CFRunLoopGetMain()
URLSession.shared.dataTask(with: URL(string: "\(address)/supplements.json")!) { data, _, error in
	DispatchQueue.main.async {
		if error == nil, let data, let decoded = try? JSONDecoder().decode([Supplement].self, from: data) {
			supplements.value = decoded
			print("supplements: \(supplements.value.count), rows \(rows())")
			for supplement in supplements.value {
				print("  \(supplement.timing.rawValue): \(supplement.name), \(supplement.dose)")
			}
		} else {
			print("supplements: failed")
		}
		URLSession.shared.dataTask(with: URL(string: "\(address)/drifted.json")!) { drifted, _, _ in
			DispatchQueue.main.async {
				let again = drifted.flatMap { try? JSONDecoder().decode([Supplement].self, from: $0) }
				print("drifted: \(again == nil ? "refused" : "accepted"), rows \(rows())")
				CFRunLoopStop(main)
			}
		}.resume()
	}
}.resume()
CFRunLoopRun()
window.close()
