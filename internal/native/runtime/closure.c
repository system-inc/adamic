// closure.c: closures and the cells they capture.

#include "adamic.h"

static void view_adapter_initialize(adamic_closure *closure);

adamic_cell *adamic_cell_new(adamic_value value, bool references) {
	adamic_cell *cell = adamic_allocate(sizeof *cell, adamic_kind_cell);
	cell->references = references;
	cell->ready = true;
#ifdef ADAMIC_CANONICAL_CLOSURES
	cell->owner = NULL;
#endif
	cell->value = value;
	return cell;
}

adamic_closure *(adamic_closure_new)(adamic_code code, size_t count) {
	adamic_closure *closure = adamic_allocate(sizeof *closure + count * sizeof closure->cells[0], adamic_kind_closure);
	view_adapter_initialize(closure);
	closure->code = code;
#ifdef ADAMIC_CLOSURE_CONVENTION
 closure->counted = false;
#endif
#ifdef ADAMIC_CLOSURE_RECEIVERS
 closure->receiver = false;
#endif
	closure->count = count;
#ifdef ADAMIC_CANONICAL_CLOSURES
	closure->canonical_owner = NULL;
	closure->canonical_previous = NULL;
	closure->canonical_next = NULL;
#endif
	for (size_t index = 0; index < count; index++) {
		closure->cells[index] = NULL;
	}
	return closure;
}

#ifdef ADAMIC_CANONICAL_CLOSURES
adamic_environment *adamic_environment_new(size_t count) {
	adamic_environment *environment = adamic_allocate(sizeof *environment + count * sizeof environment->cells[0], adamic_kind_environment);
	environment->count = count;
	environment->functions = NULL;
	for (size_t index = 0; index < count; index++) {
		environment->cells[index] = (adamic_cell){.heap = {.kind = adamic_kind_cell}, .owner = environment};
	}
	return environment;
}

// A live value is unique for its code and activation. If no owner kept it, a
// later reference can recreate it: no surviving reference can observe that gap.
// Captured cells keep the frame alive; its cache never keeps a closure alive.
adamic_closure *(adamic_closure_canonical)(adamic_cell *identity, adamic_code code, size_t count, adamic_cell *const cells[]) {
	adamic_environment *owner = identity->owner;
	for (adamic_closure *held = owner->functions; held != NULL; held = held->canonical_next) {
#ifdef ADAMIC_CLOSURE_CONVENTION
        if (held->counted) { continue; }
#endif
		if (held->code == code) {
			return adamic_retain(held);
		}
	}
	adamic_closure *closure = adamic_closure_new(code, count);
	for (size_t index = 0; index < count; index++) {
		closure->cells[index] = adamic_retain(cells[index]);
	}
	closure->canonical_owner = owner;
	closure->canonical_next = owner->functions;
	if (owner->functions != NULL) {
		owner->functions->canonical_previous = closure;
	}
	owner->functions = closure;
	return closure;
}


#ifdef ADAMIC_CLOSURE_CONVENTION
adamic_closure *(adamic_counted_closure_canonical)(adamic_cell *identity, adamic_counted_code code, size_t count, adamic_cell *const cells[]) {
 adamic_environment *owner = identity->owner;
 for (adamic_closure *held = owner->functions; held != NULL; held = held->canonical_next) {
 if (held->counted && held->counted_code == code) { return adamic_retain(held); }
 }
 adamic_closure *closure = adamic_counted_closure_new(code, count);
 for (size_t index = 0; index < count; index++) { closure->cells[index] = adamic_retain(cells[index]); }
 closure->canonical_owner = owner;
 closure->canonical_next = owner->functions;
 if (owner->functions != NULL) { owner->functions->canonical_previous = closure; }
 owner->functions = closure;
 return closure;
}
#endif

void adamic_closure_uncache(adamic_closure *closure) {
	adamic_environment *owner = closure->canonical_owner;
	if (owner == NULL) {
		return;
	}
	if (closure->canonical_previous != NULL) {
		closure->canonical_previous->canonical_next = closure->canonical_next;
	} else {
		owner->functions = closure->canonical_next;
	}
	if (closure->canonical_next != NULL) {
		closure->canonical_next->canonical_previous = closure->canonical_previous;
	}
}

#endif

static void view_adapter_initialize(adamic_closure *closure) {
	closure->flags = 0;
	closure->view_underlying = NULL;
	closure->view_key = NULL;
	closure->view_next = 0;
}

// One weak table for all views. Hidden links keep LeakSanitizer from treating
// this non-owning cache as a root. Every traversal and mutation holds the mutex.
#ifndef ADAMIC_TARGET_WASI
#include <pthread.h>
static pthread_mutex_t view_adapter_lock = PTHREAD_MUTEX_INITIALIZER;
#endif
static uintptr_t view_adapter_first;
#define VIEW_ADAPTER_HIDDEN ((uintptr_t)0xa5a5a5a5a5a5a5a0u)

static uintptr_t view_adapter_hide(adamic_closure *value) {
	return value == NULL ? 0 : (uintptr_t)value ^ VIEW_ADAPTER_HIDDEN;
}

static adamic_closure *view_adapter_reveal(uintptr_t value) {
	return value == 0 ? NULL : (adamic_closure *)(value ^ VIEW_ADAPTER_HIDDEN);
}

static void view_adapter_lock_table(void) {
#ifndef ADAMIC_TARGET_WASI
	if (pthread_mutex_lock(&view_adapter_lock) != 0) {
		static const char message[] = "cannot lock callable adapter cache";
		adamic_panic(message, sizeof message - 1);
	}
#endif
}

static void view_adapter_unlock_table(void) {
#ifndef ADAMIC_TARGET_WASI
	if (pthread_mutex_unlock(&view_adapter_lock) != 0) {
		static const char message[] = "cannot unlock callable adapter cache";
		adamic_panic(message, sizeof message - 1);
	}
#endif
}

// On main's single-thread heap this is a try-retain under the table lock, not
// a claim that the underlying heap's plain counts are thread-safe. Condition 3
// supplies the shared-count try-retain when runtime's concurrency slice lands.
static adamic_closure *view_adapter_try_retain(adamic_closure *value) {
	if (value->heap.references == 0) {
		return NULL;
	}
	return adamic_retain(value);
}

static adamic_closure *view_adapter_find_locked(adamic_closure *underlying, const void *view_key) {
	for (adamic_closure *held = view_adapter_reveal(view_adapter_first); held != NULL; held = view_adapter_reveal(held->view_next)) {
		if (held->view_underlying == underlying && held->view_key == view_key) {
			adamic_closure *retained = view_adapter_try_retain(held);
			if (retained != NULL) { return retained; }
		}
	}
	return NULL;
}

adamic_closure *adamic_view_adapter_underlying(adamic_closure *value) {
	return value != NULL && (value->flags & ADAMIC_CLOSURE_VIEW_ADAPTER) != 0 ? value->view_underlying : value;
}

adamic_closure *adamic_view_adapter_intern(adamic_closure *underlying, const void *view_key, adamic_closure *(*make)(adamic_closure *, const void *)) {
	underlying = adamic_view_adapter_underlying(underlying);
	view_adapter_lock_table();
	adamic_closure *held = view_adapter_find_locked(underlying, view_key);
	view_adapter_unlock_table();
	if (held != NULL) { return held; }

	// A factory may itself read a view. It never executes while the lock is held.
	adamic_closure *fresh = make(underlying, view_key);
	if (fresh == NULL || fresh == underlying || fresh->heap.kind != adamic_kind_closure || fresh->heap.references != 1 || (fresh->flags & ADAMIC_CLOSURE_VIEW_ADAPTER) != 0
#ifdef ADAMIC_CANONICAL_CLOSURES
		|| fresh->canonical_owner != NULL
#endif
	) {
		static const char message[] = "callable adapter factory must return a fresh noncanonical counted closure";
		adamic_panic(message, sizeof message - 1);
	}
	fresh->flags |= ADAMIC_CLOSURE_VIEW_ADAPTER;
	fresh->view_underlying = adamic_retain(underlying);
	fresh->view_key = view_key;
	fresh->view_next = 0;

	view_adapter_lock_table();
	held = view_adapter_find_locked(underlying, view_key);
	if (held == NULL) {
		fresh->view_next = view_adapter_first;
		view_adapter_first = view_adapter_hide(fresh);
	}
	view_adapter_unlock_table();
	if (held != NULL) {
		// Release the losing candidate only after unlocking: its free path locks.
		adamic_release(fresh);
		return held;
	}
	return fresh;
}

void adamic_view_adapter_forget(adamic_closure *adapter) {
	view_adapter_lock_table();
	uintptr_t *entry = &view_adapter_first;
	while (*entry != 0) {
		adamic_closure *held = view_adapter_reveal(*entry);
		if (held == adapter) {
			*entry = held->view_next;
			break;
		}
		entry = &held->view_next;
	}
	adapter->view_next = 0;
	view_adapter_unlock_table();
}
