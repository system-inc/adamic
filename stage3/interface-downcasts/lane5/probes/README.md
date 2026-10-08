# Callable integration witnesses

These witnesses target the real source lowering. They are not registered as
passing shape-check oracles while the recorded-signature and shared-read hooks
are absent. The helper tests use the runtime separately; no census pair is
certified by those tests.

`number-good.a` must print `8`, and `optional-absent.a` must print `undefined`,
matching source Node. `number-wrong-value.a` must stop at the read with exit 70,
naming `run`, its declared function type and number. `number-wrong-arity.a` must
stop at the read with exit 70, naming `run`, the declared type and arity 0,
before the callback is stored or called. Stock Node accepts and calls that
zero-arity function, printing `7`; this deliberate type-contract failure needs
an independent pinned Adamic assertion. A compile-time cast refusal does not
satisfy either negative witness. Final integrated diagnostics will use the
shared expression text and declared type, and will be pinned in an oracle.
