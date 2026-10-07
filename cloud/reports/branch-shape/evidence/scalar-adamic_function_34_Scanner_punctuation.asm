
/workspace/scratch/branch-shape/scalar:     file format elf64-x86-64


Disassembly of section .init:

Disassembly of section .plt:

Disassembly of section .plt.got:

Disassembly of section .text:

0000000000039780 <adamic_function_34_Scanner_punctuation>:
   39780:	push   %rbp
   39781:	mov    %rsp,%rbp
   39784:	push   %r14
   39786:	push   %rbx
   39787:	lea    0x9f9f2(%rip),%rax        # d9180 <adamic_stack_limit>
   3978e:	cmp    %rbp,(%rax)
   39791:	ja     39fc2 <adamic_function_34_Scanner_punctuation+0x842>
   39797:	cvttsd2si %xmm0,%rax
   3979c:	cvtsi2sd %rax,%xmm1
   397a1:	ucomisd %xmm1,%xmm0
   397a5:	mov    $0x80000000,%ecx
   397aa:	cmovne %rcx,%rax
   397ae:	cmovp  %rcx,%rax
   397b2:	movsd  0x4a8c6(%rip),%xmm1        # 84080 <_IO_stdin_used+0x80>
   397ba:	ucomisd %xmm0,%xmm1
   397be:	cmovb  %rcx,%rax
   397c2:	ucomisd 0x4aa7e(%rip),%xmm0        # 84248 <_IO_stdin_used+0x248>
   397ca:	cmovb  %rcx,%rax
   397ce:	add    $0xffffffffffffffdf,%rax
   397d2:	cmp    $0x5d,%rax
   397d6:	ja     39c69 <adamic_function_34_Scanner_punctuation+0x4e9>
   397dc:	mov    %rdi,%rbx
   397df:	lea    0x4b39e(%rip),%rcx        # 84b84 <adamic_math_exp.ln2LO+0x94>
   397e6:	movslq (%rcx,%rax,4),%rax
   397ea:	add    %rcx,%rax
   397ed:	jmp    *%rax
   397ef:	movsd  0x4a819(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   397f7:	mov    %rbx,%rdi
   397fa:	mov    $0x1,%esi
   397ff:	call   34640 <adamic_function_22_Scanner_code>
   39804:	ucomisd 0x4aa4c(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   3980c:	jne    39d49 <adamic_function_34_Scanner_punctuation+0x5c9>
   39812:	jp     39d49 <adamic_function_34_Scanner_punctuation+0x5c9>
   39818:	movsd  0x4a818(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39820:	mov    %rbx,%rdi
   39823:	mov    $0x1,%esi
   39828:	call   34640 <adamic_function_22_Scanner_code>
   3982d:	ucomisd 0x4aa23(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39835:	jne    39d49 <adamic_function_34_Scanner_punctuation+0x5c9>
   3983b:	jp     39d49 <adamic_function_34_Scanner_punctuation+0x5c9>
   39841:	lea    0x82c28(%rip),%rdi        # bc470 <adamic_string_76>
   39848:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   3984d:	lea    0x8259c(%rip),%rdi        # bbdf0 <adamic_string_120>
   39854:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39859:	lea    0x82390(%rip),%rdi        # bbbf0 <adamic_string_122>
   39860:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39865:	movsd  0x4a7a3(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   3986d:	mov    %rbx,%rdi
   39870:	mov    $0x1,%esi
   39875:	call   34640 <adamic_function_22_Scanner_code>
   3987a:	ucomisd 0x4a9ee(%rip),%xmm0        # 84270 <_IO_stdin_used+0x270>
   39882:	jne    39c75 <adamic_function_34_Scanner_punctuation+0x4f5>
   39888:	jp     39c75 <adamic_function_34_Scanner_punctuation+0x4f5>
   3988e:	movsd  0x4a7a2(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39896:	mov    %rbx,%rdi
   39899:	mov    $0x1,%esi
   3989e:	call   34640 <adamic_function_22_Scanner_code>
   398a3:	ucomisd 0x4a9ad(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   398ab:	jne    39c75 <adamic_function_34_Scanner_punctuation+0x4f5>
   398b1:	jp     39c75 <adamic_function_34_Scanner_punctuation+0x4f5>
   398b7:	lea    0x83bb2(%rip),%rdi        # bd470 <adamic_string_98>
   398be:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   398c3:	lea    0x823a6(%rip),%rdi        # bbc70 <adamic_string_123>
   398ca:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   398cf:	movsd  0x4a739(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   398d7:	mov    %rbx,%rdi
   398da:	mov    $0x1,%esi
   398df:	call   34640 <adamic_function_22_Scanner_code>
   398e4:	ucomisd 0x4a96c(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   398ec:	lea    0x82e7d(%rip),%rax        # bc770 <adamic_string_97>
   398f3:	lea    0x83a76(%rip),%rdi        # bd370 <adamic_string_96>
   398fa:	jmp    39f34 <adamic_function_34_Scanner_punctuation+0x7b4>
   398ff:	movsd  0x4a709(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39907:	mov    %rbx,%rdi
   3990a:	mov    $0x1,%esi
   3990f:	call   34640 <adamic_function_22_Scanner_code>
   39914:	ucomisd 0x4a964(%rip),%xmm0        # 84280 <_IO_stdin_used+0x280>
   3991c:	jne    39caa <adamic_function_34_Scanner_punctuation+0x52a>
   39922:	jp     39caa <adamic_function_34_Scanner_punctuation+0x52a>
   39928:	movsd  0x4a708(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39930:	mov    %rbx,%rdi
   39933:	mov    $0x1,%esi
   39938:	call   34640 <adamic_function_22_Scanner_code>
   3993d:	ucomisd 0x4a913(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39945:	jne    39caa <adamic_function_34_Scanner_punctuation+0x52a>
   3994b:	jp     39caa <adamic_function_34_Scanner_punctuation+0x52a>
   39951:	lea    0x83e98(%rip),%rdi        # bd7f0 <adamic_string_81>
   39958:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   3995d:	lea    0x8240c(%rip),%rdi        # bbd70 <adamic_string_119>
   39964:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39969:	movsd  0x4a69f(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39971:	mov    %rbx,%rdi
   39974:	mov    $0x1,%esi
   39979:	call   34640 <adamic_function_22_Scanner_code>
   3997e:	ucomisd 0x4a8ca(%rip),%xmm0        # 84250 <_IO_stdin_used+0x250>
   39986:	jne    39cdf <adamic_function_34_Scanner_punctuation+0x55f>
   3998c:	jp     39cdf <adamic_function_34_Scanner_punctuation+0x55f>
   39992:	movsd  0x4a69e(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   3999a:	mov    %rbx,%rdi
   3999d:	mov    $0x1,%esi
   399a2:	call   34640 <adamic_function_22_Scanner_code>
   399a7:	ucomisd 0x4a8a9(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   399af:	jne    39cdf <adamic_function_34_Scanner_punctuation+0x55f>
   399b5:	jp     39cdf <adamic_function_34_Scanner_punctuation+0x55f>
   399bb:	lea    0x83dae(%rip),%rdi        # bd770 <adamic_string_113>
   399c2:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   399c7:	lea    0x82622(%rip),%rdi        # bbff0 <adamic_string_124>
   399ce:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   399d3:	lea    0x82496(%rip),%rdi        # bbe70 <adamic_string_121>
   399da:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   399df:	movsd  0x4a629(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   399e7:	mov    %rbx,%rdi
   399ea:	mov    $0x1,%esi
   399ef:	call   34640 <adamic_function_22_Scanner_code>
   399f4:	ucomisd 0x4a87c(%rip),%xmm0        # 84278 <_IO_stdin_used+0x278>
   399fc:	jne    39d14 <adamic_function_34_Scanner_punctuation+0x594>
   39a02:	jp     39d14 <adamic_function_34_Scanner_punctuation+0x594>
   39a08:	movsd  0x4a628(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39a10:	mov    %rbx,%rdi
   39a13:	mov    $0x1,%esi
   39a18:	call   34640 <adamic_function_22_Scanner_code>
   39a1d:	ucomisd 0x4a833(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39a25:	jne    39d14 <adamic_function_34_Scanner_punctuation+0x594>
   39a2b:	jp     39d14 <adamic_function_34_Scanner_punctuation+0x594>
   39a31:	lea    0x838b8(%rip),%rdi        # bd2f0 <adamic_string_85>
   39a38:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39a3d:	lea    0x8362c(%rip),%rdi        # bd070 <adamic_string_125>
   39a44:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39a49:	lea    0x822a0(%rip),%rdi        # bbcf0 <adamic_string_118>
   39a50:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39a55:	movsd  0x4a5b3(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39a5d:	mov    %rbx,%rdi
   39a60:	mov    $0x1,%esi
   39a65:	call   34640 <adamic_function_22_Scanner_code>
   39a6a:	ucomisd 0x4a7e6(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39a72:	lea    0x82d77(%rip),%rax        # bc7f0 <adamic_string_80>
   39a79:	lea    0x83970(%rip),%rdi        # bd3f0 <adamic_string_79>
   39a80:	jmp    39f34 <adamic_function_34_Scanner_punctuation+0x7b4>
   39a85:	lea    0x832e4(%rip),%rdi        # bcd70 <adamic_string_117>
   39a8c:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39a91:	movsd  0x4a577(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39a99:	mov    %rbx,%rdi
   39a9c:	mov    $0x1,%esi
   39aa1:	call   34640 <adamic_function_22_Scanner_code>
   39aa6:	ucomisd 0x4a7aa(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39aae:	jne    39def <adamic_function_34_Scanner_punctuation+0x66f>
   39ab4:	jp     39def <adamic_function_34_Scanner_punctuation+0x66f>
   39aba:	lea    0x836af(%rip),%rdi        # bd170 <adamic_string_89>
   39ac1:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39ac6:	lea    0x825a3(%rip),%rdi        # bc070 <adamic_string_53>
   39acd:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39ad2:	movsd  0x4a536(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39ada:	mov    %rbx,%rdi
   39add:	mov    $0x1,%esi
   39ae2:	call   34640 <adamic_function_22_Scanner_code>
   39ae7:	ucomisd 0x4a769(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39aef:	jne    39d79 <adamic_function_34_Scanner_punctuation+0x5f9>
   39af5:	jp     39d79 <adamic_function_34_Scanner_punctuation+0x5f9>
   39afb:	movsd  0x4a535(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39b03:	mov    %rbx,%rdi
   39b06:	mov    $0x1,%esi
   39b0b:	call   34640 <adamic_function_22_Scanner_code>
   39b10:	ucomisd 0x4a740(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39b18:	jne    39d79 <adamic_function_34_Scanner_punctuation+0x5f9>
   39b1e:	jp     39d79 <adamic_function_34_Scanner_punctuation+0x5f9>
   39b24:	lea    0x828c5(%rip),%rdi        # bc3f0 <adamic_string_102>
   39b2b:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39b30:	lea    0x82639(%rip),%rdi        # bc170 <adamic_string_106>
   39b37:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39b3c:	movsd  0x4a4cc(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39b44:	mov    %rbx,%rdi
   39b47:	mov    $0x1,%esi
   39b4c:	call   34640 <adamic_function_22_Scanner_code>
   39b51:	ucomisd 0x4a6af(%rip),%xmm0        # 84208 <_IO_stdin_used+0x208>
   39b59:	jne    39dae <adamic_function_34_Scanner_punctuation+0x62e>
   39b5f:	jp     39dae <adamic_function_34_Scanner_punctuation+0x62e>
   39b65:	movsd  0x4a4cb(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39b6d:	mov    %rbx,%rdi
   39b70:	mov    $0x1,%esi
   39b75:	call   34640 <adamic_function_22_Scanner_code>
   39b7a:	ucomisd 0x4a686(%rip),%xmm0        # 84208 <_IO_stdin_used+0x208>
   39b82:	jne    39dae <adamic_function_34_Scanner_punctuation+0x62e>
   39b88:	jp     39dae <adamic_function_34_Scanner_punctuation+0x62e>
   39b8e:	lea    0x823db(%rip),%rdi        # bbf70 <adamic_string_95>
   39b95:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39b9a:	movsd  0x4a46e(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39ba2:	mov    %rbx,%rdi
   39ba5:	mov    $0x1,%esi
   39baa:	call   34640 <adamic_function_22_Scanner_code>
   39baf:	ucomisd 0x4a6a1(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39bb7:	jne    39e1f <adamic_function_34_Scanner_punctuation+0x69f>
   39bbd:	jp     39e1f <adamic_function_34_Scanner_punctuation+0x69f>
   39bc3:	lea    0x83626(%rip),%rdi        # bd1f0 <adamic_string_92>
   39bca:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39bcf:	lea    0x83d1a(%rip),%rdi        # bd8f0 <adamic_string_66>
   39bd6:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39bdb:	movsd  0x4a42d(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39be3:	mov    %rbx,%rdi
   39be6:	mov    $0x1,%esi
   39beb:	call   34640 <adamic_function_22_Scanner_code>
   39bf0:	ucomisd 0x4a668(%rip),%xmm0        # 84260 <_IO_stdin_used+0x260>
   39bf8:	jne    39dba <adamic_function_34_Scanner_punctuation+0x63a>
   39bfe:	jp     39dba <adamic_function_34_Scanner_punctuation+0x63a>
   39c04:	movsd  0x4a42c(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39c0c:	mov    %rbx,%rdi
   39c0f:	mov    $0x1,%esi
   39c14:	call   34640 <adamic_function_22_Scanner_code>
   39c19:	ucomisd 0x4a637(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39c21:	jne    39dba <adamic_function_34_Scanner_punctuation+0x63a>
   39c27:	jp     39dba <adamic_function_34_Scanner_punctuation+0x63a>
   39c2d:	lea    0x83c3c(%rip),%rdi        # bd870 <adamic_string_107>
   39c34:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39c39:	movsd  0x4a3cf(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39c41:	mov    %rbx,%rdi
   39c44:	mov    $0x1,%esi
   39c49:	call   34640 <adamic_function_22_Scanner_code>
   39c4e:	ucomisd 0x4a602(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39c56:	lea    0x83013(%rip),%rax        # bcc70 <adamic_string_112>
   39c5d:	lea    0x83a8c(%rip),%rdi        # bd6f0 <adamic_string_111>
   39c64:	jmp    39f34 <adamic_function_34_Scanner_punctuation+0x7b4>
   39c69:	lea    0x842f8(%rip),%rdi        # bdf68 <adamic_string_1>
   39c70:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39c75:	movsd  0x4a393(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39c7d:	mov    %rbx,%rdi
   39c80:	mov    $0x1,%esi
   39c85:	call   34640 <adamic_function_22_Scanner_code>
   39c8a:	ucomisd 0x4a5c6(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39c92:	jne    39e4f <adamic_function_34_Scanner_punctuation+0x6cf>
   39c98:	jp     39e4f <adamic_function_34_Scanner_punctuation+0x6cf>
   39c9e:	lea    0x8254b(%rip),%rdi        # bc1f0 <adamic_string_99>
   39ca5:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39caa:	movsd  0x4a35e(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39cb2:	mov    %rbx,%rdi
   39cb5:	mov    $0x1,%esi
   39cba:	call   34640 <adamic_function_22_Scanner_code>
   39cbf:	ucomisd 0x4a591(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39cc7:	jne    39e7f <adamic_function_34_Scanner_punctuation+0x6ff>
   39ccd:	jp     39e7f <adamic_function_34_Scanner_punctuation+0x6ff>
   39cd3:	lea    0x83916(%rip),%rdi        # bd5f0 <adamic_string_82>
   39cda:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39cdf:	movsd  0x4a329(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39ce7:	mov    %rbx,%rdi
   39cea:	mov    $0x1,%esi
   39cef:	call   34640 <adamic_function_22_Scanner_code>
   39cf4:	ucomisd 0x4a55c(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39cfc:	jne    39eaf <adamic_function_34_Scanner_punctuation+0x72f>
   39d02:	jp     39eaf <adamic_function_34_Scanner_punctuation+0x72f>
   39d08:	lea    0x83961(%rip),%rdi        # bd670 <adamic_string_114>
   39d0f:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39d14:	movsd  0x4a2f4(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39d1c:	mov    %rbx,%rdi
   39d1f:	mov    $0x1,%esi
   39d24:	call   34640 <adamic_function_22_Scanner_code>
   39d29:	ucomisd 0x4a527(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39d31:	jne    39edc <adamic_function_34_Scanner_punctuation+0x75c>
   39d37:	jp     39edc <adamic_function_34_Scanner_punctuation+0x75c>
   39d3d:	lea    0x8352c(%rip),%rdi        # bd270 <adamic_string_86>
   39d44:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39d49:	movsd  0x4a2bf(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39d51:	mov    %rbx,%rdi
   39d54:	mov    $0x1,%esi
   39d59:	call   34640 <adamic_function_22_Scanner_code>
   39d5e:	ucomisd 0x4a4f2(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39d66:	lea    0x82f83(%rip),%rax        # bccf0 <adamic_string_78>
   39d6d:	lea    0x825fc(%rip),%rdi        # bc370 <adamic_string_77>
   39d74:	jmp    39f34 <adamic_function_34_Scanner_punctuation+0x7b4>
   39d79:	movsd  0x4a28f(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39d81:	mov    %rbx,%rdi
   39d84:	mov    $0x1,%esi
   39d89:	call   34640 <adamic_function_22_Scanner_code>
   39d8e:	ucomisd 0x4a4c2(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39d96:	jne    39f09 <adamic_function_34_Scanner_punctuation+0x789>
   39d9c:	jp     39f09 <adamic_function_34_Scanner_punctuation+0x789>
   39da2:	lea    0x82547(%rip),%rdi        # bc2f0 <adamic_string_103>
   39da9:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39dae:	lea    0x8213b(%rip),%rdi        # bbef0 <adamic_string_64>
   39db5:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39dba:	movsd  0x4a24e(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39dc2:	mov    %rbx,%rdi
   39dc5:	mov    $0x1,%esi
   39dca:	call   34640 <adamic_function_22_Scanner_code>
   39dcf:	ucomisd 0x4a489(%rip),%xmm0        # 84260 <_IO_stdin_used+0x260>
   39dd7:	jne    39f46 <adamic_function_34_Scanner_punctuation+0x7c6>
   39ddd:	jp     39f46 <adamic_function_34_Scanner_punctuation+0x7c6>
   39de3:	lea    0x83186(%rip),%rdi        # bcf70 <adamic_string_108>
   39dea:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39def:	movsd  0x4a219(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39df7:	mov    %rbx,%rdi
   39dfa:	mov    $0x1,%esi
   39dff:	call   34640 <adamic_function_22_Scanner_code>
   39e04:	ucomisd 0x4a414(%rip),%xmm0        # 84220 <_IO_stdin_used+0x220>
   39e0c:	lea    0x8275d(%rip),%rax        # bc570 <adamic_string_91>
   39e13:	lea    0x82a56(%rip),%rdi        # bc870 <adamic_string_90>
   39e1a:	jmp    39f34 <adamic_function_34_Scanner_punctuation+0x7b4>
   39e1f:	movsd  0x4a1e9(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39e27:	mov    %rbx,%rdi
   39e2a:	mov    $0x1,%esi
   39e2f:	call   34640 <adamic_function_22_Scanner_code>
   39e34:	ucomisd 0x4a3ec(%rip),%xmm0        # 84228 <_IO_stdin_used+0x228>
   39e3c:	lea    0x827ad(%rip),%rax        # bc5f0 <adamic_string_94>
   39e43:	lea    0x82aa6(%rip),%rdi        # bc8f0 <adamic_string_93>
   39e4a:	jmp    39f34 <adamic_function_34_Scanner_punctuation+0x7b4>
   39e4f:	movsd  0x4a1b9(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39e57:	mov    %rbx,%rdi
   39e5a:	mov    $0x1,%esi
   39e5f:	call   34640 <adamic_function_22_Scanner_code>
   39e64:	ucomisd 0x4a404(%rip),%xmm0        # 84270 <_IO_stdin_used+0x270>
   39e6c:	lea    0x8227d(%rip),%rax        # bc0f0 <adamic_string_101>
   39e73:	lea    0x82af6(%rip),%rdi        # bc970 <adamic_string_100>
   39e7a:	jmp    39f34 <adamic_function_34_Scanner_punctuation+0x7b4>
   39e7f:	movsd  0x4a189(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39e87:	mov    %rbx,%rdi
   39e8a:	mov    $0x1,%esi
   39e8f:	call   34640 <adamic_function_22_Scanner_code>
   39e94:	ucomisd 0x4a3e4(%rip),%xmm0        # 84280 <_IO_stdin_used+0x280>
   39e9c:	lea    0x82ccd(%rip),%rax        # bcb70 <adamic_string_84>
   39ea3:	lea    0x82f46(%rip),%rdi        # bcdf0 <adamic_string_83>
   39eaa:	jmp    39f34 <adamic_function_34_Scanner_punctuation+0x7b4>
   39eaf:	movsd  0x4a159(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39eb7:	mov    %rbx,%rdi
   39eba:	mov    $0x1,%esi
   39ebf:	call   34640 <adamic_function_22_Scanner_code>
   39ec4:	ucomisd 0x4a384(%rip),%xmm0        # 84250 <_IO_stdin_used+0x250>
   39ecc:	lea    0x82d1d(%rip),%rax        # bcbf0 <adamic_string_116>
   39ed3:	lea    0x82f96(%rip),%rdi        # bce70 <adamic_string_115>
   39eda:	jmp    39f34 <adamic_function_34_Scanner_punctuation+0x7b4>
   39edc:	movsd  0x4a12c(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39ee4:	mov    %rbx,%rdi
   39ee7:	mov    $0x1,%esi
   39eec:	call   34640 <adamic_function_22_Scanner_code>
   39ef1:	ucomisd 0x4a37f(%rip),%xmm0        # 84278 <_IO_stdin_used+0x278>
   39ef9:	lea    0x827f0(%rip),%rax        # bc6f0 <adamic_string_88>
   39f00:	lea    0x82769(%rip),%rdi        # bc670 <adamic_string_87>
   39f07:	jmp    39f34 <adamic_function_34_Scanner_punctuation+0x7b4>
   39f09:	movsd  0x4a0ff(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39f11:	mov    %rbx,%rdi
   39f14:	mov    $0x1,%esi
   39f19:	call   34640 <adamic_function_22_Scanner_code>
   39f1e:	ucomisd 0x4a342(%rip),%xmm0        # 84268 <_IO_stdin_used+0x268>
   39f26:	lea    0x831c3(%rip),%rax        # bd0f0 <adamic_string_105>
   39f2d:	lea    0x825bc(%rip),%rdi        # bc4f0 <adamic_string_104>
   39f34:	cmovne %rax,%rdi
   39f38:	cmovp  %rax,%rdi
   39f3c:	call   6ecf0 <adamic_retain>
   39f41:	pop    %rbx
   39f42:	pop    %r14
   39f44:	pop    %rbp
   39f45:	ret
   39f46:	movsd  0x4a0c2(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39f4e:	mov    %rbx,%rdi
   39f51:	mov    $0x1,%esi
   39f56:	call   34640 <adamic_function_22_Scanner_code>
   39f5b:	ucomisd 0x4a2a5(%rip),%xmm0        # 84208 <_IO_stdin_used+0x208>
   39f63:	jne    39fb6 <adamic_function_34_Scanner_punctuation+0x836>
   39f65:	jp     39fb6 <adamic_function_34_Scanner_punctuation+0x836>
   39f67:	movsd  0x4a0c9(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39f6f:	mov    %rbx,%rdi
   39f72:	mov    $0x1,%esi
   39f77:	call   34640 <adamic_function_22_Scanner_code>
   39f7c:	lea    0x8306d(%rip),%rdi        # bcff0 <adamic_string_109>
   39f83:	ucomisd 0x4a13d(%rip),%xmm0        # 840c8 <_IO_stdin_used+0xc8>
   39f8b:	jb     39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39f8d:	movsd  0x4a0a3(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39f95:	mov    %rdi,%r14
   39f98:	mov    %rbx,%rdi
   39f9b:	mov    $0x1,%esi
   39fa0:	call   34640 <adamic_function_22_Scanner_code>
   39fa5:	mov    %r14,%rdi
   39fa8:	movsd  0x4a120(%rip),%xmm1        # 840d0 <_IO_stdin_used+0xd0>
   39fb0:	ucomisd %xmm0,%xmm1
   39fb4:	jb     39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39fb6:	lea    0x82f33(%rip),%rdi        # bcef0 <adamic_string_110>
   39fbd:	jmp    39f3c <adamic_function_34_Scanner_punctuation+0x7bc>
   39fc2:	call   7f280 <adamic_stack_overflow>

Disassembly of section .fini:
