// A program's own object typed as a match's groups.
type Groups = NonNullable<RegExpMatchArray['groups']>;
const own: Groups = { year: '2024' };
console.log(own.toString());
