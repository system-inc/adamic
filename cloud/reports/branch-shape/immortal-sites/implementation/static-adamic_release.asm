
/workspace/scratch/immortal-sites/static:     file format elf64-x86-64


Disassembly of section .init:

Disassembly of section .plt:

Disassembly of section .plt.got:

Disassembly of section .text:

00000000000730c0 <adamic_release>:
   730c0:	48 85 ff             	test   %rdi,%rdi
   730c3:	74 3f                	je     73104 <adamic_release+0x44>
   730c5:	83 7f 08 05          	cmpl   $0x5,0x8(%rdi)
   730c9:	75 0b                	jne    730d6 <adamic_release+0x16>
   730cb:	48 8b 47 18          	mov    0x18(%rdi),%rax
   730cf:	48 85 c0             	test   %rax,%rax
   730d2:	48 0f 45 f8          	cmovne %rax,%rdi
   730d6:	53                   	push   %rbx
   730d7:	83 7f 0c 00          	cmpl   $0x0,0xc(%rdi)
   730db:	78 12                	js     730ef <adamic_release+0x2f>
   730dd:	48 8b 07             	mov    (%rdi),%rax
   730e0:	48 85 c0             	test   %rax,%rax
   730e3:	74 1e                	je     73103 <adamic_release+0x43>
   730e5:	48 ff c8             	dec    %rax
   730e8:	48 89 07             	mov    %rax,(%rdi)
   730eb:	75 16                	jne    73103 <adamic_release+0x43>
   730ed:	eb 0f                	jmp    730fe <adamic_release+0x3e>
   730ef:	48 89 fb             	mov    %rdi,%rbx
   730f2:	e8 f9 f6 ff ff       	call   727f0 <adamic_graph_release_last>
   730f7:	48 89 df             	mov    %rbx,%rdi
   730fa:	84 c0                	test   %al,%al
   730fc:	74 05                	je     73103 <adamic_release+0x43>
   730fe:	e8 0d 00 00 00       	call   73110 <release_last>
   73103:	5b                   	pop    %rbx
   73104:	c3                   	ret

Disassembly of section .fini:
