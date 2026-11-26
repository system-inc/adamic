# narrowed_arm_property.go mutants (#8nj5pn4)

TestNarrowedArmPropertyRead holds each. Each is an exact edit of internal/lower/narrowed_nullish.go or narrowed_arm_property.go; restore it after.

1. The property dispatch dropped. In narrowedArmReadGap, `if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {` becomes `if false && (node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression) {`. The five stop rows lower again, and the lie rows print 1 natively where Node prints NaN.
2. The write test dropped. In narrowedArmPropertyReadGap, `if l.propertyWriteMayRunBetween(node, name, numeric, observation.from, observation.target, l.appendsBetween(observation, node, numeric)) {` becomes `if false && l.propertyWriteMayRunBetween(node, name, numeric, observation.from, observation.target, l.appendsBetween(observation, node, numeric)) {`. The five stop rows lower, since each has an observation.
3. The silent test dropped. In narrowedArmPropertyReadGap, `if !l.armReadMaySilentlyMisread(node, narrowed, declared) {` becomes `if false && !l.armReadMaySilentlyMisread(node, narrowed, declared) {`. The same-representation row stops, where it must lower.
4. The owner test dropped. In armReadMaySilentlyMisread, `return owner == nil || l.checker.IsTypeAssignableTo(owner, declared)` becomes `return true`. The unrelated declarer row stops, where it must lower.
