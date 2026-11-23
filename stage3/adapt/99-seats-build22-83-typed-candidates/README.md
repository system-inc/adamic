# Type the output directory candidates

At `src/compiler/moduleNameResolver.ts:2977:21` the first push reads an
evolving empty array as `any[]`. Declare the local `candidateDirectories`
as `string[]` in the private `getOutputDirectoriesForBaseDirectory`.

The declaration-directory and output-directory pushes both take the result
of `toAbsolutePath(combineDirectoryPath(...))`. The combination supplies a
string, so the string overload of `toAbsolutePath` returns a string. The
output-directory push runs only when an output directory exists and differs
from the declaration directory. The existing push order remains the same.

Compiler accepts this type-only edit under the Nov 22 ruling for family 83
(#bbssdcy). The inferred-`any` and refusal rules in `docs/rulings.md:39` and
`:282` apply to the old receiver. The declared array has one element type,
consistent with `:333`. The declaration stays inside nested private functions, within the
public-API limit at `:397`.
