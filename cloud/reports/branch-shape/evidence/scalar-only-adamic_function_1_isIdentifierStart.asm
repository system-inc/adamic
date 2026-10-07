
/workspace/scratch/branch-shape/scalar-only:     file format elf64-x86-64


Disassembly of section .init:

Disassembly of section .plt:

Disassembly of section .plt.got:

Disassembly of section .text:

0000000000037920 <adamic_function_1_isIdentifierStart>:
   37920:	push   %rbp
   37921:	mov    %rsp,%rbp
   37924:	push   %r14
   37926:	push   %rbx
   37927:	sub    $0x10,%rsp
   3792b:	lea    0xa184e(%rip),%rax        # d9180 <adamic_stack_limit>
   37932:	cmp    %rbp,(%rax)
   37935:	ja     379fa <adamic_function_1_isIdentifierStart+0xda>
   3793b:	movsd  0x4c8b5(%rip),%xmm2        # 841f8 <_IO_stdin_used+0x1f8>
   37943:	movapd %xmm0,%xmm1
   37947:	cmplepd %xmm2,%xmm1
   3794c:	movsd  0x4c7a4(%rip),%xmm2        # 840f8 <_IO_stdin_used+0xf8>
   37954:	cmplepd %xmm0,%xmm2
   37959:	andpd  %xmm1,%xmm2
   3795d:	movd   %xmm2,%eax
   37961:	mov    $0x1,%bl
   37963:	test   $0x1,%al
   37965:	jne    379ef <adamic_function_1_isIdentifierStart+0xcf>
   3796b:	ucomisd 0x4c88d(%rip),%xmm0        # 84200 <_IO_stdin_used+0x200>
   37973:	jne    37977 <adamic_function_1_isIdentifierStart+0x57>
   37975:	jnp    379ef <adamic_function_1_isIdentifierStart+0xcf>
   37977:	ucomisd 0x4c849(%rip),%xmm0        # 841c8 <_IO_stdin_used+0x1c8>
   3797f:	jne    37983 <adamic_function_1_isIdentifierStart+0x63>
   37981:	jnp    379ef <adamic_function_1_isIdentifierStart+0xcf>
   37983:	movsd  0x4c865(%rip),%xmm2        # 841f0 <_IO_stdin_used+0x1f0>
   3798b:	movapd %xmm0,%xmm1
   3798f:	cmplepd %xmm2,%xmm1
   37994:	movsd  0x4c73c(%rip),%xmm2        # 840d8 <_IO_stdin_used+0xd8>
   3799c:	cmplepd %xmm0,%xmm2
   379a1:	andpd  %xmm1,%xmm2
   379a5:	movd   %xmm2,%eax
   379a9:	test   $0x1,%al
   379ab:	jne    379ef <adamic_function_1_isIdentifierStart+0xcf>
   379ad:	ucomisd 0x4c67b(%rip),%xmm0        # 84030 <_IO_stdin_used+0x30>
   379b5:	jae    379bb <adamic_function_1_isIdentifierStart+0x9b>
   379b7:	xor    %ebx,%ebx
   379b9:	jmp    379ef <adamic_function_1_isIdentifierStart+0xcf>
   379bb:	movapd %xmm0,-0x20(%rbp)
   379c0:	cmpb   $0x0,0x89e09(%rip)        # c17d0 <adamic_ready_0>
   379c7:	je     379ff <adamic_function_1_isIdentifierStart+0xdf>
   379c9:	mov    0x89df8(%rip),%rdi        # c17c8 <adamic_global_0_identifierStart>
   379d0:	call   6ed50 <adamic_retain>
   379d5:	mov    %rax,%r14
   379d8:	mov    %rax,%rdi
   379db:	movapd -0x20(%rbp),%xmm0
   379e0:	call   37ac0 <adamic_function_0_contains>
   379e5:	mov    %eax,%ebx
   379e7:	mov    %r14,%rdi
   379ea:	call   6ed70 <adamic_release>
   379ef:	mov    %ebx,%eax
   379f1:	add    $0x10,%rsp
   379f5:	pop    %rbx
   379f6:	pop    %r14
   379f8:	pop    %rbp
   379f9:	ret
   379fa:	call   7f2e0 <adamic_stack_overflow>
   379ff:	lea    0x5012a(%rip),%rdi        # 87b30 <adamic_function_1_isIdentifierStart.message>
   37a06:	mov    $0x45,%esi
   37a0b:	call   67c10 <adamic_panic>

Disassembly of section .fini:
