# type-projection

ABI-v1 raw checker operations on live type identities, with no lint predicates.
Every answer begins with version 1 and mode type-projection.

- `type-projection\n<ID>\nproperty\n<name>`: two identities, the named property type
  at the queried node and the string-index type. Zero means absent. Native code
  chooses whether the fallback applies.
- `type-projection\n<ID>\nindex`: the number-index type identity, or zero.
- `type-projection\n<ID>\nsignatures`: count, then each call signature's return
  identity, parameter identities, and rest flag. Parameter types use the exact
  queried location. Native code decides whether a callback is callable/thenable.
- `type-projection\nheritage`: class-only, heritage type identities in source
  clause order. This reads the checker's tree because the current native parser
  represents implements entries as TypeReference rather than the Go tree's
  ExpressionWithTypeArguments. It returns types, not interface-contract verdicts.

Identity spelling is canonical decimal and must belong to the live program.
Unknown operations, incorrect arity, bad identities and invalid heritage-node
kinds refuse. Native framing checks consume all fields; released programs refuse
before any operation runs. Both Go and Adamic question files are isolated; the
shared checker switch changes by one physical line per question.
