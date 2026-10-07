# Helpers from lint wave 1 slot 06

`new_theme.a` ports Go cohere's `collapse.NewTheme`. Every call returns a fresh record and values map, empty Prefix, nil keyOrder and zero deadKeys. The order view carries an explicit `nil` bit and a values list, because this base's stage 0 refuses `string[] | null`; `nil: true` represents Go's nil slice, distinct from an allocated empty slice with `nil: false`; this constructor does not append, reslice or resolve theme entries. ThemeValue retains the literal value and complete numeric options. Map entries and the record are mutable, matching the upstream store.

The constructor has no inputs. Tests invoke the actual Go constructor for every captured consumer fixture and exercise fresh-record and fresh-map independence by writing a source-derived key, prefix, order and dead counter into one instance while observing two later instances. Those writes are test controls, not implementations of Theme.Add or another worker's helper. The source is hex-encoded to make the observation keys printable without excluding NUL or Unicode.

The owned oracle overlay exports constructor state without changing cohere's worktree. Consumer capture is adapted from slot03's capture procedure, retains its runtime hooks before external-engine skips, and records separate rule-package failures. Captured sources prove fixture coverage, not that every Go helper call in each rule was instrumented. All consuming-rule findings parity remains the rule workers' responsibility.

```bash
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/from-wave1-06/testdata/regenerate.py > /tmp/helpers06-regenerate.log 2>&1
go test ./stage1/cohere/lint/helpers/from-wave1-06 -count=1 -v -timeout=20m > /tmp/helpers06-test.log 2>&1
```

The helper runs as the original .a source on Node, emitted JavaScript on Node and native with ASan/UBSan. Dirty initial Prefix, allocated empty order instead of nil, and shared values map are clean-executing byte-comparison mutants.
