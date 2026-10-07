
/workspace/scratch/immortal-sites/fields:     file format elf64-x86-64


Disassembly of section .init:

Disassembly of section .plt:

Disassembly of section .plt.got:

Disassembly of section .text:

0000000000072e60 <adamic_retain>:
   72e60:	48 89 f8             	mov    %rdi,%rax
   72e63:	48 85 ff             	test   %rdi,%rdi
   72e66:	74 28                	je     72e90 <adamic_retain+0x30>
   72e68:	83 78 08 05          	cmpl   $0x5,0x8(%rax)
   72e6c:	48 89 c7             	mov    %rax,%rdi
   72e6f:	75 0b                	jne    72e7c <adamic_retain+0x1c>
   72e71:	48 8b 78 18          	mov    0x18(%rax),%rdi
   72e75:	48 85 ff             	test   %rdi,%rdi
   72e78:	48 0f 44 f8          	cmove  %rax,%rdi
   72e7c:	83 7f 0c 00          	cmpl   $0x0,0xc(%rdi)
   72e80:	78 0f                	js     72e91 <adamic_retain+0x31>
   72e82:	48 8b 0f             	mov    (%rdi),%rcx
   72e85:	48 85 c9             	test   %rcx,%rcx
   72e88:	74 06                	je     72e90 <adamic_retain+0x30>
   72e8a:	48 ff c1             	inc    %rcx
   72e8d:	48 89 0f             	mov    %rcx,(%rdi)
   72e90:	c3                   	ret
   72e91:	53                   	push   %rbx
   72e92:	48 89 c3             	mov    %rax,%rbx
   72e95:	e8 76 f1 ff ff       	call   72010 <adamic_graph_retain>
   72e9a:	48 89 d8             	mov    %rbx,%rax
   72e9d:	5b                   	pop    %rbx
   72e9e:	c3                   	ret

Disassembly of section .fini:
