# Mutants for a parameter stored by index into an outer array (#22h1bv2)

One mutant, an edit to internal/native, run on a Linux box: make the edit, run the tests in
internal/oracle/region_index_store_test.go, then put region.go back from a saved copy and `cmp` it.

The mutant: internal/native/region.go, escaping's `case ir.SetIndex:` arm reduced to
`derived(statement.Value)`, without `escaped = true`, so storing the parameter by index no longer marks
it escaping.

## First run, at 77a188af: survived

On Home, with the pins as they first landed, every test stayed green under the mutant. index_store.a's
keep ended `return p.x;`, and escaping counts that return as the parameter escaping on its own: derived
follows a Property to its object without asking whether the field is a scalar, and a value derived from
the parameter that a statement returns escapes. So the parameter was marked whether the SetIndex arm
fired or not, the statement got no region, and the pin witnessed nothing of the arm.

## Second run, at 77a188af: killed, with the reshaped keeps

A probe on Cloud (`internal/native`, at the same commit) put two reshaped programs beside the old one,
under the same mutant:

| Program | Unmutated | Mutant |
|---|---|---|
| index_store.a as it was, `return p.x;` | no region, `7 p7 7` | no region, `7 p7 7`: survived |
| keep returning `p.x + p.label.length;`, a Binary, which escaping consumes | no region, `9 p7 7` | a region for `total += keep(make(7));`, then ASan heap-use-after-free |
| keep returning void, `keep(make(7));` | no region, `p7 7` | a region for `keep(make(7));`, then ASan heap-use-after-free |

Both kills are the same report: a read in `adamic_retain` (runtime/adamic.h:185) of memory freed by
`adamic_region_end` (region.c:148), allocated by `object_new` (region.c:70), the point make(7) built
in the statement's region while outer[0] still held it.

Those two programs are now index_store.a and void_index_store.a. push_store.a's keep returns the same
Binary, so its pin stands on the push alone. The push isn't the SetIndex arm's: `outer.push(p)` lowers
to ir.ArrayPush, which isn't pure and so isn't a consumer (borrow.go), and escaping's general operand
rule marks it. control.a stores nothing and keeps its region either way.

## Owed

The mutant against these fixtures as tests, `TestRegionIndexStore`, `TestRegionVoidIndexStore` and
`TestRegionPushStore` in internal/oracle, hasn't run on a box yet. Expected: index_store.a and
void_index_store.a red, each with a region in its C and heap-use-after-free natively; control.a and
push_store.a green.
