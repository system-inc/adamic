let probeCount = 0;
let sink = 0;
export function burn(iterations) {
  let value = sink;
  for (let i = 0; i < iterations; i++) value = (Math.imul(value, 1664525) + 1013904223) | 0;
  sink = value;
}
export default {
  async fetch(request) {
    const path = new URL(request.url).pathname;
    if (path === '/health') return new Response('ok');
    // A reproducible first-use penalty, confined to the warmup proof path.
    if (path === '/warm-probe' && probeCount++ < 8) burn(12000000);
    return new Response(`${request.method} ${path} ${await request.text()}`, {
      headers: { 'content-type': 'text/plain; charset=utf-8', 'x-fixture': 'same' },
    });
  },
};
