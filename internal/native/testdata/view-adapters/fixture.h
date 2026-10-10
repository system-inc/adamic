#ifndef VIEW_ADAPTER_FIXTURE_H
#define VIEW_ADAPTER_FIXTURE_H
#include "adamic.h"
#include "count.h"
#include <stdio.h>
#include <stdlib.h>

static size_t made;
static const int view_a, view_b;
static adamic_value producer(adamic_closure *self, adamic_value *arguments) {
    (void)self;
    return arguments[0];
}
static adamic_value invoke(adamic_closure *self, adamic_value *arguments) {
    return adamic_closure_call(self->view->underlying, arguments, 1);
}
static adamic_closure *make_adapter(adamic_closure *underlying, const void *view_key) {
    (void)underlying; (void)view_key;
    made++;
    return adamic_closure_new(invoke, 0);
}
static void require(bool condition, const char *message) {
    if (!condition) { fprintf(stderr, "%s\n", message); exit(1); }
}
#endif
