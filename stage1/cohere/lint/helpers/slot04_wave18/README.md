# Background size and nested theme helpers

- `background_size.a` ports isBackgroundSize. Segmentation, length and percentage checks are explicit callbacks. Invalid one/two-part comma groups are skipped, while a three-part group rejects the whole value; at least one valid group is required.
- `resolve_with.a` ports Theme.ResolveWith. Key resolution, entry lookup and variable-reference construction are callbacks. It preserves nested key order, inline bit 1, absent nested entries, empty stored values and the final base value. `extraPresent` distinguishes Go's nil map on failure from its allocated empty map on success. Keys and values are retained in a Map.
- `theme_argument.a` ports resolveThemeArgument. A trailing -* invokes Resolve with present=true and options=0. Embedded -* invokes ResolveWith, and returns the last requested nested key, preserving an empty string when the key exists.

Each production file owns one helper. The nested argument callback can invoke this batch's ResolveWith; the differential driver does so. Other callbacks remain separately owned dependencies and must implement Go behavior. The private Go oracle supplies real segmentation, predicate, lookup and reference results, preserving theme prefix and options.

From the repository root with the setup environment sourced:

```
python3 stage1/cohere/lint/helpers/slot04_wave18/testdata/generate.py > /tmp/slot04-wave18-generate.log 2>&1
python3 stage1/cohere/lint/helpers/slot04_wave18/testdata/capture.py > /tmp/slot04-wave18-capture.log 2>&1
go test -count=1 -v ./stage1/cohere/lint/helpers/slot04_wave18 > /tmp/slot04-wave18-tests.log 2>&1
```

Go capture and exports use temporary overlays, without editing the shared harness or pinned cohere source. Complete rule findings and arbitrary dependency implementations are outside this comparison.
