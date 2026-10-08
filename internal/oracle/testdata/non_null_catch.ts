const map = new Map<string, number>();
try {
	console.log('before');
	console.log(`${map.get('missing')!}`);
	console.log('after');
} catch {
	console.log('caught');
} finally {
	console.log('finally');
}
