# Family 65, typed half: copied-slot presence bits

Pinned by `internal/oracle/typed_presence_test.go` (#2hzryxh, project #bbssdcy). Accepted by Adamic: a compiler change that alters what a rule admits.

## What these programs hold

A spread copies presence per slot. The copy holds what its actual source holds as own keys at the spread, never what the source's type says, and an absent spread (`...undefined`) adds nothing. `in` with a fixed key on a receiver the checker proves one object type answers from those bits, then from the prototype's names, as Node does. A receiver of type `unknown` or `object` keeps the family's stop.

## The rulings, by time and seat

- Adamic, Nov 22 23:30Z: the copied-slot presence bits are her Nov 22 ruling (a spread copies presence per slot) applied to `in` on a proven typed receiver; the `unknown` half is held.
- Language, Nov 23 03:50Z, three conditions, each a Node fact: the bits record what the copy actually holds, never the source's type (`65-getter-view.ts`); `in` still answers prototype names (`65-prototype.ts`); present-but-undefined and overwrite keep Node's answers, and an absent spread adds nothing (`65-undefined.ts`, `65-typed-presence.ts`).
- Adamic, Nov 23 03:50Z: one ruling of the three; the unknown half held (`65-unknown-presence.ts` keeps the family's stop).
- Adamic, Nov 23 05:57Z: pin both refusals with their words: `65-overwrite.ts`, refused by stock tsc with TS2783 at 1:16, and `65-overwrite-view.ts`, refused at 2:26 under `adamic/single-spread`.
- Adamic, Nov 23 07:21Z: #xy74m5x is a must on main, a stop allowed; both forms of the running test run after the fix lands.
- Adamic, Nov 24 19:32:48Z, road (b): build the typed half now. The stop (#xy74m5x) is condition 3's support today, since no program gets a wrong answer while the overwrite is refused. Both forms of `{ ...source, k: undefined }` are pinned as that stop with its words: `65-leading-overwrite.ts` (the number form, "after a spread whose source holds k as") and `../spread_overwrite/optional_number.ts` (the optional form, "its type makes k optional"), beside `copied-empty-object.ts`. Language confirms condition 3 holds under the stop.

## The running test, for #7nhzq7h

Not run here, and no skipped test stands for it. It is #7nhzq7h's done-when for family 65: when a spread writes its own layout, `65-leading-overwrite.ts` and `../spread_overwrite/optional_number.ts` each lower and print `true`, `undefined`, `true`, `undefined` with exit 0 on Node, the JavaScript backend and native, and their stop pins in `typed_presence_test.go` turn into running pins.
