interface Stamp {
	toISOString(): string;
}
function show(stamp: Stamp): void {
	console.log(stamp.toString());
}
show(new Date(0));
show({ toISOString: (): string => 'own' });
