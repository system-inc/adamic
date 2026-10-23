// crossing.swift: crossing.a's loops in Swift, the comparison (bench/apple/run.go).

import AppKit

let arguments = CommandLine.arguments
let which = arguments.count > 1 ? arguments[1] : "scalar"
let times = arguments.count > 2 ? Int(arguments[2]) ?? 0 : 0

let window = NSWindow(contentRect: NSRect(x: 0, y: 0, width: 100, height: 100), styleMask: [.titled], backing: .buffered, defer: true)
window.isReleasedWhenClosed = false
window.contentView = NSView(frame: NSRect(x: 0, y: 0, width: 100, height: 100))
let view = NSView(frame: NSRect(x: 0, y: 0, width: 10, height: 10))

var seen = 0
if which == "scalar" {
	for _ in 0..<times {
		if view.isHidden { seen += 1 }
	}
} else if which == "object" {
	for _ in 0..<times {
		if window.contentView != nil { seen += 1 }
	}
} else {
	for _ in 0..<times {
		window.title = "Adamic"
		seen += window.title.utf16.count
	}
}
print("\(which) \(times): \(seen)")
