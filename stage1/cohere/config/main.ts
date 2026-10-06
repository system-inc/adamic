// One absolute root per line, followed by tab-separated ignorePatterns. No settings are injected as
// resolved answers: this driver exercises only the project discovery seam named in discovery.ts.
import { panic, programArguments, readTextFile } from 'adamic';
import { discoverProjects } from './discovery.ts';
import { LintGlob } from './glob.ts';
const input = readTextFile(programArguments()[0] ?? panic('usage: main.ts <cases>'));
if (input.kind === 'Error') { panic(input.message); }
for (const line of input.text.split('\n')) {
	if (line === '') { continue; }
	const fields = line.split('\t');
	if (fields[0] === 'glob') {
		const matched = new LintGlob(fields[1] ?? '').matches(fields[2] ?? '');
		console.log(`glob ${matched ? '1' : '0'}`);
		continue;
	}
	const root = fields[0] ?? panic('no root');
	console.log(`root ${root}`);
	const answer = discoverProjects(root, fields.slice(1));
	if (answer.kind === 'Error') { console.log(`error ${answer.message}`); continue; }
	const found = answer.found;
	console.log(`counts ${found.submodules} ${found.ignored} ${found.neverDescended}`);
	for (const project of found.projects) { console.log(`project ${project.directory} ${project.engine}`); }
	for (const refused of found.refused) { console.log(`refused ${refused}`); }
	for (const nested of found.nestedRepositories) { console.log(`nested ${nested}`); }
}
