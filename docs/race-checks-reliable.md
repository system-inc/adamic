# Reliable concurrency race checks

This unit starts at concurrency-scaling 0f67284. It measures every existing runtime
negative control fifty times before changing the fixtures, then repeats the same
measurement after strengthening them. A race control must produce a TSan data-race
report, not merely crash or time out.

Claimed files: internal/native/parallel_test.go, internal/native/scaling_test.go,
new internal/native race-check test helpers, internal/native/testdata/parallel/*.c,
internal/native/native.go (TSan-only flag), internal/native/runtime/adamic.h,
internal/native/runtime/heap.c, internal/native/runtime/parallel.c, new runtime
TSan hook files, and internal/oracle/parallel_test.go. No lowering or JavaScript
changes. Release and ASan builds must contain no hook code; release objects are
compared byte for byte with the parent.

At this parent there are five TSan race controls, six C assertion controls and
two scaling guard/leak controls. All thirteen are measured; their detectors are
reported separately rather than calling assertion failures race reports.
