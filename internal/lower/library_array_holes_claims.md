# Array holes claims

Base: c762b555e7c0b39b105e2d9208732a5a311b2e6f. Oracle: Node 24.19.0.

This is the pre-implementation inventory. "Refused" below describes the
current base, not a claim that a hole-aware implementation has landed. The
length constructor is NotYet; array literal elisions are not lowered. Thus
none of the following operations currently accepts an array created with holes.
The existing runtime is dense and cannot distinguish a hole from undefined.
Each row must be updated with its implemented behavior or its conservative
refusal rule before length construction is enabled generally.

| Existing operation | Current hole claim | Required Node behavior |
| --- | --- | --- |
| length read | Refused at hole creation | Count holes |
| indexed read, optional indexed read, at | Refused at hole creation | Return undefined for absence |
| indexed write and compound update | Refused at hole creation | Fill an own slot |
| push | Refused at hole creation | Append own slots |
| pop | Refused at hole creation | Remove final slot; absent gives undefined |
| fill, new Array(n).fill(v) | Only complete immediate fill is built; partial constructor fill is refused | Create own slots in the range |
| slice | Refused at hole creation | Preserve holes |
| concat | Refused at hole creation | Preserve holes |
| splice | Refused at hole creation | Preserve holes in removed and shifted slots |
| reverse | Refused at hole creation | Move absence with its index |
| sort, toSorted | Refused at hole creation | sort puts holes after undefined; toSorted materializes undefined |
| map | Refused at hole creation | Skip absent indices and preserve output holes |
| forEach, filter, some, every | Refused at hole creation | Skip absent indices |
| reduce with initial value | Refused at hole creation | Skip absent indices |
| reduce without initial value | Refused independently, requires initial value | First present slot seeds accumulator; empty throws |
| find, findIndex, findLast, findLastIndex | Refused at hole creation | Visit holes as undefined |
| indexOf, lastIndexOf | Refused at hole creation | Skip holes |
| includes | Refused at hole creation | Read holes as undefined |
| join, nested join | Refused at hole creation | Empty text for holes |
| flat, flatMap | Refused at hole creation | Skip holes in flattened layers and callback input |
| copyWithin | Refused at hole creation | Copy absence, deleting destination slots |
| with, toReversed, toSpliced | Refused at hole creation | Materialize holes as undefined |
| for-of, values(), entries() | Refused at hole creation | Yield undefined for holes |
| keys() in for-of | Refused at hole creation | Yield every index, including absent slots |
| array spread and call spread | Refused at hole creation | Materialize holes as undefined |
| destructuring, rest bindings | Refused at hole creation | Iterator reads materialize undefined |
| Array.from({length}, mapper) | Built dense output; no hole input accepted by this form | Call mapper for every index |
| Array.from(iterable, optional mapper) | Refused at hole creation | Iterator reads materialize undefined |
| Array.isArray | Refused at hole creation | True for holey arrays |
| JSON.stringify | Refused at hole creation | Serialize holes as null |
| console.log / inspect | Refused at hole creation | Group consecutive holes as empty items |
| Object operations on arrays | Array operands have their own existing NotYet restrictions | Own enumeration skips holes |
| constructor and Array.prototype method name/length/typeof observations | Built; do not inspect elements | Independent of hole presence |
| new Map/Set from arrays, collection spread consumption | Refused at hole creation | Consume iterator values, including undefined |

Already refused independently of hole creation: length assignment, array `in`,
array `hasOwnProperty`, array literal elisions, delete, shift, unshift,
reduceRight, toLocaleString, and constructor forms other than immediate complete
`new Array(n).fill(v)`. Array.of has no existing lowering.

The first acceptance source is the unchanged scanner probe from
8d1b1141289fcd7340231b6950eacf6aa7f3a91f:
`stage3/drivers/scanner/probes/array-length-constructor.a`. It prints `4`.

No new runtime representation, performance result, mutant kill, or test262
agreement is claimed by this inventory commit.

## Constructor milestone

Implemented: numeric `new Array(n)`, `Array(n)`, and numeric-element generic
length constructors; length reads and indexed numeric reads and writes;
catchable invalid-length RangeError with separate nominal constructor identity.
The unchanged scanner probe is copied into the oracle fixtures. A numeric map
holds present slots, including explicitly present undefined in optional-number
arrays. No per-hole allocation is required at the maximum length.

This milestone is deliberately limited: other generic element representations
remain named NotYet. Whenever the lowered program contains ArrayHoles,
`checkArrayHoles` treats all arrays as potentially holey, including parameters,
fields and returns. It refuses all unported array methods, array for-of and
spreads, Array.from, collection constructors, object reflection, JSON,
Buffer/host calls, and spread-sensitive Math/String consumers. Dense-only
programs keep the old access and store functions. No performance measurement
or complete operation support is claimed yet.
