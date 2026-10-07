
/workspace/scratch/branch-shape/before:     file format elf64-x86-64


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
   39791:	ja     39eb1 <adamic_function_34_Scanner_punctuation+0x731>
   39797:	mov    %rdi,%rbx
   3979a:	ucomisd 0x4aac6(%rip),%xmm0        # 84268 <_IO_stdin_used+0x268>
   397a2:	jne    3981e <adamic_function_34_Scanner_punctuation+0x9e>
   397a4:	jp     3981e <adamic_function_34_Scanner_punctuation+0x9e>
   397a6:	movsd  0x4a862(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   397ae:	mov    %rbx,%rdi
   397b1:	mov    $0x1,%esi
   397b6:	call   34640 <adamic_function_22_Scanner_code>
   397bb:	ucomisd 0x4aad5(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   397c3:	jne    397f1 <adamic_function_34_Scanner_punctuation+0x71>
   397c5:	jp     397f1 <adamic_function_34_Scanner_punctuation+0x71>
   397c7:	movsd  0x4a869(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   397cf:	mov    %rbx,%rdi
   397d2:	mov    $0x1,%esi
   397d7:	call   34640 <adamic_function_22_Scanner_code>
   397dc:	ucomisd 0x4aab4(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   397e4:	jne    397f1 <adamic_function_34_Scanner_punctuation+0x71>
   397e6:	jp     397f1 <adamic_function_34_Scanner_punctuation+0x71>
   397e8:	lea    0x82c81(%rip),%rdi        # bc470 <adamic_string_76>
   397ef:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   397f1:	movsd  0x4a817(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   397f9:	mov    %rbx,%rdi
   397fc:	mov    $0x1,%esi
   39801:	call   34640 <adamic_function_22_Scanner_code>
   39806:	ucomisd 0x4aa8a(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   3980e:	lea    0x834db(%rip),%rax        # bccf0 <adamic_string_78>
   39815:	lea    0x82b54(%rip),%rdi        # bc370 <adamic_string_77>
   3981c:	jmp    39855 <adamic_function_34_Scanner_punctuation+0xd5>
   3981e:	ucomisd 0x4aa4a(%rip),%xmm0        # 84270 <_IO_stdin_used+0x270>
   39826:	jne    39867 <adamic_function_34_Scanner_punctuation+0xe7>
   39828:	jp     39867 <adamic_function_34_Scanner_punctuation+0xe7>
   3982a:	movsd  0x4a7de(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39832:	mov    %rbx,%rdi
   39835:	mov    $0x1,%esi
   3983a:	call   34640 <adamic_function_22_Scanner_code>
   3983f:	ucomisd 0x4aa51(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39847:	lea    0x82fa2(%rip),%rax        # bc7f0 <adamic_string_80>
   3984e:	lea    0x83b9b(%rip),%rdi        # bd3f0 <adamic_string_79>
   39855:	cmovne %rax,%rdi
   39859:	cmovp  %rax,%rdi
   3985d:	call   6ed70 <adamic_retain>
   39862:	pop    %rbx
   39863:	pop    %r14
   39865:	pop    %rbp
   39866:	ret
   39867:	ucomisd 0x4aa09(%rip),%xmm0        # 84278 <_IO_stdin_used+0x278>
   3986f:	jne    398eb <adamic_function_34_Scanner_punctuation+0x16b>
   39871:	jp     398eb <adamic_function_34_Scanner_punctuation+0x16b>
   39873:	movsd  0x4a795(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   3987b:	mov    %rbx,%rdi
   3987e:	mov    $0x1,%esi
   39883:	call   34640 <adamic_function_22_Scanner_code>
   39888:	ucomisd 0x4a9e8(%rip),%xmm0        # 84278 <_IO_stdin_used+0x278>
   39890:	jne    398be <adamic_function_34_Scanner_punctuation+0x13e>
   39892:	jp     398be <adamic_function_34_Scanner_punctuation+0x13e>
   39894:	movsd  0x4a79c(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   3989c:	mov    %rbx,%rdi
   3989f:	mov    $0x1,%esi
   398a4:	call   34640 <adamic_function_22_Scanner_code>
   398a9:	ucomisd 0x4a9e7(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   398b1:	jne    398be <adamic_function_34_Scanner_punctuation+0x13e>
   398b3:	jp     398be <adamic_function_34_Scanner_punctuation+0x13e>
   398b5:	lea    0x83f34(%rip),%rdi        # bd7f0 <adamic_string_81>
   398bc:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   398be:	movsd  0x4a74a(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   398c6:	mov    %rbx,%rdi
   398c9:	mov    $0x1,%esi
   398ce:	call   34640 <adamic_function_22_Scanner_code>
   398d3:	ucomisd 0x4a9bd(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   398db:	jne    3994d <adamic_function_34_Scanner_punctuation+0x1cd>
   398dd:	jp     3994d <adamic_function_34_Scanner_punctuation+0x1cd>
   398df:	lea    0x83d0a(%rip),%rdi        # bd5f0 <adamic_string_82>
   398e6:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   398eb:	ucomisd 0x4a98d(%rip),%xmm0        # 84280 <_IO_stdin_used+0x280>
   398f3:	jne    399aa <adamic_function_34_Scanner_punctuation+0x22a>
   398f9:	jp     399aa <adamic_function_34_Scanner_punctuation+0x22a>
   398ff:	movsd  0x4a709(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39907:	mov    %rbx,%rdi
   3990a:	mov    $0x1,%esi
   3990f:	call   34640 <adamic_function_22_Scanner_code>
   39914:	ucomisd 0x4a964(%rip),%xmm0        # 84280 <_IO_stdin_used+0x280>
   3991c:	jne    3997d <adamic_function_34_Scanner_punctuation+0x1fd>
   3991e:	jp     3997d <adamic_function_34_Scanner_punctuation+0x1fd>
   39920:	movsd  0x4a710(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39928:	mov    %rbx,%rdi
   3992b:	mov    $0x1,%esi
   39930:	call   34640 <adamic_function_22_Scanner_code>
   39935:	ucomisd 0x4a95b(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   3993d:	jne    3997d <adamic_function_34_Scanner_punctuation+0x1fd>
   3993f:	jp     3997d <adamic_function_34_Scanner_punctuation+0x1fd>
   39941:	lea    0x839a8(%rip),%rdi        # bd2f0 <adamic_string_85>
   39948:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   3994d:	movsd  0x4a6bb(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39955:	mov    %rbx,%rdi
   39958:	mov    $0x1,%esi
   3995d:	call   34640 <adamic_function_22_Scanner_code>
   39962:	ucomisd 0x4a90e(%rip),%xmm0        # 84278 <_IO_stdin_used+0x278>
   3996a:	lea    0x831ff(%rip),%rax        # bcb70 <adamic_string_84>
   39971:	lea    0x83478(%rip),%rdi        # bcdf0 <adamic_string_83>
   39978:	jmp    39855 <adamic_function_34_Scanner_punctuation+0xd5>
   3997d:	movsd  0x4a68b(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39985:	mov    %rbx,%rdi
   39988:	mov    $0x1,%esi
   3998d:	call   34640 <adamic_function_22_Scanner_code>
   39992:	ucomisd 0x4a8fe(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   3999a:	jne    399e3 <adamic_function_34_Scanner_punctuation+0x263>
   3999c:	jp     399e3 <adamic_function_34_Scanner_punctuation+0x263>
   3999e:	lea    0x838cb(%rip),%rdi        # bd270 <adamic_string_86>
   399a5:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   399aa:	ucomisd 0x4a88e(%rip),%xmm0        # 84240 <_IO_stdin_used+0x240>
   399b2:	jne    39a13 <adamic_function_34_Scanner_punctuation+0x293>
   399b4:	jp     39a13 <adamic_function_34_Scanner_punctuation+0x293>
   399b6:	movsd  0x4a652(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   399be:	mov    %rbx,%rdi
   399c1:	mov    $0x1,%esi
   399c6:	call   34640 <adamic_function_22_Scanner_code>
   399cb:	ucomisd 0x4a8c5(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   399d3:	jne    39a54 <adamic_function_34_Scanner_punctuation+0x2d4>
   399d5:	jp     39a54 <adamic_function_34_Scanner_punctuation+0x2d4>
   399d7:	lea    0x83792(%rip),%rdi        # bd170 <adamic_string_89>
   399de:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   399e3:	movsd  0x4a625(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   399eb:	mov    %rbx,%rdi
   399ee:	mov    $0x1,%esi
   399f3:	call   34640 <adamic_function_22_Scanner_code>
   399f8:	ucomisd 0x4a880(%rip),%xmm0        # 84280 <_IO_stdin_used+0x280>
   39a00:	lea    0x82ce9(%rip),%rax        # bc6f0 <adamic_string_88>
   39a07:	lea    0x82c62(%rip),%rdi        # bc670 <adamic_string_87>
   39a0e:	jmp    39855 <adamic_function_34_Scanner_punctuation+0xd5>
   39a13:	ucomisd 0x4a82d(%rip),%xmm0        # 84248 <_IO_stdin_used+0x248>
   39a1b:	jne    39a84 <adamic_function_34_Scanner_punctuation+0x304>
   39a1d:	jp     39a84 <adamic_function_34_Scanner_punctuation+0x304>
   39a1f:	movsd  0x4a5e9(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39a27:	mov    %rbx,%rdi
   39a2a:	mov    $0x1,%esi
   39a2f:	call   34640 <adamic_function_22_Scanner_code>
   39a34:	ucomisd 0x4a85c(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39a3c:	jne    39ae6 <adamic_function_34_Scanner_punctuation+0x366>
   39a42:	jp     39ae6 <adamic_function_34_Scanner_punctuation+0x366>
   39a48:	lea    0x837a1(%rip),%rdi        # bd1f0 <adamic_string_92>
   39a4f:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39a54:	movsd  0x4a5b4(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39a5c:	mov    %rbx,%rdi
   39a5f:	mov    $0x1,%esi
   39a64:	call   34640 <adamic_function_22_Scanner_code>
   39a69:	ucomisd 0x4a7cf(%rip),%xmm0        # 84240 <_IO_stdin_used+0x240>
   39a71:	lea    0x82af8(%rip),%rax        # bc570 <adamic_string_91>
   39a78:	lea    0x82df1(%rip),%rdi        # bc870 <adamic_string_90>
   39a7f:	jmp    39855 <adamic_function_34_Scanner_punctuation+0xd5>
   39a84:	ucomisd 0x4a79c(%rip),%xmm0        # 84228 <_IO_stdin_used+0x228>
   39a8c:	jne    39b22 <adamic_function_34_Scanner_punctuation+0x3a2>
   39a92:	jp     39b22 <adamic_function_34_Scanner_punctuation+0x3a2>
   39a98:	movsd  0x4a570(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39aa0:	mov    %rbx,%rdi
   39aa3:	mov    $0x1,%esi
   39aa8:	call   34640 <adamic_function_22_Scanner_code>
   39aad:	ucomisd 0x4a773(%rip),%xmm0        # 84228 <_IO_stdin_used+0x228>
   39ab5:	jne    39b16 <adamic_function_34_Scanner_punctuation+0x396>
   39ab7:	jp     39b16 <adamic_function_34_Scanner_punctuation+0x396>
   39ab9:	movsd  0x4a577(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39ac1:	mov    %rbx,%rdi
   39ac4:	mov    $0x1,%esi
   39ac9:	call   34640 <adamic_function_22_Scanner_code>
   39ace:	ucomisd 0x4a752(%rip),%xmm0        # 84228 <_IO_stdin_used+0x228>
   39ad6:	jne    39b16 <adamic_function_34_Scanner_punctuation+0x396>
   39ad8:	jp     39b16 <adamic_function_34_Scanner_punctuation+0x396>
   39ada:	lea    0x8248f(%rip),%rdi        # bbf70 <adamic_string_95>
   39ae1:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39ae6:	movsd  0x4a522(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39aee:	mov    %rbx,%rdi
   39af1:	mov    $0x1,%esi
   39af6:	call   34640 <adamic_function_22_Scanner_code>
   39afb:	ucomisd 0x4a745(%rip),%xmm0        # 84248 <_IO_stdin_used+0x248>
   39b03:	lea    0x82ae6(%rip),%rax        # bc5f0 <adamic_string_94>
   39b0a:	lea    0x82ddf(%rip),%rdi        # bc8f0 <adamic_string_93>
   39b11:	jmp    39855 <adamic_function_34_Scanner_punctuation+0xd5>
   39b16:	lea    0x823d3(%rip),%rdi        # bbef0 <adamic_string_64>
   39b1d:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39b22:	ucomisd 0x4a75e(%rip),%xmm0        # 84288 <_IO_stdin_used+0x288>
   39b2a:	jne    39b5e <adamic_function_34_Scanner_punctuation+0x3de>
   39b2c:	jp     39b5e <adamic_function_34_Scanner_punctuation+0x3de>
   39b2e:	movsd  0x4a4da(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39b36:	mov    %rbx,%rdi
   39b39:	mov    $0x1,%esi
   39b3e:	call   34640 <adamic_function_22_Scanner_code>
   39b43:	ucomisd 0x4a74d(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39b4b:	lea    0x82c1e(%rip),%rax        # bc770 <adamic_string_97>
   39b52:	lea    0x83817(%rip),%rdi        # bd370 <adamic_string_96>
   39b59:	jmp    39855 <adamic_function_34_Scanner_punctuation+0xd5>
   39b5e:	ucomisd 0x4a72a(%rip),%xmm0        # 84290 <_IO_stdin_used+0x290>
   39b66:	jne    39be5 <adamic_function_34_Scanner_punctuation+0x465>
   39b68:	jp     39be5 <adamic_function_34_Scanner_punctuation+0x465>
   39b6a:	movsd  0x4a49e(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39b72:	mov    %rbx,%rdi
   39b75:	mov    $0x1,%esi
   39b7a:	call   34640 <adamic_function_22_Scanner_code>
   39b7f:	ucomisd 0x4a709(%rip),%xmm0        # 84290 <_IO_stdin_used+0x290>
   39b87:	jne    39bb8 <adamic_function_34_Scanner_punctuation+0x438>
   39b89:	jp     39bb8 <adamic_function_34_Scanner_punctuation+0x438>
   39b8b:	movsd  0x4a4a5(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39b93:	mov    %rbx,%rdi
   39b96:	mov    $0x1,%esi
   39b9b:	call   34640 <adamic_function_22_Scanner_code>
   39ba0:	ucomisd 0x4a6f0(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39ba8:	jne    39bb8 <adamic_function_34_Scanner_punctuation+0x438>
   39baa:	jp     39bb8 <adamic_function_34_Scanner_punctuation+0x438>
   39bac:	lea    0x838bd(%rip),%rdi        # bd470 <adamic_string_98>
   39bb3:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39bb8:	movsd  0x4a450(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39bc0:	mov    %rbx,%rdi
   39bc3:	mov    $0x1,%esi
   39bc8:	call   34640 <adamic_function_22_Scanner_code>
   39bcd:	ucomisd 0x4a6c3(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39bd5:	jne    39c47 <adamic_function_34_Scanner_punctuation+0x4c7>
   39bd7:	jp     39c47 <adamic_function_34_Scanner_punctuation+0x4c7>
   39bd9:	lea    0x82610(%rip),%rdi        # bc1f0 <adamic_string_99>
   39be0:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39be5:	ucomisd 0x4a6ab(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39bed:	jne    39ca4 <adamic_function_34_Scanner_punctuation+0x524>
   39bf3:	jp     39ca4 <adamic_function_34_Scanner_punctuation+0x524>
   39bf9:	movsd  0x4a40f(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39c01:	mov    %rbx,%rdi
   39c04:	mov    $0x1,%esi
   39c09:	call   34640 <adamic_function_22_Scanner_code>
   39c0e:	ucomisd 0x4a682(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39c16:	jne    39c77 <adamic_function_34_Scanner_punctuation+0x4f7>
   39c18:	jp     39c77 <adamic_function_34_Scanner_punctuation+0x4f7>
   39c1a:	movsd  0x4a416(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39c22:	mov    %rbx,%rdi
   39c25:	mov    $0x1,%esi
   39c2a:	call   34640 <adamic_function_22_Scanner_code>
   39c2f:	ucomisd 0x4a661(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39c37:	jne    39c77 <adamic_function_34_Scanner_punctuation+0x4f7>
   39c39:	jp     39c77 <adamic_function_34_Scanner_punctuation+0x4f7>
   39c3b:	lea    0x827ae(%rip),%rdi        # bc3f0 <adamic_string_102>
   39c42:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39c47:	movsd  0x4a3c1(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39c4f:	mov    %rbx,%rdi
   39c52:	mov    $0x1,%esi
   39c57:	call   34640 <adamic_function_22_Scanner_code>
   39c5c:	ucomisd 0x4a62c(%rip),%xmm0        # 84290 <_IO_stdin_used+0x290>
   39c64:	lea    0x82485(%rip),%rax        # bc0f0 <adamic_string_101>
   39c6b:	lea    0x82cfe(%rip),%rdi        # bc970 <adamic_string_100>
   39c72:	jmp    39855 <adamic_function_34_Scanner_punctuation+0xd5>
   39c77:	movsd  0x4a391(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39c7f:	mov    %rbx,%rdi
   39c82:	mov    $0x1,%esi
   39c87:	call   34640 <adamic_function_22_Scanner_code>
   39c8c:	ucomisd 0x4a604(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39c94:	jne    39cbc <adamic_function_34_Scanner_punctuation+0x53c>
   39c96:	jp     39cbc <adamic_function_34_Scanner_punctuation+0x53c>
   39c98:	lea    0x82651(%rip),%rdi        # bc2f0 <adamic_string_103>
   39c9f:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39ca4:	ucomisd 0x4a5f4(%rip),%xmm0        # 842a0 <_IO_stdin_used+0x2a0>
   39cac:	jne    39cec <adamic_function_34_Scanner_punctuation+0x56c>
   39cae:	jp     39cec <adamic_function_34_Scanner_punctuation+0x56c>
   39cb0:	lea    0x824b9(%rip),%rdi        # bc170 <adamic_string_106>
   39cb7:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39cbc:	movsd  0x4a34c(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39cc4:	mov    %rbx,%rdi
   39cc7:	mov    $0x1,%esi
   39ccc:	call   34640 <adamic_function_22_Scanner_code>
   39cd1:	ucomisd 0x4a5c7(%rip),%xmm0        # 842a0 <_IO_stdin_used+0x2a0>
   39cd9:	lea    0x83410(%rip),%rax        # bd0f0 <adamic_string_105>
   39ce0:	lea    0x82809(%rip),%rdi        # bc4f0 <adamic_string_104>
   39ce7:	jmp    39855 <adamic_function_34_Scanner_punctuation+0xd5>
   39cec:	ucomisd 0x4a5b4(%rip),%xmm0        # 842a8 <_IO_stdin_used+0x2a8>
   39cf4:	jne    39d73 <adamic_function_34_Scanner_punctuation+0x5f3>
   39cf6:	jp     39d73 <adamic_function_34_Scanner_punctuation+0x5f3>
   39cf8:	movsd  0x4a310(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39d00:	mov    %rbx,%rdi
   39d03:	mov    $0x1,%esi
   39d08:	call   34640 <adamic_function_22_Scanner_code>
   39d0d:	ucomisd 0x4a593(%rip),%xmm0        # 842a8 <_IO_stdin_used+0x2a8>
   39d15:	jne    39d46 <adamic_function_34_Scanner_punctuation+0x5c6>
   39d17:	jp     39d46 <adamic_function_34_Scanner_punctuation+0x5c6>
   39d19:	movsd  0x4a317(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39d21:	mov    %rbx,%rdi
   39d24:	mov    $0x1,%esi
   39d29:	call   34640 <adamic_function_22_Scanner_code>
   39d2e:	ucomisd 0x4a562(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39d36:	jne    39d46 <adamic_function_34_Scanner_punctuation+0x5c6>
   39d38:	jp     39d46 <adamic_function_34_Scanner_punctuation+0x5c6>
   39d3a:	lea    0x83b2f(%rip),%rdi        # bd870 <adamic_string_107>
   39d41:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39d46:	movsd  0x4a2c2(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39d4e:	mov    %rbx,%rdi
   39d51:	mov    $0x1,%esi
   39d56:	call   34640 <adamic_function_22_Scanner_code>
   39d5b:	ucomisd 0x4a545(%rip),%xmm0        # 842a8 <_IO_stdin_used+0x2a8>
   39d63:	jne    39db7 <adamic_function_34_Scanner_punctuation+0x637>
   39d65:	jp     39db7 <adamic_function_34_Scanner_punctuation+0x637>
   39d67:	lea    0x83202(%rip),%rdi        # bcf70 <adamic_string_108>
   39d6e:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39d73:	ucomisd 0x4a535(%rip),%xmm0        # 842b0 <_IO_stdin_used+0x2b0>
   39d7b:	jne    39e12 <adamic_function_34_Scanner_punctuation+0x692>
   39d81:	jp     39e12 <adamic_function_34_Scanner_punctuation+0x692>
   39d87:	movsd  0x4a281(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39d8f:	mov    %rbx,%rdi
   39d92:	mov    $0x1,%esi
   39d97:	call   34640 <adamic_function_22_Scanner_code>
   39d9c:	ucomisd 0x4a4f4(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39da4:	lea    0x82ec5(%rip),%rax        # bcc70 <adamic_string_112>
   39dab:	lea    0x8393e(%rip),%rdi        # bd6f0 <adamic_string_111>
   39db2:	jmp    39855 <adamic_function_34_Scanner_punctuation+0xd5>
   39db7:	movsd  0x4a251(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39dbf:	mov    %rbx,%rdi
   39dc2:	mov    $0x1,%esi
   39dc7:	call   34640 <adamic_function_22_Scanner_code>
   39dcc:	ucomisd 0x4a454(%rip),%xmm0        # 84228 <_IO_stdin_used+0x228>
   39dd4:	jne    39ee0 <adamic_function_34_Scanner_punctuation+0x760>
   39dda:	jp     39ee0 <adamic_function_34_Scanner_punctuation+0x760>
   39de0:	movsd  0x4a250(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39de8:	mov    %rbx,%rdi
   39deb:	mov    $0x1,%esi
   39df0:	call   34640 <adamic_function_22_Scanner_code>
   39df5:	lea    0x831f4(%rip),%r14        # bcff0 <adamic_string_109>
   39dfc:	ucomisd 0x4a2c4(%rip),%xmm0        # 840c8 <_IO_stdin_used+0xc8>
   39e04:	jae    39eb6 <adamic_function_34_Scanner_punctuation+0x736>
   39e0a:	mov    %r14,%rdi
   39e0d:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39e12:	ucomisd 0x4a49e(%rip),%xmm0        # 842b8 <_IO_stdin_used+0x2b8>
   39e1a:	jne    39e99 <adamic_function_34_Scanner_punctuation+0x719>
   39e1c:	jp     39e99 <adamic_function_34_Scanner_punctuation+0x719>
   39e1e:	movsd  0x4a1ea(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39e26:	mov    %rbx,%rdi
   39e29:	mov    $0x1,%esi
   39e2e:	call   34640 <adamic_function_22_Scanner_code>
   39e33:	ucomisd 0x4a47d(%rip),%xmm0        # 842b8 <_IO_stdin_used+0x2b8>
   39e3b:	jne    39e6c <adamic_function_34_Scanner_punctuation+0x6ec>
   39e3d:	jp     39e6c <adamic_function_34_Scanner_punctuation+0x6ec>
   39e3f:	movsd  0x4a1f1(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39e47:	mov    %rbx,%rdi
   39e4a:	mov    $0x1,%esi
   39e4f:	call   34640 <adamic_function_22_Scanner_code>
   39e54:	ucomisd 0x4a43c(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39e5c:	jne    39e6c <adamic_function_34_Scanner_punctuation+0x6ec>
   39e5e:	jp     39e6c <adamic_function_34_Scanner_punctuation+0x6ec>
   39e60:	lea    0x83909(%rip),%rdi        # bd770 <adamic_string_113>
   39e67:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39e6c:	movsd  0x4a19c(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39e74:	mov    %rbx,%rdi
   39e77:	mov    $0x1,%esi
   39e7c:	call   34640 <adamic_function_22_Scanner_code>
   39e81:	ucomisd 0x4a40f(%rip),%xmm0        # 84298 <_IO_stdin_used+0x298>
   39e89:	jne    39eec <adamic_function_34_Scanner_punctuation+0x76c>
   39e8b:	jp     39eec <adamic_function_34_Scanner_punctuation+0x76c>
   39e8d:	lea    0x837dc(%rip),%rdi        # bd670 <adamic_string_114>
   39e94:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39e99:	ucomisd 0x4a1df(%rip),%xmm0        # 84080 <_IO_stdin_used+0x80>
   39ea1:	jne    39f1c <adamic_function_34_Scanner_punctuation+0x79c>
   39ea3:	jp     39f1c <adamic_function_34_Scanner_punctuation+0x79c>
   39ea5:	lea    0x82ec4(%rip),%rdi        # bcd70 <adamic_string_117>
   39eac:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39eb1:	call   7f300 <adamic_stack_overflow>
   39eb6:	movsd  0x4a17a(%rip),%xmm0        # 84038 <_IO_stdin_used+0x38>
   39ebe:	mov    %rbx,%rdi
   39ec1:	mov    $0x1,%esi
   39ec6:	call   34640 <adamic_function_22_Scanner_code>
   39ecb:	movsd  0x4a1fd(%rip),%xmm1        # 840d0 <_IO_stdin_used+0xd0>
   39ed3:	ucomisd %xmm0,%xmm1
   39ed7:	mov    %r14,%rdi
   39eda:	jb     3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39ee0:	lea    0x83009(%rip),%rdi        # bcef0 <adamic_string_110>
   39ee7:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39eec:	movsd  0x4a11c(%rip),%xmm0        # 84010 <_IO_stdin_used+0x10>
   39ef4:	mov    %rbx,%rdi
   39ef7:	mov    $0x1,%esi
   39efc:	call   34640 <adamic_function_22_Scanner_code>
   39f01:	ucomisd 0x4a3af(%rip),%xmm0        # 842b8 <_IO_stdin_used+0x2b8>
   39f09:	lea    0x82ce0(%rip),%rax        # bcbf0 <adamic_string_116>
   39f10:	lea    0x82f59(%rip),%rdi        # bce70 <adamic_string_115>
   39f17:	jmp    39855 <adamic_function_34_Scanner_punctuation+0xd5>
   39f1c:	ucomisd 0x4a39c(%rip),%xmm0        # 842c0 <_IO_stdin_used+0x2c0>
   39f24:	jne    39f34 <adamic_function_34_Scanner_punctuation+0x7b4>
   39f26:	jp     39f34 <adamic_function_34_Scanner_punctuation+0x7b4>
   39f28:	lea    0x839c1(%rip),%rdi        # bd8f0 <adamic_string_66>
   39f2f:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39f34:	ucomisd 0x4a38c(%rip),%xmm0        # 842c8 <_IO_stdin_used+0x2c8>
   39f3c:	jne    39f4c <adamic_function_34_Scanner_punctuation+0x7cc>
   39f3e:	jp     39f4c <adamic_function_34_Scanner_punctuation+0x7cc>
   39f40:	lea    0x81da9(%rip),%rdi        # bbcf0 <adamic_string_118>
   39f47:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39f4c:	ucomisd 0x4a37c(%rip),%xmm0        # 842d0 <_IO_stdin_used+0x2d0>
   39f54:	jne    39f64 <adamic_function_34_Scanner_punctuation+0x7e4>
   39f56:	jp     39f64 <adamic_function_34_Scanner_punctuation+0x7e4>
   39f58:	lea    0x81e11(%rip),%rdi        # bbd70 <adamic_string_119>
   39f5f:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39f64:	ucomisd 0x4a36c(%rip),%xmm0        # 842d8 <_IO_stdin_used+0x2d8>
   39f6c:	jne    39f7c <adamic_function_34_Scanner_punctuation+0x7fc>
   39f6e:	jp     39f7c <adamic_function_34_Scanner_punctuation+0x7fc>
   39f70:	lea    0x81e79(%rip),%rdi        # bbdf0 <adamic_string_120>
   39f77:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39f7c:	ucomisd 0x4a35c(%rip),%xmm0        # 842e0 <_IO_stdin_used+0x2e0>
   39f84:	jne    39f94 <adamic_function_34_Scanner_punctuation+0x814>
   39f86:	jp     39f94 <adamic_function_34_Scanner_punctuation+0x814>
   39f88:	lea    0x81ee1(%rip),%rdi        # bbe70 <adamic_string_121>
   39f8f:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39f94:	ucomisd 0x4a1ec(%rip),%xmm0        # 84188 <_IO_stdin_used+0x188>
   39f9c:	jne    39fac <adamic_function_34_Scanner_punctuation+0x82c>
   39f9e:	jp     39fac <adamic_function_34_Scanner_punctuation+0x82c>
   39fa0:	lea    0x81c49(%rip),%rdi        # bbbf0 <adamic_string_122>
   39fa7:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39fac:	ucomisd 0x4a204(%rip),%xmm0        # 841b8 <_IO_stdin_used+0x1b8>
   39fb4:	jne    39fc4 <adamic_function_34_Scanner_punctuation+0x844>
   39fb6:	jp     39fc4 <adamic_function_34_Scanner_punctuation+0x844>
   39fb8:	lea    0x81cb1(%rip),%rdi        # bbc70 <adamic_string_123>
   39fbf:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39fc4:	ucomisd 0x4a31c(%rip),%xmm0        # 842e8 <_IO_stdin_used+0x2e8>
   39fcc:	jne    39fdc <adamic_function_34_Scanner_punctuation+0x85c>
   39fce:	jp     39fdc <adamic_function_34_Scanner_punctuation+0x85c>
   39fd0:	lea    0x82019(%rip),%rdi        # bbff0 <adamic_string_124>
   39fd7:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39fdc:	ucomisd 0x4a30c(%rip),%xmm0        # 842f0 <_IO_stdin_used+0x2f0>
   39fe4:	jne    39ff4 <adamic_function_34_Scanner_punctuation+0x874>
   39fe6:	jp     39ff4 <adamic_function_34_Scanner_punctuation+0x874>
   39fe8:	lea    0x83081(%rip),%rdi        # bd070 <adamic_string_125>
   39fef:	jmp    3985d <adamic_function_34_Scanner_punctuation+0xdd>
   39ff4:	ucomisd 0x4a2fc(%rip),%xmm0        # 842f8 <_IO_stdin_used+0x2f8>
   39ffc:	lea    0x83f65(%rip),%rax        # bdf68 <adamic_string_1>
   3a003:	lea    0x82066(%rip),%rdi        # bc070 <adamic_string_53>
   3a00a:	jmp    39855 <adamic_function_34_Scanner_punctuation+0xd5>

Disassembly of section .fini:
