# super read inside an arrow in a subclass emits an undeclared C identifier (task rh16bxg)

**Today on origin/main 74fb6490a.** This program:

```ts
class First {
    describe(): string { return 'first'; }
}
class Second extends First {
    override describe(): string {
        const read = (): string => super.describe();
        return read() + '!';
    }
}
console.log(new Second().describe());
```

fails to build:

```
adamic: native: clang failed: exit status 1
main.c:127:73: error: use of undeclared identifier 'adamic_local_3_this'
```

The closure's body reads the method's `this` directly, but the closure never captured it. The immediately invoked form `(() => super.describe())()` and a super getter read inside an arrow fail the same way. All four reduced programs in the metamorphic report for this task build and match Node once this is fixed: class_oct6_deep (identity), class_oct6_release, class_oct6_subclass_holder and borrow_element_super_move.

**Node prints** `first!`.

**Where.** A bare `this` works because internal/lower/expression.go calls `l.touch(l.this)` at line 555 before it reads `this`. `touch`, at internal/lower/locals.go line 143, is what adds the local to every enclosing closure's `Environment`. The `super` paths skip that call:

- internal/lower/class.go, `func (l *lowering) callOrMethod` at line 312. Under `if ast.SkipParentheses(receiver).Kind == ast.KindSuperKeyword`, line 383 builds `object = ir.Read{Local: l.this, Of: ir.Object}` without a touch.
- internal/lower/class_accessors.go, `func (l *lowering) superAccessor` at line 53. It reads `l.this` at lines 89 and 94, also without a touch.

The fix is to call `l.touch(l.this)` after `useOfThis` succeeds on both paths. `superCall` (class_inheritance.go line 398) is `super(...)` in a constructor, which cannot sit inside an arrow, and needs no change.

**The fixture.** `internal/oracle/testdata/class_super_arrow.a`, registered by `internal/oracle/class_super_arrow_test.go`:

```ts
class Base {
	readonly label: string;
	constructor(label: string) {
		this.label = label;
	}
	describe(): string {
		return 'base ' + this.label;
	}
	get title(): string {
		return 'title ' + this.label;
	}
}
class Child extends Base {
	override describe(): string {
		return (() => super.describe())() + ' child';
	}
	override get title(): string {
		const read = (): string => super.title;
		return read() + '!';
	}
	later(): () => string {
		return () => super.describe() + ' later';
	}
}
const child = new Child('one');
console.log(child.describe());
console.log(child.title);
const read = child.later();
console.log(read());
```

Node prints:

```
base one child
title one!
base one later
```

**Done when.**

- `go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/class_super_arrow.a' -count=1` passes, including the leak check. `later()` returns a closure that outlives the method, which exercises the captured `this`.
- Mutant "super method uncaptured": remove the touch in `callOrMethod`. The fixture then fails with `use of undeclared identifier 'adamic_local_N_this'`.
- Mutant "super getter uncaptured": remove the touch in `superAccessor`. The fixture then fails in the same way.
- `go test ./internal/lower` passes.

**Size.** 2 lines plus the fixture. Confidence high.

**Out of scope, leave the task open for it.** One shape in the task still fails after this fix. It has a different cause, a field initializer arrow in a subclass that is never constructed:

```ts
class First { readonly label: string = 'first'; }
class Second extends First { readonly copy = (() => this.label)(); }
class Third extends Second {}
const check = (value: First): boolean => value instanceof Third;
console.log(String(check(new First())));
```

It emits `use of undeclared identifier 'adamic_local_4_this_cell'`. The same program with `new Second()` builds. Note this in the task comment rather than fixing it here.

Shared notes:

- Registration. Each fixture registers from its own new file, `internal/oracle/<fixture name>_test.go`, with an `init()` that appends to `fixtures`. Copy the pattern in internal/oracle/switch_empty_test.go lines 9 to 18 (fallthrough_test.go lines 11 to 20 is the same pattern). Do not edit the shared table in oracle_test.go.
- Names. None of the fixture names below exists on 74fb6490a. Note that `string_positions.a` is already taken by an unrelated fixture.
- Node. To run a fixture on Node directly, copy it to a `.ts` file and run `node --experimental-strip-types --no-warnings file.ts`.
- Host failures. On macOS, `go test ./internal/native` already fails on main (AddressSanitizer reports that detect_leaks is not supported on the platform, plus parseInt and cos bit mismatches against Node). Judge "the packages touched pass" on Linux, or by comparing against the same failures on main.
- Briefs 3, 4 and 5 all edit the `stringMethods` table and the `stringCall` switch in internal/lower/object.go. Land them one at a time and rebase. If brief 4 lands before brief 5, brief 5 should reuse the helper brief 4 adds.
