
/workspace/scratch/branch-shape/scalar:     file format elf64-x86-64


Disassembly of section .init:

Disassembly of section .plt:

Disassembly of section .plt.got:

Disassembly of section .text:

0000000000037900 <adamic_function_1_isIdentifierStart>:
   37900:	push   %rbp
   37901:	mov    %rsp,%rbp
   37904:	push   %r14
   37906:	push   %rbx
   37907:	sub    $0x10,%rsp
   3790b:	lea    0xa186e(%rip),%rax        # d9180 <adamic_stack_limit>
   37912:	cmp    %rbp,(%rax)
   37915:	ja     379da <adamic_function_1_isIdentifierStart+0xda>
   3791b:	movsd  0x4c8b5(%rip),%xmm2        # 841d8 <_IO_stdin_used+0x1d8>
   37923:	movapd %xmm0,%xmm1
   37927:	cmplepd %xmm2,%xmm1
   3792c:	movsd  0x4c7c4(%rip),%xmm2        # 840f8 <_IO_stdin_used+0xf8>
   37934:	cmplepd %xmm0,%xmm2
   37939:	andpd  %xmm1,%xmm2
   3793d:	movd   %xmm2,%eax
   37941:	mov    $0x1,%bl
   37943:	test   $0x1,%al
   37945:	jne    379cf <adamic_function_1_isIdentifierStart+0xcf>
   3794b:	ucomisd 0x4c88d(%rip),%xmm0        # 841e0 <_IO_stdin_used+0x1e0>
   37953:	jne    37957 <adamic_function_1_isIdentifierStart+0x57>
   37955:	jnp    379cf <adamic_function_1_isIdentifierStart+0xcf>
   37957:	ucomisd 0x4c849(%rip),%xmm0        # 841a8 <_IO_stdin_used+0x1a8>
   3795f:	jne    37963 <adamic_function_1_isIdentifierStart+0x63>
   37961:	jnp    379cf <adamic_function_1_isIdentifierStart+0xcf>
   37963:	movsd  0x4c865(%rip),%xmm2        # 841d0 <_IO_stdin_used+0x1d0>
   3796b:	movapd %xmm0,%xmm1
   3796f:	cmplepd %xmm2,%xmm1
   37974:	movsd  0x4c75c(%rip),%xmm2        # 840d8 <_IO_stdin_used+0xd8>
   3797c:	cmplepd %xmm0,%xmm2
   37981:	andpd  %xmm1,%xmm2
   37985:	movd   %xmm2,%eax
   37989:	test   $0x1,%al
   3798b:	jne    379cf <adamic_function_1_isIdentifierStart+0xcf>
   3798d:	ucomisd 0x4c69b(%rip),%xmm0        # 84030 <_IO_stdin_used+0x30>
   37995:	jae    3799b <adamic_function_1_isIdentifierStart+0x9b>
   37997:	xor    %ebx,%ebx
   37999:	jmp    379cf <adamic_function_1_isIdentifierStart+0xcf>
   3799b:	movapd %xmm0,-0x20(%rbp)
   379a0:	cmpb   $0x0,0x89e29(%rip)        # c17d0 <adamic_ready_0>
   379a7:	je     379df <adamic_function_1_isIdentifierStart+0xdf>
   379a9:	mov    0x89e18(%rip),%rdi        # c17c8 <adamic_global_0_identifierStart>
   379b0:	call   6ecf0 <adamic_retain>
   379b5:	mov    %rax,%r14
   379b8:	mov    %rax,%rdi
   379bb:	movapd -0x20(%rbp),%xmm0
   379c0:	call   37aa0 <adamic_function_0_contains>
   379c5:	mov    %eax,%ebx
   379c7:	mov    %r14,%rdi
   379ca:	call   6ed10 <adamic_release>
   379cf:	mov    %ebx,%eax
   379d1:	add    $0x10,%rsp
   379d5:	pop    %rbx
   379d6:	pop    %r14
   379d8:	pop    %rbp
   379d9:	ret
   379da:	call   7f280 <adamic_stack_overflow>
   379df:	lea    0x4d39a(%rip),%rdi        # 84d80 <adamic_function_1_isIdentifierStart.message>
   379e6:	mov    $0x45,%esi
   379eb:	call   67bb0 <adamic_panic>

Disassembly of section .fini:
