# spread_prototype_members.go's mutant

- **The refusal answers nil** (`checkSpreadPrototypeMembers` returns nil before its loop): a .a spread whose
  type declares as a field a member a class reaching it holds on its prototype lowers on, so
  `TestSpreadPrototypeField` (internal/oracle/spread_prototype_test.go) goes red on c13b_a.a, setter_a.a,
  method_field_a.a and base_getter_a.a, each wanting the spread refused; own_field_a.a and
  undeclared_getter_a.a, which the rule admits, stay green. `TestSpreadRecordsTheGettersItMayRun`
  (internal/lower/spread_getters_test.go) goes red too, on its class's and inherited rows, each wanting its .a
  twin refused. Run on Loom as its own red job.
