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
