// Independent platform implementation; it imports no handler or bridge code.
export default {
    async fetch(request: Request, bindings: Record<string, unknown>): Promise<Response> {
        const projected: Record<string, string> = {};
        for (const key of ['ZONE', 'GREETING']) {
            const candidate = bindings[key];
            if (typeof candidate !== 'string') {
                console.error(`environment binding ${JSON.stringify(key)} is missing or not a string`);
                return new Response('internal error', { status: 500 });
            }
            Object.defineProperty(projected, key, { value: candidate, enumerable: true });
        }
        const greeting = await Promise.resolve(projected.GREETING);
        return new Response([greeting, projected.ZONE, new URL(request.url).pathname, Object.keys(projected).join(',')].join(' '), {
            status: 200, headers: { 'content-type': 'text/plain; charset=utf-8' },
        });
    },
};
