# getOwnValues road (a), measurement first

Adamic, task b5w3ycg, Nov 7 00:12Z, ruled a real isArray check, filter(() => true)
on the array arm, and the original own-key loop on the narrowed dictionary arm.
The original MEASURE.md remains historical. No amendment has been built.

Base: origin/main 5a3e6607bb58bca84f65e01a4d153b3a2ba8b301 plus the requested
99z efa9c7dd7ffd4504068e6beeae26a2365eb145f1, which has not landed on main.
Merged base: b74a665d4069f959a5bc71f555da4d77ff83d4c7. Branch history and all
plain pushes are preserved. Every new commit uses the requested dependency line.
The main adaptation inputs are byte-identical to the earlier complete apply;
the measurement source copies that complete tree and applies exact 99z.
The counted compiler and harness builds both exited 0.

## Node witnesses, completed before building an amendment

The exact original getOwnValues, the proposed body with 30's ! retained, and
tsc's exact isArray helper are extracted by getownvalues-witness.cjs.
Both bodies return byte-identical Node stdout, empty stderr and exit 0 on:

| Witness | Result |
| --- | --- |
| Empty array written at index 5 | Write grows length 0 to 6; filtered length 1, x |
| [undefined, 'x'] | Filtered length 2, undefined then x |
| Empty array | Filtered length 0 |
| {'1':'a', '0':'b'} | Filtered length 2, b then a |

getownvalues-node.tar.gz retains the exact owners, sources, output bytes, base
and preparation provenance, and the excluded counter seeds. The seeds verify
that the counter distinguishes holes from present undefined, detects an extra
own property, separates a Date from a plain record, and handles VM realms.
The ordinary intrinsic length property is recorded as length, not counted as
an extra non-index array property. Symbols and non-enumerable extra properties
are included. Accessor indices are reported separately rather than invoking
getters during observation.

Step A continues with all lane shards and both backend witnesses. Step B is
conditional on the complete count and native sparse behavior. No Step B gate
is claimed here. The counter exists only in a scratch tree and is not in any
adaptation transformation.

## Checked native and JavaScript backends, completed

All four full proposed-body witnesses fail compilation identically on both
backends at fixture line 8:37, the narrowed collection read at filter.
Lower's localRead, internal/lower/locals.go:318, asks arrayPredicateDomain
(internal/lower/library_array_predicate.go:97) about every original union
member. That predicate does not accept the structural MapLike member.
A separate control replaces only the array return with [] so the record
branch can be measured: both backends stop at its for-in collection read,
fixture line 10:23, localRead at locals.go:341. Neither full arm currently
lowers in this union. No compiler rule was changed or suppressed.

The independently typed array filter and dictionary loop both compile and
match Node byte for byte, on native --sanitize and the JavaScript backend.
The array controls also match Node for [undefined, 'x'] and the empty array:
filter keeps the present undefined and returns an empty array respectively.
The dictionary control returns b then a for {'1':'a','0':'b'}.

The independent sparse-write control compiles on both backends. Both print
`before-write-length=0`, then exit 70 at a[5] = 'x', with exact stderr:

```text
adamic: panic: index 5 is outside an array of length 0
```

There is no length growth, gap fill or filter execution. Native
internal/native/runtime/array.c:504, adamic_array_set, rejects the absent
slot; the panic is at line 512's bounds message. The JavaScript backend's
adamicSetIndex helper, internal/javascript/javascript.go:111, makes the
same check before assignment. Node instead grows to length 6 with five
holes, then filter returns one element, x. Thus native cannot hold this
hole. If the full count finds any sparse arrays, Step B's explicit sparse
condition fails. The lane count is still being completed.

getownvalues-native.tar.gz retains all source, generated JavaScript, compile
and run output bytes, build provenance and relevant lowering/runtime source.
The JavaScript cases use the normal oracle/node.mjs runtime. An initial
invocation without that host runtime could not resolve package adamic; it
is retained as a setup attempt and excluded from backend verdicts.
