import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `correctness-require-response-status-check` both ways. The unit cases parse a source string and hand the rule
 symbols built by position, one occurrence at every token a case names, so each case says what the compiler would
 have resolved; a key `"<line>:<name>"` resolves that name on that line only. The three URLSession sites in
 ahraos-macos come first, as they stand (silent) and with their status check taken out (firing, or a miss the
 rule's comment names), then every shape the rule flags and every shape it leaves alone, the TypeScript rule's
 cases carried over where Swift has them. The end-to-end case builds a real package, so the symbols come from the
 index the build wrote.
 */
@Suite(.serialized)
struct CorrectnessRequireResponseStatusCheckTests {
    /* The words each finding carries, pinned here so a generator that drops or mangles them fails. */
    static let bodyMessage =
        "This reads the body of the response URLSession returned on a path where its statusCode is never read. URLSession does not throw on an HTTP error, so a 401 or 500 error page arrives here as data and is decoded, returned or kept as if the request had succeeded. Cast the response to HTTPURLResponse and check statusCode before trusting the body, and throw or return the failure when it is not a success."
    static let discardMessage =
        "This takes the response URLSession returned and drops it. Its status goes with it. URLSession does not throw on an HTTP error, so a request that fails with a 400 or 500 runs on as if it had succeeded. Keep the response, cast it to HTTPURLResponse and check statusCode, and throw or return the failure when it is not a success."

    static let dataFor =
        "s:So12NSURLSessionC10FoundationE4data3for8delegateAC4DataV_So13NSURLResponseCtAC10URLRequestV_So0A12TaskDelegate_pSgtYaKF"
    static let dataFrom =
        "s:So12NSURLSessionC10FoundationE4data4from8delegateAC4DataV_So13NSURLResponseCtAC3URLV_So0A12TaskDelegate_pSgtYaKF"
    static let uploadFor =
        "s:So12NSURLSessionC10FoundationE6upload3for4from8delegateAC4DataV_So13NSURLResponseCtAC10URLRequestV_AISo0A12TaskDelegate_pSgtYaKF"
    static let bytesFor =
        "s:So12NSURLSessionC10FoundationE5bytes3for8delegateAbCE10AsyncBytesV_So13NSURLResponseCtAC10URLRequestV_So0A12TaskDelegate_pSgtYaKF"
    static let downloadFor =
        "s:So12NSURLSessionC10FoundationE8download3for8delegateAC3URLV_So13NSURLResponseCtAC10URLRequestV_So0A12TaskDelegate_pSgtYaKF"
    static let dataTask = "c:objc(cs)NSURLSession(im)dataTaskWithRequest:completionHandler:"
    static let uploadTask = "c:objc(cs)NSURLSession(im)uploadTaskWithRequest:fromData:completionHandler:"
    static let httpResponse = "c:objc(cs)NSHTTPURLResponse"
    static let jsonData = "c:objc(cs)NSJSONSerialization(cm)dataWithJSONObject:options:error:"
    /* A client of ours with a `data(for:)` of its own, which already throws on a bad status. */
    static let ourData = "s:7Control9ApiClientC4data3forAA7PayloadVSS_tYaKF"
    /* A longer name in Foundation's extension that starts with `data(for:delegate:)`'s: one more label, so another declaration. */
    static let longerData =
        "s:So12NSURLSessionC10FoundationE4data3for8delegate7timeoutAC4DataV_So13NSURLResponseCtAC10URLRequestV_So0A12TaskDelegate_pSgSdtYaKF"

    static let urlSession: [String: String] = [
        "data": dataFor,
        "upload": uploadFor,
        "bytes": bytesFor,
        "download": downloadFor,
        "dataTask": dataTask,
        "uploadTask": uploadTask,
        "HTTPURLResponse": httpResponse,
    ]

    /* Every finding as its message (`body` or `discard`) and the text it covers, or its first line and `...` when it spans more. */
    static func findings(_ source: String, resolving: [String: String] = urlSession) -> [String] {
        let url = URL(fileURLWithPath: "/fixture/Subject.swift")
        let file = ParsedFile(
            url: url,
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        var occurrences: [FileSymbols.Occurrence] = []
        for token in file.tree.tokens(viewMode: .sourceAccurate) {
            let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            guard let symbol = resolving["\(location.line):\(token.text)"] ?? resolving[token.text] else { continue }
            occurrences.append(
                FileSymbols.Occurrence(
                    line: location.line,
                    column: location.column,
                    symbol: symbol,
                    name: token.text,
                    isReference: true,
                )
            )
        }
        let rule = CorrectnessRequireResponseStatusCheck()
        let found = rule.findings(in: file, symbols: FileSymbols(occurrences, ownedModules: ["Control"]))
        #expect(found.isEmpty || rule.applies(to: file), "the prefilter must never hide a finding")
        let lines = source.split(separator: "\n", omittingEmptySubsequences: false).map { Array($0.utf8) }
        return found.map { finding in
            #expect(finding.fixes.isEmpty && finding.suggestions.isEmpty, "the rule never fixes")
            let kind: String
            switch finding.messageId {
                case "bodyReadWithoutStatusCheck":
                    kind = "body"
                    #expect(finding.message == Self.bodyMessage)
                case "responseDiscarded":
                    kind = "discard"
                    #expect(finding.message == Self.discardMessage)
                default:
                    kind = "unknown \(finding.messageId)"
            }
            let line = lines[finding.line - 1]
            guard let endLine = finding.endLine, let endColumn = finding.endColumn, endLine == finding.line else {
                return "\(kind) \(String(decoding: line[(finding.column - 1)...], as: UTF8.self))..."
            }
            return "\(kind) \(String(decoding: line[(finding.column - 1)..<(endColumn - 1)], as: UTF8.self))"
        }
    }

    // MARK: The real sites

    /*
     ahraos-macos `Sources/AhraOs/ProfileImageCache.swift:169` on 2026-10-03, `fetchImage(from:using:)`: the portrait
     GET, which returns nil on a non-2xx so the caller negative-caches it. `check` is the guard on the cast and its
     status; the unchecked version drops it and names the response `_`.
     */
    static func portrait(response: String, check: String) -> String {
        """
        import AppKit

        @MainActor
        final class ProfileImageCache {
            private nonisolated static func fetchImage(from url: URL, using session: URLSession) async -> NSImage? {
                do {
                    let (data, \(response)) = try await session.data(from: url)
        \(check)
                    guard let image = NSImage(data: data) else {
                        portraitLog.error(
                            "portrait \\(url.path, privacy: .public) decode failed bytes=\\(data.count, privacy: .public)")
                        return nil
                    }
                    return image
                }
                catch {
                    portraitLog.error(
                        "portrait \\(url.path, privacy: .public) fetch failed: \\(error.localizedDescription, privacy: .public)")
                    return nil
                }
            }
        }

        """
    }

    static let portraitCheck = """
                    guard let http = response as? HTTPURLResponse, (200..<300).contains(http.statusCode) else {
                        return nil
                    }
        """

    @Test func portraitFetchAsItStandsChecksFirst() {
        #expect(
            Self.findings(
                Self.portrait(response: "response", check: Self.portraitCheck),
                resolving: Self.urlSession.merging(["data": Self.dataFrom]) { _, new in new },
            ).isEmpty
        )
    }

    @Test func portraitFetchWithoutItsCheckCachesAnErrorPageAsAPortrait() {
        /* A 404 or 500 body is handed to NSImage, and the decode-failure log reads it too, on its way to a normal return. */
        #expect(Self.findings(Self.portrait(response: "_", check: "")) == ["body data", "body data"])
    }

    /*
     ahraos-macos `Sources/AhraOs/AhraOsHttpClient.swift:303` on 2026-10-03, the lifecycle POST: the body is read only
     to describe a failure, on the branch the status check opened. `call` is the request's line.
     */
    static func lifecyclePost(_ call: String, check: Bool) -> String {
        let checked = """
                        if let http = response as? HTTPURLResponse, !(200..<300).contains(http.statusCode) {
                            let detail = String(decoding: data, as: UTF8.self)
                            FileHandle.standardError.write(
                                Data("[AhraOs] lifecycle POST \\(path) → HTTP \\(http.statusCode): \\(detail)\\n".utf8)
                            )
                            completion?(false)
                            return
                        }
            """
        return """
            import Foundation

            final class AhraOsHttpClient: @unchecked Sendable {
                private let baseUrl: URL
                private let urlSession: URLSession

                func postLifecycle(_ path: String, body: [String: Any], completion: (@Sendable (Bool) -> Void)?) {
                    let url = baseUrl.appendingPathComponent(path)
                    let httpBody = try? JSONSerialization.data(withJSONObject: body)
                    Task { [weak self] in
                        guard let self else { return }
                        var request = URLRequest(url: url)
                        request.httpMethod = "POST"
                        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
                        request.httpBody = httpBody
                        do {
            \(call)
            \(check ? checked : "")
                            completion?(true)
                        }
                        catch {
                            FileHandle.standardError.write(
                                Data("[AhraOs] lifecycle POST \\(path) failed: \\(error)\\n".utf8)
                            )
                            completion?(false)
                        }
                    }
                }
            }

            """
    }

    /* `JSONSerialization.data(withJSONObject:)` on line 9 is Foundation's, not URLSession's. */
    static let lifecycleResolving = urlSession.merging(["9:data": jsonData]) { _, new in new }

    @Test func lifecyclePostAsItStandsReadsTheBodyOnlyBehindTheCheck() {
        let source = Self.lifecyclePost(
            "                let (data, response) = try await self.urlSession.data(for: request)",
            check: true,
        )
        #expect(Self.findings(source, resolving: Self.lifecycleResolving).isEmpty)
    }

    @Test func lifecyclePostWithoutItsCheckReportsSuccessWhateverTheStatus() {
        let kept = Self.lifecyclePost(
            "                let (data, response) = try await self.urlSession.data(for: request)",
            check: false,
        )
        #expect(
            Self.findings(kept, resolving: Self.lifecycleResolving) == [
                "discard try await self.urlSession.data(for: request)"
            ]
        )
        let dropped = Self.lifecyclePost(
            "                _ = try await self.urlSession.data(for: request)",
            check: false,
        )
        #expect(
            Self.findings(dropped, resolving: Self.lifecycleResolving) == [
                "discard try await self.urlSession.data(for: request)"
            ]
        )
    }

    /*
     ahraos-macos `Sources/AhraOs/AhraOsHttpClient.swift:455` on 2026-10-03, `runEventStream`: the server-sent event
     stream, checked before the byte stream is read in a child `Task` the liveness monitor can cancel. Trimmed to the
     request, the check and the read.
     */
    static func eventStream(check: Bool, readInTask: Bool) -> String {
        let checked = """
                            if let http = response as? HTTPURLResponse, !(200..<300).contains(http.statusCode) {
                                httpLog.error(
                                    "SSE \\(url.lastPathComponent, privacy: .public) HTTP \\(http.statusCode, privacy: .public)")
                                throw URLError(.badServerResponse)
                            }
            """
        let read =
            readInTask
            ? """
                            let readTask = Task { () -> Void in
                                var lineBuffer = Data()
                                for try await byte in byteStream {
                                    if byte == 0x0A {
                                        onPayload(String(decoding: lineBuffer, as: UTF8.self))
                                        lineBuffer.removeAll(keepingCapacity: true)
                                    }
                                    else {
                                        lineBuffer.append(byte)
                                    }
                                }
                            }
                            try await readTask.value
            """
            : """
                            var lineBuffer = Data()
                            for try await byte in byteStream {
                                if byte == 0x0A {
                                    onPayload(String(decoding: lineBuffer, as: UTF8.self))
                                    lineBuffer.removeAll(keepingCapacity: true)
                                }
                                else {
                                    lineBuffer.append(byte)
                                }
                            }
            """
        return """
            import Foundation
            import os

            final class AhraOsHttpClient: @unchecked Sendable {
                private let urlSession: URLSession

                private func runEventStream(url: URL, expectedEvent: String, onPayload: @escaping @Sendable (String) -> Void) async {
                    while !Task.isCancelled {
                        do {
                            var request = URLRequest(url: url)
                            request.setValue("text/event-stream", forHTTPHeaderField: "Accept")
                            request.setValue("no-cache", forHTTPHeaderField: "Cache-Control")
                            let (byteStream, response) = try await urlSession.bytes(for: request)
            \(check ? checked : "")
            \(read)
                        }
                        catch {
                            await Task.sleepUnlessCancelled(nanoseconds: 1_000_000_000)
                        }
                    }
                }
            }

            """
    }

    @Test func eventStreamAsItStandsChecksBeforeTheStreamIsRead() {
        #expect(Self.findings(Self.eventStream(check: true, readInTask: true)).isEmpty)
    }

    @Test func eventStreamWithoutItsCheckIsAKnownMissWhileATaskReadsTheStream() {
        /* The child `Task` refers to the stream, so the pair is dropped whole: the graph cannot say when the task runs. */
        #expect(Self.findings(Self.eventStream(check: false, readInTask: true)).isEmpty)
    }

    @Test func eventStreamWithoutItsCheckReadInPlaceDeliversAnErrorPageAsEvents() {
        #expect(Self.findings(Self.eventStream(check: false, readInTask: false)) == ["body byteStream"])
        #expect(Self.findings(Self.eventStream(check: true, readInTask: false)).isEmpty)
    }

    /*
     ahraos-macos `Sources/AhraOsServices/ProviderProxy.swift:843` on 2026-10-03: the header fetch runs in its own task
     so the watchdog can cancel a stall, and the pair comes back through `headerFetch.value`, which the rule does not
     follow. The status is read after the cast in any case.
     */
    @Test func providerProxyHandsThePairBackThroughATask() {
        let source = """
            import Foundation

            final class ProviderProxy {
                func forward(session: URLSession, urlRequest: URLRequest, clientFd: Int32) async {
                    do {
                        let headerFetch = Task { try await session.bytes(for: urlRequest) }
                        let (byteStream, response) = try await headerFetch.value
                        guard let httpResponse = response as? HTTPURLResponse else {
                            byteStream.task.cancel()
                            return
                        }
                        if httpResponse.statusCode == 429 || httpResponse.statusCode == 529 {
                            for try await _ in byteStream {}
                            return
                        }
                        for try await byte in byteStream {
                            write(clientFd, byte)
                        }
                    }
                    catch {
                        return
                    }
                }
            }

            """
        #expect(Self.findings(source).isEmpty)
    }

    // MARK: What it reports

    @Test func aBodyDecodedAndReturned() {
        let source = """
            func load(session: URLSession, request: URLRequest) async throws -> Profile {
                let (data, _) = try await session.data(for: request)
                return try JSONDecoder().decode(Profile.self, from: data)
            }
            """
        #expect(Self.findings(source) == ["body data"])
    }

    @Test func aResponseKeptButOnlyItsUrlRead() {
        let source = """
            func load(url: URL) async throws -> Profile {
                let (data, response) = try await URLSession.shared.data(from: url)
                log(response.url, response.mimeType, response.expectedContentLength)
                return try JSONDecoder().decode(Profile.self, from: data)
            }
            """
        #expect(Self.findings(source) == ["body data"])
    }

    @Test func aCastWithNoStatusReadIsNotACheck() {
        let guarded = """
            func save(session: URLSession, request: URLRequest, body: Data) async throws -> Receipt {
                let (data, response) = try await session.upload(for: request, from: body)
                guard let http = response as? HTTPURLResponse else { throw ClientError.notHttp }
                log(http.allHeaderFields, http.value(forHTTPHeaderField: "ETag"))
                return try JSONDecoder().decode(Receipt.self, from: data)
            }
            """
        #expect(Self.findings(guarded) == ["body data"])
        let bound = """
            func save(session: URLSession, request: URLRequest, body: Data) async throws -> Receipt {
                let (data, response) = try await session.upload(for: request, from: body)
                let http = response as? HTTPURLResponse
                log(http?.suggestedFilename)
                return try JSONDecoder().decode(Receipt.self, from: data)
            }
            """
        #expect(Self.findings(bound) == ["body data"])
    }

    @Test func aBodyReadInlineOnAResponseNothingElseSees() {
        let source = """
            func load(session: URLSession, request: URLRequest) async throws -> Data {
                let first = try await session.data(for: request).0
                let second = (try await session.data(for: request)).0
                return first + second
            }
            """
        #expect(
            Self.findings(source) == [
                "body session.data(for: request).0", "body (try await session.data(for: request)).0",
            ]
        )
    }

    @Test func aCompletionHandlerThatDecodes() {
        let named = """
            func load(session: URLSession, request: URLRequest, completion: @escaping (Profile?) -> Void) {
                session.dataTask(with: request) { data, response, error in
                    guard let data, error == nil else {
                        completion(nil)
                        return
                    }
                    completion(try? JSONDecoder().decode(Profile.self, from: data))
                }.resume()
            }
            """
        #expect(Self.findings(named) == ["body data"])
        let typed = """
            func load(session: URLSession, request: URLRequest, body: Data, completion: @escaping (Data?) -> Void) {
                let task = session.uploadTask(with: request, from: body, completionHandler: { (data: Data?, response: URLResponse?, error: Error?) in
                    completion(data)
                })
                task.resume()
            }
            """
        #expect(Self.findings(typed) == ["body data"])
    }

    @Test func aBodyFieldTestedInPlaceOfTheStatus() {
        let source = """
            func load(session: URLSession, request: URLRequest) async throws -> Envelope {
                let (data, response) = try await session.data(for: request)
                let envelope = try JSONDecoder().decode(Envelope.self, from: data)
                if let message = envelope.error {
                    log((response as? HTTPURLResponse)?.statusCode)
                    throw ServerError.rejected(message)
                }
                return envelope
            }
            """
        #expect(Self.findings(source) == ["body data"])
    }

    @Test func aCheckOnOneBranchOnly() {
        let source = """
            func load(session: URLSession, request: URLRequest, verbose: Bool) async throws -> Data {
                let (data, response) = try await session.data(for: request)
                if verbose {
                    guard let http = response as? HTTPURLResponse, http.statusCode == 200 else { throw ClientError.failed }
                }
                return data
            }
            """
        #expect(Self.findings(source) == ["body data"])
    }

    @Test func aCatchThatReturnsNormally() {
        let source = """
            func load(session: URLSession, request: URLRequest) async -> Profile? {
                do {
                    let (data, _) = try await session.data(for: request)
                    return try JSONDecoder().decode(Profile.self, from: data)
                }
                catch {
                    return nil
                }
            }
            """
        #expect(Self.findings(source) == ["body data"])
    }

    @Test func aDownloadMovedIntoPlace() {
        let source = """
            func fetch(session: URLSession, request: URLRequest, destination: URL) async throws {
                let (location, _) = try await session.download(for: request)
                try FileManager.default.moveItem(at: location, to: destination)
            }
            """
        #expect(Self.findings(source) == ["body location"])
    }

    @Test func aLoopThatNeverLeavesSoOnlyTheNextTurnEndsThePair() {
        let source = """
            func poll(session: URLSession, request: URLRequest) async throws {
                while true {
                    let (data, _) = try await session.data(for: request)
                    results.append(data)
                }
            }
            """
        #expect(Self.findings(source) == ["body data"])
    }

    @Test func aPairAtFileScope() {
        let source = """
            let (data, _) = try await URLSession.shared.data(from: configurationUrl)
            let configuration = try JSONDecoder().decode(Configuration.self, from: data)
            """
        #expect(Self.findings(source) == ["body data"])
    }

    @Test func aResponseDiscarded() {
        let source = """
            func ping(session: URLSession, request: URLRequest, body: Data) async throws {
                _ = try await session.data(for: request)
                try await session.upload(for: request, from: body)
                let _ = try await session.data(for: request)
                let (_, _) = try await session.data(for: request)
                let (data, response) = try await session.data(for: request)
                log("sent")
            }
            """
        #expect(
            Self.findings(source) == [
                "discard try await session.data(for: request)",
                "discard try await session.upload(for: request, from: body)",
                "discard try await session.data(for: request)",
                "discard try await session.data(for: request)",
                "discard try await session.data(for: request)",
            ]
        )
    }

    @Test func aCompletionHandlerThatNamesNeitherBodyNorResponse() {
        let source = """
            func ping(session: URLSession, request: URLRequest, completion: @escaping (Bool) -> Void) {
                session.dataTask(with: request) { _, _, error in completion(error == nil) }.resume()
            }
            """
        #expect(
            Self.findings(source) == [
                "discard session.dataTask(with: request) { _, _, error in completion(error == nil) }"
            ]
        )
    }

    // MARK: What it leaves alone

    @Test func decodedThenCheckedBeforeTheDataIsUsed() {
        let source = """
            func load(session: URLSession, request: URLRequest) async throws -> Profile {
                let (data, response) = try await session.data(for: request)
                let profile = try JSONDecoder().decode(Profile.self, from: data)
                guard (response as? HTTPURLResponse)?.statusCode == 200 else { throw ClientError.failed }
                return profile
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func aStatusReadOnEveryPath() {
        let source = """
            func load(session: URLSession, request: URLRequest) async throws -> Data? {
                let (data, response) = try await session.data(for: request)
                if let http = response as? HTTPURLResponse, http.statusCode == 204 {
                    return nil
                }
                return data
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func checkedInTheTernaryThatReturnsIt() {
        let source = """
            func load(session: URLSession, request: URLRequest) async throws -> Data? {
                let (data, response) = try await session.data(for: request)
                return (response as? HTTPURLResponse)?.statusCode == 200 ? data : nil
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func checkedInASwitchOnTheStatus() {
        let source = """
            func load(session: URLSession, request: URLRequest) async throws -> Data {
                let (data, response) = try await session.data(for: request)
                switch (response as? HTTPURLResponse)?.statusCode {
                case 200:
                    return data
                default:
                    throw ClientError.failed
                }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func readToBeThrownNeverTrusted() {
        let source = """
            func load(session: URLSession, request: URLRequest) async throws -> Never {
                let (data, _) = try await session.data(for: request)
                throw ServerError.body(String(decoding: data, as: UTF8.self))
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func handedToAHelperThatChecks() {
        let source = """
            func load(session: URLSession, request: URLRequest) async throws -> Data {
                let (data, response) = try await session.data(for: request)
                try validate(response)
                return data
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func returnedToTheCaller() {
        let source = """
            func load(session: URLSession, request: URLRequest) async throws -> (Data, URLResponse) {
                let (data, response) = try await session.data(for: request)
                return (data, response)
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func aClosureReadsItAndTheGraphCannotSayWhen() {
        let closure = """
            func load(session: URLSession, request: URLRequest) async throws -> Data {
                let (data, response) = try await session.data(for: request)
                let describe = { response.url }
                log(describe)
                return data
            }
            """
        #expect(Self.findings(closure).isEmpty)
        let deferred = """
            func load(session: URLSession, request: URLRequest) async throws -> Data {
                let (data, response) = try await session.data(for: request)
                defer { log(response) }
                return data
            }
            """
        #expect(Self.findings(deferred).isEmpty)
    }

    @Test func aCastThatFailsIsANonHttpUrlWithNoStatus() {
        let source = """
            func load(session: URLSession, request: URLRequest) async throws -> Data {
                let (data, response) = try await session.data(for: request)
                if let http = response as? HTTPURLResponse {
                    guard http.statusCode == 200 else { throw ClientError.failed }
                }
                return data
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func aCompletionHandlerThatChecks() {
        let source = """
            func load(session: URLSession, request: URLRequest, completion: @escaping (Profile?) -> Void) {
                session.dataTask(with: request) { data, response, error in
                    if data == nil {
                        completion(nil)
                        return
                    }
                    guard error == nil, let data, let http = response as? HTTPURLResponse, (200..<300).contains(http.statusCode) else {
                        completion(nil)
                        return
                    }
                    completion(try? JSONDecoder().decode(Profile.self, from: data))
                }.resume()
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func theResponseReadInlineWithNoBody() {
        let source = """
            func head(session: URLSession, request: URLRequest) async throws -> String? {
                let response = try await session.data(for: request).1
                let (_, again) = try await session.data(for: request)
                return (again as? HTTPURLResponse)?.value(forHTTPHeaderField: "ETag") ?? response.mimeType
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func eachRequestInALoopChecked() {
        let source = """
            func loadAll(session: URLSession, requests: [URLRequest]) async throws -> [Data] {
                var results: [Data] = []
                for request in requests {
                    let (data, response) = try await session.data(for: request)
                    guard let http = response as? HTTPURLResponse, http.statusCode == 200 else { continue }
                    results.append(data)
                }
                return results
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func aMethodOfOursWithURLSessionsName() {
        let source = """
            func load(client: ApiClient, request: URLRequest) async throws -> Profile {
                let (data, _) = try await client.data(for: request)
                return try JSONDecoder().decode(Profile.self, from: data)
            }
            """
        #expect(
            Self.findings(source, resolving: Self.urlSession.merging(["data": Self.ourData]) { _, new in new }).isEmpty
        )
        #expect(
            Self.findings(source, resolving: Self.urlSession.merging(["data": Self.longerData]) { _, new in new })
                .isEmpty
        )
    }

    @Test func aCastToAnotherTypeOrAnIsTestHandsTheResponseOn() {
        let source = """
            func load(session: URLSession, request: URLRequest) async throws -> Data {
                let (data, response) = try await session.data(for: request)
                guard response is HTTPURLResponse else { throw ClientError.notHttp }
                return data
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func bindingsTheRuleDoesNotFollow() {
        let source = """
            func load(session: URLSession, request: URLRequest) async throws -> Data {
                let result = try await session.data(for: request)
                var (data, response) = try await session.data(for: request)
                async let pending = session.data(for: request)
                session.dataTask(with: request, completionHandler: handle).resume()
                return result.0 + data + (try await pending).0
            }
            """
        /* A single name, a `var`, an `async let` (whose `.0` reads `pending`, not a call) and a handler passed by name: known misses. */
        #expect(Self.findings(source).isEmpty)
    }

    @Test func theCallAloneAsABodyMayBeItsReturn() {
        let source = """
            func fetch(session: URLSession, request: URLRequest) -> Task<(Data, URLResponse), any Error> {
                Task { try await session.data(for: request) }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func theTypeOfTheCastIsResolvedNotSpelled() {
        /* An `HTTPURLResponse` the index did not resolve to Foundation's is some other type, so the cast hands the response on. */
        let source = """
            func load(session: URLSession, request: URLRequest) async throws -> Data {
                let (data, response) = try await session.data(for: request)
                guard let http = response as? HTTPURLResponse else { throw ClientError.notHttp }
                log(http.url)
                return data
            }
            """
        #expect(Self.findings(source) == ["body data"])
        #expect(Self.findings(source, resolving: Self.urlSession.filter { $0.key != "HTTPURLResponse" }).isEmpty)
    }

    // MARK: End to end

    static let packageSource = #"""
        import Foundation

        public struct Profile: Decodable, Sendable {
            public let name: String
        }

        public enum Client {
            public static func unchecked(session: URLSession, request: URLRequest) async throws -> Profile {
                let (data, _) = try await session.data(for: request)
                return try JSONDecoder().decode(Profile.self, from: data)
            }

            public static func checked(session: URLSession, request: URLRequest) async throws -> Profile {
                let (data, response) = try await session.data(for: request)
                guard let http = response as? HTTPURLResponse, (200..<300).contains(http.statusCode) else {
                    throw URLError(.badServerResponse)
                }
                return try JSONDecoder().decode(Profile.self, from: data)
            }

            public static func discarded(session: URLSession, request: URLRequest) async throws {
                _ = try await session.upload(for: request, from: Data())
            }

            public static func handler(session: URLSession, request: URLRequest, completion: @escaping @Sendable (Data?) -> Void) {
                session.dataTask(with: request) { data, _, _ in
                    completion(data)
                }.resume()
            }

            public static func streamed(session: URLSession, url: URL) async throws -> Int {
                let (stream, response) = try await session.bytes(from: url)
                var count = 0
                for try await _ in stream {
                    count += 1
                }
                guard let http = response as? HTTPURLResponse, http.statusCode == 200 else {
                    throw URLError(.badServerResponse)
                }
                return count
            }

            public static func downloaded(session: URLSession, url: URL) async throws -> URL {
                let (location, response) = try await session.download(from: url)
                guard (response as? HTTPURLResponse)?.statusCode == 200 else {
                    throw URLError(.badServerResponse)
                }
                return location
            }
        }

        """#

    @Test func theIndexResolvesURLSessionsMethodsAndHTTPURLResponse() async throws {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-typed-\(UUID().uuidString)",
            isDirectory: true,
        )
        let sources = root.appendingPathComponent("Sources/Control", isDirectory: true)
        try FileManager.default.createDirectory(at: sources, withIntermediateDirectories: true)
        try PipelineControlTests.manifest.write(
            to: root.appendingPathComponent("Package.swift"),
            atomically: true,
            encoding: .utf8,
        )
        try Self.packageSource.write(
            to: sources.appendingPathComponent("Control.swift"),
            atomically: true,
            encoding: .utf8,
        )

        /* One run builds the package and writes its index. The rule is run by hand on what the run left. */
        let options = try CommandOptions.parse(
            ["--contract", "\(EngineVersion.contract)", "--root", root.path, "--no-fix"],
            workingDirectory: root,
        )
        _ = try await Pipeline(options: options, writer: ContractWriter { _ in }, workingDirectory: root).run()
        let package = try PackageModel.load(
            root: root,
            scratchPath: Pipeline.scratchPath(for: root),
            runner: ProcessRunner(),
        )
        let parsed = await SourceParser().parse(try FileSet.build(package: package).owned)
        let rule = CorrectnessRequireResponseStatusCheck()
        let candidates = parsed.files.filter { rule.applies(to: $0) }
        let symbols = SymbolProvider(
            scratchPaths: Pipeline.symbolScratchPaths(package: package, root: root),
            runner: ProcessRunner(),
        ).symbols(for: candidates)
        #expect(symbols.fromIndex == 1, "the build's index should describe the file: \(symbols.unavailable)")
        let found = candidates.flatMap { file in
            rule.findings(in: file, symbols: symbols.symbols[file.url.path] ?? FileSymbols([])).map {
                "\($0.line):\($0.column) \($0.messageId)"
            }
        }
        /*
         `unchecked` (the decode), `discarded` (the upload's `_ =`), `handler` (the completion handler's body) and
         `streamed` (the stream counted before the status is read, and the count returned on a path that reads it
         only afterwards, so silent); not `checked` and not `downloaded`.
         */
        #expect(
            found == [
                "10:61 bodyReadWithoutStatusCheck", "22:13 responseDiscarded", "27:24 bodyReadWithoutStatusCheck",
            ],
            "\(found)",
        )
    }
}
