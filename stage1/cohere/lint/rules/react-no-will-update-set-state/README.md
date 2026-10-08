# react/no-will-update-set-state

The unchanged Go cohere rule is the oracle. The port checks the direct callee,
counts enclosing function scopes until the first lifecycle-named member, then
requires an enclosing React component. Findings underline only the callee and
have no fixes or suggestions. Shared helpers provide component recognition,
createReactClass recognition, property names and raw option decoding. The
adapter also accepts the typed options recorded by upstream tests.

The captured corpus contains 67 unique source/file/options cases. The three
owned witnesses cover member and callee spellings, nested functions under
`disallow-in-func`, and ancestry through unrelated objects and `implements`.

Go behavior deliberately preserved:

- `an UNSAFE prefixed lifecycle upstream silences with a React version setting`
  reports because Go receives no React version settings.
- Static methods and accessors named for the lifecycle report.
- A lifecycle-named object inside an unrelated component method reports;
  counting stops at the first lifecycle name rather than its component boundary.
- `implements Component` qualifies, and `createClass` does not.
- Parenthesized callee/receiver and parenthesized computed lifecycle keys are
  silent. Factory callee parentheses are accepted. Computed template keys report.

`mutant.json` replaces the ordinary lifecycle name with `componentDidUpdate`.
The rule-directory evidence records the upstream capture, toolchain timing,
selected parity checks, mutant catches and the full lint-package result.
