package comments

import "github.com/system-inc/cohere/internal/lint/rule"

// ForFile is All, computed once per file and shared by every rule that asks.
//
// The scan itself was already written once, which is what makes this easy to miss: three rules
// called one well-factored helper, and each call rescanned the file. A shared function is not a
// shared scan.
//
// The cache key moved with the package. It is the identity the cache stores under, so a stale key
// naming the old home is not cosmetic: two packages keying the same scan differently would each
// compute it, which is precisely the cost this exists to remove.
//
// Measured with `cohere --timing` before this existed: consistency-no-shouting 807ms,
// consistency-no-long-line-comment 619ms, consistency-no-single-line-jsdoc 601ms, each visiting
// exactly one node per file. Flat node counts with unequal times is the signature of work done
// outside the walk.
func ForFile(ctx rule.Context) []Comment {
	return rule.Cached(ctx.FileCache, "comments.All", func() []Comment {
		return All(ctx.SourceFile)
	})
}
