# Porting a rule

You are porting one ESLint rule to this tree. Read this document once before you
start, then work the numbered steps in order.

**Notice how it is weighted, because the weighting is the argument.** Step 6, write the
rule, is eight lines. The sections on proving your fixtures can see, on crossing the
config boundary, on running dry against real source, and on the gate are together about
a thousand. That is not neglect of the interesting part. Rules here are mostly
straightforward and the instruments that tell you whether one works are mostly not, so
almost every defect this document records was a correct-looking measurement rather than a
hard piece of logic. Read it and you should come away primed to distrust your own
instrument before you distrust the rule.

Whether that priming actually works is untested, and is the one claim here with no
measurement behind it. Agents who read this document have corrected their own findings
unprompted, but every one of them was also being checked closely and having their
corrections treated as ordinary, so the two explanations are not separated. The test that
separates them is an agent who reads this and is not watched. Until somebody runs it, read
the paragraph above as the intent of the document rather than as its demonstrated effect.

Two things to know before step 1.

**If the substrate is genuinely absent, the deliverable changes from a port to a
measurement.** Prove the absence with a compiling probe that has a control, say what
is missing and how big it is, and stop. A sized absence is a real result and a
half-built binder nobody can review is not.

**Size the rule before you accept it.** Three commands, and read the corpus and the
substrate rather than the rule file, which is the one number that is not the size.

    wc -l <rule>.ts                        the number that misleads
    wc -l tests/**/<rule>.test.ts          the corpus, which is usually the cost
    wc -l <utils-the-rule-imports>/*.ts    the substrate hiding behind a facade

Past roughly 600 rule lines over a 2,000-line corpus, an agent with 300K of headroom
returns a partial rule, and the dispatch was the error rather than the agent. Measured:
`no-unnecessary-condition` is 1,001 lines over 5,075 of corpus, and `prefer-optional-chain`
is 230 rule lines over a 1,807-line utils package and an 18,173-line corpus directory.
Say so and stop rather than delivering two-thirds of a rule.

**Cutting this document has a failure mode that cutting code does not.** Nothing goes red,
and you learn what you removed when the next person walks into the gap. It happened here: a
consolidation took this file from 3,902 lines to 2,554, the cut was right everywhere except
one section, and an agent fell into exactly that hole inside the hour. So when you remove
something, grep the file afterwards for the specific facts you believe survived rather than
trusting the shape of the result, and expect one restore.

The opposite failure is what caused the cut. Writing a finding in here had become the reward
for making one, so the same idea accumulated in twelve costumes and the procedure got harder
to read. **A finding that already has a section belongs as a sentence in that section**, not
as a thirteenth heading.

## Where things are

    cohere tree      /Users/kirkouimet/Projects/system/cohere
    rules            internal/lint/rules/<namespace>/
    shared shelf     internal/lint/ecmascript/ and internal/lint/checking/
    type graph       internal/types/program/
    write layer      internal/edit/
    vendored tsc     TypeScript/ (submodule) with shims in TypeScript-shim/
    upstream source  /tmp/lint-sources          (cloned from GitHub, see below)
    installed build  ~/Projects/ahra/node_modules    (the RUNNING rule, for ground truth)
    ahra CLI         only resolves from /Users/kirkouimet/Projects/ahra, so cd there for ahra commands

The namespaces under `internal/lint/rules/` are `base`, `core`, `next`, `nexus`,
`react`, `structure`, `tailwind`, `typescript`. Each is a Go package and each is
blank-imported by `internal/lint/registry/registry.go`.

### The upstream trees, and why both of these matter

`/tmp/lint-sources` holds the repositories that define the rules being ported.
They are shallow clones of the default branch; the React one is sparse. If a tree
is missing, reclone it:

    mkdir -p /tmp/lint-sources && cd /tmp/lint-sources
    gh repo clone eslint/eslint eslint -- --depth 1
    gh repo clone jsx-eslint/eslint-plugin-react eslint-plugin-react -- --depth 1
    gh repo clone typescript-eslint/typescript-eslint typescript-eslint -- --depth 1
    git clone --depth 1 --filter=blob:none --sparse https://github.com/facebook/react.git react
    cd react && git sparse-checkout set compiler/packages/babel-plugin-react-compiler

Where your rule lives in each:

    eslint/*              /tmp/lint-sources/eslint/lib/rules/<rule>.js
                          /tmp/lint-sources/eslint/tests/lib/rules/<rule>.js
    @typescript-eslint/*  /tmp/lint-sources/typescript-eslint/packages/eslint-plugin/src/rules/<rule>.ts
                          /tmp/lint-sources/typescript-eslint/packages/eslint-plugin/tests/rules/<rule>.test.ts
    react/*               /tmp/lint-sources/eslint-plugin-react/lib/rules/<rule>.js
                          /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/<rule>.js
    react-hooks/*         /tmp/lint-sources/react/compiler/packages/babel-plugin-react-compiler/src/
                          and the readable development bundle in node_modules, see "Know which kind of rule you have"

There is also a `/tmp/lint-sources/tse-fresh` checkout at 8.69.0, newer than the
main `typescript-eslint` clone. Check which one you are reading before you cite a
line number, and see "When the clone beats the installed build" for how to decide
between artifacts that disagree.

**A `react-hooks` rule name does not name its file, and the mapping is in one
place.** `src/CompilerError.ts` holds a switch from `ErrorCategory` to the rule's
public name, description and preset. Find your rule's name there, take the
`ErrorCategory` beside it, then grep that category through `src/Validation/` and
`src/Inference/` to find the pass that raises it. Measured examples:

    memoized-effect-dependencies    EffectDependencies            ValidateExhaustiveDependencies.ts
    exhaustive-effect-dependencies  EffectExhaustiveDependencies  ValidateExhaustiveDependencies.ts
    no-deriving-state-in-effects    EffectDerivationsOfState      ValidateNoDerivedComputationsInEffects.ts

Note the first two share a pass and the third has an `_exp` variant beside it.
**Check the `preset` field while you are there**: a rule marked
`LintRulePreset.Off` is off in upstream's own recommended config, which is
usually why it is not already enabled here and is worth a line in your report.

**A rule already in cohere may have been ported against oxc, and oxc is no longer
the authority.** This is not hypothetical: `react/no-string-refs` was already
here, ported from oxc, and the two authorities disagreed about **seven inputs**
with the old port wrong on all seven. Reading the sources was wrong about five of
them; only running the installed build settled it.

The residue is often a whole behaviour rather than a detail. That re-port found a
`.tsx`-only file gate that upstream does not have anywhere, which cost every
finding in every `.ts` file. **If your rule exists and its doc comment cites oxc,
treat it as unported.** Re-port it against the authority, measure every
divergence by driving the installed rule, and say in your commit that it was a
re-port and what changed.

**Check whether your rule is already in cohere before you start.** A rule can be
ported and enabled here and still appear on an audit of what ESLint enforces,
because the two configurations are separate surfaces. Twenty-six of seventy-eight
rules on one shortlist turned out to be in this state, and one porter was
dispatched to port a rule that had already shipped. One command settles it:

    <cohere binary> --rules | grep <your rule's last path segment>

If it is there, say so and stop: the task is wrong, not the rule. There may still
be real work (moving its registration to `<rule>_register.go`, or enabling it on
the ESLint side through `EnableRule.ts`), but that is a different task than
porting and should be reported as such.

**The clone is the source; node_modules is the oracle.** They are different
artifacts and you need both. The clone carries the original TypeScript or
JavaScript with comments, the option schema, the fixer, and above all the **test
corpus**, which the published packages do not ship at all. The installed build in
`node_modules` is the version this repository actually runs, and it is
executable: you can drive it with the ESLint Linter API on inputs you invent,
which is how you settle a question the source leaves ambiguous.

Where the clone and the installed build disagree, the installed build is what our
gate compares against, so measure on it and note the drift. The two documented
exceptions are below.

**Your scratch space is shared with every other agent.** One porter's mutation
script was overwritten mid-session by another agent's file and they nearly ran
somebody else's mutants against their own rule. Put anything you write into a
subdirectory named for your rule, never at the top of a shared temp path, and
name your throwaway packages for your rule too.

Work in this order. Do not skip ahead to writing the rule.

---

## 0b. Know which kind of rule you have before you start

Establish this in the first two minutes, because it decides what your fixture floor
is.

**The authority is whoever wrote the rule, and its source is in
`/tmp/lint-sources`.**

  | rules | authority | source |
  |---|---|---|
  | core `eslint/*` | ESLint | `/tmp/lint-sources/eslint` |
  | `@typescript-eslint/*` | typescript-eslint | `/tmp/lint-sources/typescript-eslint` |
  | `react/*` | jsx-eslint | `/tmp/lint-sources/eslint-plugin-react` |
  | `react-hooks/*`, `rules-of-hooks`, `exhaustive-deps` | React | `/tmp/lint-sources/react` |
  | `@next/next/*` | Vercel | not cloned; read the installed build |
  | `structure/*`, `nexus/*`, `base/*` | Kirk's and Kam's own TypeScript lint layers | in the ahra and api-phi-health trees, and those are the source |

**Every one of these ships a test corpus in its repository, and none of them ship it
in the published package.** That corpus is your fixture floor and it is richer than
you expect: `eqeqeq` alone carries 75 cases and 49 fix vectors. Read step 3 before
you write a fixture.

**Check where the judgment actually lives before trusting the rule file's size.** A
rule file can be a thin caller over machinery that holds the real decision:

- **ESLint's scope analysis.** A body calling `sourceCode.getScope()`,
  `context.sourceCode.getDeclaredVariables()`, or walking `variable.references` is
  delegating to `eslint-scope`. The rule file may be forty lines and the port large.
- **typescript-eslint's type services.** `getParserServices(context)` followed by
  `getTypeAtLocation` means the rule asks the type checker. We have that through the
  shim, but establish what the checker actually answers before assuming, because the
  two are not the same question.
- **React's compiler passes.** A `react-hooks/*` rule usually names a pass in
  `src/Validation/` or `src/Inference/`. The rule surface is trivial; the pass is
  the port.

**A rule that destructures an enum whose variants encode a classification is telling
you the classification happened somewhere else.** Go find where.

**Read the installed build too, and prefer running it over reading it.** The
published packages carry the compiled rule that this repository actually executes,
and the ESLint Linter API runs from `~/Projects/ahra`, so the real rule can be
driven on inputs you invent rather than reasoned about. A porter was told `08` and
`09` must not report and the dispatch was wrong: `/^0\d/u` matches them, and they
are the same banned legacy production. They found it by running the rule. Note that
ESLint defaults to module mode, where an octal is a fatal parse error and the rule
reports nothing, so ground truth needs `sourceType: "script"`.

**`eslint-plugin-react-hooks` ships two bundles and the development one is
readable.** `~/Projects/ahra/node_modules/eslint-plugin-react-hooks/cjs/eslint-plugin-react-hooks.development.js`
is compiled TypeScript with types stripped: real function names, real comments,
original variable names. Use it when you want to see what the shipped rule does; use
the clone when you want comments and tests.

**React's own pass documentation exists, and it is SECONDARY to the executable.**
`compiler/packages/babel-plugin-react-compiler/docs/passes/` has a numbered page per
pass with algorithm and edge cases, and it settled a rule's semantics in one read
for one porter. **It is also wrong.** `40-validateUseMemo.md` states that a callback
containing a bare `return;` "will trigger the error"; the lowering hardcodes
`returnVariant: 'Explicit'` regardless of argument and the case is **clean**,
measured against the running rule. A porter who found the docs early and trusted
them would ship a false positive with a confident citation. Read them for the shape
of the algorithm; take every verdict from the executable.

**`testing.RunTyped` trims your fixture, and a span assertion sliced from the Go
literal will be off by one.** `internal/lint/testing/program.go:169` writes each
fixture as `strings.TrimSpace(contents)+"\n"`. Every case copied from an upstream
tester carries a leading newline, so **the file on disk is one byte offset from the
literal in your test**. Slicing the literal to assert a span reports `"<Componen"`
for a finding correctly anchored on `"Component"`, which reads exactly like an
off-by-one in the rule. One porter chased it through the intermediate representation
and the shim before finding it in the harness.

**`testing.Run` does not trim**, so this is invisible to any rule that does not need
the checker.

## 1. Read the task

Read the whole task body before you touch anything. Parent bodies are the constraint
surface; standing rules live there.

## 1b. Read two or three rules in this tree that work the way yours will

Read them whole, including their `_test.go`. Pick by the shape of the question
rather than by the name:

    anchors on a declaration and looks for writes to it
        core/no_class_assign.go, core/no_const_assign.go, core/no_func_assign.go
    walks a regex pattern
        core/no_useless_backreference.go, core/no_regex_spaces.go
    needs the type checker
        grep -rl NeedsTypeChecker internal/lint/rules/
    proposes a fix or a suggestion
        grep -rl ReportNodeWithFixes internal/lint/rules/
    reasons about scope and bindings
        core/no_shadow_restricted_names.go, core/no_unassigned_vars.go
    gathers once over the file and then matches
        core/no_const_assign.go's KindSourceFile pre-pass

What you are taking is anchoring, `ctx` threading, type-switch layout, where
reasoning goes in the doc comment, how fixtures are tabled, how spans are asserted,
how `NeedsTypeChecker` is declared. Fourteen authors each re-invented this and no
guard caught the divergence, because every version passed its own tests.

### Read siblings for shape, never for semantics

Copy how a rule is built. Never copy what it believes about its inputs.

`no-nonoctal-decimal-escape` looks like a regex rule and is not: `/\8/` is a **pass**
in its corpus. Ported by analogy from the rule finished an hour earlier, its code and
its fixtures would have shared one wrong belief, **and every fixture would have
passed.**

When the code and the tests come from the same analogy they agree with each other and
disagree with reality. No fixture pair, no mutation sweep, and no span assertion can
see it, because none of them gets to vote on the premise. The corpus is the only
thing that can.

If you find yourself writing a fixture because a neighbouring rule had one like it,
stop and go find that case in your own corpus. If it is not there, it is not yours.

**Say in your report which rules you read and what you took from each.**

## 2. Read the reference implementation

The authority's rule file, from the clone. Read the whole rule body and every helper
it calls, not the first match.

Four things to establish explicitly before continuing:

- **`meta.schema` is the authoritative option surface.** Audit columns describing
  options have been wrong for sixteen of forty-one core rules. `schema: []` means no
  options and is checkable in one line.

- **`meta.messages`** names every distinct finding the rule can produce, which is how
  you tell one input reporting twice from two different judgments.

- **`meta.fixable`, and the fixer body.** If it says `"code"` or `"whitespace"`, the
  rule ships a repair and you are porting that too. The fixer is usually a
  `fix(fixer)` closure inside the report call. Read what it actually writes, because
  the corpus asserts the output text and your fixer has to produce it byte for byte.

- **Does it delegate?** If the body reaches for scope analysis, the type checker, or
  a compiler pass, then the rule file is a thin caller and the algorithm is
  elsewhere. A forty-line rule file can be a large port.

  **Needing the checker is not the same as the checker being sufficient.** Establish
  both what resolution answers and what it cannot. `react/jsx-no-undef` needs
  `GetSymbolAtLocation`, and resolution alone does not separate an intrinsic element
  from a typo: `<img />` and `<x-gif />` resolve to no symbol, identically to
  `<Uddefined />`. A port using resolution alone passes all eight of upstream's
  reported cases and then reports every `<div>` in the tree. The case predicate is
  load-bearing rather than an optimization, and only the corpus says so.

  **Two rules about searching, because both have cost real work.**

  *Search for the thing, not for the citation.* If somebody tells you a claim lives
  at commit X or file Y, search the whole tree for the thing itself. A search bounded
  by someone else's reference cannot find what their reference got wrong, which is
  the only case where checking was worth doing.

  *Run a control alongside any zero.* Grep for something you know is present in the
  same place with the same command. A zero from a real absence and a zero from a bad
  pattern, a wrong path, or a shim alias are the same zero. This costs one extra
  command and it is the difference between a measurement and a guess.

  **If you conclude our tree lacks the substrate it needs, prove it with a compiling
  probe before you stop.** A grep over `TypeScript-shim/` cannot see through a Go
  type alias, and the shim is full of them: `type Checker = checker.Checker` carries
  every exported upstream method without restating one. Three capabilities were
  called unavailable in one night on the evidence of a grep, and all three were
  reachable. Write a throwaway test that calls the thing, run it, delete it. Two
  minutes, and it cannot lie the way the grep did.

- **What does it deliberately not catch?** Commented-out cases and `TODO`s are
  upstream telling you what it knowingly misses. Reproduce those gaps; do not
  silently improve on them. A divergence is fine when it is stated.

**Record every divergence at the line, whichever way you resolve it.** A documented
disagreement is recoverable later; a silent choice is not.

## 2b. Fidelity is to what the rule DECIDES, not to how it obtains what it needs

The originals walk the filesystem, read `process.cwd()`, re-parse a neighbouring
file, or cache by path, **because ESLint hands them one file at a time and gives them
no program.** cohere has the program, one resident type graph, every file already
parsed.

**Reproducing a workaround for a constraint we do not have is not fidelity, it is
cargo cult.** If the reference implementation shells out to the filesystem to answer
a question the checker can answer directly, answer it directly and say so in the
commit. What you must reproduce exactly is the *decision*: which inputs report, which
do not, where the finding points, what it offers.

Three consequences, and the third is the one that ships.

**A port is not obliged to carry a defect it can see.** Two rules here report without
the fix their original ships, because those fixers drop type annotations, lose
`async`, or replace a whole `VariableDeclaration` while reporting per declarator,
which silently deletes a component. **Reporting without a fix is the subset you can
show correct.** Shipping a fix you cannot defend is worse than shipping no fix,
because a wrong fix is applied unattended.

**Where a rule has no tree exposure and upstream gives no reasoning, your own sense
of what the rule should do is the thing most likely to be wrong.** Record the
intuitive reading beside the actual one, at the line, so the next reader does not
helpfully correct it back to the wrong one.

**A surviving mutant in an arm you believe is unreachable is the arm telling you the
grammar is not the parser.** A porter emptied the getter and setter arms of
`default-param-last` and both mutants lived through all 87 corpus cases. The doc
comment beside them said those arms can never fire, because a setter takes exactly
one parameter — which is true of the *grammar* and false of the *parser*, which
recovers from illegal source and hands back a two-parameter setter. They probed the
parse instead of writing the survivor off, measured upstream reporting all four such
shapes, and added the fixtures. Two mutants caught, one doc comment corrected from an
argument into a measurement.

**And never document a limit you did not measure.** This is the third failure mode
and it is worse than the other two, because reproducing a defect leaves it findable
and silently improving leaves a difference the harness can see, while a confident doc
comment asserting "this case is a genuine divergence, reproduced as silence"
**inoculates the next reader against finding it.** They read the comment, believe the
narrowness was deliberate, and stop checking. It happened: a porter wrote exactly
that sentence about computed lifecycle keys, and upstream resolves them and reports.
Their sibling found it by running the installed rule. **If you state a divergence,
state the command that established it, or do not state it.**

## 3. Read the corpus, and take your fixtures from it

Every authority ships its tests in the repository and none of them ship it in the
published package. That corpus is your fixture floor.

    eslint/*              /tmp/lint-sources/eslint/tests/lib/rules/<rule>.js
    @typescript-eslint/*  /tmp/lint-sources/typescript-eslint/packages/eslint-plugin/tests/rules/<rule>.test.ts
    react/*               /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/<rule>.js
    react-hooks/*         alongside the pass in the compiler package, plus `__tests__` fixtures

**Read every case before writing code.** Each invalid case carries its **own**
`messageId` and, where the rule is fixable, its **own** `output`. So per-input
diagnostic counts are stated rather than recovered.

What to pull out, in order:

- **The valid cases.** These are the false positives upstream already thought about.
  Every one you skip is a class of false positive you will ship.

- **The invalid cases, with their `messageId`.** A case naming two entries in
  `errors:` reports twice, and a fixture asserting one finding per input is wrong for
  it.

- **The `options` on each case.** A rule with an option surface usually tests the same
  input under several settings, and the same code can be valid under one and invalid
  under another. `eqeqeq` does exactly this with `"smart"` and `"always"`.

- **The `output` on each invalid case, if the rule is fixable.** These are
  before-and-after pairs asserting what the repair writes rather than whether a
  finding appears, and **they are the highest-value artifact in a fixable rule's
  corpus.** `eqeqeq` carries 49 of them against 75 cases. Assert them with
  `ExpectFixedSource`, not by eye. An `output: null` means the case is reported but
  deliberately not fixed, which is itself a decision to reproduce.

- **Anything commented out.** Upstream telling you what it knowingly misses.

**A corpus file may call `RuleTester.run` more than once, and a single-slot interceptor
silently keeps the last one.** The reliable way to extract a corpus is to stub or
intercept `RuleTester` and run upstream's own test file, so its own cases produce your
fixtures rather than a parser of yours. That works, and it has one failure mode: an
interceptor written as `capturedCases = tests` rather than `capturedCases.push(...)`
keeps whichever call came last and drops the rest, with no error and a plausible-looking
count.

Measured: `no-invalid-this` makes two `run` calls, a 474-case JavaScript matrix and a
98-case TypeScript block, and an earlier single-slot interceptor kept only the second and
**lost 83% of the corpus**. `init-declarations` also makes two. Neither file looks unusual
from the outside, and neither count looks wrong once you have it.

So accumulate across calls, and **print the call count beside the case count** -- a
one-call extraction from a two-call file is the whole defect, and it is invisible in the
cases themselves:

    console.error(`captured ${cases.length} cases from ${runCalls} run() call(s)`);

Then refuse to emit anything when either number is zero, for the same reason every other
instrument here refuses: an extraction that captured nothing looks exactly like a rule
with a small corpus.

**Take the cases verbatim.** Retyping a case is how a fixture ends up testing what
you believe instead of what upstream asserts.

**Measure parenthesis behavior rather than assuming it.** `SkipParentheses` reads as
a free correctness improvement, so a porter adds it and moves on. Upstream applies it
inconsistently within a single rule, and our AST does not agree with any upstream's
about where a parenthesis survives: measured on `no-direct-mutation-state`,
`(this.state.x) = 1` reports with the span excluding the parens, `(this.state.x)++`
reports with the span including them, and `(this).state.x = 1` is silent. Two
accessors, three answers, one rule.

Most corpora write no parenthesized forms at all, so guessing wrong costs nothing at
fixture time and ships a divergence either way. **Look before you assume your corpus
is silent on parens.** If your rule matches a receiver or a callee, write `(x)` forms
into your own fixtures and run the installed rule to see which way each one falls.

**When upstream gates on source type, check whether OUR parser can produce the node in
the excluded files.** This is fidelity to a **parser difference** and it is
load-bearing rather than vestigial: the vendored compiler resolves JSDoc into real
type nodes, so `/** @type {any} */ let x = 1;` in a `.js` file produces a live
`KindAnyKeyword` that other parsers cannot produce at all. Dropping the gate as an
optimization would have reported every JSDoc `any` in every JavaScript file — **a
false-positive class no imported fixture can see.**

**The clones ship no `node_modules`, and each rule needs a different one.** Any
upstream tester or rule file that imports `ast-utils`, `eslint-visitor-keys`,
`esutils`, and friends fails to load with a *different* missing module each time, so
installing them one at a time costs a cycle per rule. Resolve them out of
`~/Projects/ahra/node_modules/.pnpm` by package name and they all land at once.

**Render `options` and `languageOptions` beside every extracted case.** A case can
carry its verdict above the rule: `no-alert` has byte-identical source in both the
valid and the invalid list, separated only by `ecmaVersion`. Read as prose that looks
like a contradiction or a duplicate in the corpus. A case appearing twice with
opposite verdicts is a version gate, not an error — and the extractor is the only
place you can see it.

## 3b. If the rule is fixable, the fixer is half the port

`meta.fixable` in the rule file tells you. If it is set, you are porting a repair as
well as a judgment, and the corpus already asserts what that repair must write.

**Upstream's fixer is a closure; ours is a value.** ESLint hands the rule a `fixer`
object and the rule calls `fixer.replaceText(node, "===")`. cohere builds the same
thing as data:

    rule.ReplaceRange(textRange, "===")
    rule.RemoveRange(textRange)

    ctx.ReportNodeWithFixes(node, message, rule.ReplaceRange(operatorRange, "==="))
    ctx.ReportNodeWithSuggestions(node, message, suggestion)

The distinction between the two report calls is not cosmetic. **A Fix is applied
unattended; a Suggestion needs a human to choose it.** Upstream draws the same line
with `fix` versus `suggest`, and it is part of what you are porting: shipping a
suggestion as a fix means the edit engine rewrites code upstream would only have
offered to rewrite. Read which one the rule uses before writing either. The engine
that applies them is `internal/edit/`.

**The `output` field in the corpus is the specification.** Assert with
`testing.ExpectFixedSource`, which compares the whole rewritten file rather than a
message id, so it catches a fixer that repairs the right span with the wrong text.

    testing.ExpectFixedSource(t, result, "a === b")

**And they are not extra coverage for a fixer. They are the ONLY coverage it has.**
This is the asymmetry to hold onto, because it is what makes skipping them feel safe:
a fixer that reports in the right place and repairs wrongly is *indistinguishable*, at
the message-id layer, from a correct one. Every id fixture stays green. The span
assertions stay green. The rule looks finished.

Measured, on `dot-notation`, which is worth stating concretely because the abstract
version of this argument does not survive contact with a deadline. The port passed all
69 of upstream's corpus verdicts, then failed **11 of 38** `output` fixtures on the
first run. Two defects, neither reachable by any id:

    5_000['prop']   ->  5_000.prop      wanted 5_000 .prop
    foo['bar']instanceof baz  ->  foo.barinstanceof baz

The first is a transcription failure: `DECIMAL_INTEGER_PATTERN` was written from what
the rule *looked like* it needed rather than read from `ast-utils.js:64`, and the real
one accepts numeric separators and a leading zero before an 8 or 9.

**The second is the one to remember.** It is not a wrong fix, it is a fix that produces
valid-looking source meaning something else. `foo.barinstanceof` is a plausible
identifier; a reviewer reading that diff sees nothing wrong, the file compiles, and the
behaviour changed. The missing piece was upstream's third yield, which inserts a space
when the repaired text would abut the next token.

So: if `meta.fixable` is set, the `output` cases are not a nice-to-have you add if there
is time. Porting the judgment without them ships a repair that runs unattended over the
whole tree with nothing able to say it is wrong.

**The decision to port a fixer at all turns on whether the repair is SPECIFIED, not on
whether fixers are frightening.** Two rules in one batch went opposite ways for the same
reason. `react/no-invalid-html-attribute`'s suggestions were declined because one of them
does a substring replacement on SOURCE text -- `value.raw.replace(value.value, '')` --
which repairs the wrong span whenever a bad token contains a good one as a prefix, and
nothing upstream pins that. `dot-notation`'s fixer was ported because 34 `output` cases
state exactly what it must write and `ExpectFixedSource` compares the whole file. Ask
which of those two situations you are in.

**`output: null` is a decision, not an omission.** It means upstream reports the case
and deliberately declines to fix it, usually because the repair would change
behaviour rather than spelling. Reproduce the decline. A fixer that repairs a case
upstream refuses to touch is a defect that no message-id fixture can see.

**A fix that is right and anchored wrong is worse than no fix.**
`ReportNodeWithFixes` anchors the finding on the node's own text, and its doc comment
says why this matters twice over: a finding reported at the wrong line while carrying
a fix means the edit lands somewhere the reader was never shown. Assert the span on
every fixable rule, not just the message.

**Read a fixable rule already in the tree before writing yours:**

    grep -rl ReportNodeWithFixes internal/lint/rules/

`internal/lint/rules/core/prefer_const.go` and
`internal/lint/rules/core/no_unused_labels.go` are both worth reading whole. They show
where the range comes from, how a multi-edit repair is assembled, and how the test
asserts the output.

**And when we decline a fixer upstream ships, say so at the line.** The tailwind rules
did this deliberately: a class literal in this codebase wraps across lines with
indentation that carries intent, and a fixer would be the first thing to reflow it.
That is a defensible decline and it is recorded where the next reader will look. An
undocumented missing fixer just reads as an unfinished port.

**Two harness facts that will look like defects in your rule.**

`/tmp` is a symlink to `/private/tmp` on macOS, so any comparison against a
`from: file` specifier path silently fails to match. Same hazard class as the
case-insensitivity warning and a different channel, and it presents as your rule
ignoring a configured allowance.

The test harness's tsconfig pins `moduleDetection: "force"`, which makes an ambient
`declare module` unresolvable. A porter's one upstream case using an `errors` package
could not be expressed because of this. They recorded it in its own test with two
controls rather than weakening the rule to make it green, which is the right shape: a
case you cannot express is a fact about the harness, and hiding it inside a relaxed
rule turns it into a fact about the rule.

## 3c. If the message interpolates anything, the ids cannot see it either

**This is 3b's argument arriving through a different door, and the pair is what makes it a
property of fixtures rather than a quirk of fixers.** There, message ids cannot see a
repair, so `output` is a fixer's only coverage. Here, a single message id cannot see a
RENDERING, so a name-and-span test is the only coverage a rule with an interpolated
message has. In both cases the fixtures are complete and the assertion layer is the wrong
one.

The trap is sharpest on a rule with exactly ONE message id, because then a fixture table
records only a count. `complexity` is the measured case: 165 imported fixtures, all green,
with three defects live and two more the corpus could not reach at all.

    an arrow assigned to a class field    rendered "Arrow function", upstream says "Method 'x'"
    an arrow in an object property        rendered "Arrow function", upstream says "Method 'b'"
    the span for a property's function     started at `function `, upstream starts at `c: function `
    a computed property key                counted inside the field's path, upstream counts it
                                           in the ENCLOSING function's
    a private method's name                rendered "Private method '#p'", upstream renders it
                                           UNQUOTED as "Private method #p"

None of the five moves a message id and none moves a count. The first three were found by
writing a table of expected NAME and SPAN per shape; the last two by mutation.

**Read `meta.messages` for placeholders, and treat each one as an assertion you owe.** A
message with `{{name}}` in it is a computed string, and the computation is usually a
helper large enough to have its own bugs: `complexity` interpolates a name built by
upstream's `getFunctionNameWithKind`, 77 lines assembling `static`, `private`, `async`,
`generator`, a kind word and a quoted name, plus a span from `getFunctionHeadLoc`. Its
corpus exercises **14 distinct renderings across 103 findings**, so the name is most of
what the rule computes and none of what its ids record.

**Test the builder DIRECTLY, not only through the rule's messages.** An end-to-end row
that fails tells you the rendering is wrong without telling you which of twenty branches
produced it, and a table keyed on the branch is what turns a red line into a location.
Doing that on `complexity` immediately found a fourth defect the end-to-end table could
not: `{ d: function named() {} }` renders `Method 'd'`, because upstream tries the
PROPERTY name first and falls back to the function's own `id` only when the property name
is not static. No corpus case writes a named function expression in a property position,
so nothing else could have caught it. The same table then exposed dead code, two accessor
arms in a branch only an arrow or a function expression can reach.

**And the parser difference cuts the other way here, so port the DECISION rather than the
dispatch.** ESTree wraps a method in a `MethodDefinition` around a `FunctionExpression`,
so upstream reads modifiers off the parent and the kind word off `parent.kind`. Measured,
our parser has no wrapper at all: `KindMethodDeclaration` is both the member and the
function. A reader porting those 77 lines faithfully will write parent-dispatch branches
this tree can never reach, and a mutation sweep will report them as survivors because they
are genuinely dead. The reachable version is smaller, and the arms that survive gutting
are the ones to delete rather than to document.

## 4. Check the shelf before writing any helper

The shared shelf is `internal/lint/ecmascript/` and `internal/lint/checking/` —
roughly a hundred non-test files across twenty packages. `ecmascript/` is subdivided
by question:

    binding      classmembers  comments   control_flow_graph   decorators
    high_level_intermediate_representation             imports  jsx
    literal      module        nextjs     property             react
    reference    regexp        regexpattern             regexsyntax
    scope        text

`checking/` holds the type-checker-facing helpers: `builtins.go`, `checker.go`,
`nodes.go`, `set.go`, `specifier.go`, `types.go`.

**Before writing a value-comparison helper, print the node's own fields.** Our parser
has often already done the work. A porter nearly wrote a numeric parser before probing
a `NumericLiteral`: its `.Text` already holds the canonical rendering of the double —
`0x10`→`16`, `1_0`→`10`, `1.0`→`1`, `1e21`→`1e+21`, `1e400`→`Infinity` — a bijection
onto the double upstream compares, and a `StringLiteral`'s `.Text` is the cooked
value. **The node is a shelf nobody thinks to search.**

**`comments.ForFile` is invisible to a search for AST helpers.** A rule whose subject
is trivia has no node kind to grep for, so the shelf's comment reader
(`internal/lint/ecmascript/comments/cache.go:19`) will not surface that way. Two
porters have now nearly reimplemented `scanner.GetLeadingCommentRanges` beside it.

**Search everything reachable, not only the shelf.** The `TypeScript-shim/` packages
expose the vendored compiler's own helpers by linkname and they are invisible to a
search of the shelf. A porter needing an edit distance nearly used
`core.GetSpellingSuggestionForStrings` — closest name on any shelf, reachable, and
**wrong**: it charges 0.1 for a case-only substitution and 2 for anything else, then
accepts within 0.4 of the name length. Measured, it returns suggestions for three of
upstream's own *passing* cases. Read the body of anything whose name matches your
question, wherever it lives.

Search the shelf, and search the rule package you are writing in, before adding a
function.

A name collision at build time means the decision already has a home; use the existing
one and read it first, because it is usually better than what you were about to write.

**Probe a shelf helper against UPSTREAM, not against its own doc comment.** The
failure mode is not "weaker than documented" — it is "different from upstream", and it
runs in both directions, **including the direction where the helper is MORE careful.**

`react.IsEs6ComponentClass` skips parentheses on the heritage receiver and its doc
advertises that skip as a correctness *feature*. It is a divergence:
`extends (React.Component)` is **silent** upstream and the helper answers true. That
is invisible from the doc comment, which is accurate about what the code does, and
invisible from the corpus, which writes no parens. A helper doing something sensible
that upstream does not do is still a defect in a port.

Both helpers a porter reached for on `no-this-in-sfc` were wrong opposite ways:
`react.IsLikelyComponentName` uses `unicode.IsUpper` where upstream tests ASCII only,
so `Фoo` and `Éoo` are silent upstream and true on the shelf; `react.IsEs5ComponentCall`
accepts `createClass` where upstream keys on `createReactClass` alone. Both live in
`internal/lint/ecmascript/react/`.

**A shelf function's doc comment is a claim, not a guarantee.** One agent found
`RegexCharElement.IsLoneSurrogate` documented as the astral-expansion mechanism with
**zero assignment sites**, validated against a control field that had three. Another
measured that `regexpattern.Walk` skips set escapes and does not recurse into
`v`-flag nested classes, contradicting what its name implies. **If a shelf function is
load-bearing for your rule, probe it on your own corpus before building on it**, and
report what you found either way. The shelf is younger than the rules it serves.
`TestShimFieldAccessorsReadTheFieldsTheyName` and the wrapped-accessor guard both
police this boundary from the other side.

**And the harder case: a shelf helper that is entirely CORRECT and answers a
different question than the one you are asking.** Every example above is a helper
that is wrong about something. This one is not, and that makes it worse: there is
nothing to find by reading it, because the code and its doc comment agree with each
other and both are accurate.

Measured on `@typescript-eslint/consistent-return`, whose void filter asks "does this
return type include `void`". The shelf has `checking.IsTypeFlagSet`, whose name is
exactly that question. It reads the type's OWN flags. typescript-eslint's
`isTypeFlagSet` first decomposes a union and ORs its constituents' flags, which is a
different question with the same name:

    void                     shelf T   upstream T   agree
    void | number            shelf F   upstream T   DISAGREE
    Promise<void | number>   shelf F   upstream T   DISAGREE   (on the awaited type)

Both disagreements are upstream PASSING cases — `function foo(flag?: boolean): number
| void` and a `Promise<void | number>` alias — so the shelf helper reports two clean
inputs. **Neither is a fixture you would think to add**, because both come from
upstream's valid list and a port failing two of nineteen valid cases reads as two
unlucky fixtures rather than as a wrong helper.

**The trap is reaching for it BY NAME.** The standard already tells you to search the
shelf before writing a helper, and that instruction is what delivers you to the wrong
function: you search for your question, you find a name that is your question, and
the match itself is the evidence you stop on. A name cannot encode which of several
readings of a question a function took, and the more natural the name the more
readings it covers.

So the rule that survives is narrower than "probe a shelf helper against upstream". It
is: **when a shelf helper's name IS your question, that is when to go read upstream's
implementation of the same-named helper and diff the two**, because a name collision
between the shelf and upstream is the one case where the collision is doing your
thinking for you.

**Write the divergent one local rather than fixing the shelf.** Correcting
`checking.IsTypeFlagSet` to decompose unions would silently change the answer for
every other caller, none of which asked for it, and the shelf helper is not wrong —
it is the other reading, which some of those callers presumably want. A local
`consistentReturnUnionFlagSet` with the measurement at the line costs one function and
moves nobody. Adding a second helper to the shelf beside the first is the option to
avoid: two shelf functions whose names differ by a word are how two callers end up
asking different questions believing they asked the same one.

> **On the shelf census.** Earlier revisions pointed at a `SHELF-CENSUS.md`, a
> whole-corpus read from 2026-08-23. It is not in this repository, and its "missing
> utilities" section was already stale in every entry checked when it was last
> consulted: `WritesToBinding` lives in `ecmascript/reference/write.go:100`,
> `EnclosingFunctionLike` in `ecmascript/scope/enclosing.go:45`, and the
> static-property-name decision it called missing shipped as `property.Name` /
> `property.NameTagged` (`ecmascript/property/name.go:82` and `:138`) rather than
> under the name it proposed. **Search the shelf by the question rather than by any
> census's proposed name**, because the name is the part that drifts.

## Our tree and upstream's tree disagree in both directions

Two facts about the parser, and they produce opposite defects. Most porters learn one.

**Our tree keeps nodes upstream's parser folds away.** Parentheses, computed names, chain
expressions. So any port reasoning about a node's immediate parent or its exact span has to
unwrap first: `({ a: (function(){}) })` reports upstream and needs the parenthesis unwrapped
here, and the fix range has to cover it or the repair leaves a stray `)`.

**Unwrap on every side of the question, not only on the node you are inspecting.** This has
now cost findings in four rules across three batches, and the miss is rarely the obvious
argument: `a || (b && c)` never reaches the `&&` in a mixed-expression walk, and
`Boolean(((a = b), b || c))` needs unwrapping on both sides of the comparison rather than on
the argument alone. Most corpora write few parenthesized forms, so guessing wrong costs
nothing until the real tree. When a walk misses a shape you expected it to reach, probe the
parent chain and read what is actually there before adjusting the walk again.

**Upstream merges shapes our tree keeps separate.** In ESTree a property, a method, a getter
and a setter are one `Property` node distinguished by a `kind` field, so upstream must guard
against accessors reaching its shorthand tests. Here each has its own syntax kind and matches
no arm, so it falls through on its own and **upstream's guard has nothing to guard.** A
mutation sweep reports such a guard as a survivor, correctly. Delete it and record why: a
documented unreachable branch is a justification, dead code is not.

**Node positions include trivia and upstream's `range` does not.** `Pos()` begins at leading
trivia and a list's `End()` sits past a trailing comma that the last element's own `End()`
stops short of. Every fixer built on raw `Pos()`/`End()` is off by whatever trivia is there,
and this has produced defects in four separate rules: a replacement span eating the space in
`var foo = 'a' + b`, a paren landing before the whitespace, a comment between two properties
reading as a comment inside the second, and a token-start scan that skipped whitespace but not
comments and so swallowed a comment's first byte.

**An `As*()` accessor is an interface conversion, not a cast.** `parent.AsCallExpression()` on
a `NewExpression` panics rather than returning nil, and a shared `case KindA, KindB:` arm is
exactly where you reach for one accessor for two kinds. This shipped once and crashed 71 real
files while all 111 of the rule's fixtures passed, because upstream's corpus has no
`new Foo(function(){})` case. Give each kind its own arm.

The walk recovers per file and takes every applicable rule down with it, so one rule's panic
costs all of them their verdict on that file, and the run still prints an ordinary summary.
The shared crash guard does not save you either: it walks every rule through the untyped
harness, so a rule declaring `NeedsTypeChecker` has its typed paths unexercised. **Your own dry
run is the only instrument that sees this class**, and only if you grep the crash line before
believing a clean run.

## 5. Write the fixtures first, from the corpus, verbatim

Before the rule. Every pass and fail case goes in **as written upstream**. This is
mandatory.

**Build the oracle as an artifact, not as working memory.** Whatever you write to extract
and replay upstream's corpus, put it somewhere durable and keep it. **The oracle outlives the
rule**: the rule is one dispatch's output, and the oracle is what lets the next person start
from your measurements rather than from scratch. It costs nothing at the time. One agent's
`one-var` oracle, controlled at 296 verdicts and 144 outputs, is what let that rule move to a
fresh agent at zero cost when it turned out to be oversized for the session that sized it, and
it was built durably because this document warns about shared scratch rather than because
anyone foresaw a handoff.

**Verify the copied strings mechanically, byte against byte.** Reading them is not
enough and this has cost three porters. One found four transcription errors its own
reading had passed over, the worst being a pass case whose `\u0020\u0020` had been
cooked into real spaces, so it would have sat in the clean list asserting the exact
opposite of upstream. Another had two fixtures pass for the wrong reason because a
heredoc collapsed `\\1` into `\1`. A third watched a mutant come back alive because
the tool writing its fixture cooked `\u0041` into `A`.

**The tool you use to write the fixture can change it, and the result still compiles
and still goes green.** Compare the bytes on disk against the bytes upstream, with a
script, not with your eyes.

**This reaches the RULE file too, not only the fixtures.** A porter wrote `String('')`
with two ASCII apostrophes into a Go doc comment and the file came back holding a
single U+201D. It compiles, `gofmt` passes, the tests stay green, and they only
noticed because the harness reported the file changed on disk.

**A second porter traced it to the editing tool.** `sed -i ''` on macOS converted a
pair of ASCII apostrophes in their Go source into a single smart quote, and the file
still compiled. Two porters, two different rules, same damage. Scan the whole file for
non-ASCII rather than grepping four characters, and run a control so a zero means
something:

    LC_ALL=C grep -n '[^\x00-\x7F]' <your file>     # should be empty
    LC_ALL=C grep -c '[^\x00-\x7F]' <a file you know has none>   # control

**`grep -c` returning 0 exits 1**, which breaks the `&&` chain this scan is naturally
written with — it reads as a failure when it is the success case. Use
`LC_ALL=C grep -c … ; echo` rather than `&&`.

The reason to copy at all: a fixture you invent encodes the same belief as the port
you are about to write, so it passes for exactly the reason the code is wrong.
Upstream's clean cases are the ones that catch you, because each was added when
somebody hit that bug.

Split them into a Fires test and a StaysSilent test. **Read
`internal/lint/rules/core/no_iterator_test.go` for the shape, not the rule file** —
the rule file will not tell you the assertion API and you will invent one that does
not exist.

The API, so you do not have to guess. The package is
`internal/lint/testing` (import it and call it `testing` or alias it; it is not named
`rule_testing` any more):

    result := testing.Run(t, YourRule, fileName, sourceText)                    // untyped
    result := testing.RunTyped(t, YourRule, fileName, sourceText)               // rules that declare NeedsTypeChecker
    result := testing.RunWithOptions(t, YourRule, fileName, sourceText, opts)
    result := testing.RunTypedWithOptions(t, YourRule, fileName, sourceText, opts)
    result := testing.RunTypedFiles(t, YourRule, files, subjectFileName)        // cross-file resolution
    testing.ExpectFindings(t, result, "yourMessageId")                          // one id per expected finding
    testing.ExpectClean(t, result)                                              // no findings; there is no ExpectNoFindings
    testing.ExpectFixedSource(t, result, wantSource)                            // applies the fixes, compares the result

`Run`, `RunWithOptions`, `ExpectFindings`, `ExpectClean` and `ExpectFixedSource` live
in `rule_testing.go`; every `RunTyped*` variant lives in `program.go` beside the
program cache.

The rule struct is `rule.Rule{Name, Run, NeedsTypeChecker, ReadsProgram,
ResolvesReactValueTypes}`. A porter once wrote a `Messages:` field from the shape of
other linters; it does not exist. Messages are `rule.Message{Id, Description}` values
passed at the report site.

**`Decode` is not on `Rule`.** It lives on `rule.Registration`, alongside
`RequiresOptions`, because `Rule` is what a rule author writes and a decoder is the
config layer's business:

    rule.Register(rule.Registration{
        Rule:   Eqeqeq,
        Decode: rule.DecodeOptionsInto[EqeqeqOptions](),
    })

If your rule takes no options, omit `Decode` entirely — that is the common case.

`Run` is `func(ctx rule.Context, options any) rule.Listeners`. **Every rule takes the
options parameter, including one with no option surface.**

**There is no `rule.OnExit` and the walk is pre-order.** Two porters reached for one
and neither found it. If your rule has to gather before it can judge, do both inside a
`KindSourceFile` listener: it fires before its children, so you can walk the tree
yourself, collect, and report at the end.
`internal/lint/rules/react/no_multi_comp.go` is the shipped example, and `internal/lint/rules/core/no_class_assign.go` uses the same shape for a
declaration anchor.

`ExpectFindings` takes one id per finding, so an input reporting twice passes two ids.

Then add cases upstream does not cover, from reading **our** code, and say in a comment
why each exists.

**Never type an escape sequence on the way to a fixture.** Byte-verification catches a
cooked escape, but the reliable fix is upstream of it: build every source string from a
list of physical lines and emit it through `json.dumps`, so no `\n` is ever typed and
nothing can cook it. A porter's `\n` sequences became real newlines through a
Python-heredoc-through-Bash path, and their piecemeal repair then silently damaged a
second case because one placeholder was a **prefix of another**. Their verification
script also asserted *before* writing, which is why the failed run left the file
untouched. Both are the pattern.

**A passing case in the imported corpus is not evidence about the thing you think it
is.** The "read siblings for shape, never for semantics" hazard also arrives through
the **imported** corpus, where it is far harder to see because the case is genuine
upstream data.

Measured on `no-this-in-sfc`: upstream ships a passing `React.createClass` case that
reads as direct evidence the namespaced factory exempts. It is not. It passes because
its `render: function () {...}` is **anonymous**, so no name can be read, and removing
the factory entirely leaves it passing. A porter treating that fixture as evidence
about factory names would have concluded the exact opposite of the truth.

**Before you let a passing case teach you a rule, change the one thing you think makes
it pass and check the verdict moves.** A case that stays green under that edit was
never testing what its shape suggests.

**A SEMANTIC ERROR decides cases above the rule too, and it looks nothing like a
suppression comment.** A porter's most misleading passing case was
`export * from '_'; export = {};` — it reads as evidence that an export assignment
suppresses the rule beside a star export. It is not evidence about the rule at all:
that input is TypeScript error 2309, and the run stops before the rule fires. **The
tell is that the OTHER diagnostics vanish too**, which you only see by reading the full
output rather than grepping for your own rule name.

**A third fixture category exists: a case whose verdict is decided ABOVE the rule.**
`testing.Run` never consults `internal/lint/suppression`, so an upstream case that is
silent only because of an `eslint-disable-next-line` comment cannot be reproduced as
silent by a rule test. A porter following step 5 literally will either break their port
to green it or quietly delete the case, and both are wrong.

**Pin it at the layer that actually decides it.** Record the case as *reporting*, with
the reasoning at the line, and say in your report that the suppression layer is what
makes it clean upstream. The porter who hit this confirmed the mechanism by deleting
the disable comment and watching the same file report.

**Driving the installed rule does not dissolve this, and there the case arrives as a
DISAGREEMENT rather than as an annotation you can read.** A porter who replayed
upstream's whole corpus through the Linter API — the method this document recommends,
because it answers what the rule does HERE rather than what the corpus file claims — got
262 of 263 rows agreeing and one row where the port reported and the oracle did not. The
natural reading of a single disagreement against a measured oracle is that the port is
wrong. It was not: the row's source carried `// eslint-disable-line`, so what the oracle
had recorded was the suppression layer removing a finding the rule had already produced.

So the Linter API is an oracle for the rule PLUS everything ESLint layers above it, and a
rule fixture wants only the first. **Grep your corpus for `eslint-disable` before
trusting any row of it**, whether the verdicts came from annotations or from a driven
run — one line, and it tells you exactly which rows are about the wrong layer:

    grep -c "eslint-disable" tests/lib/rules/<rule>.js

Confirm each hit the same way: re-run that source through the oracle with the comment
deleted. If it reports, the rule's own verdict is *reporting*, and that is what the
fixture asserts.

## 6. Write the rule until the fixtures pass

Doc comment carrying valid and invalid examples, and a message whose Description says
why the code is wrong rather than restating the rule name.

When a fixture fails, read it before changing the rule. Upstream's clean cases encode
distinctions you have not thought of yet, and the failure is usually telling you one.

## 7. Prove the fixtures can see: the mutation sweep

    internal/lint/rules/tools/score_one_mutation.sh <file> <package> "<python expression rewriting the string `source`>" [TestName]

**Mutate every discrimination the rule makes**, one at a time. **A green suite proves
nothing until something has been shown able to turn it red.**

**A survivor tells you your fixtures cannot see something. WHERE that something lives
decides which instrument can find it.** The distinction is worth holding because it tells
you which tool to reach for, and the two halves fail differently:

    the invisible thing is in YOUR CODE      a mutation sweep reaches it
        a quantifier, a guard, a branch      the mutant compiles and the corpus stays green

    the invisible thing is in THE WORLD      only a dry run or a differential reaches it
        a shape upstream never wrote,        no mutation of your rule can conjure the input
        a substrate that behaves otherwise

Measured on `@typescript-eslint/no-unsafe-enum-comparison`. Its `isNumberLike` is EVERY
union constituent over SOME intersection constituent, and flipping the union quantifier to
`some` leaves **all 85 corpus cases green**, because upstream writes no union of mixed
primitive kind. That defect lives in the rule's own logic, so a sweep found it; every other
blind-corpus instance in this document needed the real tree, because those defects were
beliefs about the substrate rather than errors in a branch.

The repair is the same either way and it is not "add a fixture": **build the input that
separates the two readings, confirm the shipped rule and the mutant disagree on it, and
add it with a control.** Confirming both halves is what distinguishes a fixture that kills
the mutant from one that merely happens to pass.

**The fork above is the easy case. The common one is a sequence, and it needs both.** Read
as a fork, this section tells you to pick a tool; the defect that most wants finding is the
one where the first tool changes the code and the second discovers the change is untested.

Measured on `@typescript-eslint/prefer-regexp-exec`. Three steps, three instruments:

    1. dry run        the `const` guard over-reported nothing but was SILENT on
                      `let r = /x/; a.match(r)`, which upstream reports. A belief
                      about the world, and only real files showed it.

    2. sweep          the replacement guard was written, and a mutation deleting it
                      SURVIVED all 37 corpus rows. Upstream never binds a regex to a
                      `let`, so no imported case can judge the thing just written.

    3. oracle         upstream's own behaviour settled what the guard should say, and
                      the test written from it is the only thing that kills the mutant.

Neither instrument alone arrives. The dry run finds the belief wrong and hands you a
correction it cannot check; the sweep finds the correction untested and cannot tell you
what it should have been. **A fix made in response to a dry run is code no fixture has ever
judged**, because the fixtures are exactly what failed to raise it, so the sweep is not
optional afterwards. Run it again after every dry-run repair, on the guard you just touched.

**The script takes an expression, not statements.** `import re; source = re.sub(...)`
fails with a bare `SyntaxError` that names nothing and reads like a broken script. Keep
the rewrite to a single expression. Separately, gofmt's alignment silently no-ops a
rewrite on the first and last entries of a `switch` case list (they end in `:` rather
than `,`), and map literals have the same hazard.

**The `./` on the package path is mandatory and its absence fails confidently.**
`score_one_mutation.sh x internal/lint/rules/core ...` reports that the package does
not compile, because Go resolves a bare path against the standard library:
`package internal/lint/rules/core is not in std`. Two porters hit this in one wave, and
the message primes you to go hunting a sibling's half-written file, which is exactly
the wrong place to look. Write `./internal/lint/rules/core`.

The tool refuses a rewrite that changed no bytes, refuses a mutant that does not
compile, and refuses to score at all unless the package is green before the mutation.
All three refusals are correct and **none is a pass.** Before the byte check existed, a
mutation changing one space inside a comment came back "caught by thirty-two lines".

**`CAUGHT: N` counts OUTPUT LINES CONTAINING "FAIL", not assertions, and reporting it as
assertions overstates it about threefold.** The tool's own line is honest -- it says
"failing line(s)" -- and it is the reader who converts that into a claim about coverage.

Decomposed on a real run, where the tool reported 6:

    --- FAIL: TestSomethingFires                 a test header
        --- FAIL: TestSomethingFires/invalid-20  a subtest header
    --- FAIL: TestSomethingElse                  another test header
    FAIL                                         a bare summary line
    FAIL  github.com/...                         the package line
    FAIL                                         another bare one

Two actual assertion failures under six matching lines. So a sweep on a rule with no
subtests and one failing assertion reports 4, and one with fifty failing subtests reports
53; the number moves with test STRUCTURE as much as with coverage.

Two consequences. **Quote it as the tool does, "CAUGHT" or "CAUGHT: N failing lines", not
as a count of assertions** -- a figure in a report that does not survive re-measurement is
the decay class this document keeps correcting, and this one was introduced by an author
who had just written a section about exactly that. And **when you want the assertion count,
grep for the file-and-line prefix** (`_test\.go:[0-9]+:`), which is one line per
`t.Errorf` and cannot be inflated by structure.

**None of which weakens a catch.** One assertion firing on the exact discriminating input
is a complete pin, and a larger number is not a better one; it usually just means the
mutation broke something structural. The count is a diagnostic, not a score.

**"Does not compile" is now a narrower message than it used to be, and this is worth
knowing.** Earlier revisions told you to read that refusal as a question rather than a
verdict, because the check was tree-wide and could not tell your broken mutation from a
sibling's half-written file. The script now runs `go vet` **on the package under test
only**, both for the baseline and for the mutant, and says so in its own output: "This
checked only $package, so the breakage is your rewrite rather than a sibling's file."
So the refusal is trustworthy about whose fault it is. The general habit still holds:
check whether the package compiles without your mutation before believing anything.

The fourth argument scopes both the baseline and the mutant to one rule's tests, which
is how you proceed when the package is held by other agents. **State the cost when you
use it:** a scoped run cannot see a guard in another package, so a mutation renaming a
rule reads as a survivor while the name guard in `internal/lint/registry` would have
caught it. Anything resting on a cross-package guard needs one unscoped run. The script
carries this warning in its own comments.

A sweep against a red package is not a weaker measurement. It is not a measurement. A
scoped sweep is a real measurement of a smaller thing.

**"Unscoped" does not mean "sees everything", and the difference has a measured
instance.** The fourth argument scopes the run to one rule's tests, and dropping it
widens the run to the whole PACKAGE — not to the tree. The tool builds and tests one
package either way, so any guard living elsewhere is invisible to both spellings.

Scored on `@typescript-eslint/consistent-return`, mutating the rule's `Name` string:

    scoped to TestConsistentReturn        SURVIVED
    unscoped over ./internal/lint/rules/typescript   SURVIVED
    go test ./internal/lint/registry/     CAUGHT, twice

Twice, because two different guards see it: `TestEveryRuleShipsAFixturePair` loses the
rule's test file, and `TestEveryRegisteredRuleIsReachableFromTheLiveConfig` reports the
renamed name as unreachable from the config. Neither lives in the rule's package and
neither can.

So the reflex "re-run it unscoped to be sure" does not close this gap. **Anything
resting on a cross-package guard is scored by running that package's suite against the
mutant directly**, which for a rename means editing the file, running
`go test ./internal/lint/registry/`, and restoring it. Say in your report which of the
two you did, because a survivor from a package sweep and a survivor from the registry
suite are different claims.

**A survival verdict can be a sampling artifact, so re-run a survivor before believing
it.** A porter's map-iteration mutant was caught 1 run in 8: Go randomises map order,
and with two entries it often preserves the original order by chance. A single run
reported "survived" and the honest answer was "survives most of the time". If your rule
iterates a map and the mutant reorders it, score it several times before writing the
equivalence argument.

### Mutation results are evidence about FIXTURES, not about the parser

Everything above establishes that a sweep proves a fixture set can see. Here is the next
thing, and it is a different claim: **a sweep proves nothing whatsoever about the
substrate.** Reading it as though it does produces a confident wrong belief about how the
AST behaves, and one that survives its own author's review, because mutation evidence
feels like measurement.

The shape first, because it generalises past the instance:

    a SURVIVING mutant   your fixtures cannot discriminate this change
    a CAUGHT mutant      your fixtures can discriminate this change

Both are statements about the discriminating power of a corpus. Neither is a statement
about what the parser returns, what a node kind contains, or what a helper answers. The
only instrument for those is a probe that asks the parser and prints what comes back.

**Measured, and the author had it backwards for two commits.** A rule needed a private
class member's name spelled the way `exceptMethods` entries are written, `#foo`. Upstream
builds it as `hashIfNeeded + name` because ESTree's `key.name` omits the hash, so the port
guarded with `if !strings.HasPrefix(text, "#")`. Two mutations were then run:

    remove the guard entirely          SURVIVED
    add the hash unconditionally       CAUGHT

The author read that pair as "the condition is dead and the body is load-bearing,
therefore `Text()` omits the hash", simplified the code to `"#" + name.Text()`, and broke
two fixtures. A five-line probe rule printing `name.Text()` for a private identifier
answered `"#foo"`: the hash was already there, the guard was correct, and its condition is
simply always true in this tree.

**Both mutation results are consistent with either belief**, which is why the pair could
not settle it and why the wrong reading felt supported. The survivor says the fixtures
cannot see the guard removed; that is equally true whether the guard is redundant or
whether removing it happens to leave behavior unchanged for another reason. The caught
mutant says the fixtures can see a doubled hash; that is equally true whether `Text()`
supplies the hash or the code does.

So when a sweep tempts you into a claim of the form "therefore the AST does X", stop and
write the probe. It is two minutes, it cannot be misread, and the alternative is a
simplification that compiles, reads better, and is wrong.

**The corollary for equivalence arguments.** A survivor you intend to write off as
equivalent needs a reason stated in terms of the CODE -- this branch is unreachable, these
two expressions compute the same value -- and that reason has to be established
independently. "It survived, so it must be redundant" is the inference this section exists
to stop. Elsewhere in this document a porter emptied two arms of `default-param-last`,
watched both survive, and found the arms genuinely reachable because the parser recovers
from illegal source. Same lesson from the other direction: the mutant was right that the
fixtures were blind, and wrong about why.

## The one lesson this document keeps relearning

Every instrument failure recorded here is the same failure, and it has appeared in a
fixture set, a mutation sweep, a differential sample, an extractor, a grep, and a shell.

**An instrument that cannot express disagreement reports agreement.**

So before believing a result, ask what input would have produced a different one, and
check that your instrument could have seen it. **If you cannot name such an input,
nothing has been measured.** The question is available before the conclusion, which is
what makes it worth more than the instances below.

And there is one repair, which every instance below received:

**Make the instrument refuse rather than answer.**

A missing thing errors and you notice. A wrong thing responds, plausibly, in the right
format, at the right magnitude, which is why none of the failures below announced itself.
So arrange for the wrong answer to be impossible rather than detectable: a harness that
errors on a mismatched denominator, a filter checked against output it must match, a
deleted build target so a failed build leaves nothing to read, a command that exits
non-zero when it is standing in the wrong repository. Six unrelated places got that same
repair in one night, which is the sign it is the statement and everything below is a face
of it.

Six worked examples, each of which cost real time:

**A corpus blind by construction.** `prefer-destructuring` passed all 103 of its imported
fixtures while reporting nothing on any modern file, because its guard tested
`NodeFlagsAwaitUsing` as a bit when it is a composite (`NodeFlagsConst | NodeFlagsUsing`)
and so matched every `const`. Upstream's corpus is written in `var`, so no imported case
could see the difference and no mutation of the rule could conjure one.

**A differential that agrees exactly.** A cross-check took the 30 files at the head of a
findings list and got 30 findings against 30. Every one of those files had exactly one
finding, so the sample could not express a disagreement about count. Re-sampling the 25
densest files gave 1213 against 1205, and the eight-finding gap was a real defect. **If a
per-file breakdown of your sample is all ones, the sample cannot fail.**

**A filter defeated by output shape rather than content.** A grep for `    --- FAIL` could
not see a failure printed at top level; two oracles filtering on `!fatal` counted
"Definition for rule X was not found" and "Unused eslint-disable directive" as findings,
both having a null rule id. **A filter is a probe and needs a control like any other:**
check the pattern against an output you know should match. And filter positively, on the
rule id you are testing, rather than negatively on the categories you happened to think of.

**A uniform verdict.** A sweep wrapper reported `REFUSED` for all twelve mutations of one
rule; running one by hand gave `CAUGHT: 89`. A uniform result across every input is the
same signal as a confident zero: the instrument is answering without looking.

**A wrong denominator.** A differential harness silently dropped six files on unknown-rule
directives, and the smaller sample read as agreement. An audit's counts were taken with
inline `eslint-disable` active, so 203 there is 204 here. **Make the harness refuse on a
wrong denominator rather than warn**, because a warning beside a plausible numerator is
read past.

**A command answering about the wrong repository.** A `git grep` finding nothing, a build
failing for want of a module, a rules count of zero, all from a working directory that had
drifted. A wrong-directory zero is indistinguishable from a right-directory zero. Assert
where you are standing before you believe a count:

    [ "$(git rev-parse --show-toplevel)" = "$expected" ] || { echo "wrong repo"; exit 1; }

The repair is the same in every case and it is not "add a fixture". **Build the input that
separates the two readings, confirm the shipped code and the mutant disagree on it, and add
it with a control.** Confirming both halves is what separates a fixture that kills a mutant
from one that merely happens to pass.

### Which instrument reaches which defect, and why you often need two

A surviving mutant tells you your fixtures cannot see something. Where that something lives
decides which tool can reach it:

    in your own code                  a mutation sweep reaches it
      a quantifier, a guard, a branch     the mutant compiles and the corpus stays green

    in the world                      only a dry run or a differential reaches it
      a shape upstream never wrote,       no mutation of your rule can conjure the input
      a substrate that behaves otherwise

That fork is the easy case. **The common case is a sequence and it needs both.** Measured on
`prefer-regexp-exec`: the dry run found a `const` guard silent on `let r = /x/; a.match(r)`,
which upstream reports; the replacement guard was then written, and a mutation deleting it
survived all 37 corpus rows, because upstream never binds a regex to a `let`; only upstream's
own behaviour settled what the guard should say.

**A fix made in response to a dry run is code no fixture has ever judged**, because the
fixtures are precisely what failed to raise it. So sweep after a dry-run repair, on the guard
you just touched rather than on the rule at large: re-sweeping everything is slower and buries
the one mutation that matters among a dozen already green.

### A control is only a control if you break it and watch it fail

Two rules, and they apply to every measurement above.

**Prove the probe can fail before you trust that it passes.** A control that cannot fail is
worth nothing, and the failure mode is not exotic: a sed pattern that matches nothing, a test
that recovers from the panic it exists to surface, a guard whose expectation agrees with any
value at all. Plant the defect, watch the probe fire, restore, watch it pass.

**Two broken instruments that share an assumption corroborate each other.** Agreement between
two measurements is evidence only if they could have disagreed. Two runs of the same wrong
grep agree perfectly. Prefer instruments that fail differently: the strongest confirmation in
this document is a node count matching across two agents' independent full-tree runs, because
nothing was shared between them but the tree.

### Your own comment is not evidence, and neither is a justification you wrote

A comment asserting a fact is a claim, not a measurement, and the cheapest thing in this
document is checking one. A doc comment here asserted that upstream's operator pattern
`/^[<>!=]?={0,2}$/` required the `=`; `{0,2}` permits zero, so bare `<` and `>` match.
Running upstream's own pattern settled it in one command.

**And a justification you write will shape the fixtures you write next.** Having argued why a
guard is right, the cases you then reach for are the ones that agree with the argument. One
guard here withheld 50 of 74 repairs, defended by "zero of the outputs begin with a semicolon"
counted from the string start; five of them do, mid-string after a newline, where the check
could not look. **The question was shaped like the answer it got.** When you find yourself
defending a decision rather than measuring it, that is the moment to build the discriminating
input instead.

### Check the thing you were just credited for

An audit carries counts and reasons, and they fail differently. A count decays, and
re-measuring produces a number that visibly disagrees. **A reason does not decay: it is either
true or it was never true**, and re-running the count leaves a false reason standing beside a
freshly confirmed number, which makes the whole row look verified.

`sort-vars` is audited No for two stated reasons: 2 violations, and "overlaps a rule we
already enforce in-house". The count is right. The overlap does not exist, and never did.

An audit field can also be accurate about upstream and wrong about what a port needs:
`prefer-arrow-callback` is listed "needs type information: no", which is true of ESLint, which
gets scope from eslint-scope, and false here, where telling a genuine self-reference from a
shadowed one is name resolution reached through the checker.

**A threshold quoted in prose decays exactly like a count.** This document's own sizing example
was 970 lines over 4,316 when written and is 1,001 over 5,075 today.

## 7b. If the rule's judgment is a matcher, the corpus is not enough

**A corpus tests a matcher only on inputs upstream thought of.** That sentence is the
whole section, and it is a different claim from the mutation sweep's. The sweep proves
your fixtures can SEE. This is about what they are pointed AT: a fixture set can be
demonstrably able to fail and still be blind, because every input in it is well-formed.

Measured on `no-restricted-imports`, whose `patterns` option matches a module specifier
with the `ignore` npm package — gitignore semantics, which nothing in this tree had, so
the port had to reimplement them. The differential was built the way this document asks:
every `group` configuration in upstream's corpus crossed with every module specifier
appearing in it, driven through the real `ignore` package. 864 pairs, 90 of them
matching, so neither degenerate answer could pass.

**It passed 864 of 864, the rule passed all 263 corpus rows, and three genuine defects
were still live:**

    a `./` trim on the candidate    made ["foo/bar"] match "./foo/bar"      upstream: false
    a `./` trim on the pattern      made ["./types"] match "types"          upstream: false
    `*` allowed to match nothing    made ["foo/*"] match "foo/" and         upstream: false
                                    "foo//bar"

None of the three is reachable from the corpus, and the reason is structural rather than
bad luck: **every specifier upstream wrote is a well-formed module name.** No corpus case
imports from `"./foo/bar"`, or `"foo/"`, or `"foo//bar"`, so no corpus case can separate a
matcher that is right about them from one that is not.

**The second instrument is the same cross product against inputs the corpus never
contains.** Take the configurations you already have and cross them with adversarial
inputs: a leading `./`, a trailing separator, a doubled separator, `../` and `../../`, a
bare `.` and `..`, a leading `#`, a trailing dot, a case-flipped name, the empty string.
540 more pairs, 76 matching, and all three defects fell out at once. Both tables now live
in the rule's test.

The general shape, for any rule whose verdict comes from a matcher rather than from the
tree:

- The corpus proves you agree with upstream **where upstream looked.**
- The adversarial cross product proves you agree **where it did not.**
- Neither is optional, and the second is the one that finds things.

This applies wherever a port reimplements a library rather than reading the AST: a glob
matcher, a regular-expression translation, a path normaliser, a name-mangling scheme. The
corpus tests the RULE. It does not test the library, because upstream never had to.

**The third defect is the one worth reading twice, because re-deriving it from first
principles gets it wrong the same way.** The bug was `foo/*` matching `foo/` and
`foo//bar`, which reads exactly like `*` needing to be one-or-more instead of
zero-or-more. It is not. Measured against the installed `ignore`:

    ["f*"]     against "f"          TRUE     `*` really is zero-or-more
    ["foo/*"]  against "foo/"       false
    ["foo/*"]  against "foo//bar"   false

So `*` is zero-or-more and correct as written, and the real rule is that **an empty path
SEGMENT never matches.** The constraint belongs on the segment, not on the wildcard, and
"fixing" the wildcard breaks `f*` while appearing to fix the symptom. Write the comment
that says so at the line, because the next reader's instinct will be the wrong one and
the fix that follows from it passes the failing case.

## 7c. Your fixtures never cross the config boundary, so they cannot test your decoder

**Every fixture reaches your decoder with bytes the TEST built. The config layer builds
different bytes.** That sentence is the whole section, and it names a blindness neither
of the two instruments above can reach.

The mutation sweep proves your fixtures can SEE. The adversarial cross product proves
they are pointed at enough INPUTS. **Neither can see this one at all**, because the
defect does not live in the rule or in its corpus. It lives on the boundary between the
config layer and the rule, and a fixture never crosses that boundary: it hands the
decoder a literal it wrote itself, so the decoder is exercised on the shape the fixture
author already believed in.

The measured instance was one line. `internal/lint/configuration/configuration.go` parsed a
rule setting as `["severity", <options>]` and kept exactly one element, `tuple[1]`, while
ESLint's `context.options` is **every** element after the severity. Every rule whose upstream
option surface has more than one element was configured differently from upstream, and each
decoder had to accept a cohere-only shape or be wrong on every real config while right on every
fixture. The two rules below are what that cost.

That line is gone. The config layer now keeps every element, and **the rule declares its
arity at registration**: `Decode` for a rule taking one option element (it is handed that
element), `DecodeOptionList` for a rule whose upstream `meta.schema` takes more than one (it is
handed every element as one JSON array, upstream's `context.options` exactly). A second element
for a `Decode` rule, or any element for a rule with no decoder, is refused at startup naming the
rule and the element. `internal/lint/registry/option_elements_test.go` pins all of it through the
real config layer. The blindness this section names has not gone anywhere: a fixture still hands
your decoder bytes the test built.

Measured, on two rules ported together, and the pair is the point because **they fail in
opposite directions:**

    id-denylist   ["error", "data", "err", "cb"]              upstream's variadic spelling
                  decoder receives  "data"                     a bare string, not a list
                  -> rule configuration: rule id-denylist: decoding []string:
                     json: cannot unmarshal string into Go value of type []string

    id-match      ["error", "^[a-z]+$", { "properties": true }]
                  decoder receives  "^[a-z]+$"                 the flag object is GONE
                  -> no error. The pattern arrives, the rule runs, reports, and looks
                     configured, with every option silently off.

**The loud one is the lucky one.** `id-denylist` refused to start, which is a failure
nobody can ship past. `id-match` is the shape this repository exists to refuse: a rule
that registers, passes a complete fixture pair in both directions, appears in `--rules`,
and enforces something other than what the config says. Its corpus has 100 cases and 52
of them carry `properties: true`; every one still passed, because every one reached the
decoder through the test's own bytes.

**Scored, rather than asserted**, because the first draft of this section claimed the
blindness without measuring it and was wrong about which mutant demonstrates it. Restoring
the original defect -- a decoder that understands only upstream's flat tuple and not the
nested shape a config delivers -- gives:

    against the 100 imported corpus rows          SURVIVED
    against the same rows plus the contract test  CAUGHT, 4 failing lines

`id-denylist` scores the same way on its own defect, and the symmetry is the point: the
loud failure and the silent one are one blindness, and the noise was luck rather than a
difference in kind.

    against the 143 imported corpus rows          SURVIVED
    against the same rows plus the contract test  CAUGHT, 4 failing lines

That is the section in two measurements. The corpus cannot see it and a table written
specifically for the boundary can, so the boundary needs its own table.

**The instrument is a dry run against the real tree, through the real config layer**, and
it is the only one that sees this. Not `go test`. Not the sweep. Seed a small tree with a
config that actually exercises your options, and read what comes back:

    <scratchpad>/probetree/CohereSettings.json     {"rules": {"<your rule>": ["error", <options>]}}
    <scratchpad>/probetree/tsconfig.json           the harness tsconfig, copied
    <scratchpad>/probetree/source/Probe1.ts        source that must report
    <scratchpad>/probetree/source/Probe2.ts        source that must not

    cd <scratchpad>/probetree && <your binary> --no-fix --lint

The second file is not optional. A configured rule that reports on everything and one
that reports on the right things look identical from the first file alone, and a decoder
that silently dropped your options usually still reports SOMETHING.

Then write the contract into the test suite, because a dry run is a thing you did once
and a fixture is a thing that keeps being true. For a `DecodeOptionList` rule the rows to pin
are the lists the config layer delivers, which are upstream's own spellings minus the severity:

    ["data", "err"]                 upstream's variadic form, every element read
    ["data"]                        one element
    nil                             a bare severity, which must decode to upstream's default
    a shape that is neither         must ERROR, never decode to an empty configuration

and add the rule to `optionElementRules` and `optionElementCases` in
`internal/lint/registry/option_elements_test.go`, with a control that omits the second element
and a treatment that writes it, whose finding counts differ.

That last row matters more than it looks. A decoder that answers "no options" to a
malformed config produces exactly the inert rule this section is about, one layer further
in. Refuse instead.

**The general form, for any rule with an option surface:** check `meta.schema` for how many
elements it declares. One element is `Decode`. More than one, or an `anyOf` of lists, or a
variadic `type: "array"` schema, is `DecodeOptionList`, and the decoder reads upstream's list as
written. A key or element your port cannot honour is refused by name rather than accepted and
ignored.

**Decode every option object with `rule.UnmarshalOptions`, and tag every field with upstream's
key.** `json.Unmarshal` drops an unknown key and matches a key case-insensitively, so a config with
`ignorePatern` loads clean and the option does nothing. `rule.DecodeOptionsInto` already uses it; a
hand-written decoder must too, including inside a custom `UnmarshalJSON` and inside the object arm
of a value that may also be a string or a boolean. There, do not fall back to the default when the
object is refused: that drops the whole setting where the lenient decode dropped one key, so return
the error. Where upstream's schema leaves an object open (no `additionalProperties: false`), keep
`json.Unmarshal`, say so at the site, and name the site in `lenientOptionDecodes`.
`internal/lint/registry/option_decoding_test.go` refuses the rest by reading the tree with types,
and `option_refusal_test.go` hands every registered decoder an unknown key (#4a4yse4).

### A presence check followed by a truthiness fallback is a family, and a written zero splits it

Not one rule's quirk. Found in `max-depth` and `complexity`, and the shape is common
enough in ESLint's option handling that the next rule carrying it should be expected
rather than discovered:

    if (typeof option === "object" && (hasOwn(option, "maximum") || hasOwn(option, "max")))
        threshold = option.maximum || option.max;

The guard tests PRESENCE and the `||` tests TRUTHINESS. Those agree on every value except
zero, and on zero they do something a reading of either half alone does not predict.
Measured against the installed rules, reading each limit out of the rule's own message
rather than inferring it from a count:

    2                        limit 2      the ordinary spelling
    0                        limit 0      a bare zero reports everything
    {"max": 0}               limit 0      `undefined || 0` is 0, so the zero WINS
    {"maximum": 0}           SILENT       `0 || undefined` is undefined, and every
                                          `count > undefined` is false, so the rule
                                          reports NOTHING at all
    {"max": 0, "maximum": 3} limit 3      the `||` consults maximum first
    {}                       the default  neither key present, so the branch is never
                                          entered: 4 for max-depth, 20 for complexity

The same six rows hold for both rules, checked on each rather than generalised from one.
**And the last row carries a trap of its own worth naming**, because it caught this
document: probing `{}` against source that does not exceed the DEFAULT reports nothing,
which reads exactly like the silencing row above it. Two rows that mean opposite things
produce identical output unless the probe input is chosen to exceed the default. Use
source past the threshold you are testing, or the measurement quietly merges two cases.

**The two zero spellings do opposite things**, and one of them disables the rule entirely.
That is not expressible as an integer threshold, so a decoder needs either a flag or a
sentinel; collapsing the rows into "zero means zero" is wrong on both.

**The mechanical tell, so this is recognised rather than rediscovered:** a presence check
gating a truthiness pick means a written zero and an absent option take DIFFERENT paths.
Any decoder that treats "absent" and "zero" as the same thing is wrong on at least one
row, and whether a fixture catches it is luck. `max-depth`'s corpus happens to carry both
zero spellings, which is the only reason the first draft of its decoder was caught.
`complexity`'s carries neither, and the rows are pinned there by hand from measurement.

So when a decoder's upstream is `hasOwn(...)` then `a || b`, write the option table out and
drive the installed rule once per row before writing any Go. It is six inputs and it
settles what no amount of reading the expression will.

### The blindness is not only about decoders, and the second instance is the sharper one

Everything above is a rule wrong about a shape the CONFIG produces, which at least lives
in your own code where a reader could find it. The same blindness has a worse form: **a
true statement about a constant in another package**, which no fixture and no mutation of
your rule can reach at all.

Measured on `prefer-destructuring`. It skips `using` and `await using` declarations,
because a destructured `using` is a parse error and a finding there is unactionable.
Upstream tests two declaration kinds, so the obvious port is two flag tests:

    list.Flags&ast.NodeFlagsUsing != 0 || list.Flags&ast.NodeFlagsAwaitUsing != 0

The second half is wrong, and the reason is one line in the vendored compiler:

    NodeFlagsAwaitUsing = NodeFlagsConst | NodeFlagsUsing      ast/nodeflags.go:51

It is a COMPOSITE rather than a bit of its own, so the mask is non-zero for **every
`const` declaration** and the guard skipped all of them. The rule reported nothing on any
modern source file, and:

    upstream's 103 imported fixtures     ALL PASS
    18 fix vectors, 30 declines          ALL PASS
    every mutation of the rule           caught or equivalent, as expected
    a dry run on a seeded tree           0 findings, control reporting 3

**The corpus is blind by construction rather than by bad luck.** Upstream's cases are
written in `var`, because the rule predates `const` being ubiquitous in test corpora, so
no imported case can distinguish a rule that handles `const` from one that skips it. And
unlike the decoder cases above, there is nothing in the rule to mutate: the code is
correct about the operation it performs and wrong about what the constant means.

**Three things follow, and the third is the one that generalises.**

Reaching for a named constant is not free the way it looks. Check whether it is a single
bit before masking against it, and prefer the narrowest flag that answers your question —
`NodeFlagsUsing` alone is right for both spellings, since `await using` carries it too.

A corpus written in one dialect is silent about the others. Upstream's is `var`-era
JavaScript; this tree is `const`-era TypeScript. That gap is the same one that made every
`.tsx` file gate invisible, and it is yours to close with cases of your own.

And the instrument that finds this class is the same one section 7c already names: **a dry
run through the real config layer against a seeded tree, with a control that reports.**
Not `go test`, not the sweep. The rule here reported zero and the control reported three,
which is the only shape that separates "this tree is clean" from "this rule cannot fire."
Run it on every port, including the ones with no options at all — this rule's defect had
nothing to do with its option surface, and a porter who reads 7c as being about decoders
will skip the run on exactly the rules where it is the only instrument left.

## 7d. A screen for an ABSENT thing is most confident when it is broken

**An audit's zero cannot tell a clean tree from a rule that cannot fire here.** Several
rules carry the verdict "Yes, violations: none" while being structurally incapable of
reporting in this tree, and the count reads identically either way. The two causes look
the same from outside and are completely different:

    the rule is configuration-driven      it reads its OWN `context.options`, which the
                                          config layer hands to rule.Context. Port it;
                                          the zero means "nothing is configured yet".
    the rule reads a SHARED setting       it reads `context.settings.*`, which
                                          rule.Context has no path to and which ahra
                                          configures nowhere. That arm cannot work here.

`react/prefer-exact-props` is the second kind, measured independently by two agents weeks
apart with identical results: all 12 of its reporting cases go silent with no settings,
because `getExactPropWrapperFunctions(context)` is the only input its reporting branch
has. The ruling is at `internal/lint/rules/react/forbid_prop_types.go:158-183`.
`id-denylist` and `id-match` are the first kind. Both zeros looked the same in the audit.

**The obvious screen is to grep the rule file for `context.settings`, and it is broken.**
This is the part to read twice, because a reader who re-derives the screen will write the
broken one:

    grep -cE "context\.settings" .../rules/id-denylist.js             0    correct
    grep -cE "context\.settings" .../rules/forbid-prop-types.js       0    WRONG

`forbid-prop-types` reads that setting **indirectly**, through
`eslint-plugin-react/lib/util/propWrapper.js:21`:

    return new Set(context.settings.propWrapperFunctions || []);

So it passes a direct scan clean and then registers a port that cannot fire. Run that
screen across a queue and you get a confident clean sweep over exactly the rules it
cannot see.

**And this is the general hazard, not a detail about one plugin.** A grep can only find
the names you thought of. That is the same failure as a corpus containing only the inputs
upstream thought of (7b), one level out: in both cases the instrument's blind spot is
shaped like the author's imagination, and in both cases the output of a blind instrument
is a clean result. **A screen looking for an absent thing gives its most confident answer
when it is broken**, because "found nothing" is simultaneously the success signal and the
failure signal.

**The executable check needs no guess about which helper name to look for.** Drive the
INSTALLED rule twice over every corpus case, once with a settings block and once without,
and compare. A rule whose reporting path reaches `context.settings` moves; one that reads
only its own options does not. Measured across three instruments on the two rules above:

    the rule file names a shared setting          0        0        control: 0  (missed it)
    its transitive imports do                     0        0        control: 1  (found it)
    corpus cases that move with a settings block  0/145    0/100    control: 0 -> 1

Only the third needs no guess, and only the third is worth running across a queue.

**The refusal is the load-bearing piece, not a nicety.** Require a control that MOVES
before believing a subject that does not, and make the probe exit non-zero when the
control stays flat. Without it a flat subject result is a number rather than a
measurement, and the two are indistinguishable in the output:

    if (controlMoved === 0) {
      console.log('REFUSING: the control did not move, so this probe is not measuring ' +
                  'a settings dependency');
      process.exit(1);
    }

That refusal fired twice while this was being built, and both causes were the instrument
rather than the rules:

    a settings block of the wrong SHAPE    propWrapperFunctions wants
                                           [{property: 'exact', exact: true}], and a bare
                                           string array leaves the control silent
    a hand-written control case            it never triggered the rule at all

Both were fixed by taking the settings block and a reporting case **verbatim from the
control rule's own corpus** rather than inventing them, which is step 3's rule applied to
the instrument instead of to the fixtures. Only then did the control move 0 -> 1 and the
subjects' flat 0-of-245 become a result.

**So the screen, in order:** read `meta.schema` for the rule's own option surface, then
drive the installed rule with and without settings over its whole corpus, with a control
that is known to depend on settings and is required to move. If any part of a rule's
reporting path reaches outside `meta.schema`, that arm cannot work here, and saying so
with the measurement is a better outcome than a port that registers and never fires.

## 8. Assert spans, not just message ids

**`ExpectFindings` asserts message ids and count and nothing else.** A rule whose
defect is where it points, or what it rewrites, passes a complete fixture pair while
being wrong.

**Assert the span of every rule, whether or not it carries a repair.** An author
derived file positions from a string literal's cooked value while the file holds raw
text (`"[á]"` is six bytes raw and four cooked), so every finding pointed at the wrong
place. All 187 verbatim fixtures were green, and the rule carried no fix at all, so
gating span checks on "does it have a fix" would have shipped it. Slice the source with
the finding's own range and compare the text:

    reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]

**When two upstream implementations disagree about where to point, the span is the only
fixture that records which one you ported.** One rule had two plausible spans, with an
identical message id. Every id fixture stayed green over the wrong choice.

**Assert the message text exactly.** An author's rule rendered `'##unusedField'` with a
doubled hash, because a private name node's `Text()` already carries the `#`. The
fixture used `strings.Contains(got, "#forgotten")`, and a two-hash string contains a
one-hash needle, so it stayed green while the message was wrong. The general form is
worth more than the instance: **a fixture whose predicate is weaker than the property
it guards is not a guard.**

**If the rule carries a fix, assert it by applying the repair and comparing the
resulting source**, never by comparing the fix's text. A fix writing the right string
over the wrong span passes a text comparison, and that is a real defect we shipped.

**Route option fixtures through the rule's own exported decoder**, not by building the
options struct directly. That is what puts a default inversion or a serde alias under
test, and those are the two lines most likely to have no upstream counterpart.

### If you score a rule in two directions, pin both

A rule measured against a corpus usually gets two numbers: how often it fires where it
should, and how often it fires where it should not. **Pinning one exactly and asserting
only `> 0` on the other is not two instruments. It is one instrument and a formality.**

The asymmetry runs one way, which is why it survives review. A change that trades true
positives away for false-positive removals moves the pinned number in the direction
that reads as success, while the loose side stays green because it is still non-zero.

Measured, on `preserve-manual-memoization`: freezing component parameters in the
aliasing graph took the clean-fixture rate from 31 of 70 to 19, a 39% improvement
against the number that rule was being judged on. The golden count went from 15 of 33
to 9 in the same run. Six programs upstream reports on had gone silent underneath a
result that looked like progress, and the author had the 19 and was composing the
commit message before checking the other side.

So: assert both counts exactly, and make each failure message say which direction means
what. "This went down, so either the rule improved and you lower it deliberately, or it
stopped reporting things upstream reports on" is the sentence that does the work. A
ceiling cannot distinguish a fix from a rule that went quiet.

### A narrow assertion throws away corpus coverage you already have

Two mutants survived a sweep not because the corpus lacked the cases, but because the
assertion only checked the message PREFIX. The union text in the message was never
compared, so the rule could print anything after the prefix and the fixtures stayed
green. Widening the assertion to the whole message killed both mutants immediately,
with no new fixtures.

That is the cheapest coverage in existence: it was already paid for and being discarded
at the last step. Before adding a fixture to kill a survivor, check whether the fixture
you have is being fully read.

The same sweep then found two real defects behind that gap: a parenthesised type was
not unwrapped, so `(/* before */ string /* after */)` printed the author's parentheses
AND the comments inside them instead of `string`; and a parenthesised union counted as
one member instead of flattening.

**Assert the whole message, not its prefix, whenever the message carries computed
text.** A message id proves which branch fired. The text proves what the rule computed,
and that is usually the part with the bug.

## 8b. If your rule anchors on a declaration and looks for writes to it

The shape is: resolve the declaration's symbol once, then for each candidate identifier
resolve it and LOOP over `symbol.Declarations` looking for the anchored declaration by
NODE IDENTITY. Never index `[0]`.

**Three traps, all of which cost real work and none of which the imported corpus can
catch.**

**Identity alone over-reports.** `A.x = 0` and `foo(A)` both resolve to the anchor and
neither reassigns it. You need structural write detection as well, and report only
where both agree. `reference.WritesToBinding`
(`internal/lint/ecmascript/reference/write.go:100`) is the shelf's answer to the second
half.

**Comparing declaration KIND instead of identity survives the entire upstream corpus.**
A kind mutant passed all 25 upstream cases plus thirteen invented ones, because
upstream's only same-kind pair has a write that resolves to nothing and exits early.
The distinguishing input is a nested redeclaration:

    class A {} { class A {} A = 1; }        two anchors, same kind, identity reports once

Write the equivalent for your declaration type, and mutate to confirm it catches.
Without it the trap ships and nothing tells you.

**This trap is not confined to declaration-anchor rules.** It bit a flow-analysis rule
for **223 false positives on the real tree**, fifty in one file, while all 71 imported
clean cases stayed green. Any rule that resolves an identifier through the checker can
hit it.

**A shorthand destructuring target needs a different accessor.** `({A} = 0)` resolves
through `GetSymbolAtLocation` to the *property's own* symbol, which reads exactly like
a correctly-declined shadow. Use
`ctx.TypeChecker.GetShorthandAssignmentValueSymbol(identifier.Parent)`, which lives in
the vendored compiler's `checker/services.go` rather than `checker.go`, so a grep
scoped to the wrong file reports it absent.

**Cost is settled: dispatch freely.** The per-file checker lock measured `setup` 0.52ms
for a class anchor against 0.51ms for a much rarer anchor over identical files. It is
flat and does not scale with what the rule does with the checker.

## 9. Register and enable

**Write `<your_rule>_register.go`. Do not open the package's `register.go`.**

    // internal/lint/rules/core/eqeqeq_register.go
    package core

    import "github.com/system-inc/cohere/internal/lint/rule"

    func init() {
        rule.Register(rule.Registration{Rule: Eqeqeq})
    }

    // with options:
    //   rule.Register(rule.Registration{Rule: Eqeqeq, Decode: rule.DecodeOptionsInto[EqeqeqOptions]()})

`rule.Register` appends into a map keyed by name and is callable from any file's
`init`, so a package may hold as many of these as it has rules. This is what makes a
port **three new files and no shared edit**: the rule, its test, and its registration.
Two authors landing at once touch nothing in common. There are over two hundred
`_register.go` files in the tree; this is the live convention, not a suggestion.

The `register.go` still present in most packages is a historical list, not a
requirement. Adding to it works and costs you a conflict with every other author in
your package, which is the whole reason not to. A duplicate name panics at startup
rather than one rule silently losing, so a collision is a crash in the first second
rather than a gap nobody measures.

Then add the rule name to `/Users/kirkouimet/Projects/ahra/CohereSettings.json` in the
rules block. **Not `.oxlintrc.json`**, which is oxlint's own config and which cohere no
longer reads. The live guards name `CohereSettings.json` at
`internal/lint/registry/live_wiring_test.go:12` and `internal/lint/configuration/live_test.go`.

**Grep that file for your rule name first.** Many rules are already enabled there.
Adding a second entry produces valid JSON with a duplicate key, one silently wins, and
nothing complains. If it is already present, that step is done.

> **`rule-inventory.json` is gone.** Earlier revisions of this document devoted most of
> this section to it: how to edit it as text rather than round-tripping it through a
> JSON library (a re-serialise silently reformatted all 2,156 lines from spaces to
> tabs), how to keep its `total` / `byNamespace` / `bySeverity` / `byEnabledBy` headers
> consistent with `TestInventoryCountersMatchItsOwnRules`, how an entry had already
> been lost to a concurrent write and gone unnoticed for two waves, and how to
> re-read the file afterwards to confirm your entry survived.
>
> None of that applies now. The file was deleted and its guards with it. What survives
> is the reasoning, and it generalises to any shared tracked file: **a shared file
> edited in place by several agents loses entries silently, a round-tripped
> serialisation produces an unreviewable diff that passes every check, and a derived
> counter that nothing asserts is a number that decays.** Prefer a per-rule file to a
> shared list wherever the choice exists, which is exactly why registration moved to
> `<rule>_register.go`.
>
> Where you used to ask the inventory a question, ask the binary: `--rules` lists what
> is actually registered, and `--lint` names every config key that resolves to nothing.

**Done means enforcing.** A rule registered and enabled nowhere passes every fixture
and lints zero files. `TestEveryRegisteredRuleIsReachableFromTheLiveConfig` catches
this, and it is the one you are most likely to skip, because your own `_test.go` is
green and that feels like the answer. It is not. Your fixtures prove the rule is
correct; only the live config proves it is reachable.

**Seeing `ctx.scoping()` upstream is not evidence you need the checker.** That table
answers two different kinds of question and only one of them costs anything:

    scope flags        is_set_accessor, is_function, is_static, is_private
                       properties of the enclosing construct. The AST already knows.
                       No checker.
    name resolution    which declaration does this identifier bind to
                       needs the checker. `resolvesToAGlobal` in
                       no_new_native_nonconstructor.go is the shipped example.

**But `resolvesToAGlobal` is a trap for any rule about `undefined` or `globalThis`.**
It returns false when `len(symbol.Declarations) == 0`, and the real `undefined` global
has zero declarations — so it answers **false on every case such a rule must report**.
An author following this exemplar ships a rule that is silently inert and whose
fixtures all pass, which is the worst combination available. The guard is
`symbol == nil || len(symbol.Declarations) == 0`. For `undefined`, `window`, and
`globalThis` alike, the correct predicate is the **complement** — ask whether the name
is declared in source, and treat "not declared" as the global. One rule had to report
every `globalThis.RegExp` case and silently reported none for exactly this reason.

`no-setter-return` reaches for `ctx.scoping()` and asks it only the first kind, so it
declares nothing. An author nearly declared the checker on the strength of seeing the
call. **Determine which one your rule asks, by probe rather than by reading the call
site**, and say in your report how you determined it.

**If your rule declares `NeedsTypeChecker`, two things bite here.**

Your fixtures must use `testing.RunTyped` rather than `testing.Run`. The plain harness
hands the rule a nil checker.

**What happens next depends on whether the rule guards.** Measured both ways: `no-eval`
guards, so it goes *completely silent* — every StaysSilent case passes vacuously and
every Fires case fails in a way that looks like a rule bug. A rule that dereferences a
checker result **without** a nil check may panic and take the whole run down — **but
the shim tolerates more than you would expect.** Measured: `GetSymbolAtLocation` on a
nil checker returns nil rather than crashing, so for that call the failure mode is
**silence**, which is the more dangerous of the two. Keep the guard regardless; just
know that a missing guard often buys you a vacuous green rather than an obvious crash.
An author found their own rule doing exactly that, and a sweep of the tree found
**three shipped typed rules with no guard at all**.

And the registry guard is per-file textual: it looks for a `.TypeChecker` selector in
the rule's own file. If you reach the checker only through a shelf helper in another
file, the guard reports that you over-declared. That message is wrong and the
declaration is right.

### Registering without enabling needs an exemption entry, in the same commit

Most ports land registered and not enabled, because the tree has a cost the rule would
flag and enabling it is a scheduled cleanup rather than a wiring step. That is a normal
outcome and it is not finished until the guard knows about it.

`TestEveryRegisteredRuleIsReachableFromTheLiveConfig` fails for any registered rule the
live config cannot resolve, because such a rule passes its own fixtures and lints nothing.
Add an entry to the `deliberatelyNotEnabled` map naming the measured count and why:

    "no-continue": "ported and registered, not enabled: 1,191 findings over 3,540 files,
                    which is a control-flow convention for the whole tree rather than a
                    defect class",

**The entry and the rule go in the same commit.** A rule committed without its entry leaves
HEAD red for everybody else, and it has happened here. That map lives in a file several
agents edit at once, so read §12 before staging it.

## Ship the judgment, decline the fixer, and record the split as a decision

A rule whose judgment is portable and whose repair is not should ship as two decisions rather
than waiting for both. Reporting without a fix is the subset you can show correct, and the
reason has teeth: **a rule that reports correctly and repairs wrongly is strictly worse than
one that only reports**, because the wrong repair is applied unattended.

That is measured, not hypothetical. `object-shorthand` shipped eleven defects and every one was
in a repair, never in a count; nine produced source that parses fine and means something else,
including dropped generic type parameters, so `key: <T>(): void => {}` became `key(): void {}`
with four findings unchanged.

**The distinguishing question is whether upstream's `output` fixtures fully specify the
repair.** If they do, port it: `prefer-template` was worth porting whole despite five repair
defects, because each was findable by byte-exact comparison against upstream's own output. If
the repair reconstructs a reference or synthesises a node the corpus never pins, the fixer is a
separate job.

**Read the `output` ratio before accepting any judgment/fixer split.** Line counts mislead here.
`prefer-optional-chain` looks like the ideal candidate until you count: the fixer is 323 lines
of 1,807, and 729 of its 739 corpus cases carry an `output`. Halving that rule along the
judgment line discards 98% of its coverage, whatever the source division suggests.

**And record the split as a decision, not as an absence.** An undocumented missing fixer reads
as an unfinished port, and the next person re-derives the whole analysis before discovering
somebody already made the call. Put the reason and the boundary in the rule's doc and leave the
repair cases in the corpus so the fixer stays separately dispatchable.

## 10. Run it dry against the real tree

    go build -o <your scratchpad>/cohere-<yourname> ./command/cohere
    <your scratchpad>/cohere-<yourname> --no-fix --lint --timing 2>&1 | grep <your-rule>

**`--no-fix` is not optional, and `--fix` is not what causes writing.** The fix phase
runs by default, so a bare `cohere --timing` mutates the tree. A seeded probe file was
rewritten this way. `--no-fix` mutates nothing; `--lint` also skips the type phase,
which you do not need for a rule timing and which costs a second.

**And `--no-fix` WITHOUT `--lint` does not lint, which is a different way to get a
confident zero.** An author measured a rule at zero findings on the whole tree this way and
briefly believed it. The binary is not quiet about it:

    types: 17 diagnostics over 3540 files in 547ms
    phases: fix skipped (--no-fix) - types ran in 547ms - lint did not run
            (types bailed: 17 type diagnostics - lint findings against wrong
            semantics are noise) - unused skipped
    this run did not check everything - the phases above say what was not checked

Two separate lines announce it, one of which exists purely to say the run was partial. So
the failure was not a silent tool; it was **reading the count without reading the statement
of what ran**, which is smaller and far more common. The same rule measured 508 findings
once `--lint` was passed.

The general form is worth more than the flag: **a summary number is only meaningful
together with the statement of what produced it.** Checking for the `lint:` summary line
works precisely because its absence is the tool telling you, and the phase line is the tool
telling you in sentences.

This tells you three things fixtures cannot: whether it fires on real code and whether
those findings are right, what it costs (one rule was 64.5% of all rule time because it
rebuilt a map per file), and whether it explodes on files larger than a fixture.

**A zero is three different things and the file count does not separate them.** An
author saw thousands of files and "registered no listener on any of them", which read
as inert and was not: the tree's tsconfig includes only TypeScript extensions and their
rule declines JavaScript, exactly as upstream does. Two moves resolve it: compare
against a control rule offered the same files and known to be in the same state, and
build a small tree that should trigger the rule.

The run's own output separates the two cases for you, and the wording is deliberate:

    note: rule @typescript-eslint/no-base-to-string was offered no files — nothing
    wired it, so its silence says nothing about the tree

**A non-zero registration count with zero findings is the positive tell that a rule
works.** `was offered no files` is the failure signal. That separates "inert" from
"correctly declining" in about ten seconds without building a probe tree.

**A rule that declines on a compiler option or a missing precondition calls `ctx.Skip`.**
`ctx.Skip("strictNullChecks is off")`, once, then return no listeners. The run names it
by default and under `--coverage`, as `rule: skipped on N files, reason`. Without it, a
rule that cannot judge reads exactly like one that judged and found nothing:
no-useless-default-assignment declined every file in ahra and www while every run read
green, because a shared helper read an unset `strict` as off (#6ar414z,
#pa7k7zv). A file the rule is simply not about, a `.d.ts` or a file outside the directory
it guards, is scope rather than a skip, and stays silent.

**Except for a rule with a decoder, where it is not.** An author's rule reported 3,407
registrations over 3,407 files and was completely broken. A rule configured as bare
`"error"` is handed **nil** options: `rule.DecodeOptionsInto` errors on empty input,
the options registry turns that into nil for a non-required rule, and `options.(T)` on
nil yields the zero value. Every fixture reached the rule through the decoder, so
nothing in the suite could see it, and 21 of 21 imported cases passed. **Give your rule
an explicit nil-options fallback and a fixture that bypasses the decoder.** If the rule
genuinely cannot work without its options, set `RequiresOptions: true` on the
registration, whose doc comment records the inert-rule defect that field exists for.

**A probe tree's config must be named `CohereSettings.json` and puts rules at the JSON
top level.** A wrong filename is worse than a wrong shape: cohere exits before linting
and prints nothing your grep will catch, so you get a zero that was never a
measurement. An author seeded `{"cohere":{"rules":...}}` and got `was offered no
files`, which reads exactly like the inert case. The shape is
`{ "rules": { "react/your-rule": "error" } }`.

**Name probe files by index, never by the identifier under test.** macOS has a
case-insensitive filesystem, so `d_getStaticProps.tsx` and `d_getstaticprops.tsx`
collapse into one file and the last write wins. An author measured a case-only change
as "does not report" and nearly recorded a rule-level divergence; their control fired,
which is the only reason they went looking. This produces a confident **wrong**
measurement rather than a zero.

**Unless the rule gates on the FILENAME, in which case index-prefixing IS the identifier
under test.** An author's `h1_document.tsx` probes silently failed upstream's
`starts_with("_document.")` and all twelve came back with the wrong message arm — a
coherent-looking result, not an error. **When a rule reads the path, vary the DIRECTORY
and keep the basename exact.**

**A path-bearing option specifier makes the FIXTURE FILENAME part of the corpus.** An
author's imported clean case failed and the rule was right: a `from: file` specifier
with a `path` compares against the **declaring file's absolute name**, so that case is
clean only because upstream named its fixture `file.ts`, while three reporting cases
depend on the same comparison failing for `tests/fixtures/file.ts`. Varying only the
filename reproduced all three rows.

**A panic in your rule costs every rule that file.** The walk recovers per *file*, not
per rule, so one nil dereference does not lose your verdict on one file — it takes the
whole file away from every registered rule. A single unguarded
`ast.SkipParentheses(memberAccessObject(callee))` in `no-alert` crashed **167 files,
about five percent of the tree**, and the run still printed green: crashed files are
named in the output but a summary line reading `lint: 126 findings` looks identical
either way.

So grep the crash line before you believe a clean run:

    cohere --lint 2>&1 | grep -c "crashed:"     # must be 0

**And assert where you are standing before you believe a count.** The measurement runs
against a tree in another directory, so every command in this section is one `cd` away
from answering about the wrong repository, and it answers quietly: a `--rules` count of
zero, a `git grep` that finds nothing, a `go build` that fails for want of a module. None
of those looks like an error. A wrong-directory zero is indistinguishable from a
right-directory zero, and the reading that follows is confident and wrong.

    [ "$(git rev-parse --show-toplevel)" = "$expected" ] || { echo "wrong repo"; exit 1; }

One line, and it converts a plausible zero into a loud failure. It is the general repair
applied to the shell rather than to a test. Three separate people hit this in
one night from three directions -- an extractor counting "No matching configuration found"
as a finding, a harness silently dropping six files, and a coordinator measuring from the
wrong checkout five times -- so the pattern is not a personal lapse and a reminder will not
fix it. Only the refusal does.

Two things make this class hard to find from fixtures. **`ast.SkipParentheses`
dereferences its argument**, so any helper that can answer nil is a panic one call
later — `memberAccessObject` answers nil for every callee that is neither a property
nor an element access. And **the shapes that reach it are ordinary**: `super(...)` and
`import(...)` are call expressions whose callee is a bare keyword, and every subclass
constructor carries a `super()`. Two independent authors hit this same class in one
wave, in different rules. `TestAPanickingRuleLosesOnlyItsFile` and the crash corpus in
`internal/lint/registry/` are the standing guards.

**Where a nil check is unnecessary, say why at the line.** `no-useless-call` and
`prefer-spread` both take the identical shape and are safe by a gate established
fifteen lines earlier. Both carry a comment naming the invariant *and* the consequence
of widening it. That is the difference between a guard someone deletes as dead and a
guard someone keeps.

**And a fixture that passes against the unfixed rule proves nothing.** The `no-alert`
fixtures never reach the nil path in the rule harness; the proof was an A/B on the real
tree, same flags, differing only in that file — 167 crashed before, 0 after. When your
unit test cannot reach the defect, say so in the test comment and name what did prove
it, rather than letting a green suite imply coverage it does not have.

### The binary that answered may not be the one you built

**Restored after a consolidation cut it. This has been the highest-cost trap in this setup.**

`go build -o <path>` leaves the previous binary in place when the compile fails, so a broken
tree does not clear the artifact. The stale binary then answers, and it answers plausibly.
Measured tonight, with another agent's rule mid-write:

    go build -o $S/probe-bin ./command/cohere   # compile error printed
    stat -f%z $S/probe-bin                      # unchanged, byte for byte
    $S/probe-bin --rules | wc -l                # 449, confident and stale

**And it composes with a second trap.** `$?` after a pipe is the exit of the last stage, not
the build, so `go build ... 2>&1 | head -2` reports success while the build failed. Redirect
to a file and read the code:

    rm -f "$S/cohere-$NAME"                                    # so a failure leaves nothing
    go build -o "$S/cohere-$NAME" ./command/cohere > "$S/build.log" 2>&1; echo "exit=$?"

The `rm -f` is the whole check, and it is the general repair rather than a rule of its own:
with nothing at the path, a failed build leaves nothing to read. Verifying a size or a
timestamp afterwards works too and depends on you remembering to look.

Three things must go wrong for a false reading: the pipe hides the failure, the artifact
survives, and the count looks right. All three happened in one command tonight, and the agent
nearly reported a clean finish from it.

**The other face of this is the binary the tree resolves.** The ahra tree reaches a linter
through `node_modules/.bin/`, a symlink to something somebody built earlier, and rebuilding
into your scratchpad does not update it. The gap between the two reads as a false negative in
your rule rather than as two binaries answering, because it moves by exactly the one or two
rules you just added. Two authors once found that symlink pointing at a binary inside their
own session's scratchpad which they had never built: the path looked like theirs, which is
exactly what disarms the check.

**The question is provenance, not proximity.** "Measure the thing" is necessary and not
sufficient, and the agent who found this corrected their own first statement of it: they did
read the count from a binary, which is a thing. It was the wrong thing and nothing in the
reading said so. The check that works is narrower: **is this artifact the one my last command
produced?** A size, a timestamp, a grep for the symbol you just added.

**And what makes this class dangerous is not that the wrong source is nearby, it is that it
answers**, which is the property the section above is built on. That is true of the stale
binary, of the symlink in `node_modules`, and of a `git grep` run in the wrong repository, so
they take the same repair as everything there.

## 11. The gate

    go build ./... && go test -count=1 ./...

Every package green. **Never pipe a build or test through `head` when you care about the
exit code**: a pipe reports the exit of the last command, not the first.

**Run the dry-run from the ahra tree, not the cohere tree.** Pointed at the wrong root
it emits a one-line failure that greps as silence — zero findings, which is the same
thing a clean rule produces. It cost one author a cycle twice. Confirm the run offered
files before reading its findings as a result.

**A zero-length measurement reads exactly like agreement.** An author nearly drew a
conclusion from two empty result files. Check that your measurement produced output at
all before comparing outputs.

**An oracle can be wrong in the PESSIMISTIC direction, and that is rarer and just as
misleading.** The usual failure is an oracle that reports nothing and reads as clean.
This one reported *fewer cases than it had* and read as a smaller corpus.

Measured on `dot-notation`. Five of its reporting cases are legacy octals -- `01['prop']`,
`08['prop']` and friends -- and driven through `@typescript-eslint/parser` all five come
back as FATAL PARSE ERRORS rather than verdicts, because a legacy octal is a syntax error
in TypeScript. The oracle said "linted 64 of 69" and the five silences looked exactly like
five clean cases. Driving the same corpus through espree in script mode moved it to 69 of
69, and the findings from 34 to 39.

The part worth copying is what came next. A fatal parse error in the ORACLE says nothing
about OUR parser, so the five were probed here directly: our parser reads all five and
reports on every one, agreeing with espree. The port was stronger than the oracle could
show, and taking the oracle's number would have recorded a limitation this rule does not
have.

So when an oracle reports fewer cases than the corpus contains, find out which ones and
why before treating the remainder as the measurement. `linted N of M` with `M > N` is a
result about the harness, not about the rule.

**And a rule can report zero because the substrate removed what it asks about.**
`unicode-bom` asks whether a file starts with a byte order mark. Every source file in a
real run is read through the vendored virtual filesystem, and that read strips a leading
mark before any rule sees the text; the vendored compiler's own test asserts exactly
that. So the rule was offered every file, registered a listener on every file, and
reported zero, and a seeded tree holding a genuinely marked file reported zero too.

Its author probed with two controls — a file whose first bytes on disk are `ef bb bf`
arrives three bytes shorter, an unmarked control of the same length is unchanged, and a
mark written mid-file survives — which localizes the removal to position zero, the only
position the rule asks about. Registered rather than abandoned, because the port is
correct against upstream and the missing piece is one line elsewhere.

The general shape: **when a rule reports zero on a tree, ask whether it could report
anything.** Seed the violation and re-run. A rule that stays silent on a seeded tree is
not clean, it is blind, and a fixture suite that never touches the filesystem cannot
tell you which.

**And an ORACLE that batches cases into one program lies about any rule comparing types
by identity.** An author building a typed replay harness put many upstream cases in a
single program. Several case files each declared `type Type`, the checker interned one
per file, reference identity failed, and **two of upstream's own failing cases came back
clean**. It reads exactly like a rule that cannot see object types.

One file per program reproduced all twenty. `RunTyped` is naturally one-file-per-call,
so fixtures are safe — this bites the *oracle you build to check them*, which is the
instrument you trust most and scrutinize least. If your rule asks the checker whether
two types are the same object rather than the same shape, batch nothing.

### The six kinds of zero

They are collected here because the class is the point, not any one instance.

1. **A run that never ran.** An author grepped a `go test` log for failures, got zero,
   and reported the corpus green. The log was fourteen lines of a *sibling's* compile
   error — the package never built, so no test in it ever executed. Zero failures, zero
   tests, and the grep cannot tell those apart. Check that the log ends in `ok` or
   `FAIL` for your package before you read a count out of it.

2. **A probe that only logs.** An author's discrimination probe ran 23 cases, printed
   `AGREE` or `*** DISAGREE ***` per case with `t.Logf`, and never called `t.Errorf` —
   so `go test` said `ok` while the log carried **seven disagreements**, five of them
   false negatives on the most common shapes the rule had to catch. The author read the
   green and reported the discriminator as working. An instrument that cannot report
   failure is not measuring.

3. **`go test -run <wrong name>`.** A proven fixer was nearly filed as unproven: mutated,
   run with `-run SingleLineJsdoc`, zero failures. The tests are named `...JsDoc...`,
   capital D. The pattern matched nothing, and Go reports that as **`ok` on the same
   line as `[no tests to run]`**. Grepping for `FAIL` finds nothing, which reads exactly
   like a passing mutation. **Run the BASELINE through the same `-run` pattern first.**
   If the unmutated code also reports `ok` with no test count, your pattern is wrong.

4. **A truncated read.** Reading through `tail -2` cut the first line off a two-line
   refusal message and left only the continuation, which read as a different verdict
   entirely. One coordinator concluded three times that a guard was not firing while it
   fired every time. `head -5` on a short message costs nothing; a truncated read costs
   the conclusion.

5. **A grep pattern that matches nothing**, or a mutant that would not compile, or a lint
   run from the wrong directory. All read as clean.

6. **A hanging fixture.** The five above all produce output you can read. This one
   produces none. A test that never returns emits no failure line and no summary, so
   `grep -c FAIL` reads 0 and a run that hung looks exactly like a run that passed. Found
   on `RegExp('\b')`: `SkipPatternEscape` returns a LENGTH while `ClassEnd` returns an
   INDEX. Conflating them spun the scanner forever. **Check the units on shelf helpers
   that answer the same shape of question:** length and index are both "a number about a
   position" and neither name says which.

### Make the instrument refuse, rather than remembering to check

A run that measured NOTHING looks identical to a run that found nothing wrong. In every
one of the cases above the fix is not "remember to check" — the coordinator did
remember, three times, and slipped the fourth. The fix is to make the tool refuse.

A lint wrapper that ends with:

    if ! grep -qE "^lint:" "$LOG"; then
      echo "REFUSING TO REPORT: no summary line, so this run linted nothing"
      exit 1
    fi

Then MUTATE THE WRAPPER to prove the refusal fires: point it at a tree with nothing to
lint and confirm it exits non-zero. An unproven guard is the thing it guards against.

The general form, for any measurement script you write:

    find the line that only appears when real work happened
    refuse to print a result if it is missing
    break it once and watch the refusal fire

A number with no evidence behind it should be impossible to obtain, not merely
discouraged. `score_one_mutation.sh` is built on exactly this principle and its header
comment is worth reading as an example of the shape.

### Two guards that catch what your own tests cannot

**`TestEveryRuleShipsAFixturePair` matches the CALL NAME, not the strength of your
assertion.** It walks for `ExpectFindings` specifically. A test that asserts counts,
spans and rendered message text but never names that helper reads to the guard as a rule
nothing proves can fire — one author shipped a rule that way and the guard caught it a
commit later. Assert whatever else you like, but call `ExpectFindings` at least once.

**Scratch packages are not free.** Building a one-rule isolation package to test against
is good practice, and every one of them must be gone before you commit. A scratch copy
that registers a real rule name breaks `TestEveryRuleIsRegistered` for every other agent
in the tree, and the failure names a package they have never heard of and cannot find in
their own diff. `TestNoScratchPackagesRemain` and `TestNoScratchFilesRemain` catch the
leftovers. Delete file by file; `-r` is blocked at the harness.

Before you commit, confirm both halves:

    grep '"<rule-name>"' /Users/kirkouimet/Projects/ahra/CohereSettings.json   # returns a hit
    go test -count=1 ./internal/lint/registry/                                 # green

Run `EnableRule.ts` from `/Users/kirkouimet/Projects/ahra` rather than hand-editing the
JSON: it writes all three surfaces, routes frontend rules to the structure config and
universal ones to nexus, and translates `typescript/` to `@typescript-eslint/` so the two
spellings do not drift.

**After you wire a rule, confirm the line is still in the config.** `EnableRule.ts` wires
a rule by read / parse / mutate / write. Whole file, no lock, no compare-and-swap. With
several agents running, two near-simultaneous writes could in principle drop one rule's
line.

Stated carefully, because a night watch got this wrong once: this race is a real property
of the code, but it has **not** been observed firing. It was diagnosed from a config line
that appeared at one checkpoint and was gone at the next, which turned out to be an agent
DELIBERATELY removing its own wiring and recording the rule as intentionally-off. Two
observations plus a plausible mechanism is not evidence. Before concluding your wiring
was clobbered, read whether somebody made a decision.

The check is worth doing either way, and it is cheap. Do it as the last step before you
report done. If the line is missing and you did not remove it yourself, say so rather
than restoring it silently, so the cause gets established rather than papered over.

## 12. Commit by pathspec, and stage and commit in one motion

Name your files. `git add -A` in a tree where several agents are working takes work
that is not yours.

**The pathspec on `git commit` limits which files are committed, not which hunks.**
`git add <path>` on a file another agent has edited stages their in-progress work along
with yours, and committing with a pathspec does not undo that. On a shared file, split
the diff by hunk and apply only yours:

    git diff -- <shared file> > /tmp/f.patch     # then keep your hunk only
    git apply --cached /tmp/f.patch
    git diff --cached --name-only                # read this before committing

**And staging is itself shared state.** Anything you stage is exposed to every other
agent's commit until you commit it, so `git diff --cached` verifies the past rather than
the future. Both of one night's mismatches happened in that window, in opposite
directions: one agent staged cleanly, checked, and had their work taken in the gap before
their own commit; another swept a section in without noticing. **Stage and commit in one
motion.** It is the only form that closes the window.

`PortingARule.md` is the one file here that several agents edit in the same session.
Treat it as the special case it is.

**Write the commit message from the staged diff, not from your notes**, and run
`git show --stat` afterwards. An agent's report is a statement about their working tree
when they wrote it; your commit is a statement about the index when you ran it, and those
drift. If a commit did take someone else's work, record it rather than rebasing: the
content is correct and in the tree, and attribution is not worth rewriting history for.

**Put scratch outside the rules tree from the first minute.** A probe package under
`internal/` compiles with everything else and breaks the build for every other agent. Use
your own scratchpad directory, never a shared name, and remove probes before reporting
unless one is a standing guard worth keeping.

## 13. Report back

Say what you ported, what you measured, and which of the two you are claiming. Name the
binary you drove, whether your control fired, and any divergence you recorded with the
command that established it.

---

## Facts about this substrate that cost somebody an hour

### Proving a zero: count the subject matter, and use the parser to do it

A rule reporting zero findings is either correct or inert, and the number cannot tell you
which. Three levels of proof, in increasing strength.

**Weakest: a seeded file.** Hand the rule a case you built to violate it and confirm it
fires. This proves the rule is not dead, and nothing more. Necessary, not sufficient.

**Better: a differential zero.** Run the reference implementation over the same files and
confirm it also reports zero. Now two implementations agree, though both could be missing
the same thing.

**Strongest, and the bar to aim for: count the subject matter.** Parse the tree and count
the constructs the rule EXISTS to judge. `unified-signatures` did this and found 11 real
overload sets across 10 files, one of them 96 signatures deep, then spot-checked two to
explain why each correctly declined — one pair differing in RETURN type, which fails the
unification gate before parameters are compared.

That converts "found nothing" into "examined eleven real candidates and correctly
declined all of them", which is a different and much stronger claim.

**Do the count with the PARSER, not with grep.** The author tried grep first and both
attempts failed, in opposite directions:

    first pattern    returned zero across the whole tree
    second pattern   returned hundreds of false hits

The first is the dangerous one, and it is the reason to always seed a control. **An
unconstrained pattern that matches nothing is a broken pattern, not a result** — it looks
exactly like a clean tree. They caught it only because the seeded file also came back
empty. The second failed because `condition ? a : b;` has precisely the shape of a
bodiless signature to a regex.

So: count with the parser, and control the counter against a seeded file, or you have
replaced one unproven zero with another one wearing a lab coat.

**Then write it where it is durable, not in a probe.** The probe directory gets deleted.
The rule's own doc comment and its test survive. Anyone reading that rule later should see
what was searched and what was found.

### When two agents need the same collector, scope and STATE the boundary

Two rules in one wave both needed "what propTypes does this component declare". The author
noticed before writing, and the resolution is worth reusing.

The wrong answers first. Writing it twice gives two rules two chances to disagree about one
question. And serialising the agents so one lands the collector first costs more than the
duplicate, while lifting ~600 lines of shared machinery inside a rule diff is the
unreviewable packaging warned about above.

**The right answer: scope to what is reachable now, land it, and STATE the boundary in the
commit.** Name which shapes you cover, which you decline, and where the shared collector
belongs when someone lifts it. A stated boundary is a handoff. A silent one is a trap for
the next reader.

Three things that made this concrete rather than vague, and they generalise:

**Check the package first.** `internal/lint/rules/react/` already carried
`componentPropsTypeNode`, `collectTypeMembers` and `resolvedTypeDeclarationBody`.
Same-package sharing is the pattern, not a workaround.

**Check whether the lifting target already exists.** `internal/lint/ecmascript/react/` was
already there with `component.go` and `names.go`, carrying nothing about propTypes. Naming
an existing empty home in the commit beats "someone should extract this".

**Check whether the overlap is total or partial.** It was partial: both rules need
declarations, but one needs defaultProps and the other needs a 578-line usage pass nothing
in the tree has. Partial overlap means both agents still have real independent work, which
is what makes scoping viable instead of merely polite.

### Reusing a sibling rule's helpers is encouraged; importing another rule PACKAGE is not

The guard forbids a rule package importing another rule package. That says nothing about
reusing helpers from a sibling file INSIDE your own package, and the difference matters.

`react/no-unstable-nested-components` was made tractable by exactly this.
`no_multi_comp.go` already carries over a thousand lines reproducing upstream's
`Components.js`, and the new rule calls straight into it:

    collectDetectedComponents   isComponentClass       returnsJsx
    createElementCallCounts     isCreateReactClassCall semanticParentOf
    skipParenthesesOptional     parentComponentOf  (from no_this_in_sfc.go)

Eight helpers, no new package, guards green. Re-deriving that component detection would have
given two rules two chances to disagree about one question.

**So the decision tree is:**

    helper in a sibling file, same package   just call it
    helper in ANOTHER rule package           lift the judgment into
                                             internal/lint/ecmascript/..., and have BOTH
                                             rules call it, as classmembers did
    logic that is genuinely one rule's       keep it in your file

Before writing a helper, grep your own package. In `internal/lint/rules/react/` in
particular, a great deal of upstream's shared machinery is already ported.

### Two bugs, not nine: read the failure PATTERN before debugging cases

Nine fixtures failed, all under-reporting, clustered 12-15 then 23, 27, 30-32. That shape —
a contiguous run plus scattered ones — was two bugs, because upstream orders its corpus by
shape and a contiguous block is usually one cause.

Both were the same KIND of error:

    isInsideRenderMethod                read as "anywhere inside a render method"
                                        upstream tests the node's DIRECT parent
    isPropertyOfObjectExpressionMatcher read as a grandparent test, on the strength
                                        of its NAME; it is `node.parent`

The author had written a confident doc comment asserting the intuitive reading was wrong.
The comment was what was wrong.

**A function's name is not its contract.** Both were settled by evaluating upstream's own
condition at every validated node rather than by reading the identifier. When a cluster of
fixtures under-reports, suspect one predicate read too broadly before suspecting many
separate bugs.

### A nil-receiver method call can be correct, so check the method before calling it a bug

A latent panic was nearly filed against `subject.Alias().Symbol()`, because two other call
sites in the same file guard `alias == nil` first and this one does not.

`TypeAlias.Symbol()` opens with `if a == nil { return nil }`. In Go a method with a
nil-receiver guard is safe to call through a nil pointer, so the unguarded version is correct
AND shorter — it delegates the nil case to the method that already handles it.

Two lessons. Read the callee before calling a missing guard a bug. And when your code is
correct but looks wrong beside its neighbours, leave a one-line comment saying why, or the
next reader will "fix" it.

### Type.Alias, not Type.AliasSymbol

`AliasSymbol()` does not exist on `Type` in the shim or upstream. The shim's only
`AliasSymbol` spelling is `AliasSymbolLinks`, an unrelated link store, which makes the wrong
name look plausible when you grep for it.

The real chain, and the shim already exposes it so no extension is needed:

    func (t *Type) Alias() *TypeAlias
    func (a *TypeAlias) Symbol() *ast.Symbol
    func Type_alias(v *checker.Type) *checker.TypeAlias

    if alias := checker.Type_alias(subject); alias != nil {
        symbol = alias.Symbol()
    }

Guard the nil: `Type_alias` returns nil for every type that is not an alias, which is most of
them.

### NeedsTypeChecker does not save you from a nil checker

`NeedsTypeChecker: true` governs the REGISTRATION path: the engine will not hand your rule a
file without a checker. It says nothing about the HARNESS path, where a `rule.Context` can be
built by hand — and `TestNoRegisteredRuleCrashesOnAbsentOptionalNodes` does exactly that.

A rule declared it correctly and still panicked with `resolveEntityName(0x0, ...)`, a nil
receiver, because the declaration felt like the guard. It is not. Write both:

    Run: func(ctx rule.Context, options any) rule.Listeners {
        if ctx.TypeChecker == nil {
            return nil
        }

Put it in `Run`, not at the top of each listener: it declines the file once instead of once
per node. `internal/lint/rules/typescript/await_thenable.go` and
`internal/lint/rules/typescript/consistent_generic_constructors.go` are the precedent, and
`await_thenable`'s comment carries the reasoning.

**Pin it with a test in your own package**, covering both the declaration and the decline.
`await_thenable` does this "so a later revert fails loudly rather than going vacuously
green." A guard nobody tests is one refactor away from being deleted as dead code, and the
panic returns with nothing left to catch it.

**And this is the concrete reason a crash outranks a wrong finding.** The walk recovers per
FILE, not per rule. One nil dereference costs every rule on that file. A wrong finding is
visible and arguable; a panic silently deletes other rules' coverage and looks like clean
output.

### A scope rule sees an EMPTY symbol table under Run and a populated one under RunTyped

The binder populates a per-scope symbol table reachable through `ast.IsLocalsContainer` plus
`ast.GetLocals`. If you are porting a rule that upstream writes against eslint-scope
(`findVariable`, `scope.upper`, `scope.variables`), you very likely do NOT need to build a
scope graph. The binder already did it.

But the table is only populated when the binder has run, and the binder runs with the
PROGRAM. Under `testing.Run` the locals table is EMPTY. Under `testing.RunTyped` it is
populated. An author who probes with `Run`, sees nothing, and concludes the substrate is
missing has measured an unbound tree, not an absent scope graph. That is a confident zero in
a new costume, and it is the trap in minute five of any scope-shaped port.

Consequence for registration: such a rule needs `NeedsTypeChecker: true` **for binding**, not
because it asks any type question. Declaring it false hands you a nil checker, and the usual
nil-guard then silently takes the wrong branch on every file.

**`ast.GetLocals` PANICS on a node that is not a locals container.** It does not return nil.
An enum declaration reaches it. Guard with `ast.IsLocalsContainer` first, every time. Same
hazard class as `ast.SkipParentheses`.

Two shapes that a walk written from upstream's structure misses silently:

- An enum declaration is NOT a locals container. Its members live on the enum symbol's
  `Exports` table, so an enum-shadowing check is portable only through a different accessor
  than everything else in the rule.
- A function expression IS a locals container, but its own name is not in it. Upstream's
  `functionExpressionName` scope has no equivalent here, so the `var a = function a() {}`
  exemption has to come from AST shape rather than from the scope table.

Better than upstream, once you are there: eslint-scope carries `isValueVariable` /
`isTypeVariable` as a boolean pair, while the symbol here carries real `SymbolFlags`
(`0x2` BlockScopedVariable, `0x40000` TypeParameter, `0x80000` TypeAlias, `0x200000` Alias).
All seven of upstream's DefinitionType variants map onto them.

### Some corpora are vendored in-repo, not under /tmp/lint-sources

The standing instruction is to import fixtures verbatim from `/tmp/lint-sources`. For React
Compiler rules that instruction points at the wrong place, and a provenance check written
against it returns a confident zero on a rule whose fixtures are impeccable.

`react/preserve-manual-memoization` draws from a vendored corpus at
`internal/lint/rules/react/conformance/testdata/fixtures/`. Grepping its test file for
`/tmp/lint-sources` or `babel-plugin-react-compiler` finds nothing, which reads as invented
fixtures and is not.

Two things make the vendored path the stronger option, and they are worth copying:

- The test file records the upstream path each fixture came from, per case.
- The conformance package DIFFS the copies against the originals, byte for byte, per file
  (`TestVendoredCorpusPairsAreComplete`, `TestCorpusIsTheSizeItClaimsToBe`, and the
  attribution tests beside them). The copy cannot silently drift from its source, which is
  more than a one-time verbatim paste guarantees.

So before concluding a rule invented its fixtures, find out where its corpus actually lives.
Check for a self-diffing test first; a rule that ships one has better provenance than a rule
that merely claims a verbatim import.

### A core rule and its typescript-eslint extension can both be on, and both will fire

Several rules in the live config are enabled twice: as the bare ESLint core rule AND as
`@typescript-eslint/<same name>`. Seeded probes on two of them emit two findings at the
identical line and column.

Two ways this bites, both of which look like a bug in your own rule:

**Your findings count looks doubled.** You port the core rule, measure it on the tree, and
the number reads as though the rule fires twice per site. It does, because the extension is
separately enabled. Check `CohereSettings.json` for `@typescript-eslint/` plus your rule's
bare name before you write the count down, and say so in your report.

**The temptation is to narrow your rule so it stops overlapping.** Do not. Your rule is
measured against upstream's corpus; narrowing it to dodge a config overlap makes it disagree
with the corpus, which is a real defect traded for a cosmetic one. Port faithfully and let
the duplicate be a config question.

Turning either rule off is Kirk's decision, not an author's, and not a night-watch one.
Report the overlap; do not resolve it.

### A flaky test is worse than no test

A mutation on map iteration order was caught by a two-group fixture only about 8% of the
time — measured, 22 reversals in 200 runs — because Go's randomisation happens to favour
insertion order when there are only two keys.

The author did the right thing: put the determinism in the RULE, and wrote at the line
that the fixture is not the guard. A test that fails 8% of the time reads as green, gets
trusted, and then blames something unrelated on the run where it does fail.

If a control only fires probabilistically, it is not a control. Fix the code and say so
at the line.

### "auto-fixable: yes" in an audit can mean one arm out of nine

Same rule. An audit said auto-fixable and a coordinator repeated that in a dispatch.
Measured against the installed build, exactly ONE of nine arms carries an applicable fix:

    !!foo             fix (and only when Boolean resolves to the global)
    +foo              suggestion only
    1 * foo           suggestion only
    foo - 0           suggestion only
    '' + foo          suggestion only
    foo + ''          suggestion only
    foo += ''         suggestion only
    -(-foo)           suggestion only
    ~foo.indexOf(x)   neither, reports bare

The corpus confirms it: 4 top-level non-null outputs, all `Boolean(...)`, against 44
explicit `output: null`. Counting `output:` lines flatly gives 43 and is WRONG, because
outputs nested inside a `suggestions:` array look identical to a grep. Match on
indentation or parse it.

**This changes what "prove the fixer" means.** `ExpectFixedSource` can only cover the 4 fix
cases. The other 44 need assertions reaching into `Diagnostics[].Suggestions`, the way
`internal/lint/rules/core/array_callback_return_test.go` already does, and the two surfaces
must be mutated SEPARATELY — a fix mutation cannot see a suggestion defect.

### When the clone beats the installed build: a same-tag strict superset

The usual guidance is that the INSTALLED build is the oracle, because the clone runs ahead
and porting to unreleased behaviour makes us disagree with the tool we are replacing.
`react/no-unknown-property` is a measured exception, and the conditions are narrow enough
to state exactly.

The author first read the data tables from the installed build and ended up carrying four
corpus cases at a verdict that contradicted upstream's own expectation. They reversed it:

    both artifacts declare 7.37.5      the clone is unreleased commits on the SAME
                                       tag, not a newer release we lag
    onScrollEnd, onScrollEndCapture,   present in the clone's rule file, ABSENT
    closedby                           from the installed one
    onLoad                             widened to include body
    removals                           NONE. Diffing both attribute tables in
                                       both directions returns nothing missing
    repairs                            none of the four carries a fix

So the clone is a strict superset: four additions, no removals, no behaviour reversals, and
nothing that rewrites source.

**That combination is what makes it safe.** A superset of a permit-list only ever makes the
rule quieter, so the worst case is a missed finding rather than a false one, and with no
repair attached there is no way for the difference to rewrite somebody's code. Reverse any
of those conditions and the installed build wins again.

**The general test, before you take the clone over the release:** same version string on
both, additions only in both directions of the diff, and no fixer on the differing entries.
Check all three. Two out of three is not the exception.
