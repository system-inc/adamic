// Independent platform implementation: no imports from the Adamic handler or bridge.
class HandlerPanic extends Error {}
async function pause(): Promise<void> { await Promise.resolve(); }
async function rejectError(): Promise<string> { return await Promise.reject(new Error('helper rejected')); }
async function throwError(): Promise<string> { throw new Error('helper threw'); }
async function rejectPanic(): Promise<string> {
    await pause();
    throw new HandlerPanic('helper panic');
}
export default {
    async fetch(request: Request): Promise<Response> {
        const route = new URL(request.url).pathname;
        const payload = await request.text();
        try {
            if (route === '/panic') {
                await pause();
                throw new HandlerPanic('after await');
            }
            if (route === '/reject') return new Response(await rejectError());
            if (route === '/throw-error') return new Response(await throwError());
            if (route === '/helper-panic') return new Response(await rejectPanic());
            await pause();
            return new Response([request.method, route, payload].join(' '), {
                status: 200, headers: { 'content-type': 'text/plain; charset=utf-8' },
            });
        } catch (error) {
            if (!(error instanceof HandlerPanic)) throw error;
            console.error(error.message);
            return new Response('internal error', { status: 500 });
        }
    },
};
