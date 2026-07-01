//go:build goexperiment.simd && amd64

#include "textflag.h"

// func adjustBrightnessAVX2(dst, src []byte, shiftVec *[32]byte, isSub bool)
TEXT ·adjustBrightnessAVX2(SB), NOSPLIT, $0
    MOVQ dst_base+0(FP), DI
    MOVQ src_base+24(FP), SI
    MOVQ dst_len+8(FP), CX      // Длина в байтах
    MOVQ shiftVec+48(FP), DX
    MOVB isSub+56(FP), AL       // Флаг вычитания

    // Загружаем вектор сдвига
    VMOVDQU (DX), Y1

loop:
    CMPQ CX, $32
    JL tail                     // Если осталось меньше 32 байт, идем в хвост

    VMOVDQU (SI), Y0            // Читаем 32 байта
    
    CMPB AL, $0
    JE add_branch

    VPSUBUSB Y1, Y0, Y0         // Вычитание с насыщением (0)
    JMP store

add_branch:
    VPADDUSB Y1, Y0, Y0         // Сложение с насыщением (255)

store:
    VMOVDQU Y0, (DI)            // Пишем 32 байта

    ADDQ $32, SI
    ADDQ $32, DI
    SUBQ $32, CX
    JMP loop

tail:
    CMPQ CX, $0
    JE done
    
    // Обработка хвоста побайтово
    MOVB (SI), BL               // Читаем байт из src
    MOVB (DX), BH               // Читаем сдвиг
    
    CMPB AL, $0
    JE tail_add
    
    // Вычитание с clamp до 0
    SUBB BH, BL
    JMP tail_store
    
tail_add:
    // Сложение с clamp до 255
    ADDB BH, BL
    JNC tail_store
    MOVB $255, BL              // Если перенос, ставим 255

tail_store:
    MOVB BL, (DI)               // Пишем байт в dst
    
    INCQ SI
    INCQ DI
    DECQ CX
    JMP tail

done:
    VZEROUPPER
    RET
