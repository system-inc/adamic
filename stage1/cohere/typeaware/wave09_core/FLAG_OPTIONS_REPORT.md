Built: threaded allowConstructorFlags through the owned handed-node frontend and upstream Go option decoder.
Commits: follows pushed 931d319f5 on area b46914832/current main c7991b900; no new claims.
Commands/output: default frontend, three option profiles and handed-node regression PASS, including sanitized native.
Mutants: clean-running ignored-options mutant differs from Go at byte 49; handed report-refetch remains caught at byte 58.
Not covered: native pattern compilation, shared JSON/registry option installation, arbitrary invalid config error text or a full gate.

The isolated driver accepts one --allow-flag= argument per
allowConstructorFlags array item and passes the array to NoInvalidRegExp.
The owned flag_options.a module rejects duplicate array items before loading
the checker. This transport does not install a shared field-5 JSON adapter.
Shared registry installation remains integration work; the options guard is
unchanged and no bypass was introduced. No shared files changed.

The independent Go oracle marshals the array into upstream
NoInvalidRegexpOptions and calls DecodeNoInvalidRegexpOptions, then passes
the decoded object to the unmodified production rule. It no longer always
passes nil. Default controls retain identical outputs. No Go rule or option
decoder was edited. Native flag judgments are the existing rule implementation;
this change completes their option input path without adding a regex matcher.

After sourcing /workspace/adamic-tools/env.sh:

```
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_frontend.py > /workspace/wave-09-flag-options-frontend.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_options.py > /workspace/wave-09-flag-options-test.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_handed.py > /workspace/wave-09-flag-options-handed.log 2>&1
```

The original 40 controls retain 35 findings/6020 serialized bytes in normal
and ASan/UBSan/LeakSanitizer executions. Sixteen new controls use nonliteral
patterns so they exercise flag options independently of the unavailable
pattern compiler. They cover joined strings, case sensitivity, intrinsic
flag duplicates, u/v exclusion, extra-flag duplicates, astral Unicode flags,
empty items and unknown characters. Across the three profiles, full finding,
fix and suggestion bytes match Go: default 13 findings/2216 bytes, joined
options 11/1998, case-sensitive options 13/2218. Native normal and sanitized
executions have empty stderr. Duplicate option items explicitly fail on both
sides (Go 2, native 70); diagnostic formatting for configuration failures is
not claimed byte-identical.

The new mutant discards the parsed allowed array in the driver. It compiles,
exits zero with empty stderr and differs only in the independently compared
findings at byte 49. This proves that comparing two implementations which
ignore options cannot satisfy this check. The handed-node verifier was
updated solely to resolve the new owned import in its temporary probe.
Its unrelated callback index zero retains all 6020 Go bytes, and its existing
report-refetch mutant still compiles/runs cleanly and differs at byte 58.

Joined-profile whole-process native/Go time: 0.114613/0.114365 seconds;
sanitized 0.114435 seconds. Case profile 0.066757/0.064080 seconds.
These single concurrent runs are not isolated throughput evidence. Successful
88-second setup was reused, nproc 5. Exact streams, logs, measurements and
archived executable mutant probes are in validation-flag-options.

Both regex claims remain incomplete because native dynamic RegExp lowering
is still unavailable. No new parser, matcher, bridge question or claim was
added. Existing corpus, released-handle, required parser, bridge and other
component checks retain LANDING_LEGACY_REPORT.md/LANDING_C799_REPORT.md
evidence; those unchanged checks were not repeated for this option path.
No full repository gate or unrelated required checks were run. Publication
is exclusively to codex/typeaware-wave-09.
