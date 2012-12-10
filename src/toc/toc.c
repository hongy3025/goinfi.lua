#include <stdio.h>
#include <string.h>

#include "toc.h"

int add(int a, int b) {
	return a+b;
}

void print(char * s, int n) {
	char ss[256];
	strncpy(ss, s, 256);
	printf("%s[%d]\n", ss, n);
}
