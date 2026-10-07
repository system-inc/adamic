# Slot 03 sixth helper batch

Three Collapse helpers, one .a file each. AtRule and StyleRule create fresh CSS node records with exactly the Go kind and unmodified name/params or selector. Unrelated text fields are empty, flags are false and the context map is nil. css_node.a is a type-only adapter: contextPresent records nil-map presence; child indices preserve pointer identity, with -1 for nil children.

ChildList explicitly represents a Go slice header: values is the shared backing storage, length and capacity describe the view, and present distinguishes nil from a nonnil empty slice. Each constructor copies the header into a fresh view and keeps the same values storage. Element writes are shared; changing the returned view's length or capacity does not change the caller's header. A caller supplies a valid view with 0 <= length <= capacity <= values.length. Slice growth/reslicing operations remain the AST adapter's responsibility, rather than JavaScript array push semantics. No separate source field or identity is invented.

variantNextOrder takes the non-null registry state as an explicit dependency. A present group returns its order, including zero and negative orders; otherwise it returns lastOrder+1. It never changes state. Exact JavaScript-safe integers are the numeric boundary, with lastOrder+1 also required to be safe when no group is present. The pinned framework orders are small; full Go int64 overflow behavior is outside this adapter.

The owned real-Go oracle uses public constructors and an oracle-only export of the private registry method. All six consuming rules and 157 runtime inputs are captured before external-engine skips. Full sources, lexical tokens, U+0000..U+00FF, NUL, whitespace and supplementary Unicode are constructor inputs. Child controls cover nil, nonnil empty, spare capacity, duplicate and nil pointers, shared writes, independent header truncation and fresh node allocation. Order controls cover safe-integer boundaries, absent/present/zero/negative groups, caller-derived states and repeated reads.

```
python3 stage1/cohere/lint/helpers/slot03/batch6/testdata/regenerate.py > /tmp/slot03-batch6-capture.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch6' -count=1 -v -timeout=20m > /tmp/slot03-batch6-tests.log 2>&1
```

Source /workspace/adamic-tools/env.sh first. Compare actual Go against source Node, emitted JavaScript through oracle/node.mjs and native under ASan/UBSan. Eleven compiling mutants prove kind tags, shared backing storage, independent headers, fresh nodes, group precedence, increment and read-only behavior. The failed external Tailwind live/corpus gate is recorded separately. These remove eighteen dependency edges across six rules, with no final helper blocker removed by this batch alone. No shared harness, registration generator, compiler or other worker directory changes. See REPORT.md and readiness.json for exact consumers and limits.
