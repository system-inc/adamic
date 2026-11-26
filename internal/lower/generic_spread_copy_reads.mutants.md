# generic_spread_copy_reads.go's mutants

- **The stop answers nil** (`checkGenericSpreadCopyReads` returns nil before its walk): generic_method_call.ts and
  private_read.ts lower again, so `TestGenericSpreadPrototypeMember`
  (internal/oracle/generic_spread_prototype_test.go) goes red on them, and on generic_method_typeof.ts and
  generic_setter.ts, each wanting the read stopped. generic_getter.ts stays stopped, by spreadCopyRead's words
  rather than these, so it goes red on the words alone. Built natively, private_read.ts then dies on the runtime
  backstop, "runtime invariant broken: a class method entered with a this that isn't an instance of its class"
  (`adamic_method_this`, internal/native_runtime/c/class_inheritance.c), never printing a number.
- **The stop answers nil and the backstop is dropped too** (`methodOwner`'s guard in
  internal/native/class_inheritance.go's `callCode` never taken): private_read.ts prints 14 natively, ASan and
  UBSan silent, where Node throws, since Scale's weigh reads #grams and extra at offsets the copy lays out
  differently. That is main's wrong answer, and the backstop's red.

Each runs on Loom as its own red job.
