# Helpers from lint wave 1 slot 05

Branch: codex/lint-helpers-from-lint-wave1-05. Base: origin/codex/lint-helpers.
Prior rule work is pushed through 8815a3f9; its shared integration blockers remain explicit in the rule reports.

1. github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.leadingInteger
   File: collapse_leading_integer.a. Six frozen remaining consumers; zero final blockers alone.
   All 389 origin refs and five distinct helper claim blobs inspected after explicit all-heads fetch. Comments are already delivered/reserved by HELPERS.md; all higher consumer-count concrete symbols are reserved. This helper ties the highest unclaimed count at six. Choose the standalone decimal-prefix parser, preserving Go machine-int overflow and ASCII scanning. Compare unchanged Go on inputs extracted from every consuming rule's tests and bounded controls, source Node, emitted JavaScript and sanitized native. Run a compiling semantic mutant. Claim pushed before code.

2. github.com/system-inc/cohere/internal/lint/rules/tailwind.FindEntryPoint
   File: tailwind_find_entry_point.a. Six remaining consumers, zero final blockers alone.
   leadingInteger is tested and pushed in c3b9b8af. All origin refs fetched again and all helper claim blobs checked. This search ties the highest unclaimed count at six. Preserve ordered POSIX filepath.Join probes and stop on first existing file; filesystem predicate is an explicit dependency. Actual Go consumer fixtures and all existence-mask/path controls decide behavior. Node, emitted JavaScript and sanitized native, with a compiling priority mutant. Claim pushed before code.

Withdrawn: FindEntryPoint. The all-origin refresh revealed an earlier claim on codex/lint-helpers-from-codex/lint-wave1-11. The selection assertion failed, but the shell continued to the claim commit. No implementation was written; earlier ownership wins.

2. github.com/system-inc/cohere/internal/lint/rules/tailwind.findTailwindPackageRoot
   File: tailwind_find_package_root.a. Six remaining consumers, zero final blockers alone.
   All 405 origin refs and 16 distinct claim blobs checked. This upward search remains unclaimed at the highest count, six. Preserve POSIX Join/Dir, exact index.css probes including duplicate cleaned probes, first match and termination at root. Filesystem predicate is explicit. Actual Go over all six consumer fixtures and path/existence controls, source Node, emitted JavaScript and sanitized native, with a compiling semantic mutant. Claim pushed before code.

3. github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.CompareBreakpoints
   File: collapse_compare_breakpoints.a. Six remaining consumers, zero final blockers alone.
   Both prior helpers are tested and pushed (c3b9b8af, e998e268). All origin refs refreshed and all helper claim blobs checked again. This comparator remains unclaimed at the highest count, six. Preserve equality, byte-string bucket ordering, raw-string fallback, ascending/descending and Go signed-int subtraction overflow. Already-owned breakpointBucket is an explicit callback dependency; validation obtains its facts from the actual Go helper. Compare every consumer fixture and pair controls on source Node, emitted JavaScript and sanitized native, with a compiling direction mutant. Claim pushed before code.

4. github.com/system-inc/cohere/internal/lint/rules/tailwind.DesignSystemForProgram
   Intended file: tailwind_design_system_for_program.a. Six remaining consumers.
   Third helper is tested and pushed in 3f8a48dc; entire helper package PASS in 72.676s. Refetched all 418 origin refs and inspected 17 distinct claim blobs. This program-keyed cache is unclaimed and ties the highest count at six. Claim precedes implementation or a concrete blocker report. Must preserve nil-program errors, identity rather than path keys, failed-result caching and synchronization across parallel callers; the full Program/filesystem/load-result contracts are prerequisites, not placeholders.

Withdrawn: DesignSystemForProgram. No cache implementation delivered. Program/recording-filesystem/load-result and mutex APIs are absent from this base; the raw probe fails TS2305. See ../wave05/CACHE_BLOCKER.md. This reservation is released so a worker with those prerequisites can take it.

Final active ownership: leadingInteger, findTailwindPackageRoot and CompareBreakpoints, all tested and pushed. FindEntryPoint and DesignSystemForProgram are withdrawn. No further helper is claimed.

## Parking continuation on current main

Rule branch is parked, rebased and independently green at 54cbb04b25d8d5553f5df56f8b976b9e48370acc on main f8013f0b, with missing shared registration/context/Diagnostic integration named in its owned PARKING_REPORT.md. Helper branch a71683ad4 is green on the same main. No main or area branch is pushed.

4. Reclaim github.com/system-inc/cohere/internal/lint/rules/tailwind.DesignSystemForProgram.
   Intended file: tailwind_design_system_for_program.a. Six remaining consumers, zero final blockers alone. All 532 origin refs and all 17 distinct helper claim blobs were inspected. The previous slot-05 reservation was explicitly released; no other exact symbol claim exists. loadDesignSystemForProgram is separately owned and is not this symbol. Delivered comment helpers are excluded. This released six-consumer helper ranks above the remaining unclaimed four-consumer control-flow helpers. Claim pushed before retrying code or prerequisites on current main.

The exact Go contract remains identity-keyed, synchronized build-once caching of successful and failed load results, with nil-program rejection and recorded filesystem reads. The older probe failed on missing Mutex; current-main support must be measured again before shipping code. Do not substitute a serial/path-keyed cache or mark consumers ready without that contract. The six consumers are enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes under better-tailwindcss.

Current-main retry outcome: blocked. Claim 821b11f19 preceded the probe; both native-build and emitted-JavaScript compilation exit 1 with TS2305 for missing Mutex. No cache implementation or consumer readiness is claimed. Reservation retained as blocked; see wave05/CACHE_BLOCKER.md. No next helper is claimed before this prerequisite is available.
