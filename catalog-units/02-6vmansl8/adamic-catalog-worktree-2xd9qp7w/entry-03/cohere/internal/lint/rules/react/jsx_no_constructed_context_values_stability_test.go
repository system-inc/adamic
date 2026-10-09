package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The fixtures for cohere's stricter half of the rule: a memoized provider value is stable only if
// every input is. Upstream has no corpus for this, so the reporting cases are the four sites the
// 2026-10-01 rule review found in ahra (DialogRoot, DrawerRoot, TableRoot, Field), reduced to the
// shape that makes them unstable and written in the memoized form the obvious fix would produce. The
// silent cases are the shapes those same files hold that ARE stable, plus every source of identity
// the analysis deliberately trusts.

// TestJsxNoConstructedContextValuesReportsAnUnstableMemoDependency covers each way a memo's input is
// proven to change every render.
func TestJsxNoConstructedContextValuesReportsAnUnstableMemoDependency(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"DialogRoot: a theme merged during render", `
function mergeTheme<T extends object>(baseTheme: T, overrideTheme?: Partial<T>): T {
    if(!overrideTheme) {
        return baseTheme;
    }
    const merged = { ...baseTheme } as Record<string, unknown>;
    return merged as T;
}
const StructureDialogTheme = { variants: {} };
declare function useTheme(): { Dialog?: Partial<typeof StructureDialogTheme> } | undefined;
declare function useControlledState(open: boolean | undefined): [boolean, (open: boolean) => void];
export function DialogRoot(properties: { open?: boolean; modal?: boolean }) {
    const theme = useTheme();
    const dialogTheme = mergeTheme(StructureDialogTheme, theme?.Dialog);
    const [open, onOpenChange] = useControlledState(properties.open);
    const contextValue = React.useMemo(
        function() {
            return { open, onOpenChange, dialogTheme, modal: properties.modal };
        },
        [open, onOpenChange, dialogTheme, properties.modal],
    );
    return <DialogContext.Provider value={contextValue}>{null}</DialogContext.Provider>;
}`},
		{"DrawerRoot: an inline merge under useCallback", `
declare const base: { a: number };
function merge(override?: { a: number }) {
    return override ? { ...base, ...override } : base;
}
export function DrawerRoot(properties: { override?: { a: number } }) {
    const drawerTheme = merge(properties.override);
    const contextValue = useCallback(function() { return drawerTheme; }, [drawerTheme]);
    return <DrawerContext.Provider value={contextValue} />;
}`},
		{"TableRoot: a Map built by a helper, cast at the provider", `
function buildColumnPinMap(columns: readonly string[]): Map<string, number> {
    const pinMap = new Map<string, number>();
    for(const column of columns) pinMap.set(column, 0);
    return pinMap;
}
interface TableContextInterface { columnPinMap: Map<string, number> }
export function TableRoot(properties: { columns: string[] }) {
    const columnPinMap = buildColumnPinMap(properties.columns);
    const contextValue = React.useMemo(() => ({ columnPinMap }), [columnPinMap]);
    return <TableContext.Provider value={contextValue as TableContextInterface} />;
}`},
		{"Field: a conditional call result and a listener lookup", `
function buildFieldContextValue(properties: { fieldName: string }) {
    function readSnapshot() { return properties.fieldName; }
    return { fieldName: properties.fieldName, readSnapshot };
}
export function Field(properties: { identifier?: string }) {
    const fieldContextValue = properties.identifier
        ? buildFieldContextValue({ fieldName: properties.identifier })
        : null;
    const value = React.useMemo(() => fieldContextValue, [fieldContextValue]);
    return <FieldContext.Provider value={value} />;
}`},
		{"an arrow helper with an expression body", `
const build = (n: number) => ({ n });
export function Component(properties: { n: number }) {
    const built = build(properties.n);
    const value = useMemo(() => ({ built }), [built]);
    return <Ctx.Provider value={value} />;
}`},
		{"an object literal built during render", `
export function Component(properties: { n: number }) {
    const style = { n: properties.n };
    const value = useMemo(() => ({ style }), [style]);
    return <Ctx.Provider value={value} />;
}`},
		{"a nullish fallback array", `
export function Component(properties: { items?: string[] }) {
    const items = properties.items ?? [];
    const value = useMemo(() => ({ items }), [items]);
    return <Ctx.Provider value={value} />;
}`},
		{"a function declared during render", `
export function Component(properties: { n: number }) {
    function handle() { return properties.n; }
    const value = useMemo(() => ({ handle }), [handle]);
    return <Ctx.Provider value={value} />;
}`},
		{"a memo whose own dependency is unstable", `
export function Component(properties: { n: number }) {
    const options = { n: properties.n };
    const derived = useMemo(() => ({ options }), [options]);
    const value = useMemo(() => ({ derived }), [derived]);
    return <Ctx.Provider value={value} />;
}`},
		{"a memo written inline at the provider", `
export function Component(properties: { n: number }) {
    const options = { n: properties.n };
    return <Ctx.Provider value={useMemo(() => ({ options }), [options])} />;
}`},
		{"a custom hook that returns a fresh object", `
function useSelection() {
    const [selected, setSelected] = React.useState<string[]>([]);
    return { selected, setSelected };
}
export function Component() {
    const selection = useSelection();
    const value = useMemo(() => ({ selection }), [selection]);
    return <Ctx.Provider value={value} />;
}`},
		{"an async helper returns a new promise every call", `
const stored = { a: 1 };
async function load() { return stored; }
export function Component() {
    const pending = load();
    const value = useMemo(() => ({ pending }), [pending]);
    return <Ctx.Provider value={value} />;
}`},
		{"a generator returns a new iterator every call", `
const stored = [1];
function* walk() { yield* stored; }
export function Component() {
    const iterator = walk();
    const value = useMemo(() => ({ iterator }), [iterator]);
    return <Ctx.Provider value={value} />;
}`},
		{"the left side of a logical expression constructs", `
declare const fallback: { a: number };
export function Component(properties: { flag: boolean }) {
    const chosen = (properties.flag && { a: 1 }) || fallback;
    const value = useMemo(() => ({ chosen }), [chosen]);
    return <Ctx.Provider value={value} />;
}`},
		{"the false branch of a conditional constructs", `
declare const stable: { a: number };
export function Component(properties: { flag: boolean }) {
    const chosen = properties.flag ? stable : { a: 1 };
    const value = useMemo(() => ({ chosen }), [chosen]);
    return <Ctx.Provider value={value} />;
}`},
		{"a comma expression evaluates to its constructed right side", `
declare function log(): void;
export function Component() {
    const options = (log(), { a: 1 });
    const value = useMemo(() => ({ options }), [options]);
    return <Ctx.Provider value={value} />;
}`},
		{"an assignment evaluates to its constructed right side", `
export function Component() {
    let previous: { a: number } | undefined;
    const options = (previous = { a: 1 });
    const value = useMemo(() => ({ options, previous }), [options]);
    return <Ctx.Provider value={value} />;
}`},
		{"a helper declared in render returns a render value", `
export function Component(properties: { n: number }) {
    const options = { n: properties.n };
    function current() { return options; }
    const chosen = current();
    const value = useMemo(() => ({ chosen }), [chosen]);
    return <Ctx.Provider value={value} />;
}`},
		{"a memo factory returns a render value", `
export function Component(properties: { n: number }) {
    const options = { n: properties.n };
    const value = useMemo(() => options, [options]);
    return <Ctx.Provider value={value} />;
}`},
		{"a merge that writes fields into its own copy before returning it", `
function mergeTheme(base: { a: number; b?: number }, override?: { a: number }) {
    if(!override) return base;
    const merged = { ...base };
    merged.a = override.a;
    merged['b'] = merged.a;
    if(merged === base || !merged || typeof merged !== 'object' || !(merged instanceof Object)) return base;
    if(merged) merged.a += 0;
    return merged;
}
declare const Base: { a: number };
export function DialogRoot(properties: { override?: { a: number } }) {
    const dialogTheme = mergeTheme(Base, properties.override);
    const value = useMemo(() => ({ dialogTheme }), [dialogTheme]);
    return <Ctx.Provider value={value} />;
}`},
		{"a memo factory in a nested component returns that component's value", `
export function makeComponent() {
    function Inner(properties: { n: number }) {
        const options = { n: properties.n };
        const value = useMemo(() => options, [options]);
        return <Ctx.Provider value={value} />;
    }
    return Inner;
}`},
		{"a memo factory returns a render fallback", `
export function Component(properties: { items?: string[] }) {
    const items = properties.items ?? [];
    const value = useMemo(() => items, [items]);
    return <Ctx.Provider value={value} />;
}`},
		{"a useCallback is a new function whatever its body returns", `
export function Component(properties: { n: number }) {
    const style = { n: properties.n };
    const handle = useCallback(() => properties.n, [style]);
    return <Ctx.Provider value={handle} />;
}`},
		{"a method whose body copies its state", `
class Store { private items: string[] = []; snapshot() { return [...this.items]; } }
declare const store: Store;
export function Component() {
    const items = store.snapshot();
    const value = useMemo(() => ({ items }), [items]);
    return <Ctx.Provider value={value} />;
}`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
				jsxNoConstructedContextValuesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "unstableDependencyMsg")
		})
	}
}

// TestJsxNoConstructedContextValuesTrustsWhatItCannotProveUnstable is the other direction: every
// input the analysis treats as stable, each one a shape that sits beside the reporting sites in the
// same real files.
func TestJsxNoConstructedContextValuesTrustsWhatItCannotProveUnstable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"hook results, props and destructured state", `
declare function useControlledState(open: boolean | undefined): [boolean, (open: boolean) => void];
declare function useIsMobile(): boolean;
export function DialogRoot(properties: { open?: boolean; onOpenAutoFocus?: () => void }) {
    const isMobile = useIsMobile();
    const dialogId = React.useId();
    const [open, onOpenChange] = useControlledState(properties.open);
    const contextValue = React.useMemo(
        () => ({ open, onOpenChange, isMobile, dialogId, onOpenAutoFocus: properties.onOpenAutoFocus }),
        [open, onOpenChange, isMobile, dialogId, properties.onOpenAutoFocus],
    );
    return <DialogContext.Provider value={contextValue} />;
}`},
		{"a module-scope constant and an import-like declaration", `
const Theme = { variants: {} };
declare const imported: { a: number };
export function Component() {
    const value = useMemo(() => ({ Theme, imported }), [Theme, imported]);
    return <Ctx.Provider value={value} />;
}`},
		{"a primitive compares by value, whatever its body's syntax claims", `
function count(items: string[]): number { return { n: items.length } as unknown as number; }
export function Component(properties: { items: string[] }) {
    const total = count(properties.items);
    const value = useMemo(() => ({ total }), [total]);
    return <Ctx.Provider value={value} />;
}`},
		{"a helper that returns its argument or a stored value", `
const cache = new Map<string, { a: number }>();
function pick(base: { a: number }) { return base; }
function lookup(key: string) { return cache.get(key); }
export function Component(properties: { base: { a: number }; key: string }) {
    const picked = pick(properties.base);
    const found = lookup(properties.key);
    const value = useMemo(() => ({ picked, found }), [picked, found]);
    return <Ctx.Provider value={value} />;
}`},
		{"a memo over stable inputs feeding another memo", `
export function Component(properties: { n: number }) {
    const options = useMemo(() => ({ n: properties.n }), [properties.n]);
    const handle = useCallback(() => options.n, [options]);
    const value = useMemo(() => ({ options, handle }), [options, handle]);
    return <Ctx.Provider value={value} />;
}`},
		{"a let that a later line may replace", `
declare const stable: { a: number };
export function Component(properties: { flag: boolean }) {
    let current = { a: 1 };
    if(properties.flag) current = stable;
    const value = useMemo(() => ({ current }), [current]);
    return <Ctx.Provider value={value} />;
}`},
		{"a property path dependency", `
export function Component(properties: { n: number }) {
    const holder = { inner: properties };
    const value = useMemo(() => ({ n: holder.inner.n }), [holder.inner]);
    return <Ctx.Provider value={value} />;
}`},
		{"a declared function has no body to read", `
declare function external(): { a: number };
export function Component() {
    const outside = external();
    const value = useMemo(() => ({ outside }), [outside]);
    return <Ctx.Provider value={value} />;
}`},
		{"a useMemo method on something other than React", `
declare const cache: { useMemo<T>(factory: () => T, dependencies: unknown[]): T };
export function Component(properties: { n: number }) {
    const options = { n: properties.n };
    const value = cache.useMemo(() => ({ options }), [options]);
    return <Ctx.Provider value={value} />;
}`},
		{"a helper whose fresh value is only inside a nested callback", `
function wrap(items: string[]) {
    items.forEach(function(item) { return { item }; });
    return items;
}
export function Component(properties: { items: string[] }) {
    const wrapped = wrap(properties.items);
    const value = useMemo(() => ({ wrapped }), [wrapped]);
    return <Ctx.Provider value={value} />;
}`},
		{"a logical assignment keeps whatever the target already held", `
let sharedOptions: { a: number } | undefined;
export function Component() {
    const options = (sharedOptions ||= { a: 1 });
    const value = useMemo(() => ({ options }), [options]);
    return <Ctx.Provider value={value} />;
}`},
		{"a recursive helper that bottoms out in a stable value terminates silent", `
declare const stable: { a: number };
function descend(depth: number): { a: number } { return depth > 0 ? descend(depth - 1) : stable; }
export function Component(properties: { depth: number }) {
    const found = descend(properties.depth);
    const value = useMemo(() => ({ found }), [found]);
    return <Ctx.Provider value={value} />;
}`},
		{"a caching helper that stores what it builds in a Map", `
const cache = new Map<string, { key: string }>();
function themeFor(key: string) {
    const cached = cache.get(key);
    if(cached) return cached;
    const created = { key };
    cache.set(key, created);
    return created;
}
export function Component(properties: { variant: string }) {
    const theme = themeFor(properties.variant);
    const value = useMemo(() => ({ theme }), [theme]);
    return <Ctx.Provider value={value} />;
}`},
		{"a caching helper that stores what it builds in an object property", `
const cacheObject: Record<string, { key: string }> = {};
function themeFor(key: string) {
    const cached = cacheObject[key];
    if(cached) return cached;
    const created = { key };
    cacheObject[key] = created;
    return created;
}
export function Component(properties: { variant: string }) {
    const theme = themeFor(properties.variant);
    const value = useMemo(() => ({ theme }), [theme]);
    return <Ctx.Provider value={value} />;
}`},
		{"a caching helper that returns the assignment into its cache", `
const cacheObject: Record<string, { key: string }> = {};
function themeFor(key: string) {
    return cacheObject[key] ?? (cacheObject[key] = { key });
}
export function Component(properties: { variant: string }) {
    const theme = themeFor(properties.variant);
    const value = useMemo(() => ({ theme }), [theme]);
    return <Ctx.Provider value={value} />;
}`},
		{"a caching helper whose built value escapes into a closure", `
const pending: Array<() => { key: string }> = [];
let last: { key: string } | undefined;
function themeFor(key: string) {
    if(last) return last;
    const created = { key };
    pending.push(function() { return created; });
    return created;
}
export function Component(properties: { variant: string }) {
    const theme = themeFor(properties.variant);
    const value = useMemo(() => ({ theme }), [theme]);
    return <Ctx.Provider value={value} />;
}`},
		{"a render fallback that is stored for the next render", `
const cache = new Map<string, { variant: string }>();
export function Component(properties: { variant: string }) {
    const theme = cache.get(properties.variant) ?? { variant: properties.variant };
    cache.set(properties.variant, theme);
    const value = useMemo(() => ({ theme }), [theme]);
    return <Ctx.Provider value={value} />;
}`},
		{"a caching helper whose built value registers itself through a method", `
class Entry { constructor(public key: string) {} attachTo(registry: Map<string, Entry>) { registry.set(this.key, this); } }
const registry = new Map<string, Entry>();
function entryFor(key: string) {
    const existing = registry.get(key);
    if(existing) return existing;
    const created = new Entry(key);
    created.attachTo(registry);
    return created;
}
export function Component(properties: { variant: string }) {
    const entry = entryFor(properties.variant);
    const value = useMemo(() => ({ entry }), [entry]);
    return <Ctx.Provider value={value} />;
}`},
		{"a caching helper that assigns what it builds to a captured local", `
const pending: Array<() => { key: string } | undefined> = [];
const cache = new Map<string, { key: string }>();
function themeFor(key: string) {
    const cached = cache.get(key);
    if(cached) return cached;
    let created: { key: string } | undefined;
    pending.push(function() { return created; });
    return (created = { key });
}
export function Component(properties: { variant: string }) {
    const theme = themeFor(properties.variant);
    const value = useMemo(() => ({ theme }), [theme]);
    return <Ctx.Provider value={value} />;
}`},
		{"a memoize wrapper from source", `
function memoize<A, R>(compute: (argument: A) => R) {
    const results = new Map<A, R>();
    return function(argument: A) {
        if(results.has(argument)) return results.get(argument) as R;
        const result = compute(argument);
        results.set(argument, result);
        return result;
    };
}
const themeFor = memoize((key: string) => ({ key }));
export function Component(properties: { variant: string }) {
    const theme = themeFor(properties.variant);
    const value = useMemo(() => ({ theme }), [theme]);
    return <Ctx.Provider value={value} />;
}`},
		{"a memoize wrapper from a library", `
declare function memoize<A, R>(compute: (argument: A) => R): (argument: A) => R;
const themeFor = memoize((key: string) => ({ key }));
export function Component(properties: { variant: string }) {
    const theme = themeFor(properties.variant);
    const value = useMemo(() => ({ theme }), [theme]);
    return <Ctx.Provider value={value} />;
}`},
		{"a useCallback result passed through a helper", `
function passThrough<T>(value: T): T { return value; }
export function Component(properties: { n: number }) {
    const handle = passThrough(useCallback(() => properties.n, [properties.n]));
    const value = useMemo(() => ({ handle }), [handle]);
    return <Ctx.Provider value={value} />;
}`},
		{"a memo over an unstable input that returns a stable object", `
declare const stableObject: { a: number };
export function Component(properties: { n: number }) {
    const options = { n: properties.n };
    const value = useMemo(() => stableObject, [options]);
    return <Ctx.Provider value={value} />;
}`},
		{"a nested memo over an unstable input that returns a stable object", `
declare const stableObject: { a: number };
export function Component(properties: { n: number }) {
    const options = { n: properties.n };
    const chosen = useMemo(() => stableObject, [options]);
    const value = useMemo(() => ({ chosen }), [chosen]);
    return <Ctx.Provider value={value} />;
}`},
		{"a memo with no list that returns a stable object", `
declare const stableObject: { a: number };
export function Component() {
    const value = useMemo(() => stableObject);
    return <Ctx.Provider value={value} />;
}`},
		{"a component declared inside another function keeps the outer value", `
export function makeComponent() {
    const shared = { a: 1 };
    function Inner(properties: { n: number }) {
        const value = useMemo(() => ({ shared, n: properties.n }), [shared, properties.n]);
        return <Ctx.Provider value={value} />;
    }
    return Inner;
}`},
		{"a helper returned by a factory keeps the factory's value", `
function makeHelpers() {
    const shared = { a: 1 };
    return { get: function() { return shared; } };
}
const helpers = makeHelpers();
export function Component() {
    const found = helpers.get();
    const value = useMemo(() => ({ found }), [found]);
    return <Ctx.Provider value={value} />;
}`},
		{"a dependency list passed by name", `
export function Component(properties: { n: number }) {
    const options = { n: properties.n };
    const dependencies = [options];
    const value = useMemo(() => ({ options }), dependencies);
    return <Ctx.Provider value={value} />;
}`},
		{"a memo in a lowercase helper is not a component", `
function helper(properties: { n: number }) {
    const options = { n: properties.n };
    const value = useMemo(() => ({ options }), [options]);
    return <Ctx.Provider value={value} />;
}`},
		{"a value built in a nested function is that function's, not render's", `
export function Component(properties: { n: number }) {
    const value = useMemo(function() {
        const inside = { n: properties.n };
        return inside;
    }, [properties.n]);
    return <Ctx.Provider value={value} />;
}`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
				jsxNoConstructedContextValuesFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestJsxNoConstructedContextValuesReportsAMemoWithoutDependencies pins the second new message: a
// memo with no list recomputes every render, which no input can rescue.
func TestJsxNoConstructedContextValuesReportsAMemoWithoutDependencies(t *testing.T) {
	t.Parallel()

	source := `
export function Component(properties: { n: number }) {
    const value = React.useMemo(() => ({ n: properties.n }));
    return <Ctx.Provider value={value} />;
}`
	result := rule_testing.RunTyped(t, JsxNoConstructedContextValues, jsxNoConstructedContextValuesFile, source)
	rule_testing.ExpectFindings(t, result, "memoWithoutDependenciesMsg")
}

// TestJsxNoConstructedContextValuesUnstableDependencyMessage pins the finding's anchor and text: it
// points at the dependency in the list, names it, and says where the new identity is made.
func TestJsxNoConstructedContextValuesUnstableDependencyMessage(t *testing.T) {
	t.Parallel()

	source := "function mergeTheme(base: { a: number }, override?: { a: number }) {\n" +
		"    if(!override) return base;\n" +
		"    const merged = { ...base, ...override };\n" +
		"    return merged;\n" +
		"}\n" +
		"declare const Base: { a: number };\n" +
		"export function DialogRoot(properties: { override?: { a: number } }) {\n" +
		"    const dialogTheme = mergeTheme(Base, properties.override);\n" +
		"    const contextValue = React.useMemo(() => ({ dialogTheme }), [dialogTheme]);\n" +
		"    return <DialogContext.Provider value={contextValue} />;\n" +
		"}\n"
	result := rule_testing.RunTyped(t, JsxNoConstructedContextValues, jsxNoConstructedContextValuesFile, source)
	rule_testing.ExpectFindings(t, result, "unstableDependencyMsg")

	diagnostic := result.Diagnostics[0]
	text := strings.TrimSpace(source) + "\n"
	if got := text[diagnostic.Range.Pos():diagnostic.Range.End()]; got != "dialogTheme" {
		t.Errorf("span = %q, wanted the dependency entry \"dialogTheme\"", got)
	}
	if line := strings.Count(text[:diagnostic.Range.Pos()], "\n") + 1; line != 9 {
		t.Errorf("finding on line %d, wanted line 9 where the dependency list is", line)
	}
	want := "The \"dialogTheme\" dependency of the useMemo that builds the Context provider's value " +
		"(at line 9) is the result of `mergeTheme(...)`, which can return a new value on every call " +
		"(at line 3), so the useMemo recomputes on every render and every consumer of this context " +
		"re-renders on every render of this component, exactly as if it were not memoized. A memoized " +
		"value is only stable when every input is: make \"dialogTheme\" stable too, by memoizing it or " +
		"by building it outside the component."
	if got := diagnostic.Message.Description; got != want {
		t.Errorf("message =\n%q\nwanted\n%q", got, want)
	}
}

// TestJsxNoConstructedContextValuesNamesTheNestedMemoDependency pins the transitive wording.
func TestJsxNoConstructedContextValuesNamesTheNestedMemoDependency(t *testing.T) {
	t.Parallel()

	source := `
export function Component(properties: { n: number }) {
    const options = { n: properties.n };
    const derived = useMemo(() => ({ options }), [options]);
    const value = useMemo(() => ({ derived }), [derived]);
    return <Ctx.Provider value={value} />;
}`
	result := rule_testing.RunTyped(t, JsxNoConstructedContextValues, jsxNoConstructedContextValuesFile, source)
	rule_testing.ExpectFindings(t, result, "unstableDependencyMsg")
	want := "comes from a useMemo whose own dependency \"options\" is a new object built during render (at line 2)"
	if got := result.Diagnostics[0].Message.Description; !strings.Contains(got, want) {
		t.Errorf("message =\n%q\nwanted it to contain\n%q", got, want)
	}
}

// TestJsxNoConstructedContextValuesOnlyCountsPerRenderDeclarations pins which declarations the
// identifier arm follows: only those evaluated each time the function holding the provider runs.
//
// The silent rows were each measured silent on the installed eslint-plugin-react 7.37.5; the first
// is the Providers.tsx shape that cohere reported and ESLint did not. The last reporting row is the
// one upstream misses (it resolves names only in the JSX's own block scope) and cohere reports.
func TestJsxNoConstructedContextValuesOnlyCountsPerRenderDeclarations(t *testing.T) {
	t.Parallel()

	silent := []struct {
		name       string
		sourceText string
	}{
		{"a module-scope object constant (Providers.tsx)", `
const iconContextValue: { weight: string; color: string } = { weight: 'bold', color: 'currentColor' };
export function Providers(properties: { children?: unknown }) {
    return <IconContext.Provider value={iconContextValue}>{properties.children}</IconContext.Provider>;
}`},
		{"a module-scope function declaration", `
function handle() {}
export function Component() {
    return <Ctx.Provider value={handle} />;
}`},
		{"a module-scope arrow constant", `
const handle = () => {};
export function Component() {
    return <Ctx.Provider value={handle} />;
}`},
		{"a component alias of a module-scope constant", `
const shared = { a: 1 };
export function Component() {
    const value = shared;
    return <Ctx.Provider value={value} />;
}`},
		{"an object declared in a function enclosing the component", `
export function makeComponent() {
    const shared = { a: 1 };
    function Inner() {
        return <Ctx.Provider value={shared} />;
    }
    return Inner;
}`},
	}
	for _, testCase := range silent {
		t.Run("silent "+testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
				jsxNoConstructedContextValuesFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}

	reporting := []struct {
		name       string
		sourceText string
	}{
		{"an object declared in the component", `
export function Component() {
    const value = { a: 1 };
    return <Ctx.Provider value={value} />;
}`},
		{"a component object used by a provider inside an if block (upstream misses it)", `
export function Component(properties: { show: boolean }) {
    const value = { a: 1 };
    if(properties.show) {
        return <Ctx.Provider value={value} />;
    }
    return null;
}`},
	}
	for _, testCase := range reporting {
		t.Run("reports "+testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, JsxNoConstructedContextValues,
				jsxNoConstructedContextValuesFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "withIdentifierMsg")
		})
	}
}
