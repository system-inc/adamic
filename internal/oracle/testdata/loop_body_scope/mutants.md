# Mutants for loop_body_scope (#9grybqc)

Each is one exact edit, run by hand with `go test ./internal/oracle -run TestLoopBodyScope -count=1` under `ADAMIC_GATE_UNCACHED=1`, then reverted. What each should make the test say is by reading the emitter; the run is what proves it.

## 1. The loop scope dropped

In `internal/native/emit_ownership.go`, `iteration` emits the body one level in and nothing else, so what the body owns joins the statement's list again:

```
-func (e *emitter) iteration(body func(leave func())) {
+func (e *emitter) iteration(body func(leave func())) {
+	e.indent++
+	body(func() {})
+	e.indent--
+	return
```

Expected: every program fails to build. The pack is declared inside the loop's block and released with the statement's temporaries past its closing brace, so clang refuses the C with "use of undeclared identifier", the q1 failure. This is a hard error, not a warning `-Werror` raises.

### 1b. The scope kept, its end-of-iteration release dropped

In `internal/native/emit_ownership.go`, `iteration`:

```
 	body(leave)
-	leave()
+	_ = leave
```

Expected: the C builds and the output agrees, and the leak check reports every iteration's rest array, and what it retained, on every program.

### 1c. No release before a break

In `internal/native/emit_arrays.go`, `arrayVisit`, the default case (some, every, findIndex):

```
 			e.line("\t%s = %s;", result, value)
 			e.indent++
-			leave()
+			_ = leave
 			e.line("break;")
```

Expected: search.a alone leaks, the pack of the iteration that settles `some`, `every` and `findIndex` early.

### 1d. No release before find's break

In `internal/native/emit_arrays.go`, `arrayVisit`, find's case, the `leave()` after its result is written:

```
 			e.indent++
-			leave()
+			_ = leave
 			e.line("break;")
```

Expected: search.a alone leaks.

Run: 1c and 1d each red through Loom's door, J-z0g4wq2 at d7722bee on azure-f109-20, TestLoopBodyScope/search alone with one LeakSanitizer report each. 1, 1b and 2 were red on all nine programs, J-pmpx9jt at d7722bee on azure-f109-19.

## 2. copyRest retains every slot

In `internal/native/arguments_length.go`, `restHoldsAlike`:

```
-	return element.IsReference() && slot.IsReference() || element == slot
+	return true
```

Before the edit, a slot the rest's element can't hold (q1's index, a number, in collect's rest of strings) is pushed as undefined, not as its number. That is a value Node would print differently, and it's safe only because no function that reads the rest runs with that pack. In q1 the closure map calls is the arrow written in place, `(item) => item`, which reads its one fixed slot. collect is in the call's layout only because its type is assignable to the callback's, and it can't reach the callback slot: a rest function seen through a type that hides its rest is NotYet (restHiddenByView, rest_fixed_view's r3 and r19), and where a type shows the rest, strictFunctionTypes holds each argument to its element. Lowering holds any type that admits a number or a boolean as a union (unknown, object and `{}` among them), and a union's rest boxes the number. So undefined goes only into a rest whose element doesn't admit the slot, or in place of a slot that is undefined itself. The comment at restSlot's last push says the same.

Every pack then goes through copyRest's loop, which retains each slot's word as a reference: the code before #9grybqc. Expected: the index slot's number is retained as a pointer, 1.0's bits at index 1 (a Map's value, 7.0, at its first entry in map_for_each.a), so the sanitized run stops with ASan's SEGV on an unknown address (or UBSan's report first), and the run disagrees with Node on its exit code. Index 0 reads as the null pointer, which retain passes over, so each program iterates at least twice.

## 3. The iteration dropped in reuse's map in place alone

In `internal/native/reuse.go`, `mapped`, the body is emitted one level in with no scope of its own, and every other site keeps its iteration:

```
-	e.iteration(func(func()) {
+	func(body func(func())) { e.indent++; body(func() {}); e.indent-- }(func(func()) {
```

Expected: mapped.a alone fails. Its map in `shouted` is the one the reuse plan takes over (`words` is a local of shouted's, not captured, read once by the map and dead after it), and its callback's type admits `joined` and `spread`, so the pack made in the loop joins the statement's list and its release is written past the loop's closing brace: clang refuses the C with "use of undeclared identifier". No other pin's map is taken over: map.a's `words` is a global and read again, throw.a's callback can throw and `some` is read again by the reduce, and a rest parameter of a closure (q1's and spread's `items`) is never handed over.
