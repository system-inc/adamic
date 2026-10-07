
/workspace/scratch/immortal-sites/before:     file format elf64-x86-64


Disassembly of section .init:

Disassembly of section .plt:

Disassembly of section .plt.got:

Disassembly of section .text:

0000000000074000 <adamic_release>:
   74000:	48 85 ff             	test   %rdi,%rdi
   74003:	74 3f                	je     74044 <adamic_release+0x44>
   74005:	83 7f 08 05          	cmpl   $0x5,0x8(%rdi)
   74009:	75 0b                	jne    74016 <adamic_release+0x16>
   7400b:	48 8b 47 18          	mov    0x18(%rdi),%rax
   7400f:	48 85 c0             	test   %rax,%rax
   74012:	48 0f 45 f8          	cmovne %rax,%rdi
   74016:	53                   	push   %rbx
   74017:	83 7f 0c 00          	cmpl   $0x0,0xc(%rdi)
   7401b:	78 12                	js     7402f <adamic_release+0x2f>
   7401d:	48 8b 07             	mov    (%rdi),%rax
   74020:	48 85 c0             	test   %rax,%rax
   74023:	74 1e                	je     74043 <adamic_release+0x43>
   74025:	48 ff c8             	dec    %rax
   74028:	48 89 07             	mov    %rax,(%rdi)
   7402b:	75 16                	jne    74043 <adamic_release+0x43>
   7402d:	eb 0f                	jmp    7403e <adamic_release+0x3e>
   7402f:	48 89 fb             	mov    %rdi,%rbx
   74032:	e8 f9 f6 ff ff       	call   73730 <adamic_graph_release_last>
   74037:	48 89 df             	mov    %rbx,%rdi
   7403a:	84 c0                	test   %al,%al
   7403c:	74 05                	je     74043 <adamic_release+0x43>
   7403e:	e8 0d 00 00 00       	call   74050 <release_last>
   74043:	5b                   	pop    %rbx
   74044:	c3                   	ret

Disassembly of section .fini:
