
/workspace/scratch/immortal-sites/fields:     file format elf64-x86-64


Disassembly of section .init:

Disassembly of section .plt:

Disassembly of section .plt.got:

Disassembly of section .text:

00000000000730a0 <adamic_release>:
   730a0:	48 85 ff             	test   %rdi,%rdi
   730a3:	74 3f                	je     730e4 <adamic_release+0x44>
   730a5:	83 7f 08 05          	cmpl   $0x5,0x8(%rdi)
   730a9:	75 0b                	jne    730b6 <adamic_release+0x16>
   730ab:	48 8b 47 18          	mov    0x18(%rdi),%rax
   730af:	48 85 c0             	test   %rax,%rax
   730b2:	48 0f 45 f8          	cmovne %rax,%rdi
   730b6:	53                   	push   %rbx
   730b7:	83 7f 0c 00          	cmpl   $0x0,0xc(%rdi)
   730bb:	78 12                	js     730cf <adamic_release+0x2f>
   730bd:	48 8b 07             	mov    (%rdi),%rax
   730c0:	48 85 c0             	test   %rax,%rax
   730c3:	74 1e                	je     730e3 <adamic_release+0x43>
   730c5:	48 ff c8             	dec    %rax
   730c8:	48 89 07             	mov    %rax,(%rdi)
   730cb:	75 16                	jne    730e3 <adamic_release+0x43>
   730cd:	eb 0f                	jmp    730de <adamic_release+0x3e>
   730cf:	48 89 fb             	mov    %rdi,%rbx
   730d2:	e8 f9 f6 ff ff       	call   727d0 <adamic_graph_release_last>
   730d7:	48 89 df             	mov    %rbx,%rdi
   730da:	84 c0                	test   %al,%al
   730dc:	74 05                	je     730e3 <adamic_release+0x43>
   730de:	e8 0d 00 00 00       	call   730f0 <release_last>
   730e3:	5b                   	pop    %rbx
   730e4:	c3                   	ret

Disassembly of section .fini:
