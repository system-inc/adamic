Temporary: comes out when non-null initialization and optional scanner inputs land

Plan published before implementation. Slice scanner.ts only. Replace precisely
the first var text = textInitial! with var text: string = undefined! and its
first setText(text, start, length) with setText(textInitial, start, length).
The intervening statements declare locals; none reads text. setText assigns
newText || "" before any scanner call. Keep this adaptation pending a ruling
and implementation for general non-null initializers.

languageVersion is optional: guard its two >= comparisons with !== undefined,
which preserves JavaScript's false result for omitted versions. The shebang
regex assertion is required, guaranteed by the initial #! test; retain it.
Other required-value ! sites retain their checks. Literal undefined! resets
remain uninitialized and require compiler read-before-assignment support.
Validate the token stream and complete upstream baseline before claiming done.
