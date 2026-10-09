# Counts changes

A/F/R/L/P/G = allocations, frees, retains, releases, peak live, values in regions. Compared with assigned base `4885cec50290686df487b62aac47c85d871ed40c`. Every existing changed numeric row was included in the focused differential oracle selector. Allocations, frees, peak and regions are unchanged for those rows. The TDZ fixture stops before cleanup; its counters are not a leak-free completion claim.

| Fixture | Before | After | Cause |
| --- | --- | --- | --- |
| internal/oracle/testdata/census_optional_values_forms.a | 102/102/52/169/23/0 | 102/102/51/168/23/0 | Receiver-independent literal methods use ordinary closure arguments; receiver retain/release pairs disappear. |
| internal/oracle/testdata/census_optional_values_padding.a | 42/42/73/108/12/0 | 42/42/69/104/12/0 | Receiver-independent literal methods use ordinary closure arguments; receiver retain/release pairs disappear. |
| internal/oracle/testdata/census_optional_values_system.a | 29/29/67/91/14/0 | 29/29/52/76/14/0 | Receiver-independent literal methods use ordinary closure arguments; receiver retain/release pairs disappear. |
| internal/oracle/testdata/census_unary_numeric.a | 187/187/178/349/14/0 | 187/187/169/340/14/0 | Receiver-independent literal methods use ordinary closure arguments; receiver retain/release pairs disappear. |
| internal/oracle/testdata/closure_convention_receiver_rest.a | 11/11/12/23/7/0 | 11/11/9/20/7/0 | Receiver-independent literal methods use ordinary closure arguments; receiver retain/release pairs disappear. |
| internal/oracle/testdata/hidden_wave2_method_destructuring.a | new | 26/26/23/48/10/0 | New differential fixture: identity, callbacks, captures, evaluation order and snapshot lifetime. |
| internal/oracle/testdata/host_never_branches.a | 50/50/57/92/5/0 | 50/50/56/91/5/0 | Receiver-independent literal methods use ordinary closure arguments; receiver retain/release pairs disappear. |
| internal/oracle/testdata/host_void_method.a | 5/5/14/18/5/0 | 5/5/12/16/5/0 | Receiver-independent literal methods use ordinary closure arguments; receiver retain/release pairs disappear. |
| internal/oracle/testdata/user_iterators.a | 663/663/455/904/64/0 | 663/663/408/857/64/0 | Receiver-independent literal methods use ordinary closure arguments; receiver retain/release pairs disappear. |
| internal/oracle/testdata/user_iterators_rest_tdz.a | 5/0/5/3/5/0 | 5/0/3/2/5/0 | Receiver-independent iterator entry/next no longer hold receiver arguments; TDZ panic exits before full cleanup (-2 retains, -1 release). |
| stage3/fixtures/taste/17_binder_flow.a | 51/51/0/51/7/0 | removed | Base already marks this fixture lowers=false; refresh removes its stale row. |

`logical_and_reference_maybe.a` moved within the regenerated table; its six numbers remain 8/8/11/22/4/0. The added registration file changes init ordering; this is a row relocation, not a runtime change. `stage3/fixtures/taste/17_binder_flow.a` is already registered with `lowers=false` in the assigned base's `taste_stage3_test.go`; no implementation or fixture change here caused its removal.

Command: `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts > /tmp/hidden-method/counts.log 2>&1`. Observed: package passed, 43.005s.
