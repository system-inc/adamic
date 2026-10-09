# Kinds dropped after checking main

## Any-array storage

The historical table has 30 `an array of any` sites. Current main's latent census
has zero such observations. Both actual compilation controls below emit C on
main without compiler edits:

```a
const values: any[] = [1];
console.log(String(values.length));
```

```a
function size(values: any[]): number { return values.length; }
console.log(String(size([1])));
```

This kind's storage/length operation is already implemented and is dropped from
41. These controls do not certify arbitrary element reads as typed or prove all
historical array sites now compile. Explicit any tokens within other contracts
remain in REMAINING.md. No array-storage adaptation mutant is added because
there is no new array-storage adaptation.

## With

No WithStatement syntax occurs in the compiler source. The main probe with a
with statement fails TS1101 and TS2410, so this is absent source, not a landed
compiler feature. There is nothing to adapt or restore as a corpus mutant.
Debugger is not dropped: it remains refused and is covered in DEBUGGER.md.
