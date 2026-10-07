#ifndef ADAMIC_ASYNC_H
#define ADAMIC_ASYNC_H
#include "adamic.h"

typedef struct adamic_async_promise adamic_async_promise;
typedef struct adamic_async_frame adamic_async_frame;
typedef struct adamic_async_reaction adamic_async_reaction;
struct adamic_async_frame {
    adamic_heap heap;
    unsigned state;
    adamic_async_promise *output;
    adamic_async_promise *waiting;
    void (*resume)(adamic_async_frame *, adamic_value, bool);
    void (*children)(adamic_async_frame *, void (*)(void *));
};
struct adamic_async_promise {
    adamic_heap heap;
    bool settled, rejected, references;
    adamic_value value;
    adamic_async_reaction *first, *last;
};
adamic_async_promise *adamic_async_new(void);
void adamic_async_settle(adamic_async_promise *, adamic_value, bool, bool);
void adamic_async_await(adamic_async_frame *, adamic_async_promise *);
void adamic_async_cancel(adamic_async_promise *);
void adamic_async_run(void);
void adamic_async_teardown(void);
void adamic_async_free_children(void *, void (*)(void *));
#endif
