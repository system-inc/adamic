
/workspace/scratch/branch-shape/before:     file format elf64-x86-64


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
   37935:	ja     379d2 <adamic_function_1_isIdentifierStart+0xb2>
   3793b:	ucomisd 0x4c8ad(%rip),%xmm0        # 841f0 <_IO_stdin_used+0x1f0>
   37943:	mov    $0x1,%bl
   37945:	jne    37949 <adamic_function_1_isIdentifierStart+0x29>
   37947:	jnp    379c7 <adamic_function_1_isIdentifierStart+0xa7>
   37949:	ucomisd 0x4c877(%rip),%xmm0        # 841c8 <_IO_stdin_used+0x1c8>
   37951:	jne    37955 <adamic_function_1_isIdentifierStart+0x35>
   37953:	jnp    379c7 <adamic_function_1_isIdentifierStart+0xa7>
   37955:	ucomisd 0x4c77b(%rip),%xmm0        # 840d8 <_IO_stdin_used+0xd8>
   3795d:	jb     3796d <adamic_function_1_isIdentifierStart+0x4d>
   3795f:	movsd  0x4c891(%rip),%xmm1        # 841f8 <_IO_stdin_used+0x1f8>
   37967:	ucomisd %xmm0,%xmm1
   3796b:	jae    379c7 <adamic_function_1_isIdentifierStart+0xa7>
   3796d:	ucomisd 0x4c783(%rip),%xmm0        # 840f8 <_IO_stdin_used+0xf8>
   37975:	jb     37985 <adamic_function_1_isIdentifierStart+0x65>
   37977:	movsd  0x4c881(%rip),%xmm1        # 84200 <_IO_stdin_used+0x200>
   3797f:	ucomisd %xmm0,%xmm1
   37983:	jae    379c7 <adamic_function_1_isIdentifierStart+0xa7>
   37985:	ucomisd 0x4c6a3(%rip),%xmm0        # 84030 <_IO_stdin_used+0x30>
   3798d:	jae    37993 <adamic_function_1_isIdentifierStart+0x73>
   3798f:	xor    %ebx,%ebx
   37991:	jmp    379c7 <adamic_function_1_isIdentifierStart+0xa7>
   37993:	movsd  %xmm0,-0x18(%rbp)
   37998:	cmpb   $0x0,0x89e31(%rip)        # c17d0 <adamic_ready_0>
   3799f:	je     379d7 <adamic_function_1_isIdentifierStart+0xb7>
   379a1:	mov    0x89e20(%rip),%rdi        # c17c8 <adamic_global_0_identifierStart>
   379a8:	call   6ed70 <adamic_retain>
   379ad:	mov    %rax,%r14
   379b0:	mov    %rax,%rdi
   379b3:	movsd  -0x18(%rbp),%xmm0
   379b8:	call   37aa0 <adamic_function_0_contains>
   379bd:	mov    %eax,%ebx
   379bf:	mov    %r14,%rdi
   379c2:	call   6ed90 <adamic_release>
   379c7:	mov    %ebx,%eax
   379c9:	add    $0x10,%rsp
   379cd:	pop    %rbx
   379ce:	pop    %r14
   379d0:	pop    %rbp
   379d1:	ret
   379d2:	call   7f300 <adamic_stack_overflow>
   379d7:	lea    0x50132(%rip),%rdi        # 87b10 <adamic_function_1_isIdentifierStart.message>
   379de:	mov    $0x45,%esi
   379e3:	call   67c30 <adamic_panic>

Disassembly of section .fini:
