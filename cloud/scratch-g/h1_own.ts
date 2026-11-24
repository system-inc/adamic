type Groups = NonNullable<RegExpMatchArray['groups']>;
const own: Groups = { year: '2024' };
console.log(own.toString());
console.log(own.toLocaleString());
console.log(String(own.valueOf() === own));
