# Debug.log callable namespace

The exact native-callable-namespace.a front24 probe now prints callable
namespace loaded in all three implementations. callable-namespace-identity.a
adds direct calls, error methods on the callable namespace, mutable exported
level, equal and unequal canonical function identities, and top-level merging:

```text
log:first
error:1:second
error:2:third
true:true:true
print:fourth
3:true
```

Node, native with sanitizers and emitted JavaScript agree. This is a closed
qualified subset: callable escapes, alias properties, reflection and implicit
name/prototype/length are NotYet. Their preflight pins keep a named callable
object limitation; no real callable property container is silently assumed.
Receiver methods require every invoking reference to retain the declaring
namespace; canonical identity reads cannot invoke a detached receiver.

callable-namespace-unknown-before.a reaches comparison through a mutable
callback before Debug exists. All backends stop at TypeError reading log,
exit 70. Identity comparisons retain these checks even when their boolean
result is computed from declaration identity.

Actual compiler mutants: invert symbol identity, caught by clean stdout
differences in both backends; erase comparison readiness, caught by Node's
exit 70 against both backends' incorrect true/exit 0. Temporary edits restored.

Validation: full internal/lower passes; all TestNativeAgreesWithNode and
TestNamespace oracle tests pass (211.772s); counts refreshed (31.056s).
Normalized Debug.log advances to Compiles. Debug advances to its independent
bodyless overload limitation. The original Debug merge census site is cleared
as a declaration shape, without claiming its production bodies all lower.
