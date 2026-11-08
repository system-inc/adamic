# Rule G: admission delta and the sys.ts census with C-noinfer-b

Measured by Performance on Cloud with ahra loom proof, Nov 8 2026. Base is main e3aa0bb7. Head is compiler/rule-g 2594029d; its invariance.go is dda69fa1's.

## Admission delta
The delta ran base e3aa0bb7 against head 2594029d, with npm bootstrapped from cloud/markdown-width/npm-bootstrap.json. Result: changed programs 0, negative witnesses 50, ruled run-time stops 45. The job is delta.sh, and its log is admission-delta-job.log.

## The sys.ts census, with Correctness's C-noinfer-b applied
The census script is sys-noinfer.sh. It applies main's adaptations to the pinned upstream 050880ce, then applies the C-noinfer-b diff from typescript/sys-timer-slot 949b9c3c (stage3/census/sys-timer-slot/C-noinfer-b.diff.txt). Each replacement must match exactly once. Then it runs the census and the latent census on src/compiler/sys.ts, first-error and full.

| mode | base | head | gone at head | new at head |
| --- | ---: | ---: | --- | --- |
| first-error | 62 | 60 | sys.ts:1499:13, 1597:13 | none |
| full | 440 | 438 | sys.ts:1499:13, 1597:13 | none |

- The two gone are "a function taking (...args: TArgs) => void seen as one taking (...args: NoInfer<A>) => void". They're Correctness's 1501 and 1599, at 1499 and 1597 here because the diff deletes two lines above them.
- These stay at both base and head:
  - sys.ts:1499 "reading setTimeout" (NotYet, full mode);
  - clearTimeout at 1500 and 1598, the host row's.

Without C-noinfer-b, at main's own slot, base and head are identical: 61 first-error findings and 432 full.

Not run: tsc's whole entry census.
