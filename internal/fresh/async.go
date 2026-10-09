package fresh

import "github.com/system-inc/adamic/internal/ir"

// RuntimeBreaksCycles accepts only compiler-owned protocol identities. It does not prove any
// source payload write fresh, and it must never inspect a name or structural shape. The first async
// lowering permits only primitive payloads, so no user object write can bypass ProveWrites.
func RuntimeBreaksCycles(generated *ir.GeneratedType) bool { return generated.RuntimeBreaksCycles() }
