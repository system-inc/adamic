
/workspace/scratch/immortal-sites/before:     file format elf64-x86-64


Disassembly of section .init:

Disassembly of section .plt:

Disassembly of section .plt.got:

Disassembly of section .text:

0000000000073dc0 <adamic_retain>:
   73dc0:	48 89 f8             	mov    %rdi,%rax
   73dc3:	48 85 ff             	test   %rdi,%rdi
   73dc6:	74 28                	je     73df0 <adamic_retain+0x30>
   73dc8:	83 78 08 05          	cmpl   $0x5,0x8(%rax)
   73dcc:	48 89 c7             	mov    %rax,%rdi
   73dcf:	75 0b                	jne    73ddc <adamic_retain+0x1c>
   73dd1:	48 8b 78 18          	mov    0x18(%rax),%rdi
   73dd5:	48 85 ff             	test   %rdi,%rdi
   73dd8:	48 0f 44 f8          	cmove  %rax,%rdi
   73ddc:	83 7f 0c 00          	cmpl   $0x0,0xc(%rdi)
   73de0:	78 0f                	js     73df1 <adamic_retain+0x31>
   73de2:	48 8b 0f             	mov    (%rdi),%rcx
   73de5:	48 85 c9             	test   %rcx,%rcx
   73de8:	74 06                	je     73df0 <adamic_retain+0x30>
   73dea:	48 ff c1             	inc    %rcx
   73ded:	48 89 0f             	mov    %rcx,(%rdi)
   73df0:	c3                   	ret
   73df1:	53                   	push   %rbx
   73df2:	48 89 c3             	mov    %rax,%rbx
   73df5:	e8 76 f1 ff ff       	call   72f70 <adamic_graph_retain>
   73dfa:	48 89 d8             	mov    %rbx,%rax
   73dfd:	5b                   	pop    %rbx
   73dfe:	c3                   	ret

Disassembly of section .fini:
