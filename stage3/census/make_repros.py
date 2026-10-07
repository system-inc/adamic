"""Small stage-0 inputs. Expectations are checked against actual observations, not assumed."""
import json
import re
from pathlib import Path

root = Path(__file__).resolve().parent / 'repro'
examples = {
'a function inside a function (a closure)': ('a', 'Implement named nested functions with recursive binding, captures and lifetime proofs.', 'function outer(): number { function inner(): number { return 1; } return inner(); } console.log(outer());'),
'a class inside a function': ('a', 'Implement local class binding and captured constructor/method environments.', 'function outer(): number { class Value { count = 1; } return new Value().count; } console.log(outer());'),
'TS1484': ('b', 'Use explicit type imports; runtime import effects must remain.', 'import { Item } from "../support.a"; const value: Item = {count: 1}; console.log(value.count);'),
'non-boolean control condition': ('b', 'Make the original truthiness test explicit; preserve JS coercion.', 'const value = 1; if (value) { console.log(value); }'),
"the non-null assertion !": ('b', 'Replace each assertion with proven narrowing or a loud checked unwrap.', 'function read(value: number | undefined): number { return value!; } console.log(read(1));'),
'TS2345': ('c', 'Required argument, optional value, array density, and generic contracts need site review.', 'function take(value: number): number { return value; } const values: number[] = []; console.log(take(values[0]));'),
'TS1294': ('c', 'Enum/namespace/parameter-property policy; see the prior complete syntax census.', 'enum Kind { First, Second } console.log(Kind.First);'),
'enum': ('c', 'Constant flags and ordinary enum runtime objects require an explicit language decision.', 'enum Kind { First, Second } console.log(Kind.First);'),
'TS7030': ('b', 'Explicit undefined return only for a declared optional-return contract; inferred contracts need review.', 'function read(flag: boolean): number | undefined { if (flag) { return 1; } } console.log(read(false));'),
'TS2322': ('c', 'Assignment buckets include optional reads, generics and narrowing; no blanket repair.', 'const values: number[] = []; const value: number = values[0]; console.log(value);'),
'a type predicate': ('c', 'Verify guards and assert functions, or narrow at use; cannot trust arbitrary predicates.', 'function isNumber(value: unknown): value is number { return typeof value === "number"; } console.log(isNumber(1));'),
'explicit any': ('b', 'Adamic intentionally forbids unproven any; observed stage-0 spelling can be NotYet.', 'let value: any = 1; console.log(value);'),
'TS2532': ('c', 'Required indexed read needs a density/bounds contract or explicit missing handling.', 'const values: number[] = []; console.log(values[0].toString());'),
'TS18048': ('c', 'Required value must be proven present; preserve legitimate absence.', 'const values: number[] = []; const value = values[0]; console.log(value.toString());'),
'TS2412': ('b', 'Truthful optional declaration may permit present undefined, preserving the write and key.', 'interface Slot { value?: number; } const slot: Slot = {}; slot.value = undefined;'),
'||=': ('b', 'Use an explicit if with identical single evaluation and truthiness semantics.', 'let value = false; value ||= true; console.log(value);'),
'TS2375': ('b', 'Represent intentionally materialized undefined in the owned optional declaration.', 'interface Slot { value?: number; } const slot: Slot = {value: undefined}; console.log(slot.value);'),
'TS2724': ('a', 'Fetch/generate upstream inputs first; this corpus has missing Diagnostics exports, not a proven language gap.', 'import { Diagnostics } from "../support.a"; console.log(Diagnostics);'),
'TS7029': ('b', 'Make intentional switch fallthrough explicit without changing control effects.', 'function read(value: number): void { switch (value) { case 1: console.log(1); case 2: console.log(2); break; default: break; } } read(1);'),
'TS2379': ('b', 'Align optional representation across owned parameter views; generic mismatch cases need review.', 'function take(value: {count?: number}): void { console.log(value.count); } take({count: undefined});'),
"the comma operator": ('b', 'Sequence statements while retaining value and evaluation order.', 'function one(): number { console.log("one"); return 1; } const value = (one(), 2); console.log(value);'),
'TS2488': ('c', 'Iterable presence, element contracts, and custom iterator representation need review.', 'const entries = new Map<string, number>(); const [key, value] = entries.entries().next().value;'),
'a namespace': ('c', 'Live namespace objects, merging, exports and initialization need a language decision.', 'namespace Values { export const count = 1; } console.log(Values.count);'),
'an index signature': ('c', 'Choose exact object-key semantics or a reviewed Map source migration.', 'interface Values { [key: string]: number; } const values: Values = {}; console.log(values);'),
'TS2339': ('c', 'Union/property access needs a real discriminant or generic contract proof.', 'function read(value: {kind: "a"; count: number} | {kind: "b"}): number { return value.count; }'),
'TS2538': ('c', 'An optional lookup key must be proved present at its original read.', 'const values: {[key: string]: number} = {}; const keys: string[] = []; console.log(values[keys[0]]);'),
'a spread after the first field': ('b', 'Only a single leading spread is sound; review hidden keys before rewriting.', 'const source = {count: 1}; const value = {label: "x", ...source}; console.log(value.count);'),
'TS2769': ('c', 'Overload selection is a bucket; iterator, element and callable contracts differ by site.', 'const values = new Map<string, number>([[1, 2]]); console.log(values.size);'),
'a label': ('b', 'Replace labeled control flow with a function preserving exits and finally effects.', 'outer: for (let index = 0; index < 1; index++) { break outer; }'),
'TS18046': ('c', 'Narrow the actual thrown value; upstream catch-any behavior is not a proof.', 'try { throw new Error("x"); } catch (error) { console.log(error.message); }'),
'TS2591': ('a', 'Host declarations and native Node host APIs are separate prerequisites.', 'console.log(process.cwd());'),
"the void operator": ('b', 'Retain evaluation and undefined result; special return adaptation is tooling-only.', 'console.log(void 0);'),
'a definite assignment assertion !': ('b', 'Initialize the declared slot or represent absence honestly.', 'class Value { count!: number; } console.log(new Value().count);'),
'in': ('c', 'Own/inherited presence and optional undefined cannot be replaced by truthiness.', 'const value = {count: 1}; console.log("count" in value);'),
'Record<string, T>': ('c', 'String records, integer-key ordering and prototype/presence semantics need a decision.', 'const values: Record<string, number> = {}; values["x"] = 1; console.log(values["x"]);'),
'an ExportDeclaration': ('b', 'Prefer declaration-site named exports; barrels also participate in module cycles.', 'const value = 1; export {value}; console.log(value);'),
'delete': ('b', 'Fixed object shapes are intentional; dynamic removal needs Map or explicit tombstones.', 'const value: {count?: number} = {count: 1}; delete value.count;'),
'&&=': ('b', 'Use an explicit if preserving evaluation and assignment semantics.', 'let value = true; value &&= false; console.log(value);'),
'debugger': ('b', 'Remove the debugger statement.', 'debugger;'),
'TS2307': ('a', 'Module input generation/resolution must precede feature work.', 'import {value} from "./missing.a"; console.log(value);'),
'TS2304': ('a', 'Missing external/global declarations are setup or host surface prerequisites.', 'console.log(missingGlobal);'),
'TS2722': ('c', 'Prove a callback present or handle its absence.', 'function read(callback: (() => number) | undefined): number { return callback(); }'),
'TS2556': ('c', 'Arity/tuple proof must justify a spread into fixed parameters.', 'function read(a: number, b: number): number { return a + b; } const values: number[] = []; console.log(read(...values));'),
'TS2684': ('c', 'Callable receiver contract must be valid at the detached call.', 'function read(this: {count: number}): number { return this.count; } console.log(read());'),
'TS2366': ('b', 'Every admitted path must meet a required return contract.', 'function read(flag: boolean): number { if (flag) { return 1; } }'),
'TS2420': ('c', 'An implements relationship must satisfy the actual required fields.', 'interface Value { count: number; } class Empty implements Value {}'),
'TS2740': ('c', 'The structural target cannot invent fields absent from the source.', 'const values: number[] = {}; console.log(values);'),
'TS7006': ('b', 'Give genuinely implicit parameters truthful explicit types.', 'function read(value) { return value; } console.log(read(1));'),
'TS7031': ('b', 'Give destructured parameters truthful explicit types.', 'function read({count}) { return count; } console.log(read({count: 1}));'),
}
root.mkdir(exist_ok=True)
(root / 'support.a').write_text('export interface Item { count: number; }\nexport const Diagnostic = 1;\n')
manifest=[]
for index,(reason,(category,action,source)) in enumerate(examples.items(),1):
 directory=root/f'r{index:02d}';directory.mkdir(exist_ok=True)
 source = re.sub(r'console\.log\((.*?)\);', r'console.log(String(\1));', source)
 (directory/'main.a').write_text(source+'\n')
 manifest.append({'reason':reason,'category':category,'action':action,'path':str((directory/'main.a').relative_to(root.parent))})
(root.parent/'data/repro_manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
