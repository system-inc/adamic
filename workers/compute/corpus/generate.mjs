import { writeFileSync } from 'node:fs';

// Fixed xorshift32 seed. No clock, ambient randomness, locale or filesystem input.
let state = 0x6a09e667;
function random(maximum) {
	state ^= state << 13; state ^= state >>> 17; state ^= state << 5;
	return (state >>> 0) % maximum;
}
const requests = [];
function add(method, path, body = null) {
	requests.push({ method, url: `https://compute.example${path}`,
		headers: [['Content-Type', 'application/json'], ['X-Corpus', 'first'], ['x-CORPUS', 'second'], ['AcCePt', '*/*']],
		body: body === null ? null : typeof body === 'string' ? body : JSON.stringify(body) });
}
const stats = (values) => add('POST', '/stats', { values });
const text = (value, limit = 10) => add('POST', '/text/top-words', { text: value, limit });
const base = { currency: 'USD', shippingZone: 'domestic', items: [{ sku: 'A-1', quantity: 1, unitPriceCents: 1000 }] };
const quote = (value = base) => add('POST', '/orders/quote', value);
add('GET', '/health');
for (const limit of ['0', '1', '2', '3', '4', '10', '1000', '100000', '4999999', '5000000', '5000001', '-0', '-1', '+1', '01', '00', '1.0', '1e2', 'NaN', 'Infinity', '', ' 2', '2 ', '2\n', '9007199254740993']) {
	add('GET', `/primes?limit=${encodeURIComponent(limit)}`);
}
for (const path of ['/primes', '/primes?x=2', '/primes?limit=2&limit=3', '/primes?limit=bad&limit=2', '/primes?limit=%32', '/primes?limit=+2', '/primes?limit=%FF', '/primes?LIMIT=2', '/primes?x=9&limit=7']) add('GET', path);
for (const values of [[1], [1, 9], [1, 2, 3], Array.from({ length: 20 }, (_, i) => i + 1), [0.1, 0.2], [1e21, 1e21 + 131072], [5e-324, 5e-324], [1e308, 1e308], [1e308, -1e308], [1e16, 1, -1e16], [-0], [0, -0, 0], [-10, -2, 1, 4], []]) stats(values);
stats(Array.from({ length: 1000 }, (_, i) => i / 7));
for (const count of [99999, 100000, 100001]) stats(Array.from({ length: count }, (_, i) => (i % 101) / 10));
for (const body of ['{"values":[-0]}', '{"values":[9007199254740993]}', '{"values":[1e400]}', '{"values":[-1e400]}', '{"values":[1],"extra":null}', '{"values":"bad","values":[1,2]}', '{"values":[1],"values":null}', '{"values":[1,"2"]}', '{"values":[null]}', '{"values":[true]}', '{"values":{}}', '{"values":null}']) add('POST', '/stats', body);
for (const subtotal of [0, 1, 4, 5, 499, 500, 501, 4999, 5000, 5001, 10000000]) {
	for (const couponCode of [undefined, 'SAVE10', 'FLAT500']) {
		for (const shippingZone of ['domestic', 'international']) quote({ ...base, shippingZone, couponCode, items: [{ sku: 'Z', quantity: 1, unitPriceCents: subtotal }] });
	}
}
for (const currency of ['USD', 'EUR', 'usd', 'GBP', null, 1]) quote({ ...base, currency });
for (const shippingZone of ['domestic', 'international', 'DOMESTIC', '', null, 1]) quote({ ...base, shippingZone });
for (const couponCode of ['SAVE10', 'FLAT500', '', 'save10', 'UNKNOWN', null, 500]) quote({ ...base, couponCode });
for (const count of [0, 1, 50, 99, 100, 101]) quote({ ...base, items: Array.from({ length: count }, (_, i) => ({ sku: `SKU-${i}`, quantity: 1000, unitPriceCents: 10000000 })) });
for (const quantity of [-1, -0, 0, 1, 2, 999, 1000, 1001, 1.5, '1', null, 9007199254740992]) quote({ ...base, items: [{ ...base.items[0], quantity }] });
for (const unitPriceCents of [-1, -0, 0, 1, 9999999, 10000000, 10000001, 1.5, '1', null]) quote({ ...base, items: [{ ...base.items[0], unitPriceCents }] });
for (const sku of ['', 'A', 'A'.repeat(31), 'A'.repeat(32), 'A'.repeat(33), 'a', 'A_', 'A\n', 'A\r', 'A B', 'é', '🌍', null, 1, '0-9']) quote({ ...base, items: [{ ...base.items[0], sku }] });
for (const key of ['currency', 'shippingZone', 'items']) { const value = { ...base }; delete value[key]; quote(value); }
for (const key of ['sku', 'quantity', 'unitPriceCents']) { const value = { ...base.items[0] }; delete value[key]; quote({ ...base, items: [value] }); }
for (const literal of ['-0', '1e400', '-1e400']) {
	for (const field of ['quantity', 'unitPriceCents']) {
		const entry = field === 'quantity' ? `"quantity":${literal},"unitPriceCents":1` : `"quantity":1,"unitPriceCents":${literal}`;
		add('POST', '/orders/quote', `{"currency":"USD","shippingZone":"domestic","items":[{"sku":"A",${entry}}]}`);
	}
}
for (const items of [null, {}, [null], ['x'], [1]]) quote({ ...base, items });
quote({ ...base, extra: [null], items: [{ ...base.items[0], extra: true }] });
add('POST', '/orders/quote', '{"currency":"GBP","currency":"EUR","shippingZone":"domestic","items":[{"sku":"OLD","sku":"NEW","quantity":1,"unitPriceCents":-0}]}');
add('POST', '/orders/quote', '{"currency":"USD","shippingZone":"domestic","items":[{"sku":"A","quantity":1,"unitPriceCents":9007199254740993}]}');
for (const value of ['', 'A a B b 10 2', 'z y x', 'Hello HELLO hello!', 'é café 中 🌍 A\u0301 a', '\ud800Alpha\udc00 BETA', 'a_b-c.d\nE\tF', 'İ I ı K K k ß SS', '123 123 001', '" { [ \\ } ]', 'a'.repeat(999999), 'a'.repeat(1000000), 'a'.repeat(1000001), '🌍'.repeat(500000), '🌍'.repeat(500001)]) text(value);
text('Alpha beta 10 🌍 '.repeat(1000), 100);
for (const limit of [0, 1, 2, 99, 100, 101, -1, 1.5, '1', null]) text('A b c a', limit);
for (const value of [null, 0, [], {}]) add('POST', '/text/top-words', { text: value, limit: 1 });
for (const body of ['{"text":"a"}', '{"limit":1}', '{"text":"old","text":"NEW new","limit":1}', '{"text":"a","limit":1,"extra":true}', '{"text":"\\ud800A\\udc00b\\ud83c\\udf0dC","limit":10}']) add('POST', '/text/top-words', body);
const malformed = ['', ' ', '{', '[', '[1,]', '{"x":1,}', 'NaN', 'Infinity', '01', '-01', '-', '1.', '1e', '1e+', 'true garbage', '{x:1}', '{"x" 1}', '{"x":}', '"unterminated', '"control\ncharacter"', '"\\x"', '"\\u12xz"', '\u00a0{}', '\ufeff{}', '\v{}', '\f{}'];
for (const path of ['/stats', '/orders/quote', '/text/top-words']) {
	for (const body of [...malformed, 'null', '[]', '42', 'true', '"string"', '{}']) add('POST', path, body);
}
// Raw-source depth, including a duplicate extra value later overwritten with zero.
for (const depth of [126, 127, 128]) {
	const nested = '['.repeat(depth) + '0' + ']'.repeat(depth);
	for (const duplicate of ['', ',"extra":0']) add('POST', '/stats', `{"values":[1],"extra":${nested}${duplicate}}`);
}
add('POST', '/stats', '{"values":[1],"__proto__":{"values":null}}');
for (const [path, correct] of [['/health', 'GET'], ['/primes', 'GET'], ['/stats', 'POST'], ['/orders/quote', 'POST'], ['/text/top-words', 'POST']]) {
	for (const method of ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS']) if (method !== correct) add(method, path);
}
for (const path of ['/', '/missing', '/health/', '/Health', '/%68ealth', '/stats/child', '/orders%2Fquote', '/text/top-words/', '/health?x=1']) {
	for (const method of ['GET', 'POST']) add(method, path);
}
while (requests.length < 600) {
	const endpoint = random(4);
	if (endpoint === 0) add('GET', `/primes?limit=${random(50001)}`);
	else if (endpoint === 1) stats(Array.from({ length: 1 + random(200) }, () => (random(200001) - 100000) / 100));
	else if (endpoint === 2) quote({ currency: random(2) ? 'USD' : 'EUR', shippingZone: random(2) ? 'domestic' : 'international',
		items: Array.from({ length: 1 + random(10) }, (_, i) => ({ sku: `P-${i}`, quantity: 1 + random(1000), unitPriceCents: random(10001) })),
		...(random(3) === 0 ? {} : { couponCode: random(2) ? 'SAVE10' : 'FLAT500' }) });
	else text(Array.from({ length: 1 + random(200) }, () => ['Alpha', 'ALPHA', 'Beta', '2', '10', '🌍', 'é', 'a-b'][random(8)]).join(' '), 1 + random(100));
}
const destination = process.argv[2] ?? new URL('./requests.jsonl', import.meta.url);
writeFileSync(destination, requests.map((request) => JSON.stringify(request)).join('\n') + '\n');
console.log(`seed 0x6a09e667: ${requests.length} requests`);
