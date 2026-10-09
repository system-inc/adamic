# `nexus/correctness-no-implicit-return`

| | |
|---|---|
| **Recommendation** | **Yes, at `error`, in a project whose tsconfig leaves `noImplicitReturns` off.** Where the flag is on, the rule stands down and says so in `--coverage`, because the type check reports the same findings |
| Findings | **api 17 before #rbvd7sv's wave, 0 after** (measured 2026-10-04), each on the same file, line and column as `tsgo --noImplicitReturns` |
| Measured precision | exact against TypeScript: the fixtures compare the rule's spans with TS7030's, built with the flag on |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes; without a checker the rule registers nothing |

## What it checks

What TypeScript's `noImplicitReturns` checks, TS7030 "Not all code paths return a value":

- A function, method, getter, function expression or arrow function with a block body, whose end is
  reachable, that returns a value somewhere and whose return type is not `void`, `any` or `undefined`
  after unwrapping a promise or generator.
- Without `strictNullChecks`, a bare `return;` in a function whose return type needs a value.

Reachability is the checker's: an exhaustive `switch` over a union ends nothing, and a call to a function
declared to return `never` ends the path it is on. A function whose declared type excludes `undefined`,
or that declares a type and never returns, is a type error whatever the flag says (TS2366, TS2355), so
the rule leaves it to the type check, as TypeScript does.

## Why it exists

Base's tsconfig leaves `noImplicitReturns` off until Kam adopts it (#tz23yx2), so nothing caught a new
implicit return in api. The ported `consistent-return` cannot stand in: it answers reachability
syntactically, so in api it reported 57 (core) and 55 (typed) against TypeScript's 17, and about 38 of
the extra are exhaustive switches and calls returning `never`, which Kirk ruled on 2026-08-25 need no
explicit return (#yj96emr).

## Options

None. The rule reports what TypeScript would, and an option would make it report something TypeScript
does not.
