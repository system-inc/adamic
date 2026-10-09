// Gap 2: throwing an Error made in another function. graphql-js's parser throws what its unexpected()
// and syntaxError() return.
function unexpected(at: number): Error {
	return new Error(`Unexpected token at ${at}.`);
}
try {
	throw unexpected(3);
} catch (error) {
	console.log(error instanceof Error ? error.message : '?');
}
