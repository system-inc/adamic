
/workspace/scratch/branch-shape/switch:     file format elf64-x86-64


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
   37915:	ja     379b2 <adamic_function_1_isIdentifierStart+0xb2>
   3791b:	ucomisd 0x4c8ad(%rip),%xmm0        # 841d0 <_IO_stdin_used+0x1d0>
   37923:	mov    $0x1,%bl
   37925:	jne    37929 <adamic_function_1_isIdentifierStart+0x29>
   37927:	jnp    379a7 <adamic_function_1_isIdentifierStart+0xa7>
   37929:	ucomisd 0x4c877(%rip),%xmm0        # 841a8 <_IO_stdin_used+0x1a8>
   37931:	jne    37935 <adamic_function_1_isIdentifierStart+0x35>
   37933:	jnp    379a7 <adamic_function_1_isIdentifierStart+0xa7>
   37935:	ucomisd 0x4c79b(%rip),%xmm0        # 840d8 <_IO_stdin_used+0xd8>
   3793d:	jb     3794d <adamic_function_1_isIdentifierStart+0x4d>
   3793f:	movsd  0x4c891(%rip),%xmm1        # 841d8 <_IO_stdin_used+0x1d8>
   37947:	ucomisd %xmm0,%xmm1
   3794b:	jae    379a7 <adamic_function_1_isIdentifierStart+0xa7>
   3794d:	ucomisd 0x4c7a3(%rip),%xmm0        # 840f8 <_IO_stdin_used+0xf8>
   37955:	jb     37965 <adamic_function_1_isIdentifierStart+0x65>
   37957:	movsd  0x4c881(%rip),%xmm1        # 841e0 <_IO_stdin_used+0x1e0>
   3795f:	ucomisd %xmm0,%xmm1
   37963:	jae    379a7 <adamic_function_1_isIdentifierStart+0xa7>
   37965:	ucomisd 0x4c6c3(%rip),%xmm0        # 84030 <_IO_stdin_used+0x30>
   3796d:	jae    37973 <adamic_function_1_isIdentifierStart+0x73>
   3796f:	xor    %ebx,%ebx
   37971:	jmp    379a7 <adamic_function_1_isIdentifierStart+0xa7>
   37973:	movsd  %xmm0,-0x18(%rbp)
   37978:	cmpb   $0x0,0x89e51(%rip)        # c17d0 <adamic_ready_0>
   3797f:	je     379b7 <adamic_function_1_isIdentifierStart+0xb7>
   37981:	mov    0x89e40(%rip),%rdi        # c17c8 <adamic_global_0_identifierStart>
   37988:	call   6ed10 <adamic_retain>
   3798d:	mov    %rax,%r14
   37990:	mov    %rax,%rdi
   37993:	movsd  -0x18(%rbp),%xmm0
   37998:	call   37a80 <adamic_function_0_contains>
   3799d:	mov    %eax,%ebx
   3799f:	mov    %r14,%rdi
   379a2:	call   6ed30 <adamic_release>
   379a7:	mov    %ebx,%eax
   379a9:	add    $0x10,%rsp
   379ad:	pop    %rbx
   379ae:	pop    %r14
   379b0:	pop    %rbp
   379b1:	ret
   379b2:	call   7f2a0 <adamic_stack_overflow>
   379b7:	lea    0x4d3a2(%rip),%rdi        # 84d60 <adamic_function_1_isIdentifierStart.message>
   379be:	mov    $0x45,%esi
   379c3:	call   67bd0 <adamic_panic>

Disassembly of section .fini:
