/* stdlib.h for nasmx86: the functions of its built-in C library, declared
   as in glibc. */
#ifndef _STDLIB_H
#define _STDLIB_H 1
#include <stddef.h>

#define EXIT_SUCCESS 0
#define EXIT_FAILURE 1
#define RAND_MAX 2147483647

void *malloc(size_t size);
void *calloc(size_t nmemb, size_t size);
void *realloc(void *ptr, size_t size);
void free(void *ptr);

void exit(int status) __attribute__((__noreturn__));
void abort(void) __attribute__((__noreturn__));

int atoi(const char *s);
long atol(const char *s);
long long atoll(const char *s);
long strtol(const char *s, char **endptr, int base);
unsigned long strtoul(const char *s, char **endptr, int base);
int abs(int j);
long labs(long j);

int rand(void);
void srand(unsigned int seed);

#endif
