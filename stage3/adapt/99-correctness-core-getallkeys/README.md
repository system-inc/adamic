# getAllKeys belongs to the fourslash harness

Adamic's ruling (b), task b5w3ycg, Nov 7 11:06Z, moves getAllKeys from compiler core to its one caller. Object.getPrototypeOf remains refused by rulings.md:83. The harness runs on Node and is never lowered. A future native requirement returns to Adamic.

The adapter removes the entire exported core declaration and its internal JSDoc. It changes the complete `const keys` statement inside TestState's constructor's memoWrap function to call a module-local function appended to fourslashImpl.ts. A named pushIfUnique import keeps the moved function body byte-identical. No compiler, library admission, fallback, or other adapter is modified.

owners.json pins the complete core declaration including its comment, the complete caller statement before and after, the new local function, and its helper import. The four pieces must all be original or all adapted. Parsing, lexical ancestry, unique bindings, namespace spelling, the following JSDoc boundary, planned-state verification, exact body equality, and rereads precede writes. Valid replay performs no writes. Contract fixtures cover owner drift, orphaned comments, duplicate locals, partial moves, no-write refusals, unrelated edits, LF and CRLF.

The pre-write search of every earlier adapt.cjs and owners JSON found no getAllKeys pin and only MultiMap's separate TestState.rangesByText owner in fourslashImpl.ts. Adapter 52 edits nearby signatures, not this caller statement. That earlier owner is untouched.

The existing reach measurement is f9d57ee2d8213572b02748c323b43b85d48d6a37, typescript/getallkeys-reach, stage3/census/getallkeys/REPORT.md: 6,566 calls, all at fourslashImpl.ts:461, and zero namespace reads. These are prior measurements, not a new instrumented count. The accepted lane uses ordinary emitted JavaScript.

supervise.py retains actual child exits and stopping reasons separately, bounds silence to 30 minutes, and cleans each process group. All command output goes to files. prepare-census.py uses the cited next-stops-after-wave procedure, modifying only physical scratch Go filenames to end in .go.txt. The ordinary compiler sources remain unchanged.
