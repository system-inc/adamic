# TestTwoImplementsTagsKeepBothInTheDeclaration's mutant

Task #4w889k2. The test holds the adapted compiler's declaration of
`testdata/implements/fixture.js` to stock TypeScript 6.0.3's, which reads
`export class Both implements A, B`.

The mutant is the reviewer's writer on #k8ed1f2, applied to
`stage3/adapt/99-readonly-family3-implements-types/adapt.cjs`:

```diff
-const after = [name + ' = [...' + name + ', implementsType];', 'readonly BaseType[]'];
+const after = [name + ' = [implementsType];', 'readonly BaseType[]'];
```

That keeps only the last implements type in `getImplementsTypes`. It checks
clean under stock TypeScript and passes the lane's upstream tests, because the
oracle never appends twice in one call (J-tzasvq2: 192 calls, none with two).

Proof J-k15ysp3, sealed on azure-d64a-1122-0748-1, at proof commit 62d8f72f
over the head 5c04cac6. The steps are on `proof/performance-4w889k2`.

- Head: `--- PASS: TestTwoImplementsTagsKeepBothInTheDeclaration (266.47s)`.
- Mutant: the adapted tree rebuilt under the mutant (`stage3-adapted-tree 18357b6be4e8 miss 268.84`), then
  `--- FAIL: TestTwoImplementsTagsKeepBothInTheDeclaration (270.46s)`. The adapted
  declaration read `export class Both implements B {` where stock's reads `export class Both implements A, B {`.
- Restored: `adapt.cjs` compared equal to its copy, and `go test ./stage3/lane/adaptedtree/` was ok in 3.4 s.
- Log sha256 0e523b4f69785b557cd2f389c50bf78f7388a4fa976cd2c046d04b2fd4766af0.
