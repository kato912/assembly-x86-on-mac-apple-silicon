/* time.h for nasmx86: the functions of its built-in C library. */
#ifndef _TIME_H
#define _TIME_H 1
#include <stddef.h>

typedef long time_t;
time_t time(time_t *t);

#endif
