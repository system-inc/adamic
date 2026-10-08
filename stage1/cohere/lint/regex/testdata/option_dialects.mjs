// Independent Node observations of the required raw-option constructor.
for (const [pattern, text] of [
  ['\\s', '\u00a0'], ['a$', 'a\n'], ['(?i)todo', 'TODO'],
  ['\\p{Greek}', 'α'], ['a\\z', 'a'], ['(?P<word>a)', 'a'],
]) {
  try { console.log(`${pattern}\t${new RegExp(pattern, 'u').test(text)}`); }
  catch (error) { if (!(error instanceof SyntaxError)) throw error; console.log(`${pattern}\tSyntaxError`); }
}
