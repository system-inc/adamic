type Groups = NonNullable<RegExpMatchArray['groups']>;
const match = '2024'.match(/(?<year>[0-9]+)/);
function found(): RegExpMatchArray | null {
	return '2024'.match(/(?<year>[0-9]+)/);
}
const tries: Array<() => unknown> = [
	() => '2024'.match(/(?<year>[0-9]+)/)!.groups!.toString(),
	() => found()!.groups!.toString(),
	() => [match!.groups!][0]!.toString(),
	() => match!.groups!['toString'](),
	() => match!.groups?.toString(),
	() => (match!.groups as Groups).toString(),
	() => match!['groups']!.toString(),
	() => (match!).groups!.valueOf(),
	() => match?.groups!.toLocaleString(),
	() => match!.groups!.hasOwnProperty('year'),
	() => match!.groups![`toString`](),
];
for (const attempt of tries) {
	try {
		console.log(String(attempt()));
	} catch (error) {
		console.log(String(error));
	}
}
