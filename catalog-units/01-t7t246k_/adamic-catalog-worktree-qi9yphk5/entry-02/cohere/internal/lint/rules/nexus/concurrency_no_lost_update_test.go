package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const concurrencyNoLostUpdateFile = "/repository/source/ConcurrencyNoLostUpdate.ts"

// concurrencyNoLostUpdatePrelude declares every name the cases use, so the checker resolves each one
// and no verdict rests on an unresolved identifier.
var concurrencyNoLostUpdatePrelude = strings.Join([]string{
	"declare function next(): Promise<number>;",
	"declare function save(value?: unknown): Promise<void>;",
	"declare function price(item?: unknown): Promise<number>;",
	"declare function load(key?: string): Promise<number[]>;",
	"declare function run(): Promise<void>;",
	"declare function use(value: unknown): void;",
	"declare function register(callback: () => void): void;",
	"declare function getPosition(): Promise<number>;",
	"declare function sign(): Promise<string>;",
	"declare function apply(callback: () => number): Promise<number>;",
	"declare function total(value: { before: number }): number;",
	"declare function convert<T>(value: number): T;",
	"declare const table: number[];",
	"declare function wrap(inner: () => number, weight: number): () => number;",
	"let render: () => number = () => 0;",
	"let mirror = 0;",
	"interface ListNode { next: ListNode | undefined }",
	"let head: ListNode | undefined;",
	"declare const flag: boolean;",
	"declare const key: string;",
	"declare const stream: AsyncIterable<number>;",
	"declare const xs: number[];",
	"declare const console: { log(...values: unknown[]): void; info(...values: unknown[]): void };",
	"interface Entry { retries: number; position: number; status: string }",
	"declare const entries: Entry[];",
	"let count = 0;",
	"let blockoutBottomRow = 0;",
	"let tokenCache: { token: string; expiresAt: number } | undefined;",
	"const state = {",
	"    count: 0, total: 0, other: 0, lastRunAt: 0, items: [] as number[],",
	"    cache: undefined as number[] | undefined, byKey: {} as Record<string, number>,",
	"};",
	"interface Account { oauth2: { token: string; ads: string } }",
	"declare function getAccount(key: string): Account;",
	"declare function updateAccount(key: string, patch: unknown): void;",
	"declare function updateProfile(key: string, patch: unknown): void;",
	"declare function loadSettings(): { theme: string; runs: number };",
	"declare function saveSettings(settings: unknown): void;",
	"declare function loadRemote(key: string): Promise<Account>;",
	"declare function saveRemote(key: string, value: unknown): Promise<void>;",
	"declare function get(key: string): number;",
	"declare function set(key: string, value: number): void;",
	"const cache = new Map<string, number>();",
	"",
}, "\n")

// concurrencyNoLostUpdatePhiSocialTerminal is `withPhiStdoutFilter` from
// `modules/phi/social/PhiSocialTerminal.ts:430` at HEAD on 2026-10-01, verbatim apart from the
// declarations its imports supply. `ArtTerminal.ts` carries the same function. It reported three times
// before the save-and-restore condition, once per restored handler, and must stay silent.
var concurrencyNoLostUpdatePhiSocialTerminal = strings.Join([]string{
	"declare const console: { log(...values: unknown[]): void; info(...values: unknown[]): void };",
	"declare const process: { stdout: { write(chunk: string | Uint8Array): boolean } };",
	"declare function stringifyArgument(value: unknown): string;",
	"declare function bytesToUtf8(bytes: Uint8Array): string;",
	"type PhiStdoutFilterHandlerInterface = (line: string) => string | undefined;",
	"export async function withPhiStdoutFilter<T>(",
	"    handler: PhiStdoutFilterHandlerInterface,",
	"    callback: () => Promise<T>,",
	"): Promise<T> {",
	"    const originalLog = console.log;",
	"    const originalInfo = console.info;",
	"    const originalStdoutWrite = process.stdout.write.bind(process.stdout);",
	"",
	"    let buffer = '';",
	"",
	"    function emit(text: string): void {",
	"        if(text.length > 0 && !text.includes('\\n')) {",
	"            const rewritten = handler(text);",
	"            if(rewritten === undefined) {",
	"                originalStdoutWrite(text);",
	"            }",
	"            else if(rewritten.length > 0) {",
	"                originalStdoutWrite(rewritten);",
	"            }",
	"            return;",
	"        }",
	"        buffer += text;",
	"        let newlineIndex = buffer.indexOf('\\n');",
	"        while(newlineIndex !== -1) {",
	"            const rawLine = buffer.slice(0, newlineIndex);",
	"            buffer = buffer.slice(newlineIndex + 1);",
	"            const rewritten = handler(rawLine);",
	"            if(rewritten === undefined) {",
	"                originalStdoutWrite(rawLine + '\\n');",
	"            }",
	"            else if(rewritten.length > 0) {",
	"                originalStdoutWrite(rewritten + '\\n');",
	"            }",
	"            newlineIndex = buffer.indexOf('\\n');",
	"        }",
	"    }",
	"",
	"    console.log = function filteredLog(...arguments_: unknown[]) {",
	"        emit(arguments_.map(stringifyArgument).join(' ') + '\\n');",
	"    };",
	"    console.info = function filteredInfo(...arguments_: unknown[]) {",
	"        emit(arguments_.map(stringifyArgument).join(' ') + '\\n');",
	"    };",
	"    process.stdout.write = function filteredWrite(chunk: string | Uint8Array) {",
	"        const text = typeof chunk === 'string' ? chunk : bytesToUtf8(chunk);",
	"        emit(text);",
	"        return true;",
	"    } as typeof process.stdout.write;",
	"",
	"    try {",
	"        return await callback();",
	"    }",
	"    finally {",
	"        if(buffer.length > 0) {",
	"            const rewritten = handler(buffer);",
	"            if(rewritten === undefined) originalStdoutWrite(buffer);",
	"            else if(rewritten.length > 0) originalStdoutWrite(rewritten + '\\n');",
	"            buffer = '';",
	"        }",
	"        console.log = originalLog;",
	"        console.info = originalInfo;",
	"        process.stdout.write = originalStdoutWrite;",
	"    }",
	"}",
	"",
}, "\n")

func concurrencyNoLostUpdateSource(lines ...string) string {
	return concurrencyNoLostUpdatePrelude + strings.Join(lines, "\n") + "\n"
}

// Each case is a whole file after the prelude and the number of findings it must produce. The
// reporting half is Kirk's four shapes from the ruling, then the paths that reach them; the silent
// half starts with the shapes upstream `require-atomic-updates` reported on ahra (39 findings, none a
// read-modify-write), reduced, then the near misses the conditions exist to separate.
func TestConcurrencyNoLostUpdate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		findings int
	}{
		// Reporting.
		{"x = x + await y", concurrencyNoLostUpdateSource(
			"export async function f() { count = count + await next(); }"), 1},
		{"x += await y", concurrencyNoLostUpdateSource(
			"export async function f() { count += await next(); }"), 1},
		{"a const read before the await, written after", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; await save(); state.count = old + 1; }"), 1},
		{"s.total = s.total + (await f())", concurrencyNoLostUpdateSource(
			"export async function f() { state.total = state.total + (await price()); }"), 1},
		{"this.count += await, in a method", concurrencyNoLostUpdateSource(
			"export class Counter { count = 0; async add() { this.count += await next(); } }"), 1},
		{"a spread of the cache before the await", concurrencyNoLostUpdateSource(
			"export class Cache { entries: Record<string, number[]> = {};",
			"    async fill() { this.entries = { ...this.entries, [key]: await load(key) }; } }"), 1},
		{"an array spread before the await", concurrencyNoLostUpdateSource(
			"export async function f() { state.items = [...state.items, ...(await load())]; }"), 1},
		{"a destructured const", concurrencyNoLostUpdateSource(
			"export async function f() { const { count: before } = state; await save(); state.count = before + 1; }"), 1},
		{"a chain of consts", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; const following = old + 1; await save(); state.count = following; }"), 1},
		{"a parameter's object, which the caller holds", concurrencyNoLostUpdateSource(
			"export async function f(entry: Entry) { entry.retries = entry.retries + await next(); }"), 1},
		{"a suspension on one branch is enough", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; if (flag) { await save(); } state.count = old + 1; }"), 1},
		{"a yield suspends", concurrencyNoLostUpdateSource(
			"export function* g() { const old = state.count; yield 1; state.count = old + 1; }"), 1},
		{"a read before a yield in the same expression", concurrencyNoLostUpdateSource(
			"export function* g(): Generator<number, void, number> { state.count = state.count + (yield 1); }"), 1},
		{"a for await loop suspends", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; for await (const x of stream) { use(x); } state.count = old + 1; }"), 1},
		{"a local a nested function writes", concurrencyNoLostUpdateSource(
			"export async function f() { let total = 0; register(() => { total = 5; }); total = total + await next(); }"), 1},
		{"the read precedes a fork that suspends", concurrencyNoLostUpdateSource(
			"export async function f() { state.count = state.count + (flag ? await next() : 1); }"), 1},
		{"a method called on the stale value", concurrencyNoLostUpdateSource(
			"export async function f() { const items = state.items; await save(); state.items = items.filter((item) => item > 0); }"), 1},
		{"through a try that suspends", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; try { await save(); } catch { use(0); } state.count = old + 1; }"), 1},
		{"a compound write through a computed key reads and writes one reference", concurrencyNoLostUpdateSource(
			"export async function f() { state.byKey[key] += await next(); }"), 1},
		{"a write to a sibling property is not a replacement", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; state.other = 0; await save(); state.count = old + 1; }"), 1},
		{"a write inside the target is not a replacement of it", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.byKey; state.byKey.extra = 1; await save(); state.byKey = { ...old, more: 2 }; }"), 1},
		{"a computed key on a parameter's object", concurrencyNoLostUpdateSource(
			"export async function f(tallies: Record<string, number>) { tallies[key] += await next(); }"), 1},
		{"a quoted key is a static path", concurrencyNoLostUpdateSource(
			"export async function f() { state['count'] = state['count'] + await next(); }"), 1},
		{"inside a callback, an outer variable", concurrencyNoLostUpdateSource(
			"export async function f() { await Promise.all(xs.map(async (x) => { count += await price(x); })); }"), 1},
		{"a const used after a suspension inside the expression", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; state.count = (await next()) + old; }"), 1},
		{"a loop writing a const from before it", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; for (const x of xs) { state.count = old + x; await save(); } }"), 1},
		{"a read inside the awaited operand", concurrencyNoLostUpdateSource(
			"export async function f() { count = (await count) + 1; }"), 1},
		{"the short-circuit side carries the stale value", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; await save(); state.count = (old + 1) || 0; }"), 1},
		{"a condition that suspends stales what came before it", concurrencyNoLostUpdateSource(
			"export async function f() { state.count = state.count + ((await next()) ? 1 : 2); }"), 1},
		{"a for await step before a write in its body", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; for await (const x of stream) { state.count = old + x; break; } }"), 1},
		{"popping a list: the head's next, read before the await", concurrencyNoLostUpdateSource(
			"export async function pop() { const following = head?.next; await save(); head = following; }"), 1},
		{"a method called on the target path", concurrencyNoLostUpdateSource(
			"export async function f() { state.items = state.items.concat(await load()); }"), 1},
		{"a const passed through a shorthand property", concurrencyNoLostUpdateSource(
			"export async function f() { const before = state.count; await save(); state.count = total({ before }); }"), 1},
		{"a computed key is evaluated before the await beside it", concurrencyNoLostUpdateSource(
			"export async function f() { count = Object.keys({ [count]: await next() }).length; }"), 1},
		{"a generator method", concurrencyNoLostUpdateSource(
			"export class Ticker { *g() { const old = state.count; yield 1; state.count = old + 1; } }"), 1},
		{"the target read as a subscript", concurrencyNoLostUpdateSource(
			"export async function f() { state.count = table[state.count] + (await next()); }"), 1},

		// Silent: the shapes upstream reported on ahra.
		{"a timestamp after an await (PhiSocialGenerator, ArtGenerator)", concurrencyNoLostUpdateSource(
			"export async function f() { const items = await load(); use(items); state.lastRunAt = Date.now(); }"), 0},
		{"restoring a saved handler in finally (ArtTerminal, PhiSocialTerminal)", concurrencyNoLostUpdateSource(
			"export async function f() {",
			"    const originalLog = console.log;",
			"    const originalInfo = console.info;",
			"    console.log = () => undefined;",
			"    console.info = () => undefined;",
			"    try { await run(); }",
			"    finally { console.log = originalLog; console.info = originalInfo; }",
			"}"), 0},
		{"PhiSocialTerminal.ts:430, a bound original restored in finally, verbatim", concurrencyNoLostUpdatePhiSocialTerminal, 0},
		{"the same restore with the replacement removed is an update", concurrencyNoLostUpdateSource(
			"declare const process: { stdout: { write(text: string): boolean } };",
			"export async function f() {",
			"    const originalStdoutWrite = process.stdout.write.bind(process.stdout);",
			"    try { await run(); }",
			"    finally { process.stdout.write = originalStdoutWrite; }",
			"}"), 1},
		{"a replacement on one path only leaves the other path an update", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; if (flag) { state.count = 0; } await save(); state.count = old + 1; }"), 1},
		{"a replacement before the suspension makes the write a restore", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; state.count = 0; await save(); state.count = old + 1; }"), 0},
		{"a callback's own parameter object in Promise.all (SomfyShadeController)", concurrencyNoLostUpdateSource(
			"export async function f() {",
			"    await Promise.all(entries.map(async (entry) => {",
			"        entry.position = await getPosition();",
			"        if (entry.position > 0) { entry.status = 'done'; }",
			"    }));",
			"}"), 0},
		{"a constant (RainbowMatrix)", concurrencyNoLostUpdateSource(
			"export async function f() { if (blockoutBottomRow > 0) { await save(); blockoutBottomRow = -1; } }"), 0},
		{"a cache replaced wholesale (GoogleClient, ApnsApi)", concurrencyNoLostUpdateSource(
			"export async function f() { if (!tokenCache) { tokenCache = { token: await sign(), expiresAt: Date.now() }; } return tokenCache; }"), 0},
		{"a parameter's fields from a response (SpotifyApi)", concurrencyNoLostUpdateSource(
			"export async function f(credentials: { accessToken: string; expiresAt: number }) {",
			"    const data = await sign();",
			"    credentials.accessToken = data;",
			"    credentials.expiresAt = Date.now() + 1000;",
			"}"), 0},

		// Silent: near misses.
		{"the read after the suspension", concurrencyNoLostUpdateSource(
			"export async function f() { await save(); state.count = state.count + 1; }"), 0},
		{"the read after the await in the same expression", concurrencyNoLostUpdateSource(
			"export async function f() { count = await next() + count; }"), 0},
		{"a local nobody else can see", concurrencyNoLostUpdateSource(
			"export async function f() { let total = 0; total += await next(); return total; }"), 0},
		{"a local a nested function only reads", concurrencyNoLostUpdateSource(
			"export async function f() { let total = 0; register(() => use(total)); total = total + await next(); }"), 0},
		{"a shadowing local of the same name", concurrencyNoLostUpdateSource(
			"export async function f() { let count = 0; count = count + await next(); return count; }"), 0},
		{"the read and the suspension on different branches", concurrencyNoLostUpdateSource(
			"export async function f() { state.count = flag ? state.count + 1 : await next(); }"), 0},
		{"a fallback that suspends does not carry the read", concurrencyNoLostUpdateSource(
			"export async function f() { state.count = state.count || await next(); state.total = state.total ?? await price(); }"), 0},
		{"the suspending branch returns", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; if (flag) { await save(); return; } state.count = old + 1; }"), 0},
		{"a const declared again on every iteration", concurrencyNoLostUpdateSource(
			"export async function f() { for (const x of xs) { const old = state.count; state.count = old + x; await save(x); } }"), 0},
		{"the await wraps the write", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; await save(state.count = old + 1); }"), 0},
		{"a logical assignment is check-then-act", concurrencyNoLostUpdateSource(
			"export async function f() { state.cache ??= await load(); }"), 0},
		{"a local object nobody else can see", concurrencyNoLostUpdateSource(
			"export async function f() { const tally = { n: 0 }; tally.n = tally.n + await next(); return tally; }"), 0},
		{"an alias read after the suspension", concurrencyNoLostUpdateSource(
			"export async function f() { const snapshot = state; await save(); state.count = snapshot.count + 1; }"), 0},
		{"a const of a different property", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.other; await save(); state.count = old + 1; }"), 0},
		{"an awaited call's result is the callee's answer (SpotifyApi, QuickBooksClient)", concurrencyNoLostUpdateSource(
			"declare function exchange(request: { body: { refresh: number } }): Promise<{ json(): Promise<{ next: number }> }>;",
			"export async function f() {",
			"    const response = await exchange({ body: { refresh: state.count } });",
			"    const tokens = await response.json();",
			"    state.count = tokens.next;",
			"}"), 0},
		{"an awaited call handed the read", concurrencyNoLostUpdateSource(
			"export async function f() { state.count = await price(state.count + 1); }"), 0},
		{"a closure reads the binding when it runs, not when it is made", concurrencyNoLostUpdateSource(
			"export async function f() { render = wrap(() => render() + 1, await next()); }"), 0},
		{"a verbatim restore through a chained assignment", concurrencyNoLostUpdateSource(
			"export async function f() { const saved = state.count; await save(); state.count = mirror = saved; }"), 0},
		{"a computed key under a plain = is not judged", concurrencyNoLostUpdateSource(
			"export async function f() { state.byKey[key] = state.byKey[key] + await next(); }"), 0},
		{"a computed key on a local nobody else can see", concurrencyNoLostUpdateSource(
			"export async function f() { const tally: Record<string, number> = {}; tally[key] += await next(); return tally; }"), 0},
		{"a computed key on a call's result is not judged", concurrencyNoLostUpdateSource(
			"export async function f() { load()[0] += await next(); }"), 0},
		{"a function that cannot suspend", concurrencyNoLostUpdateSource(
			"export function f() { count = count + 1; }"), 0},
		{"a nested function's await is not this function's", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; register(async () => { await save(); }); state.count = old + 1; }"), 0},
		{"a const that is stale only on the branch that does not carry it", concurrencyNoLostUpdateSource(
			"export async function f() { const old = state.count; const chosen = flag ? old : await next(); state.count = chosen + 1; }"), 0},
		{"a yield's value is what the caller sends", concurrencyNoLostUpdateSource(
			"export async function* g(): AsyncGenerator<number, void, number> { state.count = (yield state.count) + (await next()); }"), 0},
		{"a reassigned let is not tracked", concurrencyNoLostUpdateSource(
			"export async function f() { let old = state.count; old = 0; await save(); state.count = old + 1; }"), 0},
		{"an updater callback reads at call time", concurrencyNoLostUpdateSource(
			"export async function f() { count = await apply(() => count + 1); }"), 0},
		{"a restore through a chained assignment", concurrencyNoLostUpdateSource(
			"export async function f() {",
			"    const originalLog = console.log;",
			"    let restored: typeof console.log;",
			"    console.log = () => undefined;",
			"    try { await run(); }",
			"    finally { console.log = restored = originalLog; use(restored); }",
			"}"), 0},
		{"a comma's value is its last operand", concurrencyNoLostUpdateSource(
			"export async function f() { state.count = (use(state.count), await next()); }"), 0},
		{"a type query is not a read", concurrencyNoLostUpdateSource(
			"export async function f() { count = convert<typeof count>(await next()); }"), 0},
		{"a property of this is not the same property of a variable", concurrencyNoLostUpdateSource(
			"export class Counter { count = 0; async add() { this.count = state.count + await next(); } }"), 0},
		{"a restore through a destructured const", concurrencyNoLostUpdateSource(
			"export async function f() { const { count: saved } = state; state.count = 0; await save(); state.count = saved; }"), 0},

		// Writes through a store's write call (storeAccessOf). Reporting.
		{"a store write built from a read before the await (RedditClient)", concurrencyNoLostUpdateSource(
			"export async function f(accountKey: string) {",
			"    const account = getAccount(accountKey);",
			"    const token = await sign();",
			"    updateAccount(accountKey, { oauth2: { ...account.oauth2, token } });",
			"}"), 1},
		{"a store read inside the written value, before its await", concurrencyNoLostUpdateSource(
			"export async function f() { saveSettings({ ...loadSettings(), theme: await sign() }); }"), 1},
		{"a store with no key", concurrencyNoLostUpdateSource(
			"export async function f() { const settings = loadSettings(); await save(); saveSettings({ ...settings, runs: settings.runs + 1 }); }"), 1},
		{"a shared Map's get and set", concurrencyNoLostUpdateSource(
			"export async function f() { const old = cache.get(key) ?? 0; await save(); cache.set(key, old + 1); }"), 1},
		{"a store held on this", concurrencyNoLostUpdateSource(
			"export class Tally { entries = new Map<string, number>();",
			"    async bump() { const old = this.entries.get(key) ?? 0; await save(); this.entries.set(key, old + 1); } }"), 1},
		{"literal keys name one entry", concurrencyNoLostUpdateSource(
			"export async function f() { const account = getAccount('main'); await save(); updateAccount('main', { ...account }); }"), 1},
		{"an awaited store read, then another await", concurrencyNoLostUpdateSource(
			"export async function f() { const account = await loadRemote(key); await save(); await saveRemote(key, { ...account }); }"), 1},
		{"a suspension inside the write's own value, after the read", concurrencyNoLostUpdateSource(
			"export async function f() { const account = getAccount(key); updateAccount(key, { ...account, token: await sign() }); }"), 1},

		// Writes through a store's write call. Silent.
		{"the Reddit fix: read the account again after the await", concurrencyNoLostUpdateSource(
			"export async function f(accountKey: string) {",
			"    const token = await sign();",
			"    const account = getAccount(accountKey);",
			"    updateAccount(accountKey, { oauth2: { ...account.oauth2, token } });",
			"}"), 0},
		{"a different key is a different entry", concurrencyNoLostUpdateSource(
			"export async function f() { const account = getAccount(key); await save(); updateAccount('other', { ...account }); }"), 0},
		{"a different noun is a different store", concurrencyNoLostUpdateSource(
			"export async function f() { const account = getAccount(key); await save(); updateProfile(key, { ...account }); }"), 0},
		{"a verbatim write back is a restore", concurrencyNoLostUpdateSource(
			"export async function f() {",
			"    const saved = getAccount(key);",
			"    updateAccount(key, { oauth2: { token: '', ads: '' } });",
			"    try { await run(); }",
			"    finally { updateAccount(key, saved); }",
			"}"), 0},
		{"the saved entry written back unchanged, with nothing replaced between, is a copy", concurrencyNoLostUpdateSource(
			"export async function f() { const saved = getAccount(key); try { await run(); } finally { updateAccount(key, saved); } }"), 0},
		{"a store write before the suspension makes the later one a restore", concurrencyNoLostUpdateSource(
			"export async function f() { const saved = getAccount(key); updateAccount(key, {}); await run(); updateAccount(key, { ...saved }); }"), 0},
		{"a Map nobody else can see", concurrencyNoLostUpdateSource(
			"export async function f() { const local = new Map<string, number>(); const old = local.get(key) ?? 0; await save(); local.set(key, old + 1); return local; }"), 0},
		{"a key the function reassigns names no one entry", concurrencyNoLostUpdateSource(
			"export async function f(accountKey: string) { const account = getAccount(accountKey); await save(); accountKey = 'other'; updateAccount(accountKey, { ...account }); }"), 0},
		{"a member read as the key is not judged", concurrencyNoLostUpdateSource(
			"export async function f(options: { key: string }) { const account = getAccount(options.key); await save(); updateAccount(options.key, { ...account }); }"), 0},
		{"no suspension between the store read and the write", concurrencyNoLostUpdateSource(
			"export async function f() { const account = getAccount(key); updateAccount(key, { ...account }); await save(); }"), 0},
		{"a bare get and set name no store", concurrencyNoLostUpdateSource(
			"export async function f() { const old = get(key); await save(); set(key, old + 1); }"), 0},
		{"an awaited store read with no later suspension", concurrencyNoLostUpdateSource(
			"export async function f() { const account = await loadRemote(key); await saveRemote(key, { ...account }); }"), 0},
		{"a store read with no write", concurrencyNoLostUpdateSource(
			"export async function f() { const account = getAccount(key); await save(); use({ ...account }); }"), 0},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, ConcurrencyNoLostUpdate, concurrencyNoLostUpdateFile, testCase.source)
			if testCase.findings == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			ids := make([]string, testCase.findings)
			for index := range ids {
				ids[index] = concurrencyNoLostUpdateId
			}
			rule_testing.ExpectFindings(t, result, ids...)
		})
	}
}

// The finding spans the whole assignment and names the target as written, so the reader is not left
// to work out which write on the line loses the update.
func TestConcurrencyNoLostUpdateNamesTheTarget(t *testing.T) {
	t.Parallel()

	source := concurrencyNoLostUpdateSource(
		"export async function f() { const old = state.count; await save(); state.count = old + 1; }")
	result := rule_testing.RunTyped(t, ConcurrencyNoLostUpdate, concurrencyNoLostUpdateFile, source)
	rule_testing.ExpectFindings(t, result, concurrencyNoLostUpdateId)

	diagnostic := result.Diagnostics[0]
	text := result.SourceFile.Text()
	if got := text[diagnostic.Range.Pos():diagnostic.Range.End()]; strings.TrimSpace(got) != "state.count = old + 1" {
		t.Fatalf("span: got %q", got)
	}
	if !strings.Contains(diagnostic.Message.Description, "computed from `state.count` as it was") {
		t.Fatalf("message does not name the target: %q", diagnostic.Message.Description)
	}
}

// A store write's finding spans the write call and names the read call whose stale answer it builds
// on, since the entry has no name of its own to quote.
func TestConcurrencyNoLostUpdateNamesTheStoreRead(t *testing.T) {
	t.Parallel()

	source := concurrencyNoLostUpdateSource(
		"export async function f(accountKey: string) { const account = getAccount(accountKey); await save(); updateAccount(accountKey, { ...account }); }")
	result := rule_testing.RunTyped(t, ConcurrencyNoLostUpdate, concurrencyNoLostUpdateFile, source)
	rule_testing.ExpectFindings(t, result, concurrencyNoLostUpdateId)

	diagnostic := result.Diagnostics[0]
	text := result.SourceFile.Text()
	if got := text[diagnostic.Range.Pos():diagnostic.Range.End()]; strings.TrimSpace(got) != "updateAccount(accountKey, { ...account })" {
		t.Fatalf("span: got %q", got)
	}
	if !strings.Contains(diagnostic.Message.Description, "computed from what `getAccount(accountKey)` returned") {
		t.Fatalf("message does not name the read: %q", diagnostic.Message.Description)
	}
}
