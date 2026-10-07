# Gap descriptions and color resolution

Separate .a files implement borderSideDescription(), maskStopDescription() and resolveArmColor(value, arm, resolve). These helpers compose the already-proven batch30 colorArm/widthArm constructors. model.a defines the complete description result and the explicit external ThemeResolver dependency; UtilityArm comes from batch30/model.a.

borderSideDescription returns the present default 1px, then a color arm using --border-color and --color, then a positive-integer width arm using --border-width with px and modifier refusal. maskStopDescription has no default; its color arm uses --background-color and --color, followed by a --spacing arm inferring number then percentage, allowing percentage passthrough, using SpacingMultiplier and refusing modifiers. Each invocation allocates fresh descriptions and arrays. All other Go fields retain their zero values. Presence flags distinguish Go nil maps/slices/functions from present empty values. Empty placeholders under false presence flags must not be treated as usable dependencies.

resolveArmColor reads the supplied candidate value, the only candidate field the original helper reads. Exact inherit and transparent pass through; current becomes currentcolor. Every other spelling calls the external resolver with the exact value, candidate-present true, the unchanged arm.themeKeys array and options zero. It preserves both returned text and found flag, including empty-but-found. It ignores modifiers, arm color/refusal flags and inference fields, just as the Go helper does; those belong to its caller's dispatch. A real ThemeResolver must implement Theme.Resolve. The owned test driver supplies actual Go dependency observations, not an inferred parser or resolver. The valid contract corresponds to Go's non-nil candidate with non-nil Value; nil-pointer panic behavior is outside it.

The oracle exports unchanged private Go functions through an overlay. Each mode captures every string literal from every inventory consumer test file. Description functions take no parameters: captured strings serve as mutation probes for returned-array freshness, not as supposed constructor inputs. Color resolution uses those strings as values against eight actual Go theme configurations: nil/empty namespaces, absent entries, inline/empty-inline/reference/ordinary values, prefixing, namespace precedence and duplicates. Controls include keyword overrides, near spellings, Unicode, NUL and newline. These are helper projections, not whole-rule findings or CSS emission tests.

With the setup environment sourced:

```
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch31 -count=1 -v -timeout=20m > /tmp/lint05-batch31-final-helpers.log 2>&1
```

Set ADAMIC_SLOT05_BATCH31_EVIDENCE to an existing directory to retain generated raw cases, actual Go verdicts and coverage. Published observations are compressed losslessly. See REPORT.md for counts, mutants and limits.
