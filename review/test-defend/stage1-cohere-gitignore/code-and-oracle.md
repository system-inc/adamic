Code under test and oracle, recorded before any mutant execution.

Path row: TypeScript port path.ts functions isClean, clean, base and dir, executed on Node, compiled natively and through the JavaScript backend. Go standard-library path.Clean, path.Base and path.Dir compute expected output for 5040 inputs. pathsweep.ts is the input/output driver and is not mutated. Internal built-in mutant subcases are witnesses rather than positive comparisons.

Gap row: Adamic lowering and native emission for ten closed standalone gap programs. Node runs each original fixture and must agree with hand-written recorded stdout. Native output must agree, and leak checks run. The fixtures themselves are not mutated. D2 changes production lowering's bitwise complement enum constant to Negate. The port comparison family executes the Gitignore port against Go cohere and Git.

D1 intentionally targets the empty-path Base branch. D2 intentionally targets ~200. An earlier Array.from bound idea was rejected before execution because fresh compiler coverage showed the glob port reaches Array.from too. Mutation planning is separate from kill observations. All mutations are unconditional standalone diffs against starting origin/main, recorded in session metadata. No oracle or test edit.
