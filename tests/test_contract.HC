// Test des opcodes 0x30-0x3F (état du contrat)
I64 self   = Address();
I64 bal    = Balance(self);
I64 origin = Origin();
I64 caller = Caller();
I64 val    = CallValue();
I64 data0  = CallDataLoad(0);
I64 dsz    = CallDataSize();
I64 gsz    = CodeSize();
I64 gp     = GasPrice();
I64 rdsz   = ReturnDataSize();
I64 xsz    = ExtCodeSize(caller);
I64 xh     = ExtCodeHash(caller);

// Copies mémoire (3 args)
CallDataCopy(0, 0, 32);
CodeCopy(32, 0, 32);
ReturnDataCopy(64, 0, 32);

// ExtCodeCopy (4 args)
ExtCodeCopy(caller, 96, 0, 32);
