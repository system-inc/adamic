// messageWarningComment is the description for a comment that opens with `term`, quoting the comment.
export function messageWarningComment(term: string, quoted: string): string {
    return `This comment opens with \`${term}\`, which marks the code as knowingly unfinished: ${quoted}. A warning comment is a note to a future reader that something was deferred, so it should be tracked somewhere that gets read rather than left where only the next editor of this file will find it.`;
}
