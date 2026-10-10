package native

import "github.com/system-inc/adamic/internal/ir"

// Invocation context is thread-local and restored even when the callee returns
// an Adamic exception. Argument expressions run before entering this helper.
const viewCallableBlameRuntime = `
#include <stdlib.h>
#include <string.h>
static _Thread_local const char *adamic_view_call_site;
static inline adamic_value adamic_view_call_at(adamic_closure *closure, void *receiver, adamic_value *arguments, size_t count, size_t slots, const char *site) {
    const char *previous = adamic_view_call_site;
    if (site != NULL) adamic_view_call_site = site;
    adamic_value result = adamic_closure_receiver_call(closure, receiver, arguments, count, slots);
    adamic_view_call_site = previous;
    return result;
}
static inline _Noreturn void adamic_view_call_panic(const char *message) {
    const char *site = adamic_view_call_site == NULL ? "<runtime callback>" : adamic_view_call_site;
    size_t a = strlen(message), b = strlen(site);
    char *complete = malloc(a + b + sizeof " call at ");
    if (complete == NULL) adamic_panic(message, a);
    memcpy(complete, message, a);
    memcpy(complete + a, " call at ", sizeof " call at " - 1);
    memcpy(complete + a + sizeof " call at " - 1, site, b + 1);
    adamic_panic(complete, a + b + sizeof " call at " - 1);
}
`

func (e *emitter) viewCallableBlameRuntime() {
	for _, declaration := range e.declarations {
		if declaration == viewCallableBlameRuntime {
			return
		}
	}
	e.declarations = append(e.declarations, viewCallableBlameRuntime)
}

func viewCallableSourceSite(call ir.CallClosure) string {
	if call.CallWhere == "" || call.ArgumentCount != nil {
		return "NULL"
	}
	return cString(call.CallWhere)
}
