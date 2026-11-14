Mutant evidence for real/any-value/main.a's stage-0 status.json row (#53r37a3).

Mutant: on a scratch copy of the real bucket, restore only this row's old
NotYet outcome, full diagnostic and current_reason. Source programs, recorded
Node behavior and every other row are unchanged. The production compiler is
not mutated. The corrected row passes before rollback.

Measured command (the prepared fixtures binary uses this tree's compiler):

```sh
/tmp/53r37a3-never-refused/head8/test2json -t -p github.com/system-inc/adamic/stage3/fixtures /tmp/53r37a3-never-refused/head8/fixtures.test -test.v=test2json -test.count=1 -test.parallel=4 -test.timeout=30m '-test.run=^TestFixturesReal$' -fixtures /tmp/53r37a3-never-refused/head8/status-row-mutant
```

Result: normal real bucket exit 0; rollback mutant exit 1. Exactly one fixture
row fails: `TestFixturesReal/real/any-value/main.a`. Its `/node` check passes,
and all nine other real fixture rows pass. Ancestor failures only propagate
that one row's failure. Catch: `fixtures_test.go:493`.

```
fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/real/any-value/main.a:4:36: Adamic 0.1 refuses a value of type any: any has no proven type; type it unknown and narrow it where it's read, or give it its real type Stdout:}; recorded {Outcome:NotYet What:stage3/fixtures/real/any-value/main.a:4:36: stage 0 can't lower a value of type any yet Stdout:}
```

The any diagnostic stays at 4:36. Historical census_count 30 and the recorded
Node stdout `true:false\n`, stderr and exit are preserved. The ordinary -update
flag cannot refresh this row: it deliberately updates only currently compiling
fixtures after native/Node agreement. This record instead captures the actual
Refused diagnostic under the existing noncompiling-fixture schema.
