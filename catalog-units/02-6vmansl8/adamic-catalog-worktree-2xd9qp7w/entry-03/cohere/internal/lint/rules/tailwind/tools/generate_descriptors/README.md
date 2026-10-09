# generate_descriptors

The Phase 0 de-risk for the native Go Tailwind port. It answers one question before any Go is
written: **is a utility root's `{order, count}` reading a pure function of its value's inferred data
type?** If yes, `utilities.ts` (6,827 lines, most of it computing values `getPropertySort` never
reads) collapses to a table of a few hundred rows and Phase 3 is mechanical. If no, the shape of
Phase 3 changes.

Run it:

```
node internal/lint/rules/tailwind/tools/generate_descriptors/extract.mjs <theme.css> [--json <table.json>]
                                                [--resolve-root <dir>] [--corpus-root <dir>]
```

`<theme.css>`, the directory bare specifiers resolve from, and the tree scanned for a corpus are
three questions with one answer for a repository stylesheet, which is why they were one input.
`--resolve-root` borrows an install for a design system checked in outside any repository, the same
split `generate_collapse` took; `--corpus-root` names the tree to scan, because the derived one
is four `dirname` calls up from the entry point and will happily climb into an unrelated directory.
Both default to the old behaviour, so a repository invocation is unchanged.

Ground truth is `compileAstNodes`, an **internal** Tailwind API that gives `{order, count}` directly.
It is used because the question is diagnostic: `getClassOrder`, the public substitute, answers "did
these two classes agree" but not "what did this class declare". `getClassOrder` runs anyway, as the
end-to-end control, so a rename upstream is caught rather than silently changing the answers.

## The result

Measured on Tailwind 4.3.3 against `~/Projects/ahra` and `~/Projects/connected/www-connected-app`.
Both repos produce the same numbers and the **same exception set, class for class**, and that
agreement is weaker evidence than it reads as: both vendor the Structure submodule and import its
`global.css`, so the two stylesheets are byte-identical and the two systems contribute the same 35
utility roots and 325 theme keys. Two design systems that cannot differ cannot detect a table that
varies, which is the trap `a0f8635` documented for the invariance claims. The independent control
is `generate_descriptor_base/testdata/independent_theme.css`.

| population | measured | agreed | exceptions |
|---|---|---|---|
| registry (`getClassList()`) | 37,643 / 37,795 | 37,625 / 37,777 | 18 |
| — of which this repo's own | 35 roots, 325 theme keys | | |
| corpus (classes the repo writes) | 1,233 / 971 | 1,233 / 970 | 0 / 1 |
| modifier classes (sampled) | 40,110 / 40,325 | all | 0 |
| arbitrary sweep (327 roots x 527 shapes) | 132,203 | 132,199 | 4 |

**The model holds.** The exceptions are two named mechanisms, both understood, both countable.

## The descriptor, as measured

The claim as originally stated, `{typeList, readingPerType}`, is right in outline and incomplete in
four ways, each of which was found by a measurable population of mispredictions rather than by
reading source:

1. **The type list is ordered, and the order is per-root.** `inferDataType` returns the first match.
   `bg`'s list puts `position` before `length`, so `bg-[3px]` is a position; under declaration order
   it is a length and reads `[251]` instead of `[254]`. Fifty-eight probes turned on this one
   inversion. The order is recovered by topological sort from observation.
2. **A bare value is resolved, not inferred**, or rather both, in a fixed precedence: color
   keywords, then theme namespaces longest-first, then inference, then everything else. Getting the
   theme before inference is load-bearing: `bold` is a `--font-weight` key that also satisfies
   `family-name`, and inferring first predicts `font-bold` as a font family.
3. **A modifier is a third axis**, with three states rather than two. `text-[3px]/50` emits
   `line-height` and `shadow-[3px]/50` emits an alpha, so a modifier moves the reading; `/none` is a
   `--leading` key rather than an alpha, so it moves it differently. The modifier's value is
   otherwise not consulted: `/25`, `/[0.5]` and `/[var(--a)]` are one bucket.
4. **A type can earn its place on one axis and carry no information on another.** `color` does not
   discriminate `text-shadow` unmodified and is exactly what discriminates it under a modifier.

## The two real exceptions

**18 registry classes: per-declaration resolve-or-drop in repo `@utility` blocks.** All of the form
`<animation-root>-translate-full`, from `libraries/structure/source/theme/styles/animations.css`:

```css
@utility slide-in-from-top-* {
    --enter-translate-y: calc(--value(integer) * var(--spacing) * -1);
    --enter-translate-y: calc(--value(--percentage-*, --percentage-translate-*) * -100%);
    ...
}
```

`50` is both an integer and a `--percentage-*` key, so two declarations survive and it reads `[]#2`.
`4` is only an integer, `translate-full` only a `--percentage-translate-*` key: one declaration each,
`[]#1`. **Arity is the count of surviving declarations, and a value can satisfy several resolution
paths at once**, so it is not a function of any single data type. This is exactly the mechanism
`03-utilities-wall.md` section 4 predicted, confirmed at scale, and it is why the `@utility`
evaluator (task #4f04x54) has to model each declaration as independently resolvable. It is not a
defect in the descriptor model; it is the boundary of what the model covers.

**4 sweep probes: an upstream shadow quirk on `[16/9]` plus an arbitrary modifier.** `shadow-[16/9]`
reads `[315,316]#2` and `shadow-[16/9]/[var(--a)]` reads `#4` rather than `#3`. The shadow handler
splits the arbitrary value on `/` looking for a color slot, and on a ratio it mangles it and emits an
extra `@supports` block:

```css
--tw-shadow: 16var(--tw-shadow-color, /)9;
```

Garbage CSS from an upstream code path triggered by a `/` in the value, not by a data type. Four
probes, all `<shadow-root>-[16/9]/[<arbitrary>]`, none of which any registry contains.

## Controls, because a clean sweep is what a broken sweep looks like

- **Planted dirty descriptor.** One root's readings are shifted by one and replayed against real
  registry traffic. All 616 / 619 probes must fail, and `proven: true` reports that they did. A check
  that has never returned a positive has not been shown to be able to.
- **Volume assertions.** Registry above 20,000, roots above 200, sweep above 100,000 probes,
  property order above 300 entries, corpus above 100 classes, planted control proven. Below any of
  them the tool exits 4 and says the agreement numbers mean nothing.

- **Own-contribution assertion**, because every floor above is cleared by bare Tailwind. This
  README claimed a stylesheet importing nothing "reports 3,487 classes and exits non-zero". It does
  not: `@import "tailwindcss" source(none);` reports **23,286 registry classes, 301 functional
  roots and six figures of sweep probes**, and passed all six assertions. Tailwind registers its
  built-in roots from JavaScript rather than from CSS, so the population is real and the volumes
  are met while the `@utility` and `@theme` content that makes a design system per-repository is
  entirely absent. The same defect was found independently in the candidate parser (`a219c8f`).

  The check is the delta against a bare install built through the same borrowed engine: **35
  utility roots and 325 theme keys** are ahra's own, and a system contributing zero of both is
  refused by name. It is counted rather than named, because a hardcoded canary (`fade-in` on ahra,
  `brand-hover` on connected) is the same baked-in assumption it exists to catch. The floor is 1,
  since two roots is a real design system and a floor tuned to ahra's 35 would refuse one.

  | design system | registry | own roots | own keys | exit |
  |---|---|---|---|---|
  | ahra / connected | 37,641 | 35 | 325 | 0 |
  | `independent_theme.css` (borrowed install) | 23,391 | 2 | 4 | 0 |
  | `@import "tailwindcss" source(none);` | 23,286 | 0 | 0 | **4** |
  | real import that silently fails to resolve | 23,286 | 0 | 0 | **4** |

  The last row is the one that matters: `loadStylesheet` swallows a missing `@import` and returns
  empty content, so a broken graph produces a full-looking design system that met every volume
  floor. Registry size cannot separate these — the fixture passes 105 classes *below* the bare
  install that fails — so provenance is the only thing that does.
- **Nulls reported, never scored.** A class returning no reading is not an agreement. The registry
  has none by construction, which is why the corpus is scanned separately: that is the population the
  documented figure came from, and it holds 11 of 1,244 on ahra (`group`, `text-dark-4/70`,
  `shadow--3`) and 84 of 1,055 on connected.
- **Public-API cross-check.** Every pair of a 3,137-class sample, sorted by the predicted reading and
  by `getClassOrder`: 7 mismatches out of 4,856,848 pairs on ahra, 0 on connected, and all 7 are
  `fade-in-translate-full`, already a known exception. This control found its own bug first: an
  empty order sorts last, not first, and getting that backwards reported 365,174 mismatches while the
  model was fine.

## What Phase 3 consumes

`--json` writes `{roots: [...], statics: {...}}`, **465 KB** as raw JSON on both repos. That is the
uncompacted diagnostic shape, not the generated Go: it stores a full `{order, count}` per bucket per
root across all three modifier states, and most of those are duplicates of the root's fallback. The
Go table interns readings and drops buckets equal to the fallback.

The table is **not** checked in. It is per-repo by construction, since `readingByNamespace` is keyed on
this repo's `@theme`, and a checked-in table would bake one repo's tokens into a file claiming to
describe Tailwind, which is the bug that exists today in `internal/tailwind`'s generated tables.

## Notes for whoever picks this up

`inferDataType` is recovered from the installed bundle by probing chunks for a two-argument function
that types `url(a.png)` as image and refuses `var(...)`. The chunk filename is content-hashed. If a
future Tailwind reshapes it, the tool exits 3 rather than falling back to a hand-rolled predicate,
because a fallback would be testing the fallback.
