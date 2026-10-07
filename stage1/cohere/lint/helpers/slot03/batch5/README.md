# Slot 03 fifth helper batch

Three Collapse helpers, one .a file each. splitThemeKey returns an empty list when a key lacks two leading hyphens; otherwise it splits the suffix at every hyphen, preserving empty segments. In particular, -- returns one empty segment, --- returns two and --a- preserves the trailing empty segment. Every call creates a fresh mutable list. Go's nil empty result is represented by an empty Adamic array; consumers range over these results.

joinSegments joins the ordered input strings with one hyphen, retaining empty segments, embedded hyphens, NUL and Unicode without normalization. Empty and singleton-empty inputs both produce an empty string.

breakpointGroupOrder accepts the actual generated registrations as an explicit dependency. Its adapter retains the exact name and order fields Go reads; other generated fields are irrelevant to this helper. It returns the first exact sm registration's order with found=true, including order zero; an absent sm produces 0/false. Nil and empty registration lists are represented identically. No order constant or replacement framework table is guessed. Number-valued orders require exact JavaScript-safe integers; the pinned generated table uses small integers.

The owned oracle overlay invokes actual private Go helpers and temporarily replaces/restores the real generated registrations inside its isolated process. Actual runtime source capture covers all six consuming rules and 157 inputs. Sources, extracted tokens, actual framework names, code points U+0000..U+00FF in several positions, Unicode, repeated/trailing separators, arbitrary ordered segment lists and duplicate/missing sm controls are compared on source Node, emitted JavaScript and native under ASan/UBSan. UTF-16 output encoding preserves supplementary Unicode observations. Temporary semantic mutants prove prefix guards, empty segments, fresh-list behavior, separators, exact name matching and first-match ordering.

```
python3 stage1/cohere/lint/helpers/slot03/batch5/testdata/regenerate.py > /tmp/slot03-batch5-capture.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch5' -count=1 -v -timeout=20m > /tmp/slot03-batch5-tests.log 2>&1
```

Source /workspace/adamic-tools/env.sh first. The external Tailwind live/corpus rule-package gate fails on unavailable configured installations/corpora and is recorded separately from passing helper comparisons. These helpers remove eighteen dependency edges across six rules; none loses its final helper blocker from this batch alone. See REPORT.md and readiness.json. Arbitrary invalid UTF-8 keys and non-safe-integer registration orders are outside the adapter contract. No shared harness, registration generator or compiler source changes.
