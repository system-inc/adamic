type Groups = NonNullable<RegExpMatchArray['groups']>;
const own: Groups = { year: '2024' };
console.log(String(own.hasOwnProperty('year')));
