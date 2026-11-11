# Mutant: includesUndefined's void exclusion restored (#7yt238q)

Applied to internal/lower/expression.go's includesUndefined. Replace

		if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 || l.phantomUndefined(member) {

with

		if member.Flags()&checker.TypeFlagsUndefined != 0 || l.phantomUndefined(member) {

Then `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestVoidMemberCountsAsUndefined|TestVoidResultReadStillRefused' -count=1 -v`: witness, alias and field go red in .a and .ts (B | null | void held as one pointer again, as on main efe88776: show(undefined) prints null's line natively, and value === undefined folds to false on JavaScript). control and void_read stay green.
