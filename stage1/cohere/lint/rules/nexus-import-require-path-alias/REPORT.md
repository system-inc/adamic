Built: Direct-option path aliases, longest directory, strict roots, internal/repository exclusions, source quotes and portable compiler-path derivation. Registered by owned directory, using the authorized .ts fallback for rule modules.
Commits: claim 470b020e, implementation 7a484f451b6568c076b51aeee14f0677e6f572f1; earlier Tailwind source-token follow-up ca2a9d5a.
Commands and outputs: validate.py PASS (27 upstream cases), registered-driver PASS (74 combined cases, 21,677 bytes), raw_validate.py PASS (183 corpus observations), registry PASS and owned Go utility vet PASS.
Mutant: path_alias_longest_directory_reversed builds and runs cleanly, then only the byte comparison catches it on Node, emitted JavaScript and ASan/UBSan native.
Not covered: Ten unique compiler-derived full-rule fixtures, DecodeAt repository-root validation/config-base discovery, full gate and suppression integration. The pure derivation helper has separate parity; that does not cover project discovery.

## Observations

The unchanged Go rule is the oracle, pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. The corpus runner captures original upstream fixture sources and decoded options while running their original assertions. Source Node, emitted JavaScript and sanitized native match findings, messages, UTF-8 ranges and fixes byte for byte. The registered-driver probe runs all 74 supported fixtures through the actual generated registry, shared main driver and shared Go oracle, with filenames and explicit options preserved. No shared authored files were edited or merged.

Upstream independent-parser comparison passed on all three backends. The projected-AST corpus run passed independently, and raw_validate.py then passed the same complete captured corpora using raw sources and Adamic's own parser, without Go-projected AST input. TypeScript compiler pin is 050880ce59e30b356b686bd3144efe24f875ebc8, covering all 77 src/compiler files and 9,400,075 bytes. Stage1 is a census at capture time, not a claim that later reports or wrappers were in the earlier snapshot:

- compiler: 77 files, 9400075 bytes, no parse-diagnostic or gap exclusions.
- stage1: 161 files, 817531 bytes, no parse-diagnostic or gap exclusions.

Input names, raw-source/options hashes, output hashes, summary observations and upstream source rows are retained under evidence. Corpus observations cover each batch on all three backends. Fixtures and raw corpora had zero exits and empty stderr. Mutants were recertified after TS fallback and relocation for the import rules.

Startup-inclusive raw-source throughput, median of three runs on 200 copies of one positive fixture: native 736, Node 1,343, Go 18,884 findings/second. Native uses ASan/UBSan; these are small-fixture pipeline rates, not steady-state corpus speed claims. Projected-AST timing remains separately labeled in summary.json.

## Shared gaps and extension fallback

The current registry requires rule.ts and rejects .a mutant modules with "mutant file must be an owned TS module". The first .a descriptor attempt failed registry validation; the authorized .ts fallback made it pass. Supporting validation drivers remain .a where the standalone loader permits it. origin/codex/lint-harness-dot-a at f4d98cab was inspected, but no shared files were changed. Rule filenames use the existing context.parser.path.

The stock TestOwnedWitnesses gate fails before the new rules run: it stores the Google-font JSX witness as next-google-font-display-0.ts, then Go parsing panics with ["'>' expected.", "',' expected.", "Variable declaration expected.", "Expression expected.", "Expression expected."]. The exact output is retained in stock-witness.log. This failure is not a passing gate. The custom registered probe preserves fixture filenames and supplies explicit options, so it can exercise the supported new rules without changing the harness.

Compiler-derived aliases remain a full-rule shared-input gap: RuleContext supplies no Program/compiler options or GetPathsBasePath metadata. implementation.ts refuses aliasesFromTsconfigPaths when no provider is supplied with "NotYet: shared compiler paths provider"; it never guesses from a hardcoded tsconfig. The portable derivedAliases function takes explicit compiler-path metadata, and 127 helper cases, including 27 derivation cases, match the actual Go private helpers across all three backends. This is separate from full-rule project discovery. Ten unique derived-option fixture records are retained in excluded-derived.jsonl, rather than counted as parity. All original Go test assertions, including their firing controls, were run.

Go DecodeAt additionally needs the config directory and discovered project root, neither supplied by the shared context. fileStatus exists locally, but the prelude supplies no realpath/symlink-target operation needed for the Go pathsNest fallback. The oracle adapter consumes already decoded fixture options; config validation, exact decoder error text and root/symlink discovery are not ported. The stock witness harness also does not consume witness.options.json; this rule correctly stays silent with absent options, so a configured witness cannot pass that gate until it carries the options.

## Reproduction and toolchain

Source /workspace/adamic-tools/env.sh. Run this directory's validate.py --scratch /tmp/wave11-<rule>, then post_validate.py with the same scratch and --compiler /tmp/wave11-typescript. Redirect every command to a log. The naming directory owns raw_validate.py, benchmark_raw.py and registration_validate.py; import rules pass --source pointing to their validation.a for raw and benchmark probes. registration_validate.py uses the saved upstream corpora in /tmp/wave11-screaming, /tmp/wave11-forbidden and /tmp/wave11-alias. Alias components_validate.py has its own --scratch runner.

The existing installed cloud toolchain was reused. Go 1.27.1, clang 20.1.8 and Node 24.19.0; current nproc output is 5. Initial session setup timings were Go 0s, clang 0s, Node 0s, submodules 1s, warm 89s, done 89s. The earlier follow-up setup failure and workaround are documented in ../structure-tailwind-no-physical-direction/REPORT.md; no new successful setup timing is inferred here. Required full-repository gate was not run; scoped registry and owned Go utility vet passed, stock witness gate failed as stated.
