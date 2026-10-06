// ParseVersionRange validation. The pinned production Go oracle is a dev build, which admits pins.
import { quoted } from './json.ts';
function versionError(text: string): string {
	const plus = text.indexOf('+');
	const withoutBuild = plus < 0 ? text : text.slice(0, plus);
	const dash = withoutBuild.indexOf('-');
	const core = dash < 0 ? withoutBuild : withoutBuild.slice(0, dash);
	const parts = core.split('.');
	if (parts.length !== 3) { return `${quoted(text)} is not a version: write all three parts, like 1.4.0`; }
	for (const part of parts) {
		let digits = part;
		if (digits.startsWith('+') || digits.startsWith('-')) { digits = digits.slice(1); }
		let valid = digits !== '';
		for (const digit of digits) { if (digit < '0' || digit > '9') { valid = false; } }
		const value = Number(part);
		if (!valid || value < 0 || value >= 9223372036854775808 || (part.length > 1 && part[0] === '0')) { return `${quoted(text)} is not a version: ${quoted(part)} is not a whole number without leading zeros`; }
	}
	if (dash >= 0 && dash + 1 === withoutBuild.length) { return `${quoted(text)} is not a version: the prerelease after \`-\` is empty`; }
	return '';
}
export function versionRangeError(text: string): string {
	const trimmed = text.trim();
	if (trimmed === '') { return 'the "cohere" range is empty: write the releases this project accepts, like "^1.0.0"'; }
	for (const alternative of trimmed.split('||')) {
		const fields = alternative.split('\t').join(' ').split('\n').join(' ').split('\r').join(' ').split(' ').filter((field) => field !== '');
		if (fields.length === 0) { return `the "cohere" range ${quoted(trimmed)} has an empty alternative around \`||\``; }
		for (const field of fields) {
			if (field === '-' || ((field.includes('x') || field.includes('X') || field.includes('*')) && !field.includes('-') && !field.includes('+'))) { return `the "cohere" range ${quoted(trimmed)}: ${quoted(field)} is a wildcard or hyphen range, which this reader does not take; write the bounds, like ">=1.2.0 <2.0.0"`; }
			let rest = field;
			for (const operator of ['>=', '<=', '>', '<', '=', '^', '~']) { if (field.startsWith(operator)) { rest = field.slice(operator.length); break; } }
			const error = versionError(rest);
			if (error !== '') { return `the "cohere" range ${quoted(trimmed)}: ${error}`; }
		}
	}
	return '';
}
