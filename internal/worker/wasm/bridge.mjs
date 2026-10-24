import { WASIExit } from './wasi.mjs';

export function createWorker(host) {
  return { async fetch(request, env, context) {
    const url = new URL(request.url);
    const field = ([name, value]) => ({ name, value });
    const input = {
      method: request.method,
      url: request.url,
      path: url.pathname,
      query: Array.from(url.searchParams, field),
      headers: Array.from(request.headers, field),
      body: await request.text(),
    };
    let text;
    try {
      text = host.call(JSON.stringify(input));
    } catch (error) {
      // Panic text was already emitted by fd_write. Other failures need a log.
      if (!(error instanceof WASIExit && error.code === 70)) console.error(error);
      return new Response('internal error', { status: 500 });
    }
    const response = JSON.parse(text);
    const headers = new Headers();
    for (const { name, value } of response.headers) headers.append(name, value);
    return new Response(response.body, { status: response.status, headers });
  } };
}
