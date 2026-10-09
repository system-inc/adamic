package structure

import (
	"strings"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/sourcename"
	"github.com/system-inc/cohere/policy"
)

// The rule's messages, one handle per id, whose wording lives in
// `policy/messages/network-require-hook-request-suffix.json`.
var (
	networkRequireHookRequestSuffixHookShouldEndWithRequestText = policy.MessageOf("structure/network-require-hook-request-suffix", "hookShouldEndWithRequest")
	networkRequireHookRequestSuffixFileShouldEndWithRequestText = policy.MessageOf("structure/network-require-hook-request-suffix", "fileShouldEndWithRequest")
)

// messageHookShouldEndWithRequest is the finding, rendered when it is reported so the text comes from the current catalog.
func messageHookShouldEndWithRequest() rule.Message {
	return rule.Message{
		Id:          networkRequireHookRequestSuffixHookShouldEndWithRequestText.Id,
		Description: networkRequireHookRequestSuffixHookShouldEndWithRequestText.Render(nil),
	}
}

// messageFileShouldEndWithRequest is the finding, rendered when it is reported so the text comes from the current catalog.
func messageFileShouldEndWithRequest() rule.Message {
	return rule.Message{
		Id:          networkRequireHookRequestSuffixFileShouldEndWithRequestText.Id,
		Description: networkRequireHookRequestSuffixFileShouldEndWithRequestText.Render(nil),
	}
}

// NetworkRequireHookRequestSuffix flags a NetworkService hook, or its file, not named `*Request`.
//
//	valid:   // UserRequest.ts       export function useUserRequest() { networkService.useGraphQlQuery(...) }
//	invalid: // UserRequest.ts       export function useUser() { networkService.useGraphQlQuery(...) }
//	invalid: // User.ts             export function useUserRequest() { networkService.useGraphQlQuery(...) }
//
// Two message ids because they are two different repairs. A hook name is fixed by a rename in this
// file; a file name is fixed by moving the file and updating every importer, and a reader who sees
// one message for both has no way to tell which is being asked of them.
//
// The file finding reports on the first hook's identifier rather than on the file, which is the
// original's choice and worth keeping: a finding needs a range, and a whole-file range underlines
// everything and points at nothing. The first hook is where a reader will start reading anyway.
//
// The suggestion strips a trailing Query, Mutation, or Hook before appending Request, so
// `useUserQuery` is suggested as `useUserRequest` rather than `useUserQueryRequest`. Only one suffix
// is stripped, matching the original: chaining the strips would rewrite `useQueryHook` past what
// anyone asked for.
var NetworkRequireHookRequestSuffix = rule.Rule{
	Name:       "structure/network-require-hook-request-suffix",
	NoListener: rule.NoListenerAnswersInRun,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		analysis := NetworkFileAnalysisFor(ctx)
		if len(analysis.HookDeclarations) == 0 {
			// Declining the file rather than registering a listener that would find nothing. A file
			// with no NetworkService hooks is the overwhelming majority of files.
			return nil
		}

		for _, hook := range analysis.HookDeclarations {
			if !strings.HasSuffix(hook.Name, "Request") {
				ctx.ReportNode(hook.NameNode, messageHookShouldEndWithRequest())
			}
		}

		// The file check reports at most once, against the first hook, even when several hooks are
		// declared. The defect is the file's name, and reporting it once per hook would say the same
		// thing three times about one rename.
		if fileNeedsRequestSuffix(ctx.SourceFile.FileName().AsString()) {
			ctx.ReportNode(analysis.HookDeclarations[0].NameNode, messageFileShouldEndWithRequest())
		}

		return nil
	},
}

// fileNeedsRequestSuffix reports whether a TypeScript file's base name lacks the Request suffix.
//
// Non-TypeScript files are exempt rather than flagged. The original tests the extension explicitly,
// and the reason holds: this rule is about how a module announces itself to importers, and something
// that is not a module here is not making that announcement.
//
// Both separators are folded before taking the base name, the same normalization FileContextFor
// does and for the same reason: every test below is a suffix test, and a backslash-separated path
// would take the whole path as its base name and silently pass.
func fileNeedsRequestSuffix(fileName string) bool {
	normalizedPath := strings.ReplaceAll(fileName, `\`, "/")
	base := normalizedPath
	if index := strings.LastIndex(normalizedPath, "/"); index >= 0 {
		base = normalizedPath[index+1:]
	}

	// An Adamic `.a` hook file is named as the `.ts` it is (#kwt1htp).
	if stem, isTypeScript := strings.CutSuffix(sourcename.TreatedAs(base), ".ts"); isTypeScript {
		return !strings.HasSuffix(stem, "Request")
	}
	if stem, isTSX := strings.CutSuffix(base, ".tsx"); isTSX {
		return !strings.HasSuffix(stem, "Request")
	}
	return false
}
