function found(): RegExpMatchArray | null {
	return '2024'.match(/(?<year>[0-9]+)/);
}
found()!.groups!.toString();
