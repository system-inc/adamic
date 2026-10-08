function fail(value: number | string | undefined): void {
	console.log('before');
	console.log(`${value!}`);
}
fail(undefined);
