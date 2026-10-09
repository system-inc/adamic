package nexus

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const propertyAliasFile = "/repository/source/Thing.ts"

// Every case runs through the typed harness, because four exemptions ask the checker and production
// always has one.
//
// The first two cases are the shapes of the only two live violations in the ahra tree, both of
// which carry an `eslint-disable-next-line` for this rule. They are the closest thing to an answer
// key this rule has: it is ours, so there is no upstream corpus, and the tree reports zero findings
// precisely because those two sites are suppressed.

func TestConsistencyNoPropertyAliasFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"the real Run.ts shape", "export function run(options: OptionsInterface) {\n    const onOutputLine = options.onOutputLine;\n    return onOutputLine;\n}\n"},
		{"the real TrackedPromise shape", "export function run(tracked: TrackedInterface) {\n    const promise = tracked.promise;\n    return promise;\n}\n"},
		{"a nested reach", "export function run(state: StateInterface) {\n    const value = state.inner.value;\n    return value;\n}\n"},
		{"a long reach with no call anywhere in it", "export function run(a: AInterface) {\n    const value = a.one.two.three.value;\n    return value;\n}\n"},
		{"inside an arrow function", "export const Run = (options: OptionsInterface) => {\n    const timeout = options.timeout;\n    return timeout;\n};\n"},
		{"inside a method", "export class Thing {\n    run(options: OptionsInterface) {\n        const limit = options.limit;\n        return limit;\n    }\n}\n"},
		{"with a let binding", "export function run(options: OptionsInterface) {\n    let cursor = options.cursor;\n    return cursor;\n}\n"},
		// The hatch is for hooks, and only hooks. This rule reports when it fails to find the
		// exemption, so its polarity is inverted from most: a hatch that is too WIDE loses
		// findings quietly, and a hatch that is too NARROW manufactures accusations against
		// correct files. 117 of the 231 alias-shaped declarations on the ahra tree are exempted by
		// this hatch alone, so both directions get a case.
		// The two exemptions of #nd52037, each from the side that must still report. A write to the
		// source before the snapshot, or after its last read, leaves the reach reading what the local
		// holds, and a write to some other property does not touch the reach at all.
		{"the source is written before the declaration", "export class Runner {\n    server: ServerInterface | undefined;\n    stop() {\n        this.server = undefined;\n        const server = this.server;\n        return server;\n    }\n}\n"},
		{"the source is written after the last read", "export class Runner {\n    server: ServerInterface | undefined;\n    stop() {\n        const server = this.server;\n        close(server);\n        this.server = undefined;\n    }\n}\n"},
		{"a different property is written", "export class Runner {\n    server: ServerInterface | undefined;\n    other: number = 0;\n    stop() {\n        const server = this.server;\n        this.other = 1;\n        close(server);\n    }\n}\n"},
		{"a longer reach through the local is written", "export function run(state: StateInterface) {\n    const inner = state.inner;\n    state.inner.value = 1;\n    return inner;\n}\n"},
		{"a let only read", "export function run(options: OptionsInterface) {\n    let label = options.label;\n    use(label);\n    return label;\n}\n"},
		// The four of #r28b8he, from the side that must still report: a plain property is not a
		// getter, a local read only in its own function keeps no narrowing a closure would lose, a
		// narrowed local read nowhere nested is still an alias, and an annotation naming the
		// source's own type does no work.
		{"a plain property, not a getter", "class Account {\n    email = '';\n}\nexport function run(account: Account) {\n    const email = account.email;\n    return email;\n}\n"},
		{"read in a closure with no narrowing", "class Runner {\n    server: { close(): void } | undefined;\n    stop(queue: (callback: () => void) => void) {\n        const server = this.server;\n        queue(function() {\n            use(server);\n        });\n    }\n}\ndeclare function use(value: unknown): void;\nexport { Runner };\n"},
		{"narrowed but read only in its own function", "class Runner {\n    server: { close(): void } | undefined;\n    stop() {\n        if(this.server) {\n            const server = this.server;\n            server.close();\n        }\n    }\n}\nexport { Runner };\n"},
		{"an annotation naming the source's own type", "interface Options {\n    timeout: number;\n}\nexport function run(options: Options) {\n    const timeout: number = options.timeout;\n    return timeout;\n}\n"},
		// A member name that shares the local's spelling is not the local (#r28b8he). Each of these was
		// read as one by a walk matching names, which exempted an alias the reach replaces exactly.
		{"a closure reads the property, not the narrowed local", "class Runner {\n    server: { close(): void } | undefined;\n    stop(queue: (callback: () => void) => void) {\n        if(this.server) {\n            const server = this.server;\n            server.close();\n            queue(() => use(this.server));\n        }\n    }\n}\ndeclare function use(value: unknown): void;\nexport { Runner };\n"},
		{"a later read of the property is not a read of the snapshot", "class Runner {\n    httpServer: { close(): void } | undefined;\n    stop() {\n        const httpServer = this.httpServer;\n        use(httpServer);\n        this.httpServer = undefined;\n        use(this.httpServer);\n    }\n}\ndeclare function use(value: unknown): void;\nexport { Runner };\n"},
		{"a dependency array naming another object's property of that name", "export function run(options: OptionsInterface, other: OptionsInterface) {\n    const secret = options.secret;\n    useMemo(function() {\n        return 1;\n    }, [other.secret]);\n    return secret;\n}\n"},
		// Narrowed after the declaration but read only in its own function, where a reach narrows too.
		{"narrowed after the declaration, read outside any closure", "interface Input {\n    attribute: { key: string } | null;\n}\nexport function run(input: Input) {\n    const attribute = input.attribute;\n    if(attribute) {\n        return attribute.key;\n    }\n    return '';\n}\n"},
		{"an array argument to something that is not a hook", "export function run(options: OptionsInterface) {\n    const secret = options.secret;\n    notAHook(function() {\n        return 1;\n    }, [secret]);\n    return secret;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, ConsistencyNoPropertyAlias, propertyAliasFile, testCase.sourceText),
				"noPropertyAlias")
		})
	}
}

// The clean half is the specification. Every exemption here is a judgment rather than a narrowing,
// and each is the kind that reads as a bug until its reason is stated.

func TestConsistencyNoPropertyAliasStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"the names differ", "export function run(options: OptionsInterface) {\n    const value = options.other;\n    return value;\n}\n"},
		{"a call in the chain", "export function run() {\n    const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;\n    return timeZone;\n}\n"},
		{"a call at the head of the chain", "export function run(make: MakeInterface) {\n    const value = make().value;\n    return value;\n}\n"},
		// The call sits two links up rather than beside the property, which is what makes the
		// descent load-bearing. A mutation sweep produced this case: stopping the walk after one
		// link left every other fixture green, because each of them had its call at the head or
		// one step in. The chain still computes, so the local is still caching rather than
		// aliasing, and the exemption has to survive the extra depth.
		{"a call deeper in the chain", "export function run(a: AInterface) {\n    const value = a.make().inner.value;\n    return value;\n}\n"},
		// The call sits BELOW an element access, which is the only shape that exercises the
		// element-access arm of the chain walk. It is here because reasoning said the arm was dead
		// and a probe said otherwise: `a[k].make()` finds its call before ever descending, but
		// `make()[k]` does not, so deleting the arm turns a computed-caching case into a reported
		// alias. Argument lost to measurement.
		{"a call below an element access", "export function run(make: MakeInterface, key: string) {\n    const value = make()[key].value;\n    return value;\n}\n"},
		// Parentheses around a computing chain. The exemption has to see through them, because a
		// reader who wrapped the call in parentheses did not stop it computing.
		{"a parenthesized call", "export function run(make: MakeInterface) {\n    const value = (make()).value;\n    return value;\n}\n"},
		{"computed access", "export function run(record: RecordInterface, name: string) {\n    const key = record[name];\n    return key;\n}\n"},
		// Optional chains, which are 78 of the 231 alias-shaped declarations on the ahra tree and
		// the single largest exempted group. ESTree wraps `a?.b` in a ChainExpression, so the
		// original's `init.type !== 'MemberExpression'` test rejects it before doing anything;
		// typescript-go has no wrapper, so the exemption has to be written here or the rule
		// accuses 78 correct files.
		{"an optional access", "export function run(options: OptionsInterface) {\n    const value = options?.value;\n    return value;\n}\n"},
		// The token sits two links in rather than at the outermost access. An exemption that
		// checked only the outer link left exactly this shape reporting on the real tree.
		{"an optional link deeper in the chain", "export function run(account: AccountInterface) {\n    const displayName = account.data?.profile.displayName;\n    return displayName;\n}\n"},
		{"an optional element access", "export function run(o: OInterface, index: string) {\n    const countryCode = o.locales[index]?.countryCode;\n    return countryCode;\n}\n"},
		{"at module scope", "export const Foo = Bar.Foo;\n"},
		{"no initializer", "export function run() {\n    let value;\n    return value;\n}\n"},
		{"a destructure rather than a reach", "export function run(options: OptionsInterface) {\n    const { timeout } = options;\n    return timeout;\n}\n"},
		{"an initializer that is not a reach", "export function run() {\n    const value = 1;\n    return value;\n}\n"},
		// A snapshot taken before its source is written (#nd52037). The first is api's
		// BaseWorkerNodeRunner.ts:109, where a reach at the read would see undefined.
		{"a snapshot before its property is cleared", "export class Runner {\n    httpServer: ServerInterface | undefined;\n    stop() {\n        const httpServer = this.httpServer;\n        this.httpServer = undefined;\n        close(httpServer);\n    }\n}\n"},
		{"a snapshot before an object above it is replaced", "export function run(state: StateInterface, other: InnerInterface) {\n    const value = state.inner.value;\n    state.inner = other;\n    return value;\n}\n"},
		{"a snapshot before its root is reassigned", "export function run(options: OptionsInterface, fallback: OptionsInterface) {\n    const timeout = options.timeout;\n    options = fallback;\n    return timeout;\n}\n"},
		{"a snapshot before its property is deleted", "export function run(record: RecordInterface) {\n    const value = record.value;\n    delete record.value;\n    return value;\n}\n"},
		{"a snapshot before its property is incremented", "export function run(counter: CounterInterface) {\n    const count = counter.count;\n    counter.count++;\n    return count;\n}\n"},
		{"a snapshot before a compound write through a cast", "export function run(counter: CounterInterface) {\n    const count = counter.count;\n    (counter as WritableInterface).count += 1;\n    return count;\n}\n"},
		// A binding written again after its declaration (#nd52037): a default a branch overrides.
		{"a let reassigned in a branch", "export function run(options: OptionsInterface, compact: boolean) {\n    let label = options.label;\n    if (compact) {\n        label = 'short';\n    }\n    return label;\n}\n"},
		{"a let updated by a compound assignment", "export function run(options: OptionsInterface) {\n    let total = options.total;\n    total += 1;\n    return total;\n}\n"},
		{"a let written by destructuring", "export function run(options: OptionsInterface) {\n    let cursor = options.cursor;\n    [cursor] = next();\n    return cursor;\n}\n"},
		{"a var reassigned", "export function run(options: OptionsInterface) {\n    var limit = options.limit;\n    limit = 10;\n    return limit;\n}\n"},
		// The four of #r28b8he, locals the function makes.
		// A cast in the chain is a narrowed binding; this was the Run.ts shape, which reported and
		// carried a disable comment until the ruling on #zh8mpvp.
		{"a reach through a cast", "export function run(options: OptionsInterface) {\n    const onOutputLine = (options as ProgramOptionsInterface).onOutputLine;\n    return onOutputLine;\n}\n"},
		{"a reach through an angle-bracket cast, deeper", "export function run(value: unknown) {\n    const size = (<ShapeInterface>value).inner.size;\n    return size;\n}\n"},
		{"a getter source", "class Account {\n    get email() {\n        return 'a@b.c';\n    }\n}\nexport function run(account: Account) {\n    const email = account.email;\n    return email;\n}\n"},
		{"a getter above the property", "class Configuration {\n    get runtime() {\n        return { mode: 'test' };\n    }\n}\nexport function run(configuration: Configuration) {\n    const mode = configuration.runtime.mode;\n    return mode;\n}\n"},
		{"a narrowing a closure would lose", "class Runner {\n    server: { close(): void } | undefined;\n    stop(queue: (callback: () => void) => void) {\n        if(this.server) {\n            const server = this.server;\n            queue(function() {\n                server.close();\n            });\n        }\n    }\n}\nexport { Runner };\n"},
		{"a narrowing of an object above, read in an arrow", "interface Holder {\n    inner?: { value: number };\n}\nexport function run(holder: Holder, queue: (callback: () => void) => void) {\n    if(holder.inner) {\n        const value = holder.inner.value;\n        queue(() => use(value));\n    }\n}\ndeclare function use(value: unknown): void;\n"},
		// The narrowing comes after the declaration, on the local, and a closure keeps it (#55sfn9q).
		// api's two sites, where tsgo refuses the reach: a truthiness check, and a throw on its negation.
		{"narrowed after the declaration by if, read in a callback", "interface Input {\n    attributeToUpsert: { key: string } | null;\n}\nexport function run(input: Input, attributes: { key: string }[]) {\n    const attributeToUpsert = input.attributeToUpsert;\n    if(attributeToUpsert) {\n        return attributes.find((attribute) => attribute.key === attributeToUpsert.key);\n    }\n    return undefined;\n}\n"},
		{"narrowed after the declaration by a throw, read in a callback", "interface Response {\n    mention: { platform: string } | undefined;\n}\nexport function run(response: Response, services: { platformType: string }[]) {\n    const mention = response.mention;\n    if(!mention) throw new Error('no mention');\n    return services.find((service) => service.platformType === mention.platform);\n}\n"},
		{"an annotation widening to unknown", "interface Event {\n    data: any;\n}\nexport function run(event: Event) {\n    const data: unknown = event.data;\n    return data;\n}\n"},
		{"an annotation giving a readonly view", "interface Registry {\n    modules: string[];\n}\nexport function run(registry: Registry) {\n    const modules: readonly string[] = registry.modules;\n    return modules;\n}\n"},
		{"read in a useEffect dependency array", "export function run(options: OptionsInterface) {\n    const timeout = options.timeout;\n    React.useEffect(function() {\n        report(timeout);\n    }, [timeout]);\n}\n"},
		{"read in a bare hook dependency array", "export function run(options: OptionsInterface) {\n    const timeout = options.timeout;\n    useMemo(function() {\n        return timeout;\n    }, [timeout]);\n}\n"},
		{"read inside a longer reach in a dependency array", "export function run(options: OptionsInterface) {\n    const settings = options.settings;\n    useMemo(function() {\n        return 1;\n    }, [settings.value]);\n}\n"},
		{"read as a deep reach in a dependency array", "export function run(options: OptionsInterface) {\n    const secret = options.secret;\n    useMemo(function() {\n        return 1;\n    }, [secret.id.value]);\n}\n"},
		{"read in a dependency array inside a nested callback", "export function run(options: OptionsInterface) {\n    const secret = options.secret;\n    wrap(function() {\n        useMemo(function() {\n            return 1;\n        }, [secret]);\n    });\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, ConsistencyNoPropertyAlias, propertyAliasFile, testCase.sourceText))
		})
	}
}
