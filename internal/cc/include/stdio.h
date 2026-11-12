/* stdio.h for nasmx86: the functions of its built-in C library, declared
   as in glibc. */
#ifndef _STDIO_H
#define _STDIO_H 1
#include <stddef.h>
#include <stdarg.h>

typedef struct _IO_FILE FILE;
extern FILE *stdin;
extern FILE *stdout;
extern FILE *stderr;
#define stdin stdin
#define stdout stdout
#define stderr stderr
#define EOF (-1)

int printf(const char *format, ...);
int fprintf(FILE *stream, const char *format, ...);
int dprintf(int fd, const char *format, ...);
int sprintf(char *str, const char *format, ...);
int snprintf(char *str, size_t size, const char *format, ...);
int scanf(const char *format, ...);
int fscanf(FILE *stream, const char *format, ...);
int sscanf(const char *str, const char *format, ...);

int puts(const char *s);
int fputs(const char *s, FILE *stream);
int putchar(int c);
int putc(int c, FILE *stream);
int fputc(int c, FILE *stream);
int getchar(void);
int getc(FILE *stream);
int fgetc(FILE *stream);
int ungetc(int c, FILE *stream);
char *fgets(char *s, int size, FILE *stream);
char *gets(char *s);

FILE *fopen(const char *path, const char *mode);
int fclose(FILE *stream);
int fflush(FILE *stream);
int feof(FILE *stream);
size_t fwrite(const void *ptr, size_t size, size_t nmemb, FILE *stream);

#endif
