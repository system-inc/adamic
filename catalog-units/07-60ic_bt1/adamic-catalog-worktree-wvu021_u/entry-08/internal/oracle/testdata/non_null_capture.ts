function run(): void {
	let value: string | undefined = 'ready';
	const clear = (): void => { value = undefined; };
	if (value !== undefined) {
		clear();
		console.log('before');
		console.log(value!);
	}
}
run();
