# Method values census, October 8

Compiler base: 4721b9e180691f4f5f42b5a71166714a0c68d103.
Refusal table: d35a81d36fdafccf827bad0f572d311b2a0d4deb.

The pinned table has 108 method-value rows. 60 refer to the factory or
parenthesizer implementation families whose named bodies contain no lexical
`this`. Five accesses resolve directly to bodies that read `this`. The other
43 have unresolved runtime origins. They are not counted as this-free.

The 60 sites comprise 11 factory accesses, 45 `parenthesizer` accesses and four
`parenthesizerRules()` accesses. These are implementation-family observations,
not a checker proof about every possible object implementing those interfaces.
A compiler must preserve the distinction when implementing the ruling.

A broader name-only search finds this-free named bodies for 86 sites, but 26
are host, program or writer accesses. Matching a name does not prove the
runtime origin of a host callback. Two `trackSymbol` sites have both this-free
and this-reading candidates. Fifteen sites have no named body. The five
resolved this-reading sites plus those 43 unresolved sites account for the
remaining 48 rows.

`census.json` records every file, line, column, expression, candidate body
location and classification. Source is referenced directly through the cohere
submodule; none is copied. The source locations in all 108 rows were checked
against the AST. Analysis descends into arrows and skips nested ordinary
functions, which have their own receiver. Three parser controls distinguish
this-free arrows, arrows reading lexical this, and an ordinary nested function
reading its own this.

Reproduce after installing TypeScript 6.0.3 in a scratch directory:

```sh
METHOD_VALUES_TYPESCRIPT=/absolute/scratch/node_modules/typescript \
  node notes/method-values/ruling/census.cjs > /tmp/method-values-census.log 2>&1
```

Observed output:

```text
{"this-reading body":5,"unresolved runtime origin":43,"this-free implementation family":60}
direct bodies: 5
```

The refusal is raised in `internal/lower/refusals.go:177`. A second refusal
exists in `internal/lower/object.go`, so deleting the first one alone does not
implement extraction. Literal method closures currently receive their object
as an explicit leading argument. Class method tables contain direct-call
thunks. An extracted callable needs a separate receiver convention, preserving
identity and argument order; binding also needs an owning receiver reference.
This commit records the measurement only and changes no compiler behavior.
