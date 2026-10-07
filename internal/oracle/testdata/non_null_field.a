const box: { value: number | undefined } = { value: 1 };
function clear(): void { box.value = undefined; }
if (box.value !== undefined) {
	clear();
	console.log('before');
	console.log(`${box.value!}`);
}
