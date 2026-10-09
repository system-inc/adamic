/*
 Which declaration each name in one file resolves to, as the compiler decided it: the input a typed rule reads
 instead of guessing from a name's spelling.

 An occurrence is placed the way the contract places findings, 1-based line and 1-based UTF-8 byte column, at
 the first character of the name. Its symbol is a unified symbol name, the compiler's stable identifier for a
 declaration: `s:Sa5countSivp` is `Array.count`, and a type of ours carries its module in its name. Standard
 library declarations are spelled `s:S` (a standard substitution: `Sa` Array, `SD` Dictionary, `Sh` Set, `SS`
 String, `Sl` Collection, `ST` Sequence) or `s:s` and a length-prefixed name (`s:s14_ArrayProtocol`).

 Both sources fill the same shape: the build's index store for a file the build compiled as it stands, and
 sourcekitd's `source.request.indexsource` for one it did not.
 */
public struct FileSymbols: Sendable {
    public struct Occurrence: Sendable, Equatable {
        public var line: Int
        public var column: Int
        public var symbol: String
        public var name: String
        public var isReference: Bool
        /*
         Not written in the source but implied by it: reading `items.count` also records a call of `count`'s
         getter at the same place. A rule asks about what the source names, so implicit occurrences never answer.
         */
        public var isImplicit: Bool

        public init(line: Int, column: Int, symbol: String, name: String, isReference: Bool, isImplicit: Bool = false) {
            self.line = line
            self.column = column
            self.symbol = symbol
            self.name = name
            self.isReference = isReference
            self.isImplicit = isImplicit
        }

        /*
         The module a declaration of a module's own type lives in, read from the symbol's leading length-prefixed
         name: `s:7Control6StatusO7runningyA2CmF` is `Control`'s. Nil for anything that does not start that way:
         the standard library's, an extension's member on another module's type, an Objective-C name.
         */
        public var declaringModule: String? {
            guard symbol.hasPrefix("s:") else { return nil }
            let rest = symbol.utf8.dropFirst(2)
            let digits = rest.prefix { $0 >= UInt8(ascii: "0") && $0 <= UInt8(ascii: "9") }
            guard let length = Int(String(decoding: digits, as: UTF8.self)), length > 0 else { return nil }
            let name = rest.dropFirst(digits.count).prefix(length)
            return name.count == length ? String(decoding: name, as: UTF8.self) : nil
        }

        /*
         Declared by the Swift standard library. Its symbols name their type first, as a standard substitution
         (`s:Sa`, `s:ST`) or `s` and a length-prefixed name and a kind letter (`s:s14_ArrayProtocolP`). The prefix
         alone is not enough: a member an extension adds names the extending module next, `s:ST7ControlE6sorted`
         for our own `extension Sequence` or `s:ST10FoundationE` for Foundation's, where the standard library's
         own extensions spell that module `s` (`s:STsE5first`). Found by the sorted-first-last rule, whose
         key-path `sorted(by:)` of ours read as the standard library's.
         */
        public var isStandardLibrary: Bool {
            guard symbol.hasPrefix("s:") else { return false }
            var rest = symbol.utf8.dropFirst(2)
            /* `So` and `SC` are types imported from C and Objective-C (`s:So6CGRectV...`), not the standard library's. */
            if rest.starts(with: "So".utf8) || rest.starts(with: "SC".utf8) {
                return false
            }
            if rest.starts(with: "Sc".utf8) {
                /* The concurrency substitutions are three characters: `ScT` Task, `Sci` AsyncSequence. */
                rest = rest.dropFirst(3)
            }
            else if rest.first == UInt8(ascii: "S") {
                rest = rest.dropFirst(2)
            }
            else if rest.first == UInt8(ascii: "s") {
                rest = rest.dropFirst()
                let digits = rest.prefix { $0 >= UInt8(ascii: "0") && $0 <= UInt8(ascii: "9") }
                guard let length = Int(String(decoding: digits, as: UTF8.self)) else { return false }
                rest = rest.dropFirst(digits.count + length + 1)
            }
            else {
                return false
            }
            /*
             What follows the type: a member, or the module of an extension, a length-prefixed name and `E`. The
             standard library's underscored companion modules ship with it and are its own (`_Concurrency` adds
             AsyncSequence's `filter`, `s:Sci12_ConcurrencyE6filter`); any other module's extension is not.
             */
            let digits = rest.prefix { $0 >= UInt8(ascii: "0") && $0 <= UInt8(ascii: "9") }
            guard let length = Int(String(decoding: digits, as: UTF8.self)) else { return true }
            let module = rest.dropFirst(digits.count).prefix(length)
            guard rest.dropFirst(digits.count + length).first == UInt8(ascii: "E") else { return true }
            return module.first == UInt8(ascii: "_")
        }
    }

    private struct Position: Hashable {
        var line: Int
        var column: Int
    }

    private let byPosition: [Position: [Occurrence]]
    public let count: Int
    /*
     The modules the checked package owns, as the compiler names them (a target's name with anything that is not
     an identifier character made `_`). Vendored packages' modules are not in it. Lets a rule tell a declaration
     of ours from a dependency's.
     */
    public var ownedModules: Set<String> = []

    public init(_ occurrences: [Occurrence], ownedModules: Set<String> = []) {
        byPosition = Dictionary(grouping: occurrences) { Position(line: $0.line, column: $0.column) }
        count = occurrences.count
        self.ownedModules = ownedModules
    }

    /* Whether the declaration was written in a module this package owns. */
    public func isOwned(_ occurrence: Occurrence) -> Bool {
        occurrence.declaringModule.map(ownedModules.contains) ?? false
    }

    /* Every occurrence that starts at this place: usually one, more where one name both declares and refers. */
    public func occurrences(line: Int, column: Int) -> [Occurrence] {
        byPosition[Position(line: line, column: column)] ?? []
    }

    /* The declaration a name at this place refers to, when the compiler recorded exactly one written reference there. */
    public func reference(line: Int, column: Int) -> Occurrence? {
        let references = occurrences(line: line, column: column).filter { $0.isReference && !$0.isImplicit }
        return references.count == 1 ? references.first : nil
    }
}
