import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for `cohere-swift/legacy-constructors`, both directions, covering every example SwiftLint
 documents for `legacy_constructor`, `legacy_cggeometry_functions`, `legacy_nsgeometry_functions` and
 `legacy_random`, triggering and not, and every correction example its two geometry rules list. The failing
 cases assert the exact line and column of the call and its family's message id. The near misses are the
 shapes no C function has (wrong arity, labels, trailing closures), qualified calls, bare references, the
 names in strings and comments, and a file that declares the name itself. The fix cases check the rewrite
 text, the parentheses it adds where binding would change, nested calls fixed in one pass, and the cases
 where the fix is withheld or only suggested.
 */
struct LegacyConstructorsTests {
    static func findings(_ source: String) -> [FindingRecord] {
        let url = URL(fileURLWithPath: "/fixture/Subject.swift")
        let file = ParsedFile(
            url: url,
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        return LegacyConstructors().findings(in: file)
    }

    static func positions(_ findings: [FindingRecord]) -> [String] {
        findings.map { "\($0.line):\($0.column)" }
    }

    /* Every fix applied the way the fix phase applies them, overlaps refused. */
    static func fixed(_ source: String) -> String {
        FixApplier.apply(FixApplier.proposals(from: findings(source)), to: source).text
    }

    /* Every finding's first suggestion applied, for the rewrites that are offered and never applied. */
    static func suggested(_ source: String) -> String {
        FixApplier.apply(
            findings(source).flatMap { finding in
                (finding.suggestions.first?.fixes ?? []).map { FixApplier.Proposal(rule: finding.rule, edit: $0) }
            },
            to: source,
        )
        .text
    }

    /* SwiftLint's `legacy_constructor` triggering examples, one per line. */
    @Test func incumbentConstructorExamplesAreFound() {
        let source = """
            CGPointMake(10, 10)
            CGPointMake(xVal, yVal)
            CGPointMake(calculateX(), 10)
            CGSizeMake(10, 10)
            CGSizeMake(aWidth, aHeight)
            CGRectMake(0, 0, 10, 10)
            CGRectMake(xVal, yVal, width, height)
            CGVectorMake(10, 10)
            CGVectorMake(deltaX, deltaY)
            NSMakePoint(10, 10)
            NSMakePoint(xVal, yVal)
            NSMakeSize(10, 10)
            NSMakeSize(aWidth, aHeight)
            NSMakeRect(0, 0, 10, 10)
            NSMakeRect(xVal, yVal, width, height)
            NSMakeRange(10, 1)
            NSMakeRange(loc, len)
            UIEdgeInsetsMake(0, 0, 10, 10)
            UIEdgeInsetsMake(top, left, bottom, right)
            NSEdgeInsetsMake(0, 0, 10, 10)
            NSEdgeInsetsMake(top, left, bottom, right)
            UIOffsetMake(0, 10)
            UIOffsetMake(horizontal, vertical)
            """
        let found = Self.findings(source)
        #expect(Self.positions(found) == (1...23).map { "\($0):1" })
        #expect(found.allSatisfy { $0.messageId == "legacyConstructor" })
    }

    /* SwiftLint's `legacy_constructor` non-triggering examples: the modern initializers. */
    @Test func incumbentConstructorNonTriggeringExamplesAreNot() {
        let source = """
            CGPoint(x: 10, y: 10)
            CGPoint(x: xValue, y: yValue)
            CGSize(width: 10, height: 10)
            CGSize(width: aWidth, height: aHeight)
            CGRect(x: 0, y: 0, width: 10, height: 10)
            CGRect(x: xVal, y: yVal, width: aWidth, height: aHeight)
            CGVector(dx: 10, dy: 10)
            CGVector(dx: deltaX, dy: deltaY)
            NSPoint(x: 10, y: 10)
            NSPoint(x: xValue, y: yValue)
            NSSize(width: 10, height: 10)
            NSSize(width: aWidth, height: aHeight)
            NSRect(x: 0, y: 0, width: 10, height: 10)
            NSRect(x: xVal, y: yVal, width: aWidth, height: aHeight)
            NSRange(location: 10, length: 1)
            NSRange(location: loc, length: len)
            UIEdgeInsets(top: 0, left: 0, bottom: 10, right: 10)
            UIEdgeInsets(top: aTop, left: aLeft, bottom: aBottom, right: aRight)
            NSEdgeInsets(top: 0, left: 0, bottom: 10, right: 10)
            NSEdgeInsets(top: aTop, left: aLeft, bottom: aBottom, right: aRight)
            UIOffset(horizontal: 0, vertical: 10)
            UIOffset(horizontal: horizontal, vertical: vertical)
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* SwiftLint's `legacy_cggeometry_functions` triggering examples. */
    @Test func incumbentCoreGraphicsExamplesAreFound() {
        let source = """
            CGRectGetWidth(rect)
            CGRectGetHeight(rect)
            CGRectGetMinX(rect)
            CGRectGetMidX(rect)
            CGRectGetMaxX(rect)
            CGRectGetMinY(rect)
            CGRectGetMidY(rect)
            CGRectGetMaxY(rect)
            CGRectIsNull(rect)
            CGRectIsEmpty(rect)
            CGRectIsInfinite(rect)
            CGRectStandardize(rect)
            CGRectIntegral(rect)
            CGRectInset(rect, 10, 5)
            CGRectOffset(rect, -2, 8.3)
            CGRectUnion(rect1, rect2)
            CGRectIntersection(rect1, rect2)
            CGRectContainsRect(rect1, rect2)
            CGRectContainsPoint(rect, point)
            CGRectIntersectsRect(rect1, rect2)
            """
        let found = Self.findings(source)
        #expect(Self.positions(found) == (1...20).map { "\($0):1" })
        #expect(found.allSatisfy { $0.messageId == "legacyCGGeometryFunction" })
    }

    /* SwiftLint's `legacy_cggeometry_functions` and `legacy_nsgeometry_functions` non-triggering examples, merged. */
    @Test func incumbentGeometryNonTriggeringExamplesAreNot() {
        let source = """
            rect.width
            rect.height
            rect.minX
            rect.midX
            rect.maxX
            rect.minY
            rect.midY
            rect.maxY
            rect.isNull
            rect.isEmpty
            rect.isInfinite
            rect.standardized
            rect.integral
            rect.insetBy(dx: 5.0, dy: -7.0)
            rect.offsetBy(dx: 5.0, dy: -7.0)
            rect1.union(rect2)
            rect1.intersect(rect2)
            rect1.intersection(rect2)
            rect1.contains(rect2)
            rect.contains(point)
            rect1.intersects(rect2)
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* SwiftLint's `legacy_nsgeometry_functions` triggering examples, all but `NSEdgeInsetsEqual` (see below). */
    @Test func incumbentAppKitExamplesAreFound() {
        let source = """
            NSWidth(rect)
            NSHeight(rect)
            NSMinX(rect)
            NSMidX(rect)
            NSMaxX(rect)
            NSMinY(rect)
            NSMidY(rect)
            NSMaxY(rect)
            NSEqualRects(rect1, rect2)
            NSEqualSizes(size1, size2)
            NSEqualPoints(point1, point2)
            NSIsEmptyRect(rect)
            NSIntegralRect(rect)
            NSInsetRect(rect, 10, 5)
            NSOffsetRect(rect, -2, 8.3)
            NSUnionRect(rect1, rect2)
            NSIntersectionRect(rect1, rect2)
            NSContainsRect(rect1, rect2)
            NSPointInRect(rect, point)
            NSIntersectsRect(rect1, rect2)
            """
        let found = Self.findings(source)
        #expect(Self.positions(found) == (1...20).map { "\($0):1" })
        #expect(found.allSatisfy { $0.messageId == "legacyNSGeometryFunction" })
    }

    /* SwiftLint flags it and rewrites it to `==`, which does not compile: NSEdgeInsets is not Equatable. */
    @Test func edgeInsetsEqualIsNot() {
        #expect(Self.findings("NSEdgeInsetsEqual(insets2, insets2)").isEmpty)
    }

    /* SwiftLint's `legacy_random` examples, both directions. */
    @Test func incumbentRandomExamples() {
        let triggering = """
            arc4random()
            arc4random_uniform(83)
            drand48()
            """
        let found = Self.findings(triggering)
        #expect(Self.positions(found) == ["1:1", "2:1", "3:1"])
        #expect(found.allSatisfy { $0.messageId == "legacyRandom" && $0.fixes.isEmpty && $0.suggestions.isEmpty })
        let passing = """
            Int.random(in: 0..<10)
            Double.random(in: 8.6...111.34)
            Float.random(in: 0 ..< 1)
            """
        #expect(Self.findings(passing).isEmpty)
    }

    /* Inside larger expressions, nested in each other, across lines, and after other text on the line. */
    @Test func callsInLargerShapesAreFound() {
        let source = """
            let width = CGRectGetWidth(CGRectInset(frame, 2, 2)) + margin
            view.frame = CGRectMake(
                0, 0, NSWidth(bounds), 20)
            let index = Int(arc4random_uniform(UInt32(items.count)))
            """
        #expect(Self.positions(Self.findings(source)) == ["1:13", "1:28", "2:14", "3:11", "4:17"])
    }

    /* Shapes no C function has, qualified calls, bare references, and the names as text. */
    @Test func nearMissesAreNot() {
        let source = """
            CGPointMake(1, 2, 3)
            CGPointMake(x: 1, y: 2)
            arc4random(5)
            arc4random_uniform { 5 }
            NSWidth(rect) { }
            Darwin.arc4random()
            Foundation.NSMakeRange(0, 1)
            layout.CGRectGetWidth(rect)
            CGPointMake(_:_:)(1, 2)
            let union = frames.reduce(.null, CGRectUnion)
            let text = "CGPointMake(1, 2)"
            // NSWidth(rect)
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* A file that declares the name, anywhere and in any form, is calling its own function. */
    @Test func namesTheFileDeclaresAreNot() {
        let sources = [
            "func CGPointMake(_ x: Double, _ y: Double) -> Point { Point(x, y) }\nlet origin = CGPointMake(0, 0)",
            "let a = arc4random()\nstruct Dice { func arc4random() -> Int { 4 } }",
            "func measure(NSWidth: (CGRect) -> CGFloat) -> CGFloat { NSWidth(bounds) }",
            "let drand48 = { 0.5 }\nlet sample = drand48()",
            "let pick = { (arc4random_uniform: (Int) -> Int) in arc4random_uniform(6) }",
            "let pick = { arc4random_uniform in arc4random_uniform(6) }",
            "func `NSMakeRange`(_ start: Int, _ count: Int) -> Span { Span(start, count) }\nlet span = NSMakeRange(0, 1)",
        ]
        for source in sources {
            #expect(Self.findings(source).isEmpty, "\(source)")
        }
    }

    /* A shadow silences only its own name. */
    @Test func otherNamesInAShadowingFileAreStillFound() {
        let source = """
            func CGPointMake(_ x: Double, _ y: Double) -> Point { Point(x, y) }
            let size = CGSizeMake(1, 2)
            """
        #expect(Self.positions(Self.findings(source)) == ["2:12"])
    }

    @Test func constructorFixLabelsTheArgumentsAsWritten() {
        let source = """
            let point = CGPointMake(calculateX(), 10)
            let range = NSMakeRange(index, text.count - index)
            let insets = NSEdgeInsetsMake(
                top, // the toolbar
                0, 0, 0)
            let offset = UIOffsetMake(0, 10)
            """
        let expected = """
            let point = CGPoint(x: calculateX(), y: 10)
            let range = NSRange(location: index, length: text.count - index)
            let insets = NSEdgeInsets(
                top: top, // the toolbar
                left: 0, bottom: 0, right: 0)
            let offset = UIOffset(horizontal: 0, vertical: 10)
            """
        #expect(Self.fixed(source) == expected)
    }

    /* SwiftLint's `legacy_cggeometry_functions` correction examples. The results match, except that whitespace before a kept closing parenthesis stays for the formatter. */
    @Test func coreGraphicsFixesMatchTheIncumbentCorrections() {
        let pairs = [
            ("CGRectGetWidth( rect  )", "rect.width"),
            ("CGRectGetHeight(rect )", "rect.height"),
            ("CGRectGetMinX( rect)", "rect.minX"),
            ("CGRectGetMidX(  rect)", "rect.midX"),
            ("CGRectGetMaxX( rect)", "rect.maxX"),
            ("CGRectGetMinY(rect   )", "rect.minY"),
            ("CGRectGetMidY(rect )", "rect.midY"),
            ("CGRectGetMaxY( rect     )", "rect.maxY"),
            ("CGRectIsNull(  rect    )", "rect.isNull"),
            ("CGRectIsEmpty( rect )", "rect.isEmpty"),
            ("CGRectIsInfinite( rect )", "rect.isInfinite"),
            ("CGRectStandardize( rect)", "rect.standardized"),
            ("CGRectIntegral(rect )", "rect.integral"),
            ("CGRectInset(rect, 5.0, -7.0)", "rect.insetBy(dx: 5.0, dy: -7.0)"),
            ("CGRectOffset(rect, -2, 8.3)", "rect.offsetBy(dx: -2, dy: 8.3)"),
            ("CGRectUnion(rect1, rect2)", "rect1.union(rect2)"),
            ("CGRectIntersection( rect1 ,rect2)", "rect1.intersection(rect2)"),
            ("CGRectContainsRect( rect1,rect2     )", "rect1.contains(rect2     )"),
            ("CGRectContainsPoint(rect  ,point)", "rect.contains(point)"),
            ("CGRectIntersectsRect(  rect1,rect2 )", "rect1.intersects(rect2 )"),
            ("CGRectIntersectsRect(rect1, rect2 )\nCGRectGetWidth(rect  )", "rect1.intersects(rect2 )\nrect.width"),
        ]
        for (source, expected) in pairs {
            #expect(Self.fixed(source) == expected, "\(source)")
        }
    }

    /* Moved arguments that carry their own type are fixed, with parentheses only where binding would change. */
    @Test func fixesApplyWhereMovedArgumentsCarryTheirOwnType() {
        let source = """
            let height = CGRectGetHeight(view.frame)
            let first = CGRectGetMinX(frames!.first!)
            let named = CGRectGetWidth(super.bounds)
            let empty = CGRectIsEmpty(CGRect.zero)
            let wrapped = CGRectGetWidth((frame))
            let cast = CGRectGetWidth(value as CGRect)
            let inner = CGRectInset(bounds, margin(for: item), 1)
            if !NSEqualPoints(start, end) { }
            if NSEqualPoints(start, end) { }
            check(NSEqualSizes(size, other), flag || NSEqualPoints(start, end))
            """
        let expected = """
            let height = view.frame.height
            let first = frames!.first!.minX
            let named = super.bounds.width
            let empty = CGRect.zero.isEmpty
            let wrapped = (frame).width
            let cast = (value as CGRect).width
            let inner = bounds.insetBy(dx: margin(for: item), dy: 1)
            if !(start == end) { }
            if start == end { }
            check(size == other, flag || (start == end))
            """
        let found = Self.findings(source)
        #expect(found.allSatisfy { !$0.fixes.isEmpty && $0.suggestions.isEmpty })
        #expect(Self.fixed(source) == expected)
    }

    /* A generic return may take its type from the parameter, so `decode().width` might not compile: suggested, never applied. */
    @Test func genericReturnReceiverIsSuggestedNotFixed() {
        let source = "let width = CGRectGetWidth(decode())"
        let found = Self.findings(source)
        #expect(Self.positions(found) == ["1:13"])
        #expect(found.first?.fixes.isEmpty == true)
        #expect(found.first?.suggestions.count == 1)
        #expect(Self.fixed(source) == source)
        #expect(Self.suggested(source) == "let width = decode().width")
    }

    /* Any call, subscript, operator, generic member or literal in a moved argument withholds the fix and keeps the suggestion. */
    @Test func movedArgumentsWhoseTypeCouldBeInferredAreSuggested() {
        let source = """
            let width = CGRectGetWidth(override ?? frame)
            let inner = CGRectInset(layout(for: item).bounds, 1, 1)
            let first = CGRectGetMinX(frames[0])
            if CGRectIsEmpty(make { bounds }) { }
            let typed = CGRectGetWidth(Box<CGRect>.value)
            let made = CGRectGetWidth(CGRect(x: 0, y: 0, width: 1, height: 1))
            let same = NSEqualSizes(size, other ?? .zero)
            let level = NSEqualPoints(.zero, start)
            let near = CGRectGetWidth(layout(.zero))
            """
        let found = Self.findings(source)
        #expect(Self.positions(found) == ["1:13", "2:13", "3:13", "4:4", "5:13", "6:12", "7:12", "8:13", "9:12"])
        #expect(found.allSatisfy { $0.fixes.isEmpty && $0.suggestions.count == 1 })
        #expect(Self.fixed(source) == source)
        let suggested = """
            let width = (override ?? frame).width
            let inner = layout(for: item).bounds.insetBy(dx: 1, dy: 1)
            let first = frames[0].minX
            if (make { bounds }).isEmpty { }
            let typed = Box<CGRect>.value.width
            let made = CGRect(x: 0, y: 0, width: 1, height: 1).width
            let same = size == (other ?? .zero)
            let level = .zero == start
            let near = layout(.zero).width
            """
        #expect(Self.suggested(source) == suggested)
    }

    /* Inner calls that move nothing are still fixed; an outer one whose receiver is a call is only suggested. */
    @Test func nestedCallsFixWhatIsSafeAndSuggestTheRest() {
        let source = "let width = CGRectGetWidth(CGRectInset(CGRectMake(0, 0, side, side), 1, 1))"
        let found = Self.findings(source)
        #expect(found.map(\.fixes.isEmpty) == [true, true, false])
        #expect(
            Self.fixed(source)
                == "let width = CGRectGetWidth(CGRectInset(CGRect(x: 0, y: 0, width: side, height: side), 1, 1))"
        )
        #expect(Self.suggested(source) == "let width = CGRectMake(0, 0, side, side).insetBy(dx: 1, dy: 1).width")
    }

    /* Reported, with nothing offered: a receiver typed only by the parameter, and an edit that would delete a comment. */
    @Test func fixesAreWithheldWhereTheyCannotBeMadeSafely() {
        let source = """
            let zero = CGRectGetWidth(.zero)
            let made = CGRectGetWidth(.init(x: 0, y: 0, width: 1, height: 1))
            let empty = CGRectIsNull(nil ?? frame)
            let noted = CGRectGetWidth(/* the window */ frame)
            let union = CGRectUnion(frame, // the panel
                other)
            """
        let found = Self.findings(source)
        #expect(Self.positions(found) == ["1:12", "2:12", "3:13", "4:13", "5:13"])
        #expect(found.allSatisfy { $0.fixes.isEmpty && $0.suggestions.isEmpty })
    }

    /* The NSGeometry rewrites that change meaning are offered, never applied; the two exact ones are fixed. */
    @Test func appKitRewritesAreSuggestedUnlessExact() {
        let source = """
            let width = NSWidth(bounds)
            let inset = NSInsetRect(bounds, 2, 2)
            let inside = NSPointInRect(point, bounds)
            let same = NSEqualRects(first, second)
            let level = NSEqualPoints(start, end)
            """
        let found = Self.findings(source)
        #expect(found.map(\.fixes.isEmpty) == [true, true, true, true, false])
        #expect(found.map(\.suggestions.isEmpty) == [false, false, false, false, true])
        let suggested = """
            let width = bounds.width
            let inset = bounds.insetBy(dx: 2, dy: 2)
            let inside = bounds.contains(point)
            let same = first == second
            let level = NSEqualPoints(start, end)
            """
        #expect(Self.suggested(source) == suggested)
        #expect(
            Self.fixed(source) == source.replacingOccurrences(of: "NSEqualPoints(start, end)", with: "start == end")
        )
    }

    /* Each message names the legacy call and what to write instead. */
    @Test func messagesNameTheModernForm() {
        let found = Self.findings("CGPointMake(1, 2)\nCGRectGetWidth(rect)\nNSWidth(rect)\narc4random_uniform(6)")
        #expect(
            found.map(\.message) == [
                "CGPointMake is the C spelling from before Swift had its own initializers. Write CGPoint(x:y:), which builds the same value and reads as Swift.",
                "CGRectGetWidth is the C spelling of a CGRect member. Write rect.width, the same function called the Swift way.",
                "NSWidth is the C spelling from before Swift had its own API. Swift's form is rect.width, but it works on the standardized rectangle, so it differs when a width or height is negative or the rectangle is empty. Switch where that cannot happen.",
                "arc4random_uniform is the C random API from before Swift had its own. Write Int.random(in: 0..<upperBound), or UInt32.random(in:) where the UInt32 result matters. The result type differs, so choose the type the code needs.",
            ]
        )
    }
}
