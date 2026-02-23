// Test des opcodes 0x68-0x6F (mémoire sous-64 bits)

// Stores
MStore16(0, 0xABCD);
MStore32(8, 0xDEADBEEF);

// Loads zero-étendus
I64 a = MLoad16(0);
I64 b = MLoad32(8);

// Loads sign-étendus
I64 c = MLoad16S(0);
I64 d = MLoad32S(8);

// Sign-extension scalaire
I64 e = Sext8(0xFF);
I64 f = Sext16(0x8000);
