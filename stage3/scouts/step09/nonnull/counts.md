Counts measured with the checked non-null compiler at 93ebcf520c87eb55b47ada1d39189796959b2778.
The committed .a fixtures are refused. These rows describe their temporary .ts controls,
not an addition to the central oracle's registered fixture population.

| Fixture | Proven | Checked | Allocations | Frees | Retains | Releases | Peak | Regions | Exit |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 01_erased.a control | 1 | 0 | 2 | 2 | 0 | 2 | 2 | 0 | 0 |
| 02_checked_pass.a control | 0 | 1 | 5 | 5 | 3 | 8 | 3 | 0 | 0 |
| 03_checked_stop.a control | 0 | 1 | 2 | 1 | 1 | 1 | 2 | 0 | 70 |

The failing row terminates at the ruled panic before normal cleanup. Passing native
controls and the check-omission mutant run with ASan/UBSan and leak detection.
Intentional terminal panics use ASan/UBSan with leak detection disabled; their exact
stderr, location, expression, stdout and exit status are checked. Raw count output
and run commands are retained in observations.json.
