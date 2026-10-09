import { fileURLToPath } from 'node:url';
import { replay } from './replay.mjs';

// Each implementation meets the same recording.
const workers = ['compute/twin/worker.ts', 'compute/handler.a'];
for (const worker of workers) {
	const path = (relative) => fileURLToPath(new URL(relative, import.meta.url));
	if (!await replay(path(worker), path('compute/corpus/requests.jsonl'), '--compare', path('compute/corpus/responses.jsonl'))) {
		process.exitCode = 1;
		break;
	}
}
