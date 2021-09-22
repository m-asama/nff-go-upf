#include "textflag.h"

// func tsc() uint64
TEXT ·tsc(SB),NOSPLIT,$0-8
	RDTSC
	SHLQ	$32, DX
	ADDQ	DX, AX
	MOVQ	AX, ret+0(FP)
	RET

