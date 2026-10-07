# Helpers from lint wave 1 slot 04

Original branch: codex/lint-wave1-04, all work pushed through 3aa2bc62.
Helper branch: codex/lint-helpers-from-codex/lint-wave1-04.
Base: origin/codex/lint-helpers at 95100eb440b47f3f18e960c6f5b49cadf0dc1d9d.

1. github.com/system-inc/cohere/internal/lint/rules/tailwind.projectRootOf
   File: helpers/from_wave1_04/project_root.a.
   Six remaining consumers, tied for the highest unclaimed concrete helper count:
   better-tailwindcss/enforce-canonical-classes,
   better-tailwindcss/enforce-consistent-class-order,
   better-tailwindcss/enforce-consistent-variant-order,
   better-tailwindcss/enforce-shorthand-classes,
   better-tailwindcss/no-conflicting-classes,
   better-tailwindcss/no-unknown-classes.

Every origin head was fetched and every helper claim blob inspected before selection.
The comment bundle is already delivered on the base and is excluded. FindEntryPoint
was claimed by slot 11 during the refresh and is not claimed here. All larger named
helpers are reserved. This symbol ties the highest remaining consumer count at six.
Preserve configured-file preference, Go filepath.Dir behavior on the POSIX target,
nil/empty compiler-options fallback and lazy current-directory reads. Program options
and directory reads are explicit dependencies, not a program engine implementation.
Compare actual pinned Go on inputs from every consuming rule plus path and callback
controls, source Node, emitted JavaScript and sanitized native. Require a compiling
semantic mutant per helper. One helper per .a file; only owned harness files change.
This claim is pushed before code. Re-fetch all origin claims before the next helper.

2. github.com/system-inc/cohere/internal/lint/rules/tailwind.loadDesignSystemForProgram
   File: helpers/from_wave1_04/load_design_system_for_program.a.
   Six consumers, the same six Tailwind rules listed above.
   Selection: 428 origin refs and all 17 distinct helper claim blobs inspected.
   Tied highest unclaimed named helper count: six.
   Preserve one program filesystem acquisition, load before reads, snapshots on
   failure, every opaque result field and Go result value semantics. The loader
   and RecordingFS are explicit dependencies, owned separately. Compare the actual
   Go function on six-consumer inputs and dependency controls on all three targets.
   This claim is pushed before code; projectRootOf was pushed as 482d36bf first.

Delivery status: both claimed helpers are implemented, tested and pushed.
projectRootOf: 482d36bf; loadDesignSystemForProgram: bb0fd01e.
Five compiling semantic mutants are caught across Node, emitted JavaScript and
sanitized native execution by the actual Go comparison. See the owned REPORT.md
and evidence/both-tests.log for the exact dependency boundaries and residual
rule prerequisites. No claim of whole-rule/native CSS engine completion is made.

3. github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.trimLeadingJavaScriptSpace
   File: helpers/from_wave1_04/trim_leading_javascript_space.a.
   Four consumers, tied for the highest remaining unclaimed named helper count:
   better-tailwindcss/enforce-consistent-class-order,
   better-tailwindcss/enforce-shorthand-classes,
   better-tailwindcss/no-conflicting-classes,
   better-tailwindcss/no-unknown-classes.
   Selection: 550 fetched origin refs and all 19 distinct helper claim blobs
   inspected, plus the shared comment claim in HELPERS.md. Larger named helpers
   are delivered or reserved. Prior rule branch is rebased and parked with green
   owned comparisons, pushed as ebdc5a5c; helper branch beeb6533 is rebased, green
   and pushed on main f8013f0. No main or area branch is pushed.
   Preserve the leading run only, unchanged remainder, complete-string trimming,
   empty input and Unicode rune behavior. The separately claimed JavaScript-space
   predicate is an explicit dependency. Compare actual Go on fixture literal
   domains from all four consumers plus Unicode and boundary controls, source
   Node, emitted JavaScript and sanitized native; require compiling mutants.
   One helper per .a file. This claim is pushed before implementation.

Delivery 3: trimLeadingJavaScriptSpace is implemented and validated. All 38 original
Go consumer tests pass; 63,845 helper cases yield 883,340 identical bytes across
source Node, emitted JavaScript and sanitized native. Both compiling mutants are
caught. Complete owned helper gate PASS 26.695s, all seven semantic mutants;
vet and six uncached input probes pass. See from_wave1_04/TRIM_REPORT.md.

4. github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.scanNumber
   File: helpers/from_wave1_04/scan_number.a.
   Four consumers, the same class-order, shorthand, conflicting and unknown
   Tailwind rules listed for trimLeadingJavaScriptSpace. Highest unclaimed
   named count after fetching all 573 origin refs and reading 20 distinct
   helper claim blobs plus HELPERS.md. Both owned branches are landing-ready
   on current main c01907a7: rules 20e03b57 are parked and fully re-greened;
   helpers c023b39c are rebased, fully re-greened and pushed.
   Preserve Go's anchored grammar, required fractional digits, optional sign,
   all-or-nothing exponent, ASCII digit membership and consumed byte count.
   Successful prefixes are ASCII so their UTF-16 and byte counts coincide.
   The separately claimed digit predicate is an explicit dependency. Compare
   actual Go on all four consumer fixture domains plus bounded exhaustive
   grammar and Unicode controls, on source Node, emitted JavaScript and
   sanitized native. Require compiling semantic mutants. Claim pushed before code.

Delivery 4: scanNumber is implemented and tested. 101,547 cases, 802,813 exact
Go bytes and both compiling scanner mutants pass on source Node, emitted
JavaScript and ASan/UBSan native. Complete four-helper gate PASS 35.578s with
all nine semantic mutants, vet PASS. See from_wave1_04/SCAN_REPORT.md.

5. github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.numberWithSuffix
   File: helpers/from_wave1_04/number_with_suffix.a.
   Four consumers: class-order, shorthand, conflicting and unknown Tailwind rules.
   Highest unclaimed named count after fetching 582 origin refs and reading every
   one of twenty distinct helper claims plus HELPERS.md. Main is unchanged at
   c01907a7; rules 20e03b57 are parked and green, helpers 05ec8b0b are green and
   pushed, including the scanner this helper consumes.
   Preserve one scanner call, no match when consumption is zero, exact whole
   remaining-string membership, empty/nil suffix-list behavior and empty suffixes.
   The scanNumber callback is an explicit dependency, implemented by the held
   scanner during comparisons. Compare actual Go over all four consumers' fixture
   domains, grammar/Unicode boundaries and suffix sets on source Node, emitted
   JavaScript and sanitized native; require compiling semantic mutants.
   One helper per .a file. Push this claim before code.

Delivery 5: numberWithSuffix is implemented and tested. 1,218,564 cases,
27,938,475 exact Go result/scan-count/argument bytes and all three compiling
mutants pass on source Node, emitted JavaScript and ASan/UBSan native. Complete
five-helper gate PASS 91.597s, all twelve semantic mutants; vet PASS.
See from_wave1_04/SUFFIX_REPORT.md. No further helper is claimed yet.
