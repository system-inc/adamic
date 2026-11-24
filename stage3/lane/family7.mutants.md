# Family 7 landing printer guard mutant

In `stage3/lane/check.py`, restore the old evidence-based skip at the
start of `method_signature_errors`, after it sets `name`:

```python
table = tree / 'patch-set.md'
if not table.exists() or not any(line.startswith(f'| {name} |') for line in table.read_text().splitlines()):
    return []
```

From `stage3/lane`, run
`python3 -m unittest -v test_check.LandingLaneTests.test_family7_printer_signature_is_checked_in_byte_comparisons`.
The test restores `writeList.nodes: NodeArray<T>` and keeps that source
control in place while removing the Family 7 row or the whole table.
The mutant returns a passing verdict in both missing-evidence subtests,
which expect a failing verdict naming `writeList.nodes`. The same
subtests pass with the current guard. The sealed run belongs to the
house proof.
