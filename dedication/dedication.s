// dedication.s: write one line, exit 0.

	.text
	.globl	_main
	.p2align	2
_main:
	mov	x0, #1			// stdout
	adr	x1, msg
	mov	x2, #len
	bl	_write
	mov	x0, #0
	bl	_exit

msg:	.ascii	"In dedication, with gratitude, to Kenneth Lane Thompson, whose work we stand on. - Kirk and Ahra\n"
	.set	len, . - msg
