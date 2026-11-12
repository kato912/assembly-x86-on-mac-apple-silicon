/* ctype.h for nasmx86: the functions of its built-in C library. As in
   glibc, a true result is the character class's bit (isdigit('5') is
   2048), not 1. */
#ifndef _CTYPE_H
#define _CTYPE_H 1

int isalnum(int c);
int isalpha(int c);
int isdigit(int c);
int islower(int c);
int isupper(int c);
int isxdigit(int c);
int isspace(int c);
int isprint(int c);
int ispunct(int c);
int tolower(int c);
int toupper(int c);

#endif
