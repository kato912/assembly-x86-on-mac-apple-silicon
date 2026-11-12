/* unistd.h for nasmx86: the functions of its built-in C library. */
#ifndef _UNISTD_H
#define _UNISTD_H 1

unsigned int sleep(unsigned int seconds);
int usleep(unsigned int usec);
void _exit(int status) __attribute__((__noreturn__));

#endif
