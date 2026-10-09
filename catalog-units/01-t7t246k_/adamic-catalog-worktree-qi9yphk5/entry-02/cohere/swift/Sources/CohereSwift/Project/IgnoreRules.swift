import Foundation

/*
 What git would leave out of a repository, read from the files git itself reads, with no git process.

 Kirk's ruling (#67vevpz): cohere runs no git at all. The file set used to ask `git ls-files` which files
 git does not ignore, and the format phase asked `git rev-parse` where the repository starts. Both answers
 are on disk: the repository starts at the nearest `.git`, and what it ignores is every `.gitignore` from
 its root down, `info/exclude` in its git directory, and the global excludes file. Reading them costs no
 process launch and does not depend on whichever git is first on PATH.

 The rules are git's own (`dir.c` and `wildmatch.c`), ported rather than approximated, because a matcher
 that is nearly right drops a file silently, which is the failure the coverage line exists to prevent:
 - Blank lines and `#` comments are skipped. `\#` and `\!` begin a pattern with that character.
 - Trailing spaces are trimmed unless escaped with `\`.
 - `!` negates, and the last matching line wins. Files are consulted from the deepest `.gitignore` up to
   the root's, then `info/exclude`, then the global file, and the first file with a matching line decides.
 - A trailing `/` matches directories only.
 - A pattern with a slash before its end is anchored to its file's directory. One without matches a name
   at any depth below that directory.
 - `*`, `?` and `[...]` never match `/` in an anchored pattern. `**` crosses directories when it stands as
   a whole segment, leading, in the middle or trailing. (Spelled out rather than shown, because those
   spellings would end and open Swift block comments.)
 - A file inside an excluded directory cannot be re-included: git never looks inside, so neither does this.
 - A directory holding its own `.git` is another repository and nothing inside it is listed, the boundary
   Go's `formatfiles.HasOwnRepository` draws with the same check.
 - `core.ignorecase` folds ASCII case, as git does on macOS's default filesystem.

 Go's `formatfiles.matchesIgnore` is a deliberate subset (no negation, no anchoring rule), sized for the
 format walk's pruning. This is all of git's rules, because here it replaces git's answer outright.

 Two differences from `git ls-files --cached`, by construction, because whether a file is tracked lives in
 git's index, a binary file this does not read:
 - A tracked file that matches an ignore pattern is listed by git, since tracking outranks ignoring, and is
   excluded here. Measured before this shipped (`git ls-files -ci --exclude-standard`): neither the Presence
   worktree nor cohere's repository tracks such a Swift file, and the archive copies have no repository.
 - A tracked file deleted from the working tree is still listed by git, and is not here, since it is not on
   disk. The file set used to name such a file as one no target compiles.
 Also not read: the system config file and `include` directives, when looking for `core.excludesFile`.
 */
final class IgnoreRules {
    /* The repository holding a directory: the nearest directory above it with a `.git`, and the git directory holding its shared files. */
    struct Repository: Equatable, Sendable {
        /* The directory holding `.git`, symlinks resolved, as `git rev-parse --show-toplevel` prints it. */
        var root: URL
        /*
         Where `info/exclude` and `config` live: `.git` itself, or for a worktree the main repository's git
         directory, which the worktree's `commondir` names. Nil when a `.git` file names nothing readable.
         */
        var commonDirectory: URL?

        /*
         The repository holding `directory`, or nil when no directory from it up to `/` holds a `.git`. A
         `.git` that is a file belongs to a worktree or a submodule, and holds `gitdir: <path>`.
         */
        static func containing(_ directory: URL) -> Repository? {
            /* Walked as a path string and stopped at `/` by name: a URL's parent of `/` is `/..`, which never stops climbing. */
            var candidate = directory.resolvingSymlinksInPath().path
            while true {
                let marker = URL(fileURLWithPath: candidate, isDirectory: true).appendingPathComponent(".git")
                if FileManager.default.fileExists(atPath: marker.path) {
                    let gitDirectory = IgnoreRules.isDirectory(marker) ? marker : linkedGitDirectory(marker)
                    return Repository(
                        root: URL(fileURLWithPath: candidate, isDirectory: true),
                        commonDirectory: gitDirectory.map(commonDirectory(of:)),
                    )
                }
                let parent = (candidate as NSString).deletingLastPathComponent
                if parent == candidate || parent.isEmpty {
                    return nil
                }
                candidate = parent
            }
        }

        /* The git directory a `.git` file points at, relative to the file's own directory unless absolute. */
        static func linkedGitDirectory(_ marker: URL) -> URL? {
            guard let contents = FileManager.default.contents(atPath: marker.path) else { return nil }
            let prefix = "gitdir:"
            let lines = String(decoding: contents, as: UTF8.self).split(whereSeparator: \.isNewline)
            guard let line = lines.first(where: { $0.hasPrefix(prefix) }) else { return nil }
            let path = line.dropFirst(prefix.count).trimmingCharacters(in: .whitespaces)
            return URL(fileURLWithPath: path, relativeTo: marker.deletingLastPathComponent()).standardizedFileURL
        }

        /* A worktree's git directory names the shared one in `commondir`, relative to itself. Any other git directory is its own. */
        static func commonDirectory(of gitDirectory: URL) -> URL {
            guard
                let contents = FileManager.default.contents(
                    atPath: gitDirectory.appendingPathComponent("commondir").path
                )
            else {
                return gitDirectory
            }
            let path = String(decoding: contents, as: UTF8.self).trimmingCharacters(in: .whitespacesAndNewlines)
            guard !path.isEmpty else { return gitDirectory }
            return URL(fileURLWithPath: path, relativeTo: gitDirectory).standardizedFileURL
        }
    }

    /* One line of an ignore file, parsed as git's `parse_path_pattern` parses it. */
    struct Pattern: Equatable, Sendable {
        /* The glob, without its `!`, its trailing `/` or, when anchored, its leading `/`. */
        var glob: [UInt8]
        var negated: Bool
        var directoriesOnly: Bool
        /* No slash before the end: matched against the name alone, at any depth. */
        var matchesNameAnywhere: Bool

        init(line: [UInt8]) {
            var glob = line[...]
            negated = glob.first == Byte.exclamation
            if negated {
                glob = glob.dropFirst()
            }
            directoriesOnly = glob.last == Byte.slash
            if directoriesOnly {
                glob = glob.dropLast()
            }
            matchesNameAnywhere = !glob.contains(Byte.slash)
            if !matchesNameAnywhere && glob.first == Byte.slash {
                glob = glob.dropFirst()
            }
            self.glob = Array(glob)
        }

        /* `path` is relative to the ignore file's directory, and `name` is its last component. */
        func matches(path: [UInt8], name: [UInt8], isDirectory: Bool, foldCase: Bool) -> Bool {
            if directoriesOnly && !isDirectory {
                return false
            }
            if matchesNameAnywhere {
                return Wildcard(pattern: glob, text: name, slashIsSpecial: false, foldCase: foldCase).matches()
            }
            return Wildcard(pattern: glob, text: path, slashIsSpecial: true, foldCase: foldCase).matches()
        }

        /* Every pattern in an ignore file's bytes, in file order, as git's `add_patterns_from_buffer` reads them. */
        static func patterns(in contents: [UInt8]) -> [Pattern] {
            var body = contents[...]
            let byteOrderMark: [UInt8] = [0xEF, 0xBB, 0xBF]
            if body.starts(with: byteOrderMark) {
                body = body.dropFirst(byteOrderMark.count)
            }
            var patterns: [Pattern] = []
            for rawLine in body.split(separator: Byte.newline, omittingEmptySubsequences: false) {
                var line = rawLine
                if line.last == Byte.carriageReturn {
                    line = line.dropLast()
                }
                guard !line.isEmpty, line.first != Byte.hash else { continue }
                let trimmed = trimmingTrailingSpaces(line)
                if !trimmed.isEmpty {
                    patterns.append(Pattern(line: Array(trimmed)))
                }
            }
            return patterns
        }

        /* Trailing spaces go, tabs and escaped spaces stay: git's `trim_trailing_spaces`. */
        static func trimmingTrailingSpaces(_ line: ArraySlice<UInt8>) -> ArraySlice<UInt8> {
            var firstTrailingSpace: Int?
            var index = line.startIndex
            while index < line.endIndex {
                switch line[index] {
                    case Byte.space:
                        if firstTrailingSpace == nil {
                            firstTrailingSpace = index
                        }
                    case Byte.backslash:
                        index += 1
                        if index == line.endIndex {
                            return line
                        }
                        firstTrailingSpace = nil
                    default:
                        firstTrailingSpace = nil
                }
                index += 1
            }
            guard let firstTrailingSpace else { return line }
            return line[line.startIndex..<firstTrailingSpace]
        }
    }

    /* The bytes the parser and the matcher compare against, named. */
    enum Byte {
        static let slash = UInt8(ascii: "/")
        static let backslash = UInt8(ascii: "\\")
        static let exclamation = UInt8(ascii: "!")
        static let caret = UInt8(ascii: "^")
        static let hash = UInt8(ascii: "#")
        static let space = UInt8(ascii: " ")
        static let tab = UInt8(ascii: "\t")
        static let newline = UInt8(ascii: "\n")
        static let carriageReturn = UInt8(ascii: "\r")
        static let asterisk = UInt8(ascii: "*")
        static let question = UInt8(ascii: "?")
        static let openBracket = UInt8(ascii: "[")
        static let closeBracket = UInt8(ascii: "]")
        static let colon = UInt8(ascii: ":")
        static let hyphen = UInt8(ascii: "-")
    }

    /*
     git's `wildmatch`, ported line for line over bytes. `slashIsSpecial` is its `WM_PATHNAME`: `*`, `?` and a
     bracket never match `/`, and `**` crosses directories only as a whole segment. The two abort outcomes
     are how git stops a doomed backtrack early, and they are kept so the answers stay git's.
     */
    struct Wildcard {
        enum Outcome {
            case match
            case noMatch
            case abortAll
            case abortToDoubleAsterisk
        }

        var pattern: [UInt8]
        var text: [UInt8]
        var slashIsSpecial: Bool
        var foldCase: Bool

        func matches() -> Bool {
            match(patternFrom: 0, textFrom: 0) == .match
        }

        /* The byte at `index`, or 0 past the end, which is how git's C reads its terminating NUL. */
        private func patternByte(_ index: Int) -> UInt8 {
            index < pattern.count ? pattern[index] : 0
        }

        private func textByte(_ index: Int) -> UInt8 {
            index < text.count ? text[index] : 0
        }

        private func folded(_ byte: UInt8) -> UInt8 {
            foldCase && Self.isUpper(byte) ? byte + 32 : byte
        }

        private func match(patternFrom patternStart: Int, textFrom textStart: Int) -> Outcome {
            var patternIndex = patternStart
            var textIndex = textStart
            while patternIndex < pattern.count {
                var patternCharacter = folded(pattern[patternIndex])
                var textCharacter = textByte(textIndex)
                if textIndex >= text.count && patternCharacter != Byte.asterisk {
                    return .abortAll
                }
                textCharacter = folded(textCharacter)
                switch patternCharacter {
                    case Byte.backslash:
                        /* Literal match with the next byte. A pattern ending in `\` reads 0 here and matches nothing. */
                        patternIndex += 1
                        patternCharacter = patternByte(patternIndex)
                        if textCharacter != patternCharacter {
                            return .noMatch
                        }
                    case Byte.question:
                        if slashIsSpecial && textCharacter == Byte.slash {
                            return .noMatch
                        }
                    case Byte.asterisk:
                        let crossesSlashes: Bool
                        patternIndex += 1
                        if patternByte(patternIndex) == Byte.asterisk {
                            let beforeAsterisks = patternIndex - 2
                            patternIndex += 1
                            while patternByte(patternIndex) == Byte.asterisk {
                                patternIndex += 1
                            }
                            let next = patternByte(patternIndex)
                            if !slashIsSpecial {
                                crossesSlashes = true
                            }
                            else if (beforeAsterisks < 0 || pattern[beforeAsterisks] == Byte.slash)
                                && (next == 0 || next == Byte.slash
                                    || (next == Byte.backslash && patternByte(patternIndex + 1) == Byte.slash))
                            {
                                /* A whole-segment `**` between slashes may match no directory at all, so try that first. */
                                if next == Byte.slash
                                    && match(patternFrom: patternIndex + 1, textFrom: textIndex) == .match
                                {
                                    return .match
                                }
                                crossesSlashes = true
                            }
                            else {
                                crossesSlashes = false
                            }
                        }
                        else {
                            crossesSlashes = !slashIsSpecial
                        }
                        if patternIndex >= pattern.count {
                            /* A trailing `**` matches everything left. A trailing `*` matches only within this directory. */
                            if !crossesSlashes && text[textIndex...].contains(Byte.slash) {
                                return .abortToDoubleAsterisk
                            }
                            return .match
                        }
                        if !crossesSlashes && pattern[patternIndex] == Byte.slash {
                            /* One `*` before a slash matches the rest of this directory's name. */
                            guard let slash = text[textIndex...].firstIndex(of: Byte.slash) else {
                                return .abortAll
                            }
                            textIndex = slash + 1
                            patternIndex += 1
                            continue
                        }
                        while true {
                            if textCharacter == 0 {
                                break
                            }
                            let next = pattern[patternIndex]
                            if !Self.isGlobSpecial(next) {
                                /* A literal after the asterisk: skip ahead to where it occurs, since everything before it belongs to the asterisk. */
                                let literal = folded(next)
                                while textIndex < text.count {
                                    textCharacter = folded(text[textIndex])
                                    if (!crossesSlashes && textCharacter == Byte.slash) || textCharacter == literal {
                                        break
                                    }
                                    textIndex += 1
                                }
                                if textIndex >= text.count {
                                    textCharacter = 0
                                }
                                if textCharacter != literal {
                                    return .noMatch
                                }
                            }
                            let outcome = match(patternFrom: patternIndex, textFrom: textIndex)
                            if outcome != .noMatch {
                                if !crossesSlashes || outcome != .abortToDoubleAsterisk {
                                    return outcome
                                }
                            }
                            else if !crossesSlashes && textCharacter == Byte.slash {
                                return .abortToDoubleAsterisk
                            }
                            textIndex += 1
                            textCharacter = textByte(textIndex)
                        }
                        return .abortAll
                    case Byte.openBracket:
                        guard
                            let afterBracket = matchBracket(patternIndex: &patternIndex, textCharacter: textCharacter)
                        else {
                            return .abortAll
                        }
                        if !afterBracket || (slashIsSpecial && textCharacter == Byte.slash) {
                            return .noMatch
                        }
                    default:
                        if textCharacter != patternCharacter {
                            return .noMatch
                        }
                }
                patternIndex += 1
                textIndex += 1
            }
            return textIndex < text.count ? .noMatch : .match
        }

        /*
         A `[...]` class starting at `patternIndex`, left on its closing `]`. True when `textCharacter` is in
         it (out of it, for `[!...]` or `[^...]`), nil when the class is malformed, which git treats as a
         pattern that can match nothing more.
         */
        private func matchBracket(patternIndex: inout Int, textCharacter: UInt8) -> Bool? {
            patternIndex += 1
            var classCharacter = patternByte(patternIndex)
            if classCharacter == Byte.caret {
                classCharacter = Byte.exclamation
            }
            let negated = classCharacter == Byte.exclamation
            if negated {
                patternIndex += 1
                classCharacter = patternByte(patternIndex)
            }
            var previous: UInt8 = 0
            var matched = false
            repeat {
                if classCharacter == 0 {
                    return nil
                }
                if classCharacter == Byte.backslash {
                    patternIndex += 1
                    classCharacter = patternByte(patternIndex)
                    if classCharacter == 0 {
                        return nil
                    }
                    if textCharacter == classCharacter {
                        matched = true
                    }
                }
                else if classCharacter == Byte.hyphen && previous != 0
                    && patternByte(patternIndex + 1) != 0 && patternByte(patternIndex + 1) != Byte.closeBracket
                {
                    patternIndex += 1
                    classCharacter = patternByte(patternIndex)
                    if classCharacter == Byte.backslash {
                        patternIndex += 1
                        classCharacter = patternByte(patternIndex)
                        if classCharacter == 0 {
                            return nil
                        }
                    }
                    if textCharacter <= classCharacter && textCharacter >= previous {
                        matched = true
                    }
                    else if foldCase && Self.isLower(textCharacter) {
                        let upper = textCharacter - 32
                        if upper <= classCharacter && upper >= previous {
                            matched = true
                        }
                    }
                    classCharacter = 0
                }
                else if classCharacter == Byte.openBracket && patternByte(patternIndex + 1) == Byte.colon {
                    patternIndex += 2
                    let nameStart = patternIndex
                    while patternByte(patternIndex) != 0 && patternByte(patternIndex) != Byte.closeBracket {
                        patternIndex += 1
                    }
                    if patternByte(patternIndex) == 0 {
                        return nil
                    }
                    let nameLength = patternIndex - nameStart - 1
                    if nameLength < 0 || patternByte(patternIndex - 1) != Byte.colon {
                        /* No `:]`: the `[` was an ordinary member of the class. */
                        patternIndex = nameStart - 2
                        classCharacter = Byte.openBracket
                        if textCharacter == classCharacter {
                            matched = true
                        }
                    }
                    else {
                        let name = String(decoding: pattern[nameStart..<(nameStart + nameLength)], as: UTF8.self)
                        guard let member = Self.isInNamedClass(name, textCharacter, foldCase: foldCase) else {
                            return nil
                        }
                        if member {
                            matched = true
                        }
                        classCharacter = 0
                    }
                }
                else if textCharacter == classCharacter {
                    matched = true
                }
                previous = classCharacter
                patternIndex += 1
                classCharacter = patternByte(patternIndex)
            } while classCharacter != Byte.closeBracket
            return matched != negated
        }

        /* `[:name:]` inside a class. Nil for a name git does not know, which aborts the match as git does. */
        static func isInNamedClass(_ name: String, _ byte: UInt8, foldCase: Bool) -> Bool? {
            switch name {
                case "alnum": isAlpha(byte) || isDigit(byte)
                case "alpha": isAlpha(byte)
                case "blank": byte == Byte.space || byte == Byte.tab
                case "cntrl": byte < 0x20 || byte == 0x7F
                case "digit": isDigit(byte)
                case "graph": byte > 0x20 && byte < 0x7F
                case "lower": isLower(byte)
                case "print": byte >= 0x20 && byte < 0x7F
                case "punct": byte > 0x20 && byte < 0x7F && !isAlpha(byte) && !isDigit(byte)
                /* git's own `isspace`: no vertical tab or form feed. */
                case "space":
                    byte == Byte.space || byte == Byte.tab || byte == Byte.newline || byte == Byte.carriageReturn
                case "upper": isUpper(byte) || (foldCase && isLower(byte))
                case "xdigit":
                    isDigit(byte) || (byte >= UInt8(ascii: "a") && byte <= UInt8(ascii: "f"))
                        || (byte >= UInt8(ascii: "A") && byte <= UInt8(ascii: "F"))
                default: nil
            }
        }

        static func isGlobSpecial(_ byte: UInt8) -> Bool {
            byte == Byte.asterisk || byte == Byte.question || byte == Byte.openBracket || byte == Byte.backslash
        }

        static func isUpper(_ byte: UInt8) -> Bool {
            byte >= UInt8(ascii: "A") && byte <= UInt8(ascii: "Z")
        }

        static func isLower(_ byte: UInt8) -> Bool {
            byte >= UInt8(ascii: "a") && byte <= UInt8(ascii: "z")
        }

        static func isAlpha(_ byte: UInt8) -> Bool {
            isUpper(byte) || isLower(byte)
        }

        static func isDigit(_ byte: UInt8) -> Bool {
            byte >= UInt8(ascii: "0") && byte <= UInt8(ascii: "9")
        }
    }

    let repository: Repository
    /* The repository root as a path, symlinks resolved, every path below it is measured from. */
    private let rootPath: String
    private let foldCase: Bool
    /* `info/exclude`, then the global excludes file: consulted after every `.gitignore`, in that order. */
    private let repositoryWidePatterns: [[Pattern]]
    /* Each directory's own `.gitignore`, by its path relative to the root ("" for the root), read once. */
    private var directoryPatterns: [String: [Pattern]] = [:]

    /*
     `environment` supplies `HOME` and `XDG_CONFIG_HOME`, where the global config and excludes file live.
     Tests pass their own, so a machine's global ignore cannot leak into what they prove.
     */
    init(repository: Repository, environment: [String: String] = ProcessInfo.processInfo.environment) {
        self.repository = repository
        rootPath = repository.root.path
        let home = environment["HOME"].flatMap { $0.isEmpty ? nil : $0 } ?? NSHomeDirectory()
        let configurationDirectory =
            environment["XDG_CONFIG_HOME"].flatMap { $0.isEmpty ? nil : $0 + "/git" } ?? home + "/.config/git"

        /* git's order: the XDG file, then `~/.gitconfig`, then the repository's own; the last one setting a key wins. */
        var configurationFiles = [configurationDirectory + "/config", home + "/.gitconfig"]
        if let commonDirectory = repository.commonDirectory {
            configurationFiles.append(commonDirectory.appendingPathComponent("config").path)
        }
        var settings: [String: String] = [:]
        for file in configurationFiles {
            guard let contents = FileManager.default.contents(atPath: file) else { continue }
            for entry in Self.coreSettings(in: String(decoding: contents, as: UTF8.self)) {
                settings[entry.key] = entry.value
            }
        }
        foldCase = settings["ignorecase"].map(Self.isTrue) ?? false

        let globalExcludes: String
        if let configured = settings["excludesfile"], !configured.isEmpty {
            if configured.hasPrefix("~/") {
                globalExcludes = home + configured.dropFirst(1)
            }
            else if configured.hasPrefix("/") {
                globalExcludes = configured
            }
            else {
                /* git reads it from the top of the work tree, where it runs. */
                globalExcludes = repository.root.appendingPathComponent(configured).path
            }
        }
        else {
            globalExcludes = configurationDirectory + "/ignore"
        }
        var repositoryWide: [[Pattern]] = []
        if let commonDirectory = repository.commonDirectory {
            repositoryWide.append(Self.patterns(inFile: commonDirectory.appendingPathComponent("info/exclude").path))
        }
        repositoryWide.append(Self.patterns(inFile: globalExcludes))
        repositoryWidePatterns = repositoryWide
    }

    /* Whether git would list `path` (absolute, symlinks resolved): inside this repository, inside no other, under no ignored directory, and not ignored itself. */
    func isVisible(_ path: String, isDirectory: Bool = false) -> Bool {
        let prefix = rootPath.hasSuffix("/") ? rootPath : rootPath + "/"
        guard path.hasPrefix(prefix) else { return false }
        let components = path.dropFirst(prefix.count).split(separator: "/").map(String.init)
        var parent = ""
        for (index, name) in components.enumerated() {
            let relative = parent.isEmpty ? name : parent + "/" + name
            let entryIsDirectory = index < components.count - 1 || isDirectory
            if !admits(name: name, relative: relative, parent: parent, isDirectory: entryIsDirectory) {
                return false
            }
            parent = relative
        }
        return true
    }

    /*
     Every file git would list below `directory` (absolute, symlinks resolved), as absolute paths with
     symlinks resolved, the way the file set compares them. Pruned as git prunes: an ignored directory and
     another repository are never entered. A symlink is a file here, as it is to git, never followed.
     */
    func visibleFiles(under directory: String) -> [String] {
        guard directory == rootPath || isVisible(directory, isDirectory: true) else { return [] }
        let relative = directory == rootPath ? "" : String(directory.dropFirst(rootPath.count + 1))
        var files: [String] = []
        collect(relativeDirectory: relative, into: &files)
        return files
    }

    private func collect(relativeDirectory: String, into files: inout [String]) {
        let absolute = relativeDirectory.isEmpty ? rootPath : rootPath + "/" + relativeDirectory
        let entries: [URL]
        do {
            entries = try FileManager.default.contentsOfDirectory(
                at: URL(fileURLWithPath: absolute, isDirectory: true),
                includingPropertiesForKeys: [.isDirectoryKey, .isSymbolicLinkKey],
            )
        }
        catch {
            /* A directory that cannot be read lists nothing, as git skips one it cannot open. */
            return
        }
        for entry in entries {
            let name = entry.lastPathComponent
            let relative = relativeDirectory.isEmpty ? name : relativeDirectory + "/" + name
            let isSymbolicLink = Self.isSymbolicLink(entry)
            let isDirectory = !isSymbolicLink && Self.isDirectory(entry)
            guard admits(name: name, relative: relative, parent: relativeDirectory, isDirectory: isDirectory) else {
                continue
            }
            if isDirectory {
                collect(relativeDirectory: relative, into: &files)
            }
            else {
                let path = rootPath + "/" + relative
                files.append(isSymbolicLink ? URL(fileURLWithPath: path).resolvingSymlinksInPath().path : path)
            }
        }
    }

    /* One step down: whether git lists this entry (a file) or enters it (a directory). */
    private func admits(name: String, relative: String, parent: String, isDirectory: Bool) -> Bool {
        if name == ".git" {
            return false
        }
        if isDirectory && FileManager.default.fileExists(atPath: rootPath + "/" + relative + "/.git") {
            return false
        }
        return !isIgnored(relative: relative, parent: parent, name: name, isDirectory: isDirectory)
    }

    /* The deepest `.gitignore` with a matching line decides, then `info/exclude`, then the global file. */
    private func isIgnored(relative: String, parent: String, name: String, isDirectory: Bool) -> Bool {
        let path = Array(relative.utf8)
        let nameBytes = path[(path.count - name.utf8.count)...]
        var directory: String? = parent
        while let current = directory {
            let patterns = patterns(ownedBy: current)
            let below = current.isEmpty ? path[...] : path[(current.utf8.count + 1)...]
            if let decision = Self.decision(
                patterns,
                path: below,
                name: nameBytes,
                isDirectory: isDirectory,
                foldCase: foldCase,
            ) {
                return decision
            }
            directory = current.isEmpty ? nil : String(current[..<(current.lastIndex(of: "/") ?? current.startIndex)])
        }
        for patterns in repositoryWidePatterns {
            if let decision = Self.decision(
                patterns,
                path: path[...],
                name: nameBytes,
                isDirectory: isDirectory,
                foldCase: foldCase,
            ) {
                return decision
            }
        }
        return false
    }

    /* The last line in one file matching the path: ignored, or re-included by `!`. Nil when no line matches. */
    static func decision(
        _ patterns: [Pattern],
        path: ArraySlice<UInt8>,
        name: ArraySlice<UInt8>,
        isDirectory: Bool,
        foldCase: Bool,
    ) -> Bool? {
        guard !patterns.isEmpty else { return nil }
        let path = Array(path)
        let name = Array(name)
        let last = patterns.last { $0.matches(path: path, name: name, isDirectory: isDirectory, foldCase: foldCase) }
        return last.map { !$0.negated }
    }

    private func patterns(ownedBy relativeDirectory: String) -> [Pattern] {
        if let cached = directoryPatterns[relativeDirectory] {
            return cached
        }
        let file =
            relativeDirectory.isEmpty ? rootPath + "/.gitignore" : rootPath + "/" + relativeDirectory + "/.gitignore"
        let patterns = Self.patterns(inFile: file)
        directoryPatterns[relativeDirectory] = patterns
        return patterns
    }

    /* A missing ignore file ignores nothing, as it does for git. */
    static func patterns(inFile path: String) -> [Pattern] {
        guard let contents = FileManager.default.contents(atPath: path) else { return [] }
        return Pattern.patterns(in: Array(contents))
    }

    /*
     The `core` section's keys from a git config file, lowercased, in file order. Enough of git's config
     grammar for the two keys read here: sections, `key = value`, a bare key meaning true, quotes, escapes
     and trailing comments.
     */
    static func coreSettings(in contents: String) -> [(key: String, value: String)] {
        var inCore = false
        var settings: [(key: String, value: String)] = []
        for rawLine in contents.split(whereSeparator: \.isNewline) {
            let line = rawLine.trimmingCharacters(in: .whitespaces)
            if line.isEmpty || line.hasPrefix("#") || line.hasPrefix(";") {
                continue
            }
            if line.hasPrefix("[") {
                let header = line.dropFirst().prefix { $0 != "]" }
                inCore = header.trimmingCharacters(in: .whitespaces).lowercased() == "core"
                continue
            }
            guard inCore else { continue }
            guard let equals = line.firstIndex(of: "=") else {
                settings.append((key: line.lowercased(), value: "true"))
                continue
            }
            let key = line[..<equals].trimmingCharacters(in: .whitespaces).lowercased()
            settings.append((key: key, value: configurationValue(line[line.index(after: equals)...])))
        }
        return settings
    }

    /* A config value as git reads it: quotes removed, `\n` `\t` `\\` `\"` unescaped, a `#` or `;` outside quotes ending it, outer spaces dropped. */
    static func configurationValue(_ raw: Substring) -> String {
        var value = ""
        var pendingSpaces = ""
        var quoted = false
        var characters = raw.makeIterator()
        while let character = characters.next() {
            if character == "\\" {
                guard let escaped = characters.next() else { break }
                value += pendingSpaces
                pendingSpaces = ""
                switch escaped {
                    case "n": value.append("\n")
                    case "t": value.append("\t")
                    default: value.append(escaped)
                }
            }
            else if character == "\"" {
                quoted.toggle()
            }
            else if !quoted && (character == "#" || character == ";") {
                break
            }
            else if !quoted && character.isWhitespace {
                if !value.isEmpty {
                    pendingSpaces.append(character)
                }
            }
            else {
                value += pendingSpaces
                pendingSpaces = ""
                value.append(character)
            }
        }
        return value
    }

    static func isTrue(_ value: String) -> Bool {
        ["true", "yes", "on", "1"].contains(value.lowercased())
    }

    static func isDirectory(_ url: URL) -> Bool {
        do {
            return try url.resourceValues(forKeys: [.isDirectoryKey]).isDirectory ?? false
        }
        catch {
            /* Gone between listing and asking: not a directory to descend into. */
            return false
        }
    }

    static func isSymbolicLink(_ url: URL) -> Bool {
        do {
            return try url.resourceValues(forKeys: [.isSymbolicLinkKey]).isSymbolicLink ?? false
        }
        catch {
            /* Gone between listing and asking: nothing to resolve. */
            return false
        }
    }
}
