
/workspace/scratch/branch-shape/scalar-only:     file format elf64-x86-64


Disassembly of section .init:

Disassembly of section .plt:

Disassembly of section .plt.got:

Disassembly of section .text:

00000000000397a0 <adamic_function_34_Scanner_punctuation>:
   397a0:	push   %rbp
   397a1:	mov    %rsp,%rbp
   397a4:	push   %r14
   397a6:	push   %rbx
   397a7:	lea    0x9f9d2(%rip),%rax        # d9180 <adamic_stack_limit>
   397ae:	cmp    %rbp,(%rax)
   397b1:	ja     39ed1 <adamic_function_34_Scanner_punctuation+0x731>
   397b7:	mov    %rdi,%rbx
   397ba:	ucomisd 0x4aaa6(%rip),%xmm0        # 84268 <_IO_stdin_used+0x268>
   397c2:	jne    3983e <adamic_function_34_Scanner_punctuation+0x9e>
   397c4:	jp     3983e <adamic_function_34_Scanner_punctuation+0x9e>
   397c6:	movsd  0x4a842(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   397ce:	mov    %rbx,%rdi
   397d1:	mov    $0x1,%esi
   397d6:	call   34640 <adamic_function_22_Scanner_code>
   397db:	ucomisd 0x4aab5(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   397e3:	jne    39811 <adamic_function_34_Scanner_punctuation+0x71>
   397e5:	jp     39811 <adamic_function_34_Scanner_punctuation+0x71>
   397e7:	movsd  0x4a849(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   397ef:	mov    %rbx,%rdi
   397f2:	mov    $0x1,%esi
   397f7:	call   34640 <adamic_function_22_Scanner_code>
   397fc:	ucomisd 0x4aa94(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39804:	jne    39811 <adamic_function_34_Scanner_punctuation+0x71>
   39806:	jp     39811 <adamic_function_34_Scanner_punctuation+0x71>
   39808:	lea    0x82c61(%rip),%rdi        # bc470 <adamic_string_76>
   3980f:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39811:	movsd  0x4a7f7(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39819:	mov    %rbx,%rdi
   3981c:	mov    $0x1,%esi
   39821:	call   34640 <adamic_function_22_Scanner_code>
   39826:	ucomisd 0x4aa6a(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   3982e:	lea    0x834bb(%rip),%rax        # bccf0 <adamic_string_78>
   39835:	lea    0x82b34(%rip),%rdi        # bc370 <adamic_string_77>
   3983c:	jmp    39875 <adamic_function_34_Scanner_punctuation+0xd5>
   3983e:	ucomisd 0x4aa2a(%rip),%xmm0        # 84270 <_IO_stdin_used+0x270>
   39846:	jne    39887 <adamic_function_34_Scanner_punctuation+0xe7>
   39848:	jp     39887 <adamic_function_34_Scanner_punctuation+0xe7>
   3984a:	movsd  0x4a7be(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39852:	mov    %rbx,%rdi
   39855:	mov    $0x1,%esi
   3985a:	call   34640 <adamic_function_22_Scanner_code>
   3985f:	ucomisd 0x4aa31(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39867:	lea    0x82f82(%rip),%rax        # bc7f0 <adamic_string_80>
   3986e:	lea    0x83b7b(%rip),%rdi        # bd3f0 <adamic_string_79>
   39875:	cmovne %rax,%rdi
   39879:	cmovp  %rax,%rdi
   3987d:	call   6ed50 <adamic_retain>
   39882:	pop    %rbx
   39883:	pop    %r14
   39885:	pop    %rbp
   39886:	ret
   39887:	ucomisd 0x4a9e9(%rip),%xmm0        # 84278 <_IO_stdin_used+0x278>
   3988f:	jne    3990b <adamic_function_34_Scanner_punctuation+0x16b>
   39891:	jp     3990b <adamic_function_34_Scanner_punctuation+0x16b>
   39893:	movsd  0x4a775(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   3989b:	mov    %rbx,%rdi
   3989e:	mov    $0x1,%esi
   398a3:	call   34640 <adamic_function_22_Scanner_code>
   398a8:	ucomisd 0x4a9c8(%rip),%xmm0        # 84278 <_IO_stdin_used+0x278>
   398b0:	jne    398de <adamic_function_34_Scanner_punctuation+0x13e>
   398b2:	jp     398de <adamic_function_34_Scanner_punctuation+0x13e>
   398b4:	movsd  0x4a77c(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   398bc:	mov    %rbx,%rdi
   398bf:	mov    $0x1,%esi
   398c4:	call   34640 <adamic_function_22_Scanner_code>
   398c9:	ucomisd 0x4a9c7(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   398d1:	jne    398de <adamic_function_34_Scanner_punctuation+0x13e>
   398d3:	jp     398de <adamic_function_34_Scanner_punctuation+0x13e>
   398d5:	lea    0x83f14(%rip),%rdi        # bd7f0 <adamic_string_81>
   398dc:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   398de:	movsd  0x4a72a(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   398e6:	mov    %rbx,%rdi
   398e9:	mov    $0x1,%esi
   398ee:	call   34640 <adamic_function_22_Scanner_code>
   398f3:	ucomisd 0x4a99d(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   398fb:	jne    3996d <adamic_function_34_Scanner_punctuation+0x1cd>
   398fd:	jp     3996d <adamic_function_34_Scanner_punctuation+0x1cd>
   398ff:	lea    0x83cea(%rip),%rdi        # bd5f0 <adamic_string_82>
   39906:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   3990b:	ucomisd 0x4a96d(%rip),%xmm0        # 84280 <_IO_stdin_used+0x280>
   39913:	jne    399ca <adamic_function_34_Scanner_punctuation+0x22a>
   39919:	jp     399ca <adamic_function_34_Scanner_punctuation+0x22a>
   3991f:	movsd  0x4a6e9(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39927:	mov    %rbx,%rdi
   3992a:	mov    $0x1,%esi
   3992f:	call   34640 <adamic_function_22_Scanner_code>
   39934:	ucomisd 0x4a944(%rip),%xmm0        # 84280 <_IO_stdin_used+0x280>
   3993c:	jne    3999d <adamic_function_34_Scanner_punctuation+0x1fd>
   3993e:	jp     3999d <adamic_function_34_Scanner_punctuation+0x1fd>
   39940:	movsd  0x4a6f0(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39948:	mov    %rbx,%rdi
   3994b:	mov    $0x1,%esi
   39950:	call   34640 <adamic_function_22_Scanner_code>
   39955:	ucomisd 0x4a93b(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   3995d:	jne    3999d <adamic_function_34_Scanner_punctuation+0x1fd>
   3995f:	jp     3999d <adamic_function_34_Scanner_punctuation+0x1fd>
   39961:	lea    0x83988(%rip),%rdi        # bd2f0 <adamic_string_85>
   39968:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   3996d:	movsd  0x4a69b(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39975:	mov    %rbx,%rdi
   39978:	mov    $0x1,%esi
   3997d:	call   34640 <adamic_function_22_Scanner_code>
   39982:	ucomisd 0x4a8ee(%rip),%xmm0        # 84278 <_IO_stdin_used+0x278>
   3998a:	lea    0x831df(%rip),%rax        # bcb70 <adamic_string_84>
   39991:	lea    0x83458(%rip),%rdi        # bcdf0 <adamic_string_83>
   39998:	jmp    39875 <adamic_function_34_Scanner_punctuation+0xd5>
   3999d:	movsd  0x4a66b(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   399a5:	mov    %rbx,%rdi
   399a8:	mov    $0x1,%esi
   399ad:	call   34640 <adamic_function_22_Scanner_code>
   399b2:	ucomisd 0x4a8de(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   399ba:	jne    39a03 <adamic_function_34_Scanner_punctuation+0x263>
   399bc:	jp     39a03 <adamic_function_34_Scanner_punctuation+0x263>
   399be:	lea    0x838ab(%rip),%rdi        # bd270 <adamic_string_86>
   399c5:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   399ca:	ucomisd 0x4a86e(%rip),%xmm0        # 84240 <_IO_stdin_used+0x240>
   399d2:	jne    39a33 <adamic_function_34_Scanner_punctuation+0x293>
   399d4:	jp     39a33 <adamic_function_34_Scanner_punctuation+0x293>
   399d6:	movsd  0x4a632(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   399de:	mov    %rbx,%rdi
   399e1:	mov    $0x1,%esi
   399e6:	call   34640 <adamic_function_22_Scanner_code>
   399eb:	ucomisd 0x4a8a5(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   399f3:	jne    39a74 <adamic_function_34_Scanner_punctuation+0x2d4>
   399f5:	jp     39a74 <adamic_function_34_Scanner_punctuation+0x2d4>
   399f7:	lea    0x83772(%rip),%rdi        # bd170 <adamic_string_89>
   399fe:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39a03:	movsd  0x4a605(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39a0b:	mov    %rbx,%rdi
   39a0e:	mov    $0x1,%esi
   39a13:	call   34640 <adamic_function_22_Scanner_code>
   39a18:	ucomisd 0x4a860(%rip),%xmm0        # 84280 <_IO_stdin_used+0x280>
   39a20:	lea    0x82cc9(%rip),%rax        # bc6f0 <adamic_string_88>
   39a27:	lea    0x82c42(%rip),%rdi        # bc670 <adamic_string_87>
   39a2e:	jmp    39875 <adamic_function_34_Scanner_punctuation+0xd5>
   39a33:	ucomisd 0x4a80d(%rip),%xmm0        # 84248 <_IO_stdin_used+0x248>
   39a3b:	jne    39aa4 <adamic_function_34_Scanner_punctuation+0x304>
   39a3d:	jp     39aa4 <adamic_function_34_Scanner_punctuation+0x304>
   39a3f:	movsd  0x4a5c9(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39a47:	mov    %rbx,%rdi
   39a4a:	mov    $0x1,%esi
   39a4f:	call   34640 <adamic_function_22_Scanner_code>
   39a54:	ucomisd 0x4a83c(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39a5c:	jne    39b06 <adamic_function_34_Scanner_punctuation+0x366>
   39a62:	jp     39b06 <adamic_function_34_Scanner_punctuation+0x366>
   39a68:	lea    0x83781(%rip),%rdi        # bd1f0 <adamic_string_92>
   39a6f:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39a74:	movsd  0x4a594(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39a7c:	mov    %rbx,%rdi
   39a7f:	mov    $0x1,%esi
   39a84:	call   34640 <adamic_function_22_Scanner_code>
   39a89:	ucomisd 0x4a7af(%rip),%xmm0        # 84240 <_IO_stdin_used+0x240>
   39a91:	lea    0x82ad8(%rip),%rax        # bc570 <adamic_string_91>
   39a98:	lea    0x82dd1(%rip),%rdi        # bc870 <adamic_string_90>
   39a9f:	jmp    39875 <adamic_function_34_Scanner_punctuation+0xd5>
   39aa4:	ucomisd 0x4a77c(%rip),%xmm0        # 84228 <_IO_stdin_used+0x228>
   39aac:	jne    39b42 <adamic_function_34_Scanner_punctuation+0x3a2>
   39ab2:	jp     39b42 <adamic_function_34_Scanner_punctuation+0x3a2>
   39ab8:	movsd  0x4a550(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39ac0:	mov    %rbx,%rdi
   39ac3:	mov    $0x1,%esi
   39ac8:	call   34640 <adamic_function_22_Scanner_code>
   39acd:	ucomisd 0x4a753(%rip),%xmm0        # 84228 <_IO_stdin_used+0x228>
   39ad5:	jne    39b36 <adamic_function_34_Scanner_punctuation+0x396>
   39ad7:	jp     39b36 <adamic_function_34_Scanner_punctuation+0x396>
   39ad9:	movsd  0x4a557(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39ae1:	mov    %rbx,%rdi
   39ae4:	mov    $0x1,%esi
   39ae9:	call   34640 <adamic_function_22_Scanner_code>
   39aee:	ucomisd 0x4a732(%rip),%xmm0        # 84228 <_IO_stdin_used+0x228>
   39af6:	jne    39b36 <adamic_function_34_Scanner_punctuation+0x396>
   39af8:	jp     39b36 <adamic_function_34_Scanner_punctuation+0x396>
   39afa:	lea    0x8246f(%rip),%rdi        # bbf70 <adamic_string_95>
   39b01:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39b06:	movsd  0x4a502(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39b0e:	mov    %rbx,%rdi
   39b11:	mov    $0x1,%esi
   39b16:	call   34640 <adamic_function_22_Scanner_code>
   39b1b:	ucomisd 0x4a725(%rip),%xmm0        # 84248 <_IO_stdin_used+0x248>
   39b23:	lea    0x82ac6(%rip),%rax        # bc5f0 <adamic_string_94>
   39b2a:	lea    0x82dbf(%rip),%rdi        # bc8f0 <adamic_string_93>
   39b31:	jmp    39875 <adamic_function_34_Scanner_punctuation+0xd5>
   39b36:	lea    0x823b3(%rip),%rdi        # bbef0 <adamic_string_64>
   39b3d:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39b42:	ucomisd 0x4a73e(%rip),%xmm0        # 84288 <_IO_stdin_used+0x288>
   39b4a:	jne    39b7e <adamic_function_34_Scanner_punctuation+0x3de>
   39b4c:	jp     39b7e <adamic_function_34_Scanner_punctuation+0x3de>
   39b4e:	movsd  0x4a4ba(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39b56:	mov    %rbx,%rdi
   39b59:	mov    $0x1,%esi
   39b5e:	call   34640 <adamic_function_22_Scanner_code>
   39b63:	ucomisd 0x4a72d(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39b6b:	lea    0x82bfe(%rip),%rax        # bc770 <adamic_string_97>
   39b72:	lea    0x837f7(%rip),%rdi        # bd370 <adamic_string_96>
   39b79:	jmp    39875 <adamic_function_34_Scanner_punctuation+0xd5>
   39b7e:	ucomisd 0x4a70a(%rip),%xmm0        # 84290 <_IO_stdin_used+0x290>
   39b86:	jne    39c05 <adamic_function_34_Scanner_punctuation+0x465>
   39b88:	jp     39c05 <adamic_function_34_Scanner_punctuation+0x465>
   39b8a:	movsd  0x4a47e(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39b92:	mov    %rbx,%rdi
   39b95:	mov    $0x1,%esi
   39b9a:	call   34640 <adamic_function_22_Scanner_code>
   39b9f:	ucomisd 0x4a6e9(%rip),%xmm0        # 84290 <_IO_stdin_used+0x290>
   39ba7:	jne    39bd8 <adamic_function_34_Scanner_punctuation+0x438>
   39ba9:	jp     39bd8 <adamic_function_34_Scanner_punctuation+0x438>
   39bab:	movsd  0x4a485(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39bb3:	mov    %rbx,%rdi
   39bb6:	mov    $0x1,%esi
   39bbb:	call   34640 <adamic_function_22_Scanner_code>
   39bc0:	ucomisd 0x4a6d0(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39bc8:	jne    39bd8 <adamic_function_34_Scanner_punctuation+0x438>
   39bca:	jp     39bd8 <adamic_function_34_Scanner_punctuation+0x438>
   39bcc:	lea    0x8389d(%rip),%rdi        # bd470 <adamic_string_98>
   39bd3:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39bd8:	movsd  0x4a430(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39be0:	mov    %rbx,%rdi
   39be3:	mov    $0x1,%esi
   39be8:	call   34640 <adamic_function_22_Scanner_code>
   39bed:	ucomisd 0x4a6a3(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39bf5:	jne    39c67 <adamic_function_34_Scanner_punctuation+0x4c7>
   39bf7:	jp     39c67 <adamic_function_34_Scanner_punctuation+0x4c7>
   39bf9:	lea    0x825f0(%rip),%rdi        # bc1f0 <adamic_string_99>
   39c00:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39c05:	ucomisd 0x4a68b(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39c0d:	jne    39cc4 <adamic_function_34_Scanner_punctuation+0x524>
   39c13:	jp     39cc4 <adamic_function_34_Scanner_punctuation+0x524>
   39c19:	movsd  0x4a3ef(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39c21:	mov    %rbx,%rdi
   39c24:	mov    $0x1,%esi
   39c29:	call   34640 <adamic_function_22_Scanner_code>
   39c2e:	ucomisd 0x4a662(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39c36:	jne    39c97 <adamic_function_34_Scanner_punctuation+0x4f7>
   39c38:	jp     39c97 <adamic_function_34_Scanner_punctuation+0x4f7>
   39c3a:	movsd  0x4a3f6(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39c42:	mov    %rbx,%rdi
   39c45:	mov    $0x1,%esi
   39c4a:	call   34640 <adamic_function_22_Scanner_code>
   39c4f:	ucomisd 0x4a641(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39c57:	jne    39c97 <adamic_function_34_Scanner_punctuation+0x4f7>
   39c59:	jp     39c97 <adamic_function_34_Scanner_punctuation+0x4f7>
   39c5b:	lea    0x8278e(%rip),%rdi        # bc3f0 <adamic_string_102>
   39c62:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39c67:	movsd  0x4a3a1(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39c6f:	mov    %rbx,%rdi
   39c72:	mov    $0x1,%esi
   39c77:	call   34640 <adamic_function_22_Scanner_code>
   39c7c:	ucomisd 0x4a60c(%rip),%xmm0        # 84290 <_IO_stdin_used+0x290>
   39c84:	lea    0x82465(%rip),%rax        # bc0f0 <adamic_string_101>
   39c8b:	lea    0x82cde(%rip),%rdi        # bc970 <adamic_string_100>
   39c92:	jmp    39875 <adamic_function_34_Scanner_punctuation+0xd5>
   39c97:	movsd  0x4a371(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39c9f:	mov    %rbx,%rdi
   39ca2:	mov    $0x1,%esi
   39ca7:	call   34640 <adamic_function_22_Scanner_code>
   39cac:	ucomisd 0x4a5e4(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39cb4:	jne    39cdc <adamic_function_34_Scanner_punctuation+0x53c>
   39cb6:	jp     39cdc <adamic_function_34_Scanner_punctuation+0x53c>
   39cb8:	lea    0x82631(%rip),%rdi        # bc2f0 <adamic_string_103>
   39cbf:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39cc4:	ucomisd 0x4a5d4(%rip),%xmm0        # 842a0 <_IO_stdin_used+0x2a0>
   39ccc:	jne    39d0c <adamic_function_34_Scanner_punctuation+0x56c>
   39cce:	jp     39d0c <adamic_function_34_Scanner_punctuation+0x56c>
   39cd0:	lea    0x82499(%rip),%rdi        # bc170 <adamic_string_106>
   39cd7:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39cdc:	movsd  0x4a32c(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39ce4:	mov    %rbx,%rdi
   39ce7:	mov    $0x1,%esi
   39cec:	call   34640 <adamic_function_22_Scanner_code>
   39cf1:	ucomisd 0x4a5a7(%rip),%xmm0        # 842a0 <_IO_stdin_used+0x2a0>
   39cf9:	lea    0x833f0(%rip),%rax        # bd0f0 <adamic_string_105>
   39d00:	lea    0x827e9(%rip),%rdi        # bc4f0 <adamic_string_104>
   39d07:	jmp    39875 <adamic_function_34_Scanner_punctuation+0xd5>
   39d0c:	ucomisd 0x4a594(%rip),%xmm0        # 842a8 <_IO_stdin_used+0x2a8>
   39d14:	jne    39d93 <adamic_function_34_Scanner_punctuation+0x5f3>
   39d16:	jp     39d93 <adamic_function_34_Scanner_punctuation+0x5f3>
   39d18:	movsd  0x4a2f0(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39d20:	mov    %rbx,%rdi
   39d23:	mov    $0x1,%esi
   39d28:	call   34640 <adamic_function_22_Scanner_code>
   39d2d:	ucomisd 0x4a573(%rip),%xmm0        # 842a8 <_IO_stdin_used+0x2a8>
   39d35:	jne    39d66 <adamic_function_34_Scanner_punctuation+0x5c6>
   39d37:	jp     39d66 <adamic_function_34_Scanner_punctuation+0x5c6>
   39d39:	movsd  0x4a2f7(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39d41:	mov    %rbx,%rdi
   39d44:	mov    $0x1,%esi
   39d49:	call   34640 <adamic_function_22_Scanner_code>
   39d4e:	ucomisd 0x4a542(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39d56:	jne    39d66 <adamic_function_34_Scanner_punctuation+0x5c6>
   39d58:	jp     39d66 <adamic_function_34_Scanner_punctuation+0x5c6>
   39d5a:	lea    0x83b0f(%rip),%rdi        # bd870 <adamic_string_107>
   39d61:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39d66:	movsd  0x4a2a2(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39d6e:	mov    %rbx,%rdi
   39d71:	mov    $0x1,%esi
   39d76:	call   34640 <adamic_function_22_Scanner_code>
   39d7b:	ucomisd 0x4a525(%rip),%xmm0        # 842a8 <_IO_stdin_used+0x2a8>
   39d83:	jne    39dd7 <adamic_function_34_Scanner_punctuation+0x637>
   39d85:	jp     39dd7 <adamic_function_34_Scanner_punctuation+0x637>
   39d87:	lea    0x831e2(%rip),%rdi        # bcf70 <adamic_string_108>
   39d8e:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39d93:	ucomisd 0x4a515(%rip),%xmm0        # 842b0 <_IO_stdin_used+0x2b0>
   39d9b:	jne    39e32 <adamic_function_34_Scanner_punctuation+0x692>
   39da1:	jp     39e32 <adamic_function_34_Scanner_punctuation+0x692>
   39da7:	movsd  0x4a261(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39daf:	mov    %rbx,%rdi
   39db2:	mov    $0x1,%esi
   39db7:	call   34640 <adamic_function_22_Scanner_code>
   39dbc:	ucomisd 0x4a4d4(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39dc4:	lea    0x82ea5(%rip),%rax        # bcc70 <adamic_string_112>
   39dcb:	lea    0x8391e(%rip),%rdi        # bd6f0 <adamic_string_111>
   39dd2:	jmp    39875 <adamic_function_34_Scanner_punctuation+0xd5>
   39dd7:	movsd  0x4a231(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39ddf:	mov    %rbx,%rdi
   39de2:	mov    $0x1,%esi
   39de7:	call   34640 <adamic_function_22_Scanner_code>
   39dec:	ucomisd 0x4a434(%rip),%xmm0        # 84228 <_IO_stdin_used+0x228>
   39df4:	jne    39f00 <adamic_function_34_Scanner_punctuation+0x760>
   39dfa:	jp     39f00 <adamic_function_34_Scanner_punctuation+0x760>
   39e00:	movsd  0x4a230(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39e08:	mov    %rbx,%rdi
   39e0b:	mov    $0x1,%esi
   39e10:	call   34640 <adamic_function_22_Scanner_code>
   39e15:	lea    0x831d4(%rip),%r14        # bcff0 <adamic_string_109>
   39e1c:	ucomisd 0x4a2a4(%rip),%xmm0        # 840c8 <_IO_stdin_used+0xc8>
   39e24:	jae    39ed6 <adamic_function_34_Scanner_punctuation+0x736>
   39e2a:	mov    %r14,%rdi
   39e2d:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39e32:	ucomisd 0x4a47e(%rip),%xmm0        # 842b8 <_IO_stdin_used+0x2b8>
   39e3a:	jne    39eb9 <adamic_function_34_Scanner_punctuation+0x719>
   39e3c:	jp     39eb9 <adamic_function_34_Scanner_punctuation+0x719>
   39e3e:	movsd  0x4a1ca(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39e46:	mov    %rbx,%rdi
   39e49:	mov    $0x1,%esi
   39e4e:	call   34640 <adamic_function_22_Scanner_code>
   39e53:	ucomisd 0x4a45d(%rip),%xmm0        # 842b8 <_IO_stdin_used+0x2b8>
   39e5b:	jne    39e8c <adamic_function_34_Scanner_punctuation+0x6ec>
   39e5d:	jp     39e8c <adamic_function_34_Scanner_punctuation+0x6ec>
   39e5f:	movsd  0x4a1d1(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39e67:	mov    %rbx,%rdi
   39e6a:	mov    $0x1,%esi
   39e6f:	call   34640 <adamic_function_22_Scanner_code>
   39e74:	ucomisd 0x4a41c(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39e7c:	jne    39e8c <adamic_function_34_Scanner_punctuation+0x6ec>
   39e7e:	jp     39e8c <adamic_function_34_Scanner_punctuation+0x6ec>
   39e80:	lea    0x838e9(%rip),%rdi        # bd770 <adamic_string_113>
   39e87:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39e8c:	movsd  0x4a17c(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39e94:	mov    %rbx,%rdi
   39e97:	mov    $0x1,%esi
   39e9c:	call   34640 <adamic_function_22_Scanner_code>
   39ea1:	ucomisd 0x4a3ef(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39ea9:	jne    39f0c <adamic_function_34_Scanner_punctuation+0x76c>
   39eab:	jp     39f0c <adamic_function_34_Scanner_punctuation+0x76c>
   39ead:	lea    0x837bc(%rip),%rdi        # bd670 <adamic_string_114>
   39eb4:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39eb9:	ucomisd 0x4a1bf(%rip),%xmm0        # 84080 <_IO_stdin_used+0x80>
   39ec1:	jne    39f3c <adamic_function_34_Scanner_punctuation+0x79c>
   39ec3:	jp     39f3c <adamic_function_34_Scanner_punctuation+0x79c>
   39ec5:	lea    0x82ea4(%rip),%rdi        # bcd70 <adamic_string_117>
   39ecc:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39ed1:	call   7f2e0 <adamic_stack_overflow>
   39ed6:	movsd  0x4a15a(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39ede:	mov    %rbx,%rdi
   39ee1:	mov    $0x1,%esi
   39ee6:	call   34640 <adamic_function_22_Scanner_code>
   39eeb:	movsd  0x4a1dd(%rip),%xmm1        # 840d0 <_IO_stdin_used+0xd0>
   39ef3:	ucomisd %xmm0,%xmm1
   39ef7:	mov    %r14,%rdi
   39efa:	jb     3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39f00:	lea    0x82fe9(%rip),%rdi        # bcef0 <adamic_string_110>
   39f07:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39f0c:	movsd  0x4a0fc(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39f14:	mov    %rbx,%rdi
   39f17:	mov    $0x1,%esi
   39f1c:	call   34640 <adamic_function_22_Scanner_code>
   39f21:	ucomisd 0x4a38f(%rip),%xmm0        # 842b8 <_IO_stdin_used+0x2b8>
   39f29:	lea    0x82cc0(%rip),%rax        # bcbf0 <adamic_string_116>
   39f30:	lea    0x82f39(%rip),%rdi        # bce70 <adamic_string_115>
   39f37:	jmp    39875 <adamic_function_34_Scanner_punctuation+0xd5>
   39f3c:	ucomisd 0x4a37c(%rip),%xmm0        # 842c0 <_IO_stdin_used+0x2c0>
   39f44:	jne    39f54 <adamic_function_34_Scanner_punctuation+0x7b4>
   39f46:	jp     39f54 <adamic_function_34_Scanner_punctuation+0x7b4>
   39f48:	lea    0x839a1(%rip),%rdi        # bd8f0 <adamic_string_66>
   39f4f:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39f54:	ucomisd 0x4a36c(%rip),%xmm0        # 842c8 <_IO_stdin_used+0x2c8>
   39f5c:	jne    39f6c <adamic_function_34_Scanner_punctuation+0x7cc>
   39f5e:	jp     39f6c <adamic_function_34_Scanner_punctuation+0x7cc>
   39f60:	lea    0x81d89(%rip),%rdi        # bbcf0 <adamic_string_118>
   39f67:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39f6c:	ucomisd 0x4a35c(%rip),%xmm0        # 842d0 <_IO_stdin_used+0x2d0>
   39f74:	jne    39f84 <adamic_function_34_Scanner_punctuation+0x7e4>
   39f76:	jp     39f84 <adamic_function_34_Scanner_punctuation+0x7e4>
   39f78:	lea    0x81df1(%rip),%rdi        # bbd70 <adamic_string_119>
   39f7f:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39f84:	ucomisd 0x4a34c(%rip),%xmm0        # 842d8 <_IO_stdin_used+0x2d8>
   39f8c:	jne    39f9c <adamic_function_34_Scanner_punctuation+0x7fc>
   39f8e:	jp     39f9c <adamic_function_34_Scanner_punctuation+0x7fc>
   39f90:	lea    0x81e59(%rip),%rdi        # bbdf0 <adamic_string_120>
   39f97:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39f9c:	ucomisd 0x4a33c(%rip),%xmm0        # 842e0 <_IO_stdin_used+0x2e0>
   39fa4:	jne    39fb4 <adamic_function_34_Scanner_punctuation+0x814>
   39fa6:	jp     39fb4 <adamic_function_34_Scanner_punctuation+0x814>
   39fa8:	lea    0x81ec1(%rip),%rdi        # bbe70 <adamic_string_121>
   39faf:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39fb4:	ucomisd 0x4a1cc(%rip),%xmm0        # 84188 <_IO_stdin_used+0x188>
   39fbc:	jne    39fcc <adamic_function_34_Scanner_punctuation+0x82c>
   39fbe:	jp     39fcc <adamic_function_34_Scanner_punctuation+0x82c>
   39fc0:	lea    0x81c29(%rip),%rdi        # bbbf0 <adamic_string_122>
   39fc7:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39fcc:	ucomisd 0x4a1e4(%rip),%xmm0        # 841b8 <_IO_stdin_used+0x1b8>
   39fd4:	jne    39fe4 <adamic_function_34_Scanner_punctuation+0x844>
   39fd6:	jp     39fe4 <adamic_function_34_Scanner_punctuation+0x844>
   39fd8:	lea    0x81c91(%rip),%rdi        # bbc70 <adamic_string_123>
   39fdf:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39fe4:	ucomisd 0x4a2fc(%rip),%xmm0        # 842e8 <_IO_stdin_used+0x2e8>
   39fec:	jne    39ffc <adamic_function_34_Scanner_punctuation+0x85c>
   39fee:	jp     39ffc <adamic_function_34_Scanner_punctuation+0x85c>
   39ff0:	lea    0x81ff9(%rip),%rdi        # bbff0 <adamic_string_124>
   39ff7:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   39ffc:	ucomisd 0x4a2ec(%rip),%xmm0        # 842f0 <_IO_stdin_used+0x2f0>
   3a004:	jne    3a014 <adamic_function_34_Scanner_punctuation+0x874>
   3a006:	jp     3a014 <adamic_function_34_Scanner_punctuation+0x874>
   3a008:	lea    0x83061(%rip),%rdi        # bd070 <adamic_string_125>
   3a00f:	jmp    3987d <adamic_function_34_Scanner_punctuation+0xdd>
   3a014:	ucomisd 0x4a2dc(%rip),%xmm0        # 842f8 <_IO_stdin_used+0x2f8>
   3a01c:	lea    0x83f45(%rip),%rax        # bdf68 <adamic_string_1>
   3a023:	lea    0x82046(%rip),%rdi        # bc070 <adamic_string_53>
   3a02a:	jmp    39875 <adamic_function_34_Scanner_punctuation+0xd5>

Disassembly of section .fini:
