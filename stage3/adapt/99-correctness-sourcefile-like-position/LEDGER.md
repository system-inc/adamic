- **99-correctness-sourcefile-like-position (Adamic, task gqxd8v4, Nov 13
  22:09Z), type-only:** SourceFileLike's `@internal` member
  getPositionOfLineAndCharacter becomes an optional function-valued property,
  `((line: number, character: number, allowEdits?: true) => number) | undefined`,
  so scanner.ts:469's presence test reads a property, not a method as a value.
  This extends Adamic's Nov 1 sanction on #axochk, which superseded adaptation
  20's "no method-to-property conversion" for exactly 19 public members by name
  in adaptation 32's idiom (97d), to this one further member, by name, in a plain
  arrow type. The Nov 1 limit exists because a plain arrow drops method
  bivariance, which can move public relations and the declarations. Here neither
  moves: upstream's own hereby dts gives byte-identical typescript.d.ts and
  tsserverlibrary.d.ts, a SourceFile passed as SourceFileLike (utilities.ts:7959)
  gains no refusal, JavaScript is identical over 604 files, and the member has
  one reader. Adaptation 20's rule still governs every other member. Evidence:
  typescript/sourcefile-like-position-evidence,
  stage3/census/sourcefile-like-position.
