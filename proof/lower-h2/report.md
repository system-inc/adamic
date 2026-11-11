# Adamic H2 lowering probe against Node

Tested 2026-11-11 on Linux x86_64, at detached main commit `d690f276749f91e6b36fe3a9726a5a453fcd7d39`.

## Findings

- `h2_plain.ts`: release, sanitized and JavaScript lowering all refuse `host.read` with `unbound-method`. No native executable exists to test a null receiver. Node catches a TypeError and prints `TypeError`.
- `h2_field.ts`: all three lowering paths refuse `host.read` with `unbound-method`. No native executable exists to test a null receiver. **Node prints `undefined`, not `TypeError`**: the supplied `holder.run()` expression calls with `holder` as its receiver, and `holder.name` is absent. The supplied program does not perform a plain receiverless call after reading the field.
- `h2_bind_control.ts`: all three paths lower successfully. Release native, sanitized native, generated JavaScript and Node all print `host`, with empty stderr and exit 0. **Native does not exhibit a null-receiver read, crash, sanitizer report or garbage output in the program that lowers.** The bound receiver is preserved.

These exact programs do not demonstrate an unsafe native null-receiver read at this revision. The two unbound reads are rejected before execution; the bound control behaves like Node.

## Setup and commands

Loaded `/workspace/adamic-tools/env.sh` first (PATH for the provided tools, `GOTOOLCHAIN=auto`, `TMPDIR=/tmp/adamic-gate`). Created `/tmp/adamic-gate` outside the repository. The initial checkout could not find the requested commit locally; fetched `origin main`, then successfully checked it out detached. An initial build attempt before submodule alignment failed because the TypeScript Go module was not yet present, and the earlier checkout’s cohere revision lacked `unicode_properties`; neither setup failure is counted as a probe result. After recursive initialization and alignment, the requested build succeeded with empty stdout/stderr and exit 0:

```sh
git checkout --detach d690f276749f91e6b36fe3a9726a5a453fcd7d39
git submodule update --init --recursive
mkdir -p /tmp/adamic-gate
go build -o /tmp/adamic ./cmd/adamic
```

Submodules: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`. Tool versions: Node `v24.19.0`, Go `go1.27.1`, clang `20.1.8`. `/tmp/adamic --version` confirms the requested full commit.

For each name, commands ran from `/workspace/adamic`:

```sh
/tmp/adamic build /tmp/p/<name>.ts -o /tmp/p/<name>_release
/tmp/p/<name>_release
/tmp/adamic build /tmp/p/<name>.ts -o /tmp/p/<name>_sanitize --sanitize
ASAN_OPTIONS=detect_leaks=0 /tmp/p/<name>_sanitize
/tmp/adamic js /tmp/p/<name>.ts > /tmp/p/<name>.mjs
node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/p/<name>.mjs
node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/p/<name>.ts
```

Default `adamic build` selects release mode. The harness sets `ASAN_OPTIONS=detect_leaks=0` for native runs (including the sanitized run). Native execution and generated-JavaScript execution occur only after successful builds/lowering; failed lowering never runs an empty or stale output file. Generated JavaScript is run as `.mjs` through `oracle/node.mjs`, matching the oracle’s `onJavaScriptBackend` → `onNode` path.

In the tables below strings use JSON escaping, so `""` means an empty stream and `\n` denotes a newline. “Not run” has no process exit status.

## h2_plain.ts

Exact input saved as `/tmp/p/h2_plain.ts`:

```ts
const host = { name: 'host', read(): string { return this.name; } };
const getRead = (): () => string => host.read;
const original = getRead();
try { console.log(original()); } catch (error) { console.log(error instanceof TypeError ? 'TypeError' : 'other'); }
```

| Stage | Exit | stdout | stderr |
| --- | --- | --- | --- |
| Node source run | 0 | "TypeError\n" | "" |
| Release build | 1 | "" | Exact stop below |
| Release native run | — | Not run: lowering failed | — |
| Sanitized build | 1 | "" | Exact stop below |
| Sanitized native run | — | Not run: lowering failed | — |
| JavaScript lowering | 1 | "" | Exact stop below |
| Generated JavaScript run | — | Not run: lowering failed | — |

Release build, sanitized build and JavaScript lowering all produce this identical stderr (including its final newline), with empty stdout and exit 1:

```text
adamic: /tmp/p/h2_plain.ts:2:37: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)
```

## h2_field.ts

Exact input saved as `/tmp/p/h2_field.ts`:

```ts
interface Holder { run: () => string; }
const host = { name: 'host', read(): string { return this.name; } };
const holder: Holder = { run: host.read };
try { console.log(holder.run()); } catch (error) { console.log(error instanceof TypeError ? 'TypeError' : 'other'); }
```

| Stage | Exit | stdout | stderr |
| --- | --- | --- | --- |
| Node source run | 0 | "undefined\n" | "" |
| Release build | 1 | "" | Exact stop below |
| Release native run | — | Not run: lowering failed | — |
| Sanitized build | 1 | "" | Exact stop below |
| Sanitized native run | — | Not run: lowering failed | — |
| JavaScript lowering | 1 | "" | Exact stop below |
| Generated JavaScript run | — | Not run: lowering failed | — |

Release build, sanitized build and JavaScript lowering all produce this identical stderr (including its final newline), with empty stdout and exit 1:

```text
adamic: /tmp/p/h2_field.ts:3:31: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)
```

## h2_bind_control.ts

Exact input saved as `/tmp/p/h2_bind_control.ts`:

```ts
const host = { name: 'host', read(): string { return this.name; } };
const bound = host.read.bind(host);
console.log(bound());
```

| Stage | Exit | stdout | stderr |
| --- | --- | --- | --- |
| Node source run | 0 | "host\n" | "" |
| Release build | 0 | "" | "" |
| Release native run | 0 | "host\n" | "" |
| Sanitized build | 0 | "" | "" |
| Sanitized native run | 0 | "host\n" | "" |
| JavaScript lowering | 0 | Generated JavaScript (verbatim appendix below) | "" |
| Generated JavaScript run | 0 | "host\n" | "" |

## Generated JavaScript stdout for the bound control

The following is the complete stdout of the successful `adamic js` invocation. Its stderr is empty and its exit status is 0. Executing this module through the oracle loader produces `host\n`, empty stderr and exit 0.

<details>
<summary>Verbatim generated JavaScript</summary>

```js
// Generated by adamic from h2_bind_control.ts. Do not edit.

import { encodeJson, decodeJson, fileStatus, panic, programArguments, readDirectory, readTextFile, utf8At, utf8Length, writeTextFile } from 'adamic';

const adamicCallableRepresentationCompatible = (from, to, fromMembers = 0, toMembers = 0) => {
    if (from === 0 || to === 0) return false;
    if (from !== 10 && to !== 10 && from !== to) return false;
    if (from !== 10 && to !== 10 && (fromMembers === 0 || toMembers === 0)) return from === to;
    const given = fromMembers !== 0 || from === 10 ? fromMembers : (1 << from);
    const wanted = toMembers !== 0 || to === 10 ? toMembers : (1 << to);
    return given !== 0 && wanted !== 0 && (given & wanted) === given;
};
const adamicViewCallableSignaturesMatch = (recorded, expected) => recorded!==undefined && recorded!==null && expected!==undefined && expected!==null && recorded.result!==0 && expected.result!==0 && (expected.result===255 || adamicCallableRepresentationCompatible(recorded.result, expected.result, recorded.resultMask, expected.resultMask)) && recorded.parameters.length===expected.parameters.length && recorded.parameters.every((value,index)=>adamicCallableRepresentationCompatible(expected.parameters[index], value, expected.parameterMasks?.[index], recorded.parameterMasks?.[index]));
const adamicViewCallableShape = (value, recorded, expected, expression, optional = false) => {
    if (optional && value === undefined) return undefined;
    let found = value === null ? "null" : adamicTypeOf(value);
    if (found === "function") {
        found = "function with unknown signature";
        if (recorded !== undefined && recorded !== null && expected !== undefined && expected !== null) {
            if (recorded.parameters.length !== expected.parameters.length) found = "function with arity " + recorded.parameters.length;
            else if (recorded.result === 0 || expected.result === 0) found = "function with unknown signature";
            else if (expected.result !== 255 && !adamicCallableRepresentationCompatible(recorded.result, expected.result, recorded.resultMask, expected.resultMask)) found = "function with incompatible result representation";
            else if (adamicViewCallableSignaturesMatch(recorded, expected)) return value;
            else found = "function with incompatible parameter representations";
        }
    }
    const wanted = expected === undefined || expected === null ? "function with known signature" : expected.name;
    panic("field read failed: " + expression + " expected " + wanted + ", found " + found);
};

const adamicUntaggedPlainSelect = (value, contracts, id, expression, declared, element = false) => {
 if(value===undefined && contracts[id-1]?.Undefined) return element ? true : 0;
 const slot = (object,name) => {
  if(object===null || typeof object!=='object' || object instanceof Map) return undefined;
  const field=Object.getOwnPropertyDescriptor(object,name);
  if(field===undefined || !Object.hasOwn(field,'value') || adamicFieldReadiness.get(object)?.has(name)) return undefined;
  return field;
 };
 const active=new Map();
 const matches=(value,id,depth=0)=>{
  // A defined value can only use the non-undefined arm of this optional union.
  // Resolve it in this frame, leaving branching unions to ordinary backtracking.
  const optional=contracts[id-1];
  if(value!==undefined && optional?.Kind===4 && (!optional.Unsupported || optional.Unsupported==='untagged object union') && !optional.Nominal && optional.Members?.length===2){
   const [left,right]=optional.Members;
   if(contracts[left-1]?.Kind===7){id=right;depth++;}
   else if(contracts[right-1]?.Kind===7){id=left;depth++;}
  }
  const reference=value!==null && typeof value==='object';
  let seen=reference ? active.get(value) : undefined;
  if(seen?.has(id)) return true;
  if(reference){if(seen===undefined){seen=new Set();active.set(value,seen);}seen.add(id);}
  try {
  const contract=contracts[id-1];
  if(contract===undefined || (contract.Unsupported && contract.Unsupported!=='untagged object union') || contract.Nominal) return false;
  if(value===undefined && (contract.Undefined || contract.Kind===7)) return true;
  if(contract.Kind===1){
   const wanted=contract.Of===1 || contract.Of===7?'number':contract.Of===2 || contract.Of===9?'boolean':contract.Of===3?'string':'unknown';
   if(typeof value!==wanted) return false;
   return !contract.Allowed?.length || contract.Allowed.some(literal=>value===(literal.Of===1?literal.Number:literal.Of===2?literal.Boolean:literal.String));
  }
  if(contract.Kind===5){
   if(!contract.Result && !contract.DiscardResult) return false;
   const parameters=(contract.Parameters || []).map(id=>contracts[id-1]?.Of || 0);
   const expected={parameters,result:contract.DiscardResult?255:contracts[contract.Result-1]?.Of || 0};
   const recorded=adamicUntaggedCallableRecorded(value);
   if(contract.ProducerCertified && !(contract.Functions || []).includes(recorded?.function)) return false;
   return adamicViewCallableSignaturesMatch(recorded,expected);
  }
  if(contract.Kind===3) return Array.isArray(value) && !adamicTupleObjects.has(value);
  if(contract.Kind===6) return value===null;
  if(contract.Kind===7) return value===undefined;
  if(contract.Kind===4){for(const member of contract.Members || []){if(matches(value,member,depth+1)) return true;}return false;}
  if(contract.Kind===2 && contract.Of===10) return value!==null && (typeof value==='object' || typeof value==='function' || value instanceof AdamicClosure);
  if(contract.Kind!==2 || value===null || typeof value!=='object' || (Array.isArray(value) && !adamicTupleObjects.has(value)) || value instanceof Map) return false;
  if(contract.FixedTuple && (!Array.isArray(value) || !adamicTupleObjects.has(value) || value.length !== contract.Tuple.length)) return false;
  const fields=contract.Fields || [];
  const ownKind=fields.find(field=>field.Name==='kind' && !field.Optional && contracts[field.Contract-1]?.Kind===1 && [1,2,3].includes(contracts[field.Contract-1]?.Of));
  if(ownKind){const actual=slot(value,'kind');return actual!==undefined && matches(actual.value,ownKind.Contract,depth+1);}
  const tags=fields.filter(field=>!field.Optional && contracts[field.Contract-1]?.Kind===1 && contracts[field.Contract-1]?.Allowed?.length);
  if(tags.length) return tags.every(field=>{const actual=slot(value,field.Name);return actual!==undefined && matches(actual.value,field.Contract,depth+1);});
  for(const field of fields){const actual=slot(value,field.Name);if(!(actual===undefined ? field.Optional && !Object.hasOwn(value,field.Name) : matches(actual.value,field.Contract,depth+1))) return false;}
  return true;
  } finally {if(reference){seen.delete(id);if(seen.size===0) active.delete(value);}}
 };
 if(element) return matches(value,id);
 if(contracts[id-1].FixedTuple && matches(value,id)) return id;
 for(const member of contracts[id-1].Members || []){if(matches(value,member)) return member;}
 panic('field read failed: '+expression+' matches no member of '+declared+'; expected '+declared+', found '+(value===null?'null':value===undefined?'undefined':Array.isArray(value)?'array':value instanceof Map?'Map':adamicTypeOf(value)));
};

const adamicUntaggedCallableRecorded=(value)=>((code)=>code === function_2_bound_function ? ({...{parameters: [], result: 3, name: "bound_function", parameterMasks: [], resultMask: 0},function:2}) : undefined)(value instanceof AdamicClosure ? value.code : value);
const adamicArrayContracts=null;

const adamicArrayStorage = new WeakMap();
const adamicViewArray = (array, references, kind) => { adamicArrayStorage.set(array,{references,kind}); return array; };
// Checked per read, never cached: an alias can push between reads.
const adamicViewElement = (array,index,relative,check) => {
 if(relative){index=Math.trunc(index)||0;if(index<0)index+=array.length;}
 if(!Number.isInteger(index)||index<0||index>=array.length)return undefined;
 const contract=adamicArrayContracts?.[check.Contract-1];
 const storage=adamicArrayStorage.get(array);
 const references=contract && [3,4,5,6,8,10].includes(contract.Of);
 const fail=kind=>panic('element read failed: '+check.View+' is not a '+check.ViewType+'; expected '+check.ViewType+', found '+kind);
 if(check.Contract && storage && storage.references && !references)fail(storage.kind);
 const value=array[index];
 if(check.Contract && !adamicUntaggedPlainSelect(value,adamicArrayContracts,check.Contract,check.Where,check.ViewType,true))fail(value===null?'null':Array.isArray(value)?'array':value instanceof Map?'Map':adamicTypeOf(value));
 return value;
};
const adamicCheckedArrayRoad = (array,callback,initial,method,check,admitsUndefined,elementType) => {
 const count=array.length, mapped=[];
 let carried=initial;
 const last=method==='findLast'||method==='findLastIndex';
 const search=last||method==='find'||method==='findIndex';
 for(let step=0;step<count;step++){
  const index=last?count-1-step:step;
  if(index>=array.length){
   if(method==='map')panic('map: the array shrank while it was being mapped');
   if(!search)continue;
   if(!admitsUndefined)panic(method+': index '+index+' is undefined; element type '+elementType+' does not admit undefined');
  }
  const element=adamicViewElement(array,index,false,check);
  const answer=adamicCall(callback,method==='reduce'?[carried,element,index,array]:[element,index,array]);
  if(method==='reduce')carried=answer;
  else if(method==='map')mapped.push(answer);
  else if(method==='filter'){if(answer)mapped.push(element);}
  else if(method==='some'&&answer)return true;
  else if(method==='every'&&!answer)return false;
  else if(search&&answer){
   if(method==='find'||method==='findLast')return adamicViewElement(adamicViewArray([element],referencesForCheck(check),'reference storage'),0,false,check);
   return index;
  }
 }
 return method==='map'||method==='filter'?mapped:method==='reduce'?carried:method==='some'?false:method==='every'?true:search?(method==='find'||method==='findLast'?undefined:-1):undefined;
};
const referencesForCheck=check=>[3,4,5,6,8,10].includes(adamicArrayContracts?.[check.Contract-1]?.Of || 10);
class AdamicClosure {
	constructor(code, cells, receiver = false, length = 0) {
		this.code = code;
		this.cells = cells;
		this.receiver = receiver;
		this.length = length;
	}
}
const adamicCanonical = (identity, code, cells, length) => { const values = identity.functions ??= new Map(); if (!values.has(code)) values.set(code, new AdamicClosure(code, cells, false, length)); return values.get(code); };
const adamicTypeOf = (value) => value instanceof AdamicClosure ? 'function' : typeof value;
const adamicTupleObjects = new WeakSet();
const adamicTuple = (values) => { adamicTupleObjects.add(values); return values; };
const adamicViewObject = (value) => value !== null && typeof value === "object" && (!Array.isArray(value) || adamicTupleObjects.has(value)) && !(value instanceof Map);
const adamicFieldReadiness = new WeakMap();
const adamicFieldRepresentations = new WeakMap();
const adamicRecordFieldTypes = (object, types) => { adamicFieldRepresentations.set(object, {...adamicFieldRepresentations.get(object), ...types}); return object; };
const adamicViewWrite = (object, name, value, type, declared) => { if (adamicFieldRepresentations.get(object)?.[name] !== type && !(adamicFieldRepresentations.get(object)?.[name] === 13 && ([3,4,5,6,8,10].includes(type))) && !(adamicFieldRepresentations.get(object)?.[name] === 10 && type <= 2) && !(adamicFieldRepresentations.get(object)?.[name] === 7 && type === 1)) adamicViewField(object, name, "<write>." + name, type, declared || adamicViewTypeNames[type] || "unknown"); adamicWriteField(object, name, value); adamicRecordFieldTypes(object, {[name]: adamicFieldRepresentations.get(object)?.[name] === 10 && type <= 2 ? 10 : type}); };
const adamicUninitializedFields = (object, names) => { adamicFieldReadiness.set(object, new Set(names)); return object; };
const adamicReadField = (object, name, expression, optional = false, allowAbsent = false, fieldView = false) => {
    if (optional && (object === undefined || object === null)) return undefined;
    if (object === undefined || object === null) panic(fieldView ? "field read failed: " + expression + " is not initialized; expected " + fieldView + ", found missing" : "read before assignment: field '" + name + "' in " + expression);
    if (!Object.hasOwn(object, name) && typeof object === 'function' && adamicClassIdentities.has(Object.getPrototypeOf(object))) return adamicReadField(Object.getPrototypeOf(object), name, expression, false, allowAbsent, fieldView);
    if (allowAbsent && !Object.hasOwn(object, name)) return undefined;
    if (!Object.hasOwn(object, name) || adamicFieldReadiness.get(object)?.has(name)) panic(fieldView ? "field read failed: " + expression + " is not initialized; expected " + fieldView + ", found " + (!Object.hasOwn(object, name) ? "missing" : "uninitialized") : "read before assignment: field '" + name + "' in " + expression);
    return object[name];
};
const adamicPlaceholderNarrow = (value, type, message) => {
    const valid = type === 1 ? typeof value === "number" : type === 2 ? typeof value === "boolean" : type === 3 ? typeof value === "string" : type === 4 || type === 14 ? value !== null && typeof value === "object" && !Array.isArray(value) && !(value instanceof Map) && !ArrayBuffer.isView(value) : type === 5 ? Array.isArray(value) : type === 6 ? value instanceof Map : type === 8 ? typeof value === "function" : false;
    if (!valid) panic(message);
    return value;
};
const adamicViewTypeNames = Object.freeze({1: "number", 2: "boolean", 3: "string", 4: "object", 5: "array", 6: "Map"});
const adamicViewField = (object, name, expression, type, expected = adamicViewTypeNames[type], allowed = []) => {
    const value = adamicReadField(object, name, expression, false, false, expected);
    const valid = type === 1 ? typeof value === "number" : type === 2 ? typeof value === "boolean" : type === 3 ? typeof value === "string" : type === 4 ? adamicViewObject(value) : type === 5 ? Array.isArray(value) : type === 6 ? value instanceof Map : false;
    if (!valid) panic("field read failed: " + expression + " is not a " + expected + "; expected " + expected + ", found " + (value === undefined ? "nullish" : value === null ? "nullish" : Array.isArray(value) ? "array" : value instanceof Map ? "Map" : typeof value));
    if (allowed.length && !allowed.includes(value)) panic("field read failed: " + expression + " expected " + expected + ", found " + typeof value + " " + value);
    return value;
};
const adamicPlaceholderViewField = (object, name, expression, type, expected, allowed, optional, absent) => {
    if (optional && (object === undefined || object === null)) return undefined;
    if (object !== undefined && object !== null) {
        if (!Object.hasOwn(object, name) && typeof object === 'function' && adamicClassIdentities.has(Object.getPrototypeOf(object))) return adamicPlaceholderViewField(Object.getPrototypeOf(object), name, expression, type, expected, allowed, false, absent);
        if (absent && !Object.hasOwn(object, name)) return undefined;
        if (Object.hasOwn(object, name) && (object[name] === undefined || object[name] === null)) return object[name];
    }
    return adamicViewField(object, name, expression, type, expected, allowed);
};
const adamicCheckedViewCast = (object, field, type, allowed, message) => allowed.includes(adamicViewField(object, field, field, type)) ? object : panic(message);
const adamicDefineField = (object, name, value, enumerable, ready, type) => { if (type !== undefined) adamicRecordFieldTypes(object, {[name]: type}); Object.defineProperty(object, name, {value, writable: true, enumerable, configurable: true}); if (ready) adamicFieldReadiness.get(object)?.delete(name); else { let fields = adamicFieldReadiness.get(object); if (!fields) adamicFieldReadiness.set(object, fields = new Set()); fields.add(name); } };
const adamicSpreadFields = (object, expression) => { const result = {}; if (object !== undefined && object !== null) for (const name of Object.keys(object)) Object.defineProperty(result, name, {value: adamicReadField(object, name, expression), enumerable: true, writable: true, configurable: true}); return adamicRecordFieldTypes(result, adamicFieldRepresentations.get(object) || {}); };
const adamicWriteField = (object, name, value) => { object[name] = value; adamicFieldReadiness.get(object)?.delete(name); };
import { createHash as adamicNodeCreateHash } from 'node:crypto';
class AdamicIterator { constructor(next) { this.next = next; } }
class AdamicHostObject {}
const adamicHostMemoryUsage = usage => Object.setPrototypeOf(usage, AdamicHostObject.prototype);
class AdamicCollectionIterator extends AdamicIterator {}
const adamicCollectionIterator = (collection, part) => { const iterator = collection[part](); return new AdamicCollectionIterator(new AdamicClosure(() => { const step = iterator.next(); if (part === 'entries' && !step.done) adamicTuple(step.value); return step; }, [])); };
const adamicStringIterator = string => { const iterator = string[Symbol.iterator](); return new AdamicIterator(new AdamicClosure(() => iterator.next(), [])); };
const adamicJSONValue = (value) => {
 if (value instanceof AdamicClosure) return () => undefined;
 if (Array.isArray(value)) return value.map(adamicJSONValue);
 if (value !== null && typeof value === 'object' && !(value instanceof Map) && !(value instanceof Set)) {
  const copy = {};
  for (const key of Object.keys(value)) Object.defineProperty(copy, key, {value: adamicJSONValue(value[key]), enumerable: true, configurable: true, writable: true});
  return copy;
 }
 return value;
};
const adamicJSONStringify = (value, replacer, space) => JSON.stringify(adamicJSONValue(value), replacer, space);
const adamicDirect = (fn, values) => fn(...values);

const adamicRecordGet = (record, key) => {
 if (Object.hasOwn(record, key)) return record[key];
 if (Object.hasOwn(Object.prototype, key)) panic("record member '" + key + "' is missing; records hold own keys only");
 return undefined;
};
const adamicRecordHas = (key, record) => {
 if (Object.hasOwn(record, key)) return true;
 if (Object.hasOwn(Object.prototype, key)) panic("record member '" + key + "' is missing; records hold own keys only");
 return false;
};
const adamicRecordSet = (record, key, value) => {
 if (key === '__proto__') panic('NotYet: record assignment to __proto__ requires the Object.prototype setter; use an own data property');
 record[key] = value;
 return value;
};
const adamicBindMethod = (object, name, optional, make) => { const own = Object.hasOwn(object, name); const methods = own ? undefined : adamicStaticMethods.get(object); const fn = own ? object[name] : methods && Object.hasOwn(methods, name) ? methods[name] : object[name]; if (optional && fn == null) return undefined; const invoke = fn instanceof AdamicClosure ? function(...values) { return adamicCall(fn, fn.receiver ? [this, ...values] : values); } : function(...values) { return adamicDirect(fn, [this, ...values]); }; const bound = invoke.bind(object); return make({code: (self, values) => bound(...values)}, object); };
const adamicBoundCall = (self, values) => { const fn = self.cells[0].value; return adamicCall(fn, fn.receiver ? [self.cells[1].value, ...values] : values); };
const adamicCall = (closure, values) => closure.code(closure, values);
const adamicStaticMethods = new WeakMap();
const adamicCallee = (object, name) => { if (Object.hasOwn(object, name)) { const fn = object[name]; return fn.receiver ? { code: (closure, values) => adamicCall(fn, [object, ...values]) } : fn; } const methods = adamicStaticMethods.get(object); const fn = methods && Object.hasOwn(methods, name) ? methods[name] : object[name]; return { code: (closure, values) => fn(object, ...values) }; };
const adamicOptionalCall = (object, name, values) => object === undefined ? undefined : adamicCall(adamicCallee(object, name), values());
const adamicOptionalMethodCall = (object, name, values, optionalObject) => { if (optionalObject && object == null) return undefined; const own = Object.hasOwn(object, name); const methods = own ? undefined : adamicStaticMethods.get(object); const fn = own ? object[name] : methods && Object.hasOwn(methods, name) ? methods[name] : object[name]; if (fn == null) return undefined; return own ? adamicCall(fn, fn.receiver ? [object, ...values()] : values()) : fn(object, ...values()); };
const adamicParallelMap = (items, work) => items.map((item, index) => adamicCall(work, [item, index]));
const adamicVisit = (array, method, callback) => array[method]((element, index, all) => adamicCall(callback, [element, index, all]));
const adamicMap = (array, callback) => {
	const count = array.length;
	const mapped = [];
	for (let index = 0; index < count; index++) {
		if (index >= array.length) panic('map: the array shrank while it was being mapped');
		mapped.push(adamicCall(callback, [array[index], index, array]));
	}
	return mapped;
};
const adamicFind = (array, method, callback, admitsUndefined, elementType) => {
	const count = array.length;
	const last = method === 'findLast' || method === 'findLastIndex';
	const valueResult = method === 'find' || method === 'findLast';
	for (let step = 0; step < count; step++) {
		const index = last ? count - 1 - step : step;
		if (index >= array.length && !admitsUndefined) panic(`${method}: index ${index} is undefined; element type ${elementType} does not admit undefined`);
		const element = array[index];
		if (adamicCall(callback, [element, index, array])) return valueResult ? element : index;
	}
	return valueResult ? undefined : -1;
};
const adamicCollectionVisit = (collection, callback) => collection.forEach((value, key, all) => adamicCall(callback, [value, key, all]));
const adamicFrom = (length, callback) => Array.from({ length }, (element, index) => adamicCall(callback, [element, index]));
const adamicDefinedNull = (value, message) => value === null ? panic(message) : value;
const adamicDefined = (value, message) => value === undefined ? panic(message) : value;
const adamicSort = (array, callback) => array.sort((left, right) => adamicCall(callback, [left, right]));
const adamicReduce =(array, callback, initial) => array.reduce((carried, element, index, all) => adamicCall(callback, [carried, element, index, all]), initial);
const adamicSetIndex = (array, index, value) => {
	if (!(Number.isInteger(index) && index >= 0 && index < array.length)) panic(`index ${index} is outside an array of length ${array.length}`);
	array[index] = value;
};
const adamicCast = (object, field, allowed, message) => allowed.includes(object[field]) ? object : panic(message);
const adamicUnready = (name) => { throw new ReferenceError(`Cannot access '${name}' before initialization`); };

let local_0_host;
let ready_0 = false;
let local_1_bound;
let ready_1 = false;

function function_0_object_method(self, values) {
	let local_2_this = values[0];
	return local_2_this["name"];
}

function function_1_bind_factory(local_3_bound_value, local_4_bound_value) {
	const local_3_bound_value_cell = { value: local_3_bound_value, ready: true };
	const local_4_bound_value_cell = { value: local_4_bound_value, ready: true };
	return new AdamicClosure(function_2_bound_function, [local_3_bound_value_cell, local_4_bound_value_cell], false, 0);
}

function function_2_bound_function(self, values) {
	return adamicBoundCall(self, []);
}

local_0_host = ({"name": "host", "read": new AdamicClosure(function_0_object_method, [], true, 0)});
ready_0 = true;
local_1_bound = adamicBindMethod(local_0_host, "read", false, (target, receiver) => function_1_bind_factory(target, receiver));
ready_1 = true;
console.log(adamicCall(local_1_bound, []));
```

</details>
