/*
 The part of `swift package describe --type json` the engine reads. SwiftPM writes more; only these
 fields are decoded, so a new field in a later toolchain is ignored rather than fatal.
 */
struct DescribeOutput: Decodable {
    /* One target as `describe` writes it. `sources` are relative to `path`, and absent for some target kinds. */
    struct Target: Decodable {
        var name: String
        var type: String
        var path: String
        var sources: [String]?
    }

    /* A dependency. `fileSystem` ones carry an absolute `path`; `sourceControl` ones carry a URL the engine never reads. */
    struct Dependency: Decodable {
        var identity: String
        var type: String
        var path: String?
    }

    /* A product and the targets it is built from. An executable's build directory is named for its product. */
    struct Product: Decodable {
        var name: String
        var targets: [String]
        /* What kind of product, as `describe` writes it: an object with one key, `{"library": ["automatic"]}` or `{"executable": null}`. */
        var type: Kind?
    }

    /* A product's kind, read from the one key of the object `describe` writes, whatever its value holds. */
    struct Kind: Decodable {
        var name: String

        private struct Key: CodingKey {
            var stringValue: String
            var intValue: Int? { nil }
            init(stringValue: String) { self.stringValue = stringValue }
            init?(intValue: Int) { nil }
        }

        init(from decoder: any Decoder) throws {
            name = try decoder.container(keyedBy: Key.self).allKeys.first?.stringValue ?? ""
        }
    }

    var name: String
    var targets: [Target]
    var products: [Product]?
    var dependencies: [Dependency]?
}
