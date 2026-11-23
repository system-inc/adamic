# TypeScript upstream ledger

Write upstream findings and drafted issues in the adaptation's own
`stage3/adapt/<directory>/LEDGER.md`. Amend an entry in its owning directory.
For an adaptation with no upstream note, use the following single line.

```text
none: no upstream note was recorded when this adaptation landed.
```

Generate the combined ledger from the repository root with either command.

```sh
python3 stage3/upstream/ledger.py
python3 stage3/upstream/ledger.py /tmp/tsc-upstream-ledger.md
```

The generator reads the same directory order as apply and skips `none:` files.
Use a scratch path for the combined ledger.
