type Groups = NonNullable<RegExpMatchArray['groups']>;
const match = '2024'.match(/(?<year>[0-9]+)/);
(match!.groups as Groups).toString();
