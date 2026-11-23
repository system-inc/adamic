# Nongeneric type reference resolution name

Family 113 at `src/compiler/program.ts:1078:12`, reported from the getter
object's unit at `src/compiler/program.ts:1081:1`. The private helper now
takes `FileReference | string` and explicitly returns `string`. Its
`isString` guard and both branches retain their exact source bytes.

The getter slot accepts `FileReference | string`. The map callback accepts
elements constrained to that union and produces strings. The direct call
passes a `FileReference` and uses the resulting string as a resolution
name. The helper is private and does not appear in the public declarations.

Adamic ruled on #bbssdcy at 21:49Z on November 22, 2026. The current lowering
boundary follows the generic-body and generic-value rules in
`docs/rulings.md:127` and `:128`. Narrowing and predicate rules `:141` and
`:142` apply to the unchanged body. This is a temporary adaptation under
`:400`. Remove it when the compiler lowers the generic signature.
