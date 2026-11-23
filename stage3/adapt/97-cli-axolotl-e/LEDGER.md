## 7. tsc passes a caught value's message, undefined or a non-string, to a callback declared `(message: string)`

Status: fixed in our fork by JavaScript-changing adaptation 97e, which converts
the message to text under Adamic's Nov 1 ruling (#axochk); draft issue, not
filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/program.ts createGetSourceFile and
createWriteFileMeasuringIO, each `catch (e) { if (onError) { onError(e.message); } }`;
src/compiler/types.ts's two declarations of that callback,
WriteFileCallback's `onError?: (message: string) => void` (types.ts:4584) and
CompilerHost.getSourceFile's (types.ts:8142). Lines are in the adapted tree at
adaptation 94's tip.

Both functions wrap a supplied callback: readFile for getSourceFile, the
writer, createDirectory and directoryExists for writeFile. Nothing constrains
what that callback throws, and the catch hands `e.message` to onError as it
is. The declared `string` holds only when the thrown value is an object with a
string message. tsc's own consumers then pass that message to
formatStringFromArgs, which fills each slot through `Debug.checkDefined`
(utilities.ts:8584). An undefined message fails there with `Debug Failure.`
instead of reporting `Cannot read file` or `Could not write file`.

Node 24.14.1 counterexample: a compiler host whose readFile throws
`{ message: 42 }`, then `{}`.
Before: onError receives the number 42, then undefined. getSourceFile returns
the empty source file both times. A thrown string or number also hands onError
undefined. A thrown null or undefined throws TypeError reading message out of
getSourceFile instead of calling onError. createWriteFileMeasuringIO's catch
gives the same results when the writer throws them. Adaptation 47's README
records these as the counterexamples that declined both sites.
After: onError receives the string "42", then the string "undefined", and so
does every other thrown value that reaches it. The consumer's
`Cannot read file 'a.ts': undefined.` replaces the Debug Failure. A thrown
null or undefined still throws the same TypeError.

Our patch: stage3/adapt/97-cli-axolotl-e/adapt.cjs, under Adamic's ruling of
Nov 1 (#axochk), "convert, don't cast". Widening onError to unknown was
measured and breaks five places in the tree, and a cast to the primitive is
refused, so the declaration stays `string` and the value is converted:

```diff
+import { propertyOf } from "./hostErrors.js";
-                onError(e.message);
+                onError(String(propertyOf(e, "message")));
```

at both sites. propertyOf is adaptation 94's helper. It returns exactly what
the plain read returns, and it throws V8's own TypeError on null and undefined.
The global String then turns what it returns into text. JavaScript changes at
exactly those two lines. A string message, an Error's for instance, passes
through unchanged, and a null or undefined thrown value throws the same
TypeError, with propertyOf's frame on top. The conversion keeps the meaning,
the message as text, and turns the crash into the diagnostic tsc meant. The
witnesses, the mutant and the transpile diff are in the adaptation's README.

### Draft GitHub issue

Title: getSourceFile and writeFile pass a non-string or undefined message to their `onError(message: string)` callback

Version: TypeScript 6.0.3, reproduced on Node 24.14.1.

Minimal reproduction:

```sh
mkdir tsc-onerror-message-repro
cd tsc-onerror-message-repro
npm install --no-audit --no-fund typescript@6.0.3
```

Save this as repro.cjs and run `node repro.cjs`:

```js
const ts = require("typescript");
const host = ts.createCompilerHost({});
for (const thrown of [{ message: 42 }, {}]) {
    host.readFile = () => { throw thrown; };
    host.getSourceFile("a.ts", ts.ScriptTarget.Latest, message => console.log(typeof message, message));
}
```

Expected: onError is declared `(message: string) => void` in CompilerHost
and WriteFileCallback, so either it receives a string or the declaration says
it may not.

Actual:

```text
number 42
undefined undefined
```

The compiler host's writeFile has the same catch: replacing `ts.sys.writeFile`
with a function throwing the same values prints the same two lines. A readFile
that throws null or undefined makes getSourceFile throw
`TypeError: Cannot read properties of null (reading 'message')` from inside
its catch.

Suggested upstream change: either convert the caught value to a string before
calling onError, for example with a message read that falls back to
`String(e)`, or declare the callback's parameter as what it receives. Either
way, the declaration and the call agree. This draft makes no claim about
which one upstream prefers. Our fork converts the message read with String,
so tsc's own consumers report `Cannot read file 'a.ts': undefined.` where an
undefined message used to fail Debug.checkDefined in formatStringFromArgs.
