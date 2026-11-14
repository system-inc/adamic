// tsbuildPublic.ts:287's shape: createBuilderStatusReporter's optional pretty, captured by the
// closures of the object it returns and read on each report.
interface Reporter {
	readonly report: (message: string) => string;
	readonly mode: () => string;
}

function createReporter(clock: string, pretty?: boolean): Reporter {
	return {
		report: (message: string): string => {
			let output = pretty ? `[${clock}] ` : `${clock} - `;
			output += message;
			return output;
		},
		mode: (): string => `pretty ${pretty} ${pretty === undefined}`,
	};
}

const reporters = [
	createReporter("12:00"),
	createReporter("12:01", true),
	createReporter("12:02", false),
	createReporter("12:03", undefined),
];
for (const reporter of reporters) {
	console.log(`${reporter.report("built")}, ${reporter.mode()}`);
}
