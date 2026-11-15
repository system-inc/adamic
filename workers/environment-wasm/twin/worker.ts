// Independent platform implementation; no generated bridge, crossing or handler imports.
export default {
    async fetch(request: Request, bindings: Record<string, unknown>): Promise<Response> {
        const prefix = bindings.PREFIX;
        if (typeof prefix !== 'string') {
            console.error('environment binding "PREFIX" is missing or not a string');
            return new Response('internal error', { status: 500 });
        }
        const size = new URL(request.url).pathname.length;
        const total = size * (size + 1) / 2;
        return new Response(`${prefix}:${total}`, {
            status: 200, headers: { 'content-type': 'text/plain; charset=utf-8' },
        });
    },
};
