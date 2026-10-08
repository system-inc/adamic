interface Stats { values: number[] }
interface Item { sku: string; quantity: number; unitPriceCents: number }
interface Order {
	currency: 'USD' | 'EUR';
	shippingZone: 'domestic' | 'international';
	items: Item[];
	couponCode?: string;
}
interface TextRequest { text: string; limit: number }

function object(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null && !Array.isArray(value);
}
function integer(value: unknown, minimum: number, maximum: number): value is number {
	return typeof value === 'number' && Number.isInteger(value) && value >= minimum && value <= maximum;
}
function stats(value: unknown): value is Stats {
	return object(value) && Object.hasOwn(value, 'values') && Array.isArray(value.values)
		&& value.values.length >= 1 && value.values.length <= 100000
		&& value.values.every((entry: unknown) => typeof entry === 'number' && Number.isFinite(entry));
}
function item(value: unknown): value is Item {
	return object(value) && Object.hasOwn(value, 'sku') && Object.hasOwn(value, 'quantity')
		&& Object.hasOwn(value, 'unitPriceCents') && typeof value.sku === 'string'
		&& /^[A-Z0-9-]{1,32}(?![\s\S])/.test(value.sku)
		&& integer(value.quantity, 1, 1000) && integer(value.unitPriceCents, 0, 10000000);
}
function order(value: unknown): value is Order {
	return object(value) && Object.hasOwn(value, 'currency') && Object.hasOwn(value, 'shippingZone')
		&& Object.hasOwn(value, 'items') && (value.currency === 'USD' || value.currency === 'EUR')
		&& (value.shippingZone === 'domestic' || value.shippingZone === 'international')
		&& Array.isArray(value.items) && value.items.length >= 1 && value.items.length <= 100
		&& value.items.every(item)
		&& (!Object.hasOwn(value, 'couponCode') || typeof value.couponCode === 'string');
}
function textRequest(value: unknown): value is TextRequest {
	return object(value) && Object.hasOwn(value, 'text') && Object.hasOwn(value, 'limit')
		&& typeof value.text === 'string' && value.text.length <= 1000000 && integer(value.limit, 1, 100);
}

// JSON.parse already checked the grammar. Scan the source because dropped fields and
// overwritten duplicate values still count toward decodeJson's container limit.
function withinDepth(source: string): boolean {
	let depth = 0;
	let quoted = false;
	for (let index = 0; index < source.length; index += 1) {
		const character = source[index];
		if (quoted) {
			if (character === '\\') index += 1;
			else if (character === '"') quoted = false;
		} else if (character === '"') quoted = true;
		else if (character === '{' || character === '[') {
			depth += 1;
			if (depth > 128) return false;
		} else if (character === '}' || character === ']') depth -= 1;
	}
	return true;
}
function json(value: unknown, status = 200, allow?: string): Response {
	const headers: [string, string][] = [];
	if (allow !== undefined) headers.push(['allow', allow]);
	headers.push(['content-type', 'application/json; charset=utf-8']);
	return new Response(JSON.stringify(value), { status, headers });
}
function invalid(): Response { return json({ error: 'invalid request' }, 400); }

function primes(limit: number) {
	const composite = new Uint8Array(limit + 1);
	for (let prime = 2; prime * prime <= limit; prime += 1) {
		if (composite[prime] === 0) {
			for (let multiple = prime * prime; multiple <= limit; multiple += prime) composite[multiple] = 1;
		}
	}
	let count = 0;
	let last = 0;
	for (let value = 2; value <= limit; value += 1) {
		if (composite[value] === 0) { count += 1; last = value; }
	}
	return { limit, count, last };
}
function summarize(values: number[]) {
	const count = values.length;
	let sum = 0;
	for (let index = 0; index < count; index += 1) sum = sum + values[index]!;
	const mean = sum / count;
	const sorted = values.slice().sort((left, right) => left - right);
	const middle = Math.floor(count / 2);
	const median = count % 2 === 1 ? sorted[middle]! : (sorted[middle - 1]! + sorted[middle]!) / 2;
	const p95 = sorted[Math.ceil(0.95 * count) - 1]!;
	let squaredSum = 0;
	for (let index = 0; index < count; index += 1) {
		const difference = values[index]! - mean;
		squaredSum = squaredSum + difference * difference;
	}
	const standardDeviation = Math.sqrt(squaredSum / count);
	return { count, mean, median, p95, min: sorted[0]!, max: sorted[count - 1]!, standardDeviation };
}
function quote(value: Order) {
	let subtotalCents = 0;
	const lines = value.items.map((entry) => {
		const lineCents = Math.round(entry.quantity * entry.unitPriceCents);
		subtotalCents = subtotalCents + lineCents;
		return { sku: entry.sku, quantity: entry.quantity, lineCents };
	});
	const discountCents = value.couponCode === 'SAVE10' ? Math.round(subtotalCents * 0.1)
		: value.couponCode === 'FLAT500' ? Math.min(500, subtotalCents) : 0;
	const shippingCents = value.shippingZone === 'domestic' ? (subtotalCents >= 5000 ? 0 : 799) : 2499;
	const taxCents = value.shippingZone === 'domestic' ? Math.round((subtotalCents - discountCents) * 0.0725) : 0;
	const totalCents = subtotalCents - discountCents + shippingCents + taxCents;
	return { currency: value.currency, subtotalCents, discountCents, shippingCents, taxCents, totalCents, lines };
}
function topWords(value: TextRequest) {
	const counts = new Map<string, number>();
	for (const token of value.text.match(/[A-Za-z0-9]+/g) ?? []) {
		const word = token.toLowerCase(); // Only ASCII tokens reach this conversion.
		counts.set(word, (counts.get(word) ?? 0) + 1);
	}
	const words = [...counts].map(([word, count]) => ({ word, count }));
	words.sort((left, right) => right.count - left.count || (left.word < right.word ? -1 : left.word > right.word ? 1 : 0));
	return { words: words.slice(0, value.limit) };
}

export default {
	async fetch(request: Request): Promise<Response> {
		const url = new URL(request.url);
		const method = url.pathname === '/health' || url.pathname === '/primes' ? 'GET'
			: ['/stats', '/orders/quote', '/text/top-words'].includes(url.pathname) ? 'POST' : undefined;
		if (method === undefined) return json({ error: 'not found' }, 404);
		if (request.method !== method) return json({ error: 'method not allowed' }, 405, method);
		if (url.pathname === '/health') return new Response('ok', { headers: [['content-type', 'text/plain; charset=utf-8']] });
		if (url.pathname === '/primes') {
			// URLSearchParams.get chooses the first decoded limit; other query fields are ignored.
			const raw = url.searchParams.get('limit');
			if (raw === null || !/^(0|[1-9][0-9]*)(?![\s\S])/.test(raw) || Number(raw) > 5000000) return invalid();
			return json(primes(Number(raw)));
		}
		const source = await request.text();
		let value: unknown;
		try { value = JSON.parse(source); } catch { return invalid(); }
		if (!withinDepth(source)) return invalid();
		if (url.pathname === '/stats') return stats(value) ? json(summarize(value.values)) : invalid();
		if (url.pathname === '/orders/quote') {
			if (!order(value) || (value.couponCode !== undefined && value.couponCode !== 'SAVE10' && value.couponCode !== 'FLAT500')) return invalid();
			return json(quote(value));
		}
		return textRequest(value) ? json(topWords(value)) : invalid();
	},
};
