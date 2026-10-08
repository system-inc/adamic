// Harness-only substitute. This process is explicitly not workerd.
import http from 'node:http';
import { pathToFileURL } from 'node:url';
const worker = (await import(pathToFileURL(process.argv[2]))).default;
const server = http.createServer(async (incoming, outgoing) => {
  try {
    const chunks = [];
    for await (const chunk of incoming) chunks.push(chunk);
    const method = incoming.method;
    const request = new Request(`http://${incoming.headers.host}${incoming.url}`, {
      method, headers: incoming.headers,
      ...(method === 'GET' || method === 'HEAD' ? {} : { body: Buffer.concat(chunks) }),
    });
    const response = await worker.fetch(request, {}, { waitUntil() {} });
    outgoing.writeHead(response.status, Object.fromEntries(response.headers));
    outgoing.end(Buffer.from(await response.arrayBuffer()));
  } catch (error) {
    console.error(error);
    outgoing.writeHead(500);
    outgoing.end('adapter error');
  }
});
server.listen(Number(process.argv[3]), '127.0.0.1');
