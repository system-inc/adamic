# Quarantined parser scratch: evidence only

Never merge this branch. It preserves the miscompiling compiler integration at
33bf53aca3ce199aa05fe7bd2e7898e7c5ef2d5b unchanged, with only this README added.
It is published solely so compiler can prove that the new closure convention
turns the incompatible integration into a build error. It is not a release,
a validated compiler, or a base for native parser acceptance.

The base contains arguments-length 9534e8ab. In the recorded run, the following
function-value witness built successfully and exited 0 with empty stderr, but
native stdout was `9\n` while Node stdout was `1\n`:

```typescript
function reader(): number { return arguments.length; }
const counted: (value: number) => number = reader;
console.log(`${counted(9)}`);
```

The source witness is native-arguments-length-value.a on the parser proof branch.
The recorded run and related oracle failures are in that branch's
stage3/drivers/parser/evidence/front25/. This README reports existing evidence;
no build, mutant or acceptance run was performed when publishing this branch.

The original scratch's local SDK module-path replacements and untracked build
files are not part of the committed base and are not added here. Reproduction
must supply the repository's pinned SDK dependencies. No compiler source change
or conflict resolution is introduced by the publication commit.
