
/workspace/scratch/immortal-sites/static:     file format elf64-x86-64


Disassembly of section .init:

Disassembly of section .plt:

Disassembly of section .plt.got:

Disassembly of section .text:

0000000000072e80 <adamic_retain>:
   72e80:	48 89 f8             	mov    %rdi,%rax
   72e83:	48 85 ff             	test   %rdi,%rdi
   72e86:	74 28                	je     72eb0 <adamic_retain+0x30>
   72e88:	83 78 08 05          	cmpl   $0x5,0x8(%rax)
   72e8c:	48 89 c7             	mov    %rax,%rdi
   72e8f:	75 0b                	jne    72e9c <adamic_retain+0x1c>
   72e91:	48 8b 78 18          	mov    0x18(%rax),%rdi
   72e95:	48 85 ff             	test   %rdi,%rdi
   72e98:	48 0f 44 f8          	cmove  %rax,%rdi
   72e9c:	83 7f 0c 00          	cmpl   $0x0,0xc(%rdi)
   72ea0:	78 0f                	js     72eb1 <adamic_retain+0x31>
   72ea2:	48 8b 0f             	mov    (%rdi),%rcx
   72ea5:	48 85 c9             	test   %rcx,%rcx
   72ea8:	74 06                	je     72eb0 <adamic_retain+0x30>
   72eaa:	48 ff c1             	inc    %rcx
   72ead:	48 89 0f             	mov    %rcx,(%rdi)
   72eb0:	c3                   	ret
   72eb1:	53                   	push   %rbx
   72eb2:	48 89 c3             	mov    %rax,%rbx
   72eb5:	e8 76 f1 ff ff       	call   72030 <adamic_graph_retain>
   72eba:	48 89 d8             	mov    %rbx,%rax
   72ebd:	5b                   	pop    %rbx
   72ebe:	c3                   	ret

Disassembly of section .fini:
