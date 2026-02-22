// Test des opcodes 0x50-0x5F (pile, mémoire, stockage)

// Stockage persistant
SStore(0, 42);
I64 v = SLoad(0);

// Stockage transitoire
TStore(1, 100);
I64 tv = TLoad(1);

// Infos mémoire / gas
I64 ms  = MSize();
I64 gas = Gas();
I64 pc  = Pc();

// Copie mémoire→mémoire
MStore(0, 0xDEADBEEF);
MCopy(32, 0, 8);

// Pop
I64 x = 10 + 5;
