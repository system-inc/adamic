function fail(value: boolean | undefined): void {
	console.log('before');
	console.log(`${value!}`);
}
fail(undefined);
