# Mutant: includesUndefined's void exclusion restored (#7yt238q)

Applied to internal/lower/expression.go's includesUndefined. Replace

		if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 || l.phantomUndefined(member) {

with

		if member.Flags()&checker.TypeFlagsUndefined != 0 || l.phantomUndefined(member) {

Then `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestVoidMemberCountsAsUndefined|TestVoidResultReadStillRefused' -count=1 -v`: witness, alias, field and field_alias go red in .a and .ts (B | null | void held as one pointer again, as on main efe88776: show(undefined) prints null's line natively, and value === undefined folds to false on JavaScript). nullish and control stay green (?? and ?. read null and undefined alike, so one pointer never misread them), and so do the four H2 refusals (void_read, void_argument, void_binding, void_field), which don't ask includesUndefined.
