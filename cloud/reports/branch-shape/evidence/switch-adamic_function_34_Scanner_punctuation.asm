
/workspace/scratch/branch-shape/switch:     file format elf64-x86-64


Disassembly of section .init:

Disassembly of section .plt:

Disassembly of section .plt.got:

Disassembly of section .text:

0000000000039760 <adamic_function_34_Scanner_punctuation>:
   39760:	push   %rbp
   39761:	mov    %rsp,%rbp
   39764:	push   %r14
   39766:	push   %rbx
   39767:	lea    0x9fa12(%rip),%rax        # d9180 <adamic_stack_limit>
   3976e:	cmp    %rbp,(%rax)
   39771:	ja     39fa2 <adamic_function_34_Scanner_punctuation+0x842>
   39777:	cvttsd2si %xmm0,%rax
   3977c:	cvtsi2sd %rax,%xmm1
   39781:	ucomisd %xmm1,%xmm0
   39785:	mov    $0x80000000,%ecx
   3978a:	cmovne %rcx,%rax
   3978e:	cmovp  %rcx,%rax
   39792:	movsd  0x4a8e6(%rip),%xmm1        # 84080 <_IO_stdin_used+0x80>
   3979a:	ucomisd %xmm0,%xmm1
   3979e:	cmovb  %rcx,%rax
   397a2:	ucomisd 0x4aa9e(%rip),%xmm0        # 84248 <_IO_stdin_used+0x248>
   397aa:	cmovb  %rcx,%rax
   397ae:	add    $0xffffffffffffffdf,%rax
   397b2:	cmp    $0x5d,%rax
   397b6:	ja     39c49 <adamic_function_34_Scanner_punctuation+0x4e9>
   397bc:	mov    %rdi,%rbx
   397bf:	lea    0x4b39e(%rip),%rcx        # 84b64 <adamic_math_exp.ln2LO+0x94>
   397c6:	movslq (%rcx,%rax,4),%rax
   397ca:	add    %rcx,%rax
   397cd:	jmp    *%rax
   397cf:	movsd  0x4a839(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   397d7:	mov    %rbx,%rdi
   397da:	mov    $0x1,%esi
   397df:	call   34640 <adamic_function_22_Scanner_code>
   397e4:	ucomisd 0x4aa6c(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   397ec:	jne    39d29 <adamic_function_34_Scanner_punctuation+0x5c9>
   397f2:	jp     39d29 <adamic_function_34_Scanner_punctuation+0x5c9>
   397f8:	movsd  0x4a838(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39800:	mov    %rbx,%rdi
   39803:	mov    $0x1,%esi
   39808:	call   34640 <adamic_function_22_Scanner_code>
   3980d:	ucomisd 0x4aa43(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39815:	jne    39d29 <adamic_function_34_Scanner_punctuation+0x5c9>
   3981b:	jp     39d29 <adamic_function_34_Scanner_punctuation+0x5c9>
   39821:	lea    0x82c48(%rip),%rdi        # bc470 <adamic_string_76>
   39828:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   3982d:	lea    0x825bc(%rip),%rdi        # bbdf0 <adamic_string_120>
   39834:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39839:	lea    0x823b0(%rip),%rdi        # bbbf0 <adamic_string_122>
   39840:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39845:	movsd  0x4a7c3(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   3984d:	mov    %rbx,%rdi
   39850:	mov    $0x1,%esi
   39855:	call   34640 <adamic_function_22_Scanner_code>
   3985a:	ucomisd 0x4aa0e(%rip),%xmm0        # 84270 <_IO_stdin_used+0x270>
   39862:	jne    39c55 <adamic_function_34_Scanner_punctuation+0x4f5>
   39868:	jp     39c55 <adamic_function_34_Scanner_punctuation+0x4f5>
   3986e:	movsd  0x4a7c2(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39876:	mov    %rbx,%rdi
   39879:	mov    $0x1,%esi
   3987e:	call   34640 <adamic_function_22_Scanner_code>
   39883:	ucomisd 0x4a9cd(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   3988b:	jne    39c55 <adamic_function_34_Scanner_punctuation+0x4f5>
   39891:	jp     39c55 <adamic_function_34_Scanner_punctuation+0x4f5>
   39897:	lea    0x83bd2(%rip),%rdi        # bd470 <adamic_string_98>
   3989e:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   398a3:	lea    0x823c6(%rip),%rdi        # bbc70 <adamic_string_123>
   398aa:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   398af:	movsd  0x4a759(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   398b7:	mov    %rbx,%rdi
   398ba:	mov    $0x1,%esi
   398bf:	call   34640 <adamic_function_22_Scanner_code>
   398c4:	ucomisd 0x4a98c(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   398cc:	lea    0x82e9d(%rip),%rax        # bc770 <adamic_string_97>
   398d3:	lea    0x83a96(%rip),%rdi        # bd370 <adamic_string_96>
   398da:	jmp    39f14 <adamic_function_34_Scanner_punctuation+0x7b4>
   398df:	movsd  0x4a729(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   398e7:	mov    %rbx,%rdi
   398ea:	mov    $0x1,%esi
   398ef:	call   34640 <adamic_function_22_Scanner_code>
   398f4:	ucomisd 0x4a984(%rip),%xmm0        # 84280 <_IO_stdin_used+0x280>
   398fc:	jne    39c8a <adamic_function_34_Scanner_punctuation+0x52a>
   39902:	jp     39c8a <adamic_function_34_Scanner_punctuation+0x52a>
   39908:	movsd  0x4a728(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39910:	mov    %rbx,%rdi
   39913:	mov    $0x1,%esi
   39918:	call   34640 <adamic_function_22_Scanner_code>
   3991d:	ucomisd 0x4a933(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39925:	jne    39c8a <adamic_function_34_Scanner_punctuation+0x52a>
   3992b:	jp     39c8a <adamic_function_34_Scanner_punctuation+0x52a>
   39931:	lea    0x83eb8(%rip),%rdi        # bd7f0 <adamic_string_81>
   39938:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   3993d:	lea    0x8242c(%rip),%rdi        # bbd70 <adamic_string_119>
   39944:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39949:	movsd  0x4a6bf(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39951:	mov    %rbx,%rdi
   39954:	mov    $0x1,%esi
   39959:	call   34640 <adamic_function_22_Scanner_code>
   3995e:	ucomisd 0x4a8ea(%rip),%xmm0        # 84250 <_IO_stdin_used+0x250>
   39966:	jne    39cbf <adamic_function_34_Scanner_punctuation+0x55f>
   3996c:	jp     39cbf <adamic_function_34_Scanner_punctuation+0x55f>
   39972:	movsd  0x4a6be(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   3997a:	mov    %rbx,%rdi
   3997d:	mov    $0x1,%esi
   39982:	call   34640 <adamic_function_22_Scanner_code>
   39987:	ucomisd 0x4a8c9(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   3998f:	jne    39cbf <adamic_function_34_Scanner_punctuation+0x55f>
   39995:	jp     39cbf <adamic_function_34_Scanner_punctuation+0x55f>
   3999b:	lea    0x83dce(%rip),%rdi        # bd770 <adamic_string_113>
   399a2:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   399a7:	lea    0x82642(%rip),%rdi        # bbff0 <adamic_string_124>
   399ae:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   399b3:	lea    0x824b6(%rip),%rdi        # bbe70 <adamic_string_121>
   399ba:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   399bf:	movsd  0x4a649(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   399c7:	mov    %rbx,%rdi
   399ca:	mov    $0x1,%esi
   399cf:	call   34640 <adamic_function_22_Scanner_code>
   399d4:	ucomisd 0x4a89c(%rip),%xmm0        # 84278 <_IO_stdin_used+0x278>
   399dc:	jne    39cf4 <adamic_function_34_Scanner_punctuation+0x594>
   399e2:	jp     39cf4 <adamic_function_34_Scanner_punctuation+0x594>
   399e8:	movsd  0x4a648(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   399f0:	mov    %rbx,%rdi
   399f3:	mov    $0x1,%esi
   399f8:	call   34640 <adamic_function_22_Scanner_code>
   399fd:	ucomisd 0x4a853(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39a05:	jne    39cf4 <adamic_function_34_Scanner_punctuation+0x594>
   39a0b:	jp     39cf4 <adamic_function_34_Scanner_punctuation+0x594>
   39a11:	lea    0x838d8(%rip),%rdi        # bd2f0 <adamic_string_85>
   39a18:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39a1d:	lea    0x8364c(%rip),%rdi        # bd070 <adamic_string_125>
   39a24:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39a29:	lea    0x822c0(%rip),%rdi        # bbcf0 <adamic_string_118>
   39a30:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39a35:	movsd  0x4a5d3(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39a3d:	mov    %rbx,%rdi
   39a40:	mov    $0x1,%esi
   39a45:	call   34640 <adamic_function_22_Scanner_code>
   39a4a:	ucomisd 0x4a806(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39a52:	lea    0x82d97(%rip),%rax        # bc7f0 <adamic_string_80>
   39a59:	lea    0x83990(%rip),%rdi        # bd3f0 <adamic_string_79>
   39a60:	jmp    39f14 <adamic_function_34_Scanner_punctuation+0x7b4>
   39a65:	lea    0x83304(%rip),%rdi        # bcd70 <adamic_string_117>
   39a6c:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39a71:	movsd  0x4a597(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39a79:	mov    %rbx,%rdi
   39a7c:	mov    $0x1,%esi
   39a81:	call   34640 <adamic_function_22_Scanner_code>
   39a86:	ucomisd 0x4a7ca(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39a8e:	jne    39dcf <adamic_function_34_Scanner_punctuation+0x66f>
   39a94:	jp     39dcf <adamic_function_34_Scanner_punctuation+0x66f>
   39a9a:	lea    0x836cf(%rip),%rdi        # bd170 <adamic_string_89>
   39aa1:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39aa6:	lea    0x825c3(%rip),%rdi        # bc070 <adamic_string_53>
   39aad:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39ab2:	movsd  0x4a556(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39aba:	mov    %rbx,%rdi
   39abd:	mov    $0x1,%esi
   39ac2:	call   34640 <adamic_function_22_Scanner_code>
   39ac7:	ucomisd 0x4a789(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39acf:	jne    39d59 <adamic_function_34_Scanner_punctuation+0x5f9>
   39ad5:	jp     39d59 <adamic_function_34_Scanner_punctuation+0x5f9>
   39adb:	movsd  0x4a555(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39ae3:	mov    %rbx,%rdi
   39ae6:	mov    $0x1,%esi
   39aeb:	call   34640 <adamic_function_22_Scanner_code>
   39af0:	ucomisd 0x4a760(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39af8:	jne    39d59 <adamic_function_34_Scanner_punctuation+0x5f9>
   39afe:	jp     39d59 <adamic_function_34_Scanner_punctuation+0x5f9>
   39b04:	lea    0x828e5(%rip),%rdi        # bc3f0 <adamic_string_102>
   39b0b:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39b10:	lea    0x82659(%rip),%rdi        # bc170 <adamic_string_106>
   39b17:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39b1c:	movsd  0x4a4ec(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39b24:	mov    %rbx,%rdi
   39b27:	mov    $0x1,%esi
   39b2c:	call   34640 <adamic_function_22_Scanner_code>
   39b31:	ucomisd 0x4a6cf(%rip),%xmm0        # 84208 <_IO_stdin_used+0x208>
   39b39:	jne    39d8e <adamic_function_34_Scanner_punctuation+0x62e>
   39b3f:	jp     39d8e <adamic_function_34_Scanner_punctuation+0x62e>
   39b45:	movsd  0x4a4eb(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39b4d:	mov    %rbx,%rdi
   39b50:	mov    $0x1,%esi
   39b55:	call   34640 <adamic_function_22_Scanner_code>
   39b5a:	ucomisd 0x4a6a6(%rip),%xmm0        # 84208 <_IO_stdin_used+0x208>
   39b62:	jne    39d8e <adamic_function_34_Scanner_punctuation+0x62e>
   39b68:	jp     39d8e <adamic_function_34_Scanner_punctuation+0x62e>
   39b6e:	lea    0x823fb(%rip),%rdi        # bbf70 <adamic_string_95>
   39b75:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39b7a:	movsd  0x4a48e(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39b82:	mov    %rbx,%rdi
   39b85:	mov    $0x1,%esi
   39b8a:	call   34640 <adamic_function_22_Scanner_code>
   39b8f:	ucomisd 0x4a6c1(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39b97:	jne    39dff <adamic_function_34_Scanner_punctuation+0x69f>
   39b9d:	jp     39dff <adamic_function_34_Scanner_punctuation+0x69f>
   39ba3:	lea    0x83646(%rip),%rdi        # bd1f0 <adamic_string_92>
   39baa:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39baf:	lea    0x83d3a(%rip),%rdi        # bd8f0 <adamic_string_66>
   39bb6:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39bbb:	movsd  0x4a44d(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39bc3:	mov    %rbx,%rdi
   39bc6:	mov    $0x1,%esi
   39bcb:	call   34640 <adamic_function_22_Scanner_code>
   39bd0:	ucomisd 0x4a688(%rip),%xmm0        # 84260 <_IO_stdin_used+0x260>
   39bd8:	jne    39d9a <adamic_function_34_Scanner_punctuation+0x63a>
   39bde:	jp     39d9a <adamic_function_34_Scanner_punctuation+0x63a>
   39be4:	movsd  0x4a44c(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39bec:	mov    %rbx,%rdi
   39bef:	mov    $0x1,%esi
   39bf4:	call   34640 <adamic_function_22_Scanner_code>
   39bf9:	ucomisd 0x4a657(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39c01:	jne    39d9a <adamic_function_34_Scanner_punctuation+0x63a>
   39c07:	jp     39d9a <adamic_function_34_Scanner_punctuation+0x63a>
   39c0d:	lea    0x83c5c(%rip),%rdi        # bd870 <adamic_string_107>
   39c14:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39c19:	movsd  0x4a3ef(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39c21:	mov    %rbx,%rdi
   39c24:	mov    $0x1,%esi
   39c29:	call   34640 <adamic_function_22_Scanner_code>
   39c2e:	ucomisd 0x4a622(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39c36:	lea    0x83033(%rip),%rax        # bcc70 <adamic_string_112>
   39c3d:	lea    0x83aac(%rip),%rdi        # bd6f0 <adamic_string_111>
   39c44:	jmp    39f14 <adamic_function_34_Scanner_punctuation+0x7b4>
   39c49:	lea    0x84318(%rip),%rdi        # bdf68 <adamic_string_1>
   39c50:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39c55:	movsd  0x4a3b3(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39c5d:	mov    %rbx,%rdi
   39c60:	mov    $0x1,%esi
   39c65:	call   34640 <adamic_function_22_Scanner_code>
   39c6a:	ucomisd 0x4a5e6(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39c72:	jne    39e2f <adamic_function_34_Scanner_punctuation+0x6cf>
   39c78:	jp     39e2f <adamic_function_34_Scanner_punctuation+0x6cf>
   39c7e:	lea    0x8256b(%rip),%rdi        # bc1f0 <adamic_string_99>
   39c85:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39c8a:	movsd  0x4a37e(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39c92:	mov    %rbx,%rdi
   39c95:	mov    $0x1,%esi
   39c9a:	call   34640 <adamic_function_22_Scanner_code>
   39c9f:	ucomisd 0x4a5b1(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39ca7:	jne    39e5f <adamic_function_34_Scanner_punctuation+0x6ff>
   39cad:	jp     39e5f <adamic_function_34_Scanner_punctuation+0x6ff>
   39cb3:	lea    0x83936(%rip),%rdi        # bd5f0 <adamic_string_82>
   39cba:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39cbf:	movsd  0x4a349(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39cc7:	mov    %rbx,%rdi
   39cca:	mov    $0x1,%esi
   39ccf:	call   34640 <adamic_function_22_Scanner_code>
   39cd4:	ucomisd 0x4a57c(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39cdc:	jne    39e8f <adamic_function_34_Scanner_punctuation+0x72f>
   39ce2:	jp     39e8f <adamic_function_34_Scanner_punctuation+0x72f>
   39ce8:	lea    0x83981(%rip),%rdi        # bd670 <adamic_string_114>
   39cef:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39cf4:	movsd  0x4a314(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39cfc:	mov    %rbx,%rdi
   39cff:	mov    $0x1,%esi
   39d04:	call   34640 <adamic_function_22_Scanner_code>
   39d09:	ucomisd 0x4a547(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39d11:	jne    39ebc <adamic_function_34_Scanner_punctuation+0x75c>
   39d17:	jp     39ebc <adamic_function_34_Scanner_punctuation+0x75c>
   39d1d:	lea    0x8354c(%rip),%rdi        # bd270 <adamic_string_86>
   39d24:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39d29:	movsd  0x4a2df(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39d31:	mov    %rbx,%rdi
   39d34:	mov    $0x1,%esi
   39d39:	call   34640 <adamic_function_22_Scanner_code>
   39d3e:	ucomisd 0x4a512(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39d46:	lea    0x82fa3(%rip),%rax        # bccf0 <adamic_string_78>
   39d4d:	lea    0x8261c(%rip),%rdi        # bc370 <adamic_string_77>
   39d54:	jmp    39f14 <adamic_function_34_Scanner_punctuation+0x7b4>
   39d59:	movsd  0x4a2af(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39d61:	mov    %rbx,%rdi
   39d64:	mov    $0x1,%esi
   39d69:	call   34640 <adamic_function_22_Scanner_code>
   39d6e:	ucomisd 0x4a4e2(%rip),%xmm0        # 84258 <_IO_stdin_used+0x258>
   39d76:	jne    39ee9 <adamic_function_34_Scanner_punctuation+0x789>
   39d7c:	jp     39ee9 <adamic_function_34_Scanner_punctuation+0x789>
   39d82:	lea    0x82567(%rip),%rdi        # bc2f0 <adamic_string_103>
   39d89:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39d8e:	lea    0x8215b(%rip),%rdi        # bbef0 <adamic_string_64>
   39d95:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39d9a:	movsd  0x4a26e(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39da2:	mov    %rbx,%rdi
   39da5:	mov    $0x1,%esi
   39daa:	call   34640 <adamic_function_22_Scanner_code>
   39daf:	ucomisd 0x4a4a9(%rip),%xmm0        # 84260 <_IO_stdin_used+0x260>
   39db7:	jne    39f26 <adamic_function_34_Scanner_punctuation+0x7c6>
   39dbd:	jp     39f26 <adamic_function_34_Scanner_punctuation+0x7c6>
   39dc3:	lea    0x831a6(%rip),%rdi        # bcf70 <adamic_string_108>
   39dca:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39dcf:	movsd  0x4a239(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39dd7:	mov    %rbx,%rdi
   39dda:	mov    $0x1,%esi
   39ddf:	call   34640 <adamic_function_22_Scanner_code>
   39de4:	ucomisd 0x4a434(%rip),%xmm0        # 84220 <_IO_stdin_used+0x220>
   39dec:	lea    0x8277d(%rip),%rax        # bc570 <adamic_string_91>
   39df3:	lea    0x82a76(%rip),%rdi        # bc870 <adamic_string_90>
   39dfa:	jmp    39f14 <adamic_function_34_Scanner_punctuation+0x7b4>
   39dff:	movsd  0x4a209(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39e07:	mov    %rbx,%rdi
   39e0a:	mov    $0x1,%esi
   39e0f:	call   34640 <adamic_function_22_Scanner_code>
   39e14:	ucomisd 0x4a40c(%rip),%xmm0        # 84228 <_IO_stdin_used+0x228>
   39e1c:	lea    0x827cd(%rip),%rax        # bc5f0 <adamic_string_94>
   39e23:	lea    0x82ac6(%rip),%rdi        # bc8f0 <adamic_string_93>
   39e2a:	jmp    39f14 <adamic_function_34_Scanner_punctuation+0x7b4>
   39e2f:	movsd  0x4a1d9(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39e37:	mov    %rbx,%rdi
   39e3a:	mov    $0x1,%esi
   39e3f:	call   34640 <adamic_function_22_Scanner_code>
   39e44:	ucomisd 0x4a424(%rip),%xmm0        # 84270 <_IO_stdin_used+0x270>
   39e4c:	lea    0x8229d(%rip),%rax        # bc0f0 <adamic_string_101>
   39e53:	lea    0x82b16(%rip),%rdi        # bc970 <adamic_string_100>
   39e5a:	jmp    39f14 <adamic_function_34_Scanner_punctuation+0x7b4>
   39e5f:	movsd  0x4a1a9(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39e67:	mov    %rbx,%rdi
   39e6a:	mov    $0x1,%esi
   39e6f:	call   34640 <adamic_function_22_Scanner_code>
   39e74:	ucomisd 0x4a404(%rip),%xmm0        # 84280 <_IO_stdin_used+0x280>
   39e7c:	lea    0x82ced(%rip),%rax        # bcb70 <adamic_string_84>
   39e83:	lea    0x82f66(%rip),%rdi        # bcdf0 <adamic_string_83>
   39e8a:	jmp    39f14 <adamic_function_34_Scanner_punctuation+0x7b4>
   39e8f:	movsd  0x4a179(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39e97:	mov    %rbx,%rdi
   39e9a:	mov    $0x1,%esi
   39e9f:	call   34640 <adamic_function_22_Scanner_code>
   39ea4:	ucomisd 0x4a3a4(%rip),%xmm0        # 84250 <_IO_stdin_used+0x250>
   39eac:	lea    0x82d3d(%rip),%rax        # bcbf0 <adamic_string_116>
   39eb3:	lea    0x82fb6(%rip),%rdi        # bce70 <adamic_string_115>
   39eba:	jmp    39f14 <adamic_function_34_Scanner_punctuation+0x7b4>
   39ebc:	movsd  0x4a14c(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39ec4:	mov    %rbx,%rdi
   39ec7:	mov    $0x1,%esi
   39ecc:	call   34640 <adamic_function_22_Scanner_code>
   39ed1:	ucomisd 0x4a39f(%rip),%xmm0        # 84278 <_IO_stdin_used+0x278>
   39ed9:	lea    0x82810(%rip),%rax        # bc6f0 <adamic_string_88>
   39ee0:	lea    0x82789(%rip),%rdi        # bc670 <adamic_string_87>
   39ee7:	jmp    39f14 <adamic_function_34_Scanner_punctuation+0x7b4>
   39ee9:	movsd  0x4a11f(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39ef1:	mov    %rbx,%rdi
   39ef4:	mov    $0x1,%esi
   39ef9:	call   34640 <adamic_function_22_Scanner_code>
   39efe:	ucomisd 0x4a362(%rip),%xmm0        # 84268 <_IO_stdin_used+0x268>
   39f06:	lea    0x831e3(%rip),%rax        # bd0f0 <adamic_string_105>
   39f0d:	lea    0x825dc(%rip),%rdi        # bc4f0 <adamic_string_104>
   39f14:	cmovne %rax,%rdi
   39f18:	cmovp  %rax,%rdi
   39f1c:	call   6ed10 <adamic_retain>
   39f21:	pop    %rbx
   39f22:	pop    %r14
   39f24:	pop    %rbp
   39f25:	ret
   39f26:	movsd  0x4a0e2(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39f2e:	mov    %rbx,%rdi
   39f31:	mov    $0x1,%esi
   39f36:	call   34640 <adamic_function_22_Scanner_code>
   39f3b:	ucomisd 0x4a2c5(%rip),%xmm0        # 84208 <_IO_stdin_used+0x208>
   39f43:	jne    39f96 <adamic_function_34_Scanner_punctuation+0x836>
   39f45:	jp     39f96 <adamic_function_34_Scanner_punctuation+0x836>
   39f47:	movsd  0x4a0e9(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39f4f:	mov    %rbx,%rdi
   39f52:	mov    $0x1,%esi
   39f57:	call   34640 <adamic_function_22_Scanner_code>
   39f5c:	lea    0x8308d(%rip),%rdi        # bcff0 <adamic_string_109>
   39f63:	ucomisd 0x4a15d(%rip),%xmm0        # 840c8 <_IO_stdin_used+0xc8>
   39f6b:	jb     39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39f6d:	movsd  0x4a0c3(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39f75:	mov    %rdi,%r14
   39f78:	mov    %rbx,%rdi
   39f7b:	mov    $0x1,%esi
   39f80:	call   34640 <adamic_function_22_Scanner_code>
   39f85:	mov    %r14,%rdi
   39f88:	movsd  0x4a140(%rip),%xmm1        # 840d0 <_IO_stdin_used+0xd0>
   39f90:	ucomisd %xmm0,%xmm1
   39f94:	jb     39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39f96:	lea    0x82f53(%rip),%rdi        # bcef0 <adamic_string_110>
   39f9d:	jmp    39f1c <adamic_function_34_Scanner_punctuation+0x7bc>
   39fa2:	call   7f2a0 <adamic_stack_overflow>

Disassembly of section .fini:
