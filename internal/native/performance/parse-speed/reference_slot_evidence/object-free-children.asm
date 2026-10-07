
/workspace/scratch/reference-slots/baseline:     file format elf64-x86-64


Disassembly of section .init:

Disassembly of section .plt:

Disassembly of section .plt.got:

Disassembly of section .text:

000000000006e750 <adamic_object_free_children>:
   6e750:	push   %r15
   6e752:	push   %r14
   6e754:	push   %r12
   6e756:	push   %rbx
   6e757:	push   %rax
   6e758:	mov    %rsi,%rbx
   6e75b:	mov    %rdi,%r14
   6e75e:	mov    0x18(%rdi),%r15
   6e762:	test   %r15,%r15
   6e765:	jne    6e788 <adamic_object_free_children+0x38>
   6e767:	mov    0x10(%r14),%rax
   6e76b:	mov    (%rax),%rcx
   6e76e:	test   %rcx,%rcx
   6e771:	je     6e7fe <adamic_object_free_children+0xae>
   6e777:	xor    %r15d,%r15d
   6e77a:	jmp    6e7d8 <adamic_object_free_children+0x88>
   6e77c:	nopl   0x0(%rax)
   6e780:	mov    (%r15),%r15
   6e783:	test   %r15,%r15
   6e786:	je     6e7fe <adamic_object_free_children+0xae>
   6e788:	mov    0x8(%r15),%rax
   6e78c:	mov    0x10(%r15),%r12
   6e790:	cmp    %rax,%r12
   6e793:	jbe    6e780 <adamic_object_free_children+0x30>
   6e795:	mov    0x10(%r14),%rcx
   6e799:	mov    0x10(%rcx),%rcx
   6e79d:	cmpb   $0x1,-0x1(%rcx,%r12,1)
   6e7a3:	lea    -0x1(%r12),%r12
   6e7a8:	jne    6e790 <adamic_object_free_children+0x40>
   6e7aa:	mov    0x28(%r14,%r12,8),%rdi
   6e7af:	test   %rdi,%rdi
   6e7b2:	je     6e790 <adamic_object_free_children+0x40>
   6e7b4:	cmpq   $0x0,(%rdi)
   6e7b8:	je     6e790 <adamic_object_free_children+0x40>
   6e7ba:	call   *%rbx
   6e7bc:	mov    0x8(%r15),%rax
   6e7c0:	jmp    6e790 <adamic_object_free_children+0x40>
   6e7c2:	data16 data16 data16 data16 cs nopw 0x0(%rax,%rax,1)
   6e7d0:	inc    %r15
   6e7d3:	cmp    %rcx,%r15
   6e7d6:	jae    6e7fe <adamic_object_free_children+0xae>
   6e7d8:	mov    0x10(%rax),%rdx
   6e7dc:	cmpb   $0x1,(%rdx,%r15,1)
   6e7e1:	jne    6e7d0 <adamic_object_free_children+0x80>
   6e7e3:	mov    0x28(%r14,%r15,8),%rdi
   6e7e8:	test   %rdi,%rdi
   6e7eb:	je     6e7d0 <adamic_object_free_children+0x80>
   6e7ed:	cmpq   $0x0,(%rdi)
   6e7f1:	je     6e7d0 <adamic_object_free_children+0x80>
   6e7f3:	call   *%rbx
   6e7f5:	mov    0x10(%r14),%rax
   6e7f9:	mov    (%rax),%rcx
   6e7fc:	jmp    6e7d0 <adamic_object_free_children+0x80>
   6e7fe:	add    $0x8,%rsp
   6e802:	pop    %rbx
   6e803:	pop    %r12
   6e805:	pop    %r14
   6e807:	pop    %r15
   6e809:	ret

Disassembly of section .fini:
