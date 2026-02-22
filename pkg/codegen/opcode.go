package codegen

import "fmt"

type Opcode byte

const (
	// Les 16 opcodes arithmétiques de la VM
	OP_STOP       Opcode = 0x00 // Arrête l'exécution
	OP_ADD        Opcode = 0x01 // a, b → a+b          (wrapping)            gas=3
	OP_MUL        Opcode = 0x02 // a, b → a*b          (64 bits bas)         gas=5
	OP_SUB        Opcode = 0x03 // a, b → a-b                                gas=3
	OP_DIV        Opcode = 0x04 // a, b → a/b          (non signé, 0 si b=0) gas=5
	OP_SDIV       Opcode = 0x05 // a, b → a/b          (signé)               gas=5
	OP_MOD        Opcode = 0x06 // a, b → a%b          (non signé)           gas=5
	OP_SMOD       Opcode = 0x07 // a, b → a%b          (signé)               gas=5
	OP_ADDMOD     Opcode = 0x08 // a, b, m → (a+b)%m   (intermédiaire 65b)   gas=8
	OP_MULMOD     Opcode = 0x09 // a, b, m → (a*b)%m   (intermédiaire 128b)  gas=8
	OP_EXP        Opcode = 0x0A // a, b → a^b          (wrapping)            gas=8+dyn
	OP_SIGNEXTEND Opcode = 0x0B // b, x → sext(x, b)   (extension signe)     gas=5
	OP_MULHI      Opcode = 0x0C // a, b → hi64(a*b)    (64 bits hauts)       gas=5
	OP_MODEXP     Opcode = 0x0D // base, exp, mod → base^exp%mod             gas=20
	OP_ADDCARRY   Opcode = 0x0E // a, b, cin → sum, cout                     gas=5
	OP_FIXMUL18   Opcode = 0x0F // a, b → (a*b)/10^18  (virgule fixe)        gas=5

	// Opcodes de comparaison et logique
	OP_LT     Opcode = 0x10 // a, b → a<b (non signé)                  gas=3
	OP_GT     Opcode = 0x11 // a, b → a>b (non signé)                  gas=3
	OP_SLT    Opcode = 0x12 // a, b → a<b (signé)                      gas=3
	OP_SGT    Opcode = 0x13 // a, b → a>b (signé)                      gas=3
	OP_EQ     Opcode = 0x14 // a, b → a==b                             gas=3
	OP_ISZERO Opcode = 0x15 // a → a==0                                gas=3
	OP_AND    Opcode = 0x16 // a, b → a&b                              gas=3
	OP_OR     Opcode = 0x17 // a, b → a|b                              gas=3
	OP_XOR    Opcode = 0x18 // a, b → a^b                              gas=3
	OP_NOT    Opcode = 0x19 // a → ^a                                  gas=3
	OP_BYTE   Opcode = 0x1A // i, x → byte i de x                     gas=3
	OP_SHL    Opcode = 0x1B // shift, val → val << shift               gas=3
	OP_SHR    Opcode = 0x1C // shift, val → val >> shift               gas=3
	OP_SAR     Opcode = 0x1D // shift, val → val >>> shift (signé)      gas=3
	OP_CLZ      Opcode = 0x1E // a → clz(a)             Count Leading Zeros   gas=3
	OP_FIXDIV18 Opcode = 0x1F // a, b → (a*10^18)/b    Division virgule fixe gas=5

	// Opcodes hash et manipulation de bits
	OP_HASH   Opcode = 0x20 // offset, size → h      Keccak-256            gas=30+dyn
	OP_ROL    Opcode = 0x21 // shift, val → rotl(val,shift) Rotation gauche gas=3
	OP_ROR    Opcode = 0x22 // shift, val → rotr(val,shift) Rotation droite gas=3
	OP_POPCNT Opcode = 0x23 // a → popcount(a)       Nombre de bits à 1    gas=3
	OP_BSWAP  Opcode = 0x24 // a → bswap(a)          Inversion octets      gas=3

	// Opcodes d'état du contrat
	OP_ADDRESS        Opcode = 0x30 // → addr            Adresse du contrat courant    gas=2
	OP_BALANCE        Opcode = 0x31 // addr → bal        Solde d'une adresse           gas=700
	OP_ORIGIN         Opcode = 0x32 // → addr            Origine de la transaction     gas=2
	OP_CALLER         Opcode = 0x33 // → addr            Appelant direct               gas=2
	OP_CALLVALUE      Opcode = 0x34 // → val             Valeur wei de l'appel         gas=2
	OP_CALLDATALOAD   Opcode = 0x35 // i → data          8 octets calldata à offset i  gas=3
	OP_CALLDATASIZE   Opcode = 0x36 // → size            Taille des calldata           gas=2
	OP_CALLDATACOPY   Opcode = 0x37 // dst,src,size →    Copie calldata en mémoire     gas=3+dyn
	OP_CODESIZE       Opcode = 0x38 // → size            Taille du bytecode courant    gas=2
	OP_CODECOPY       Opcode = 0x39 // dst,src,size →    Copie bytecode en mémoire     gas=3+dyn
	OP_GASPRICE       Opcode = 0x3A // → price           Prix du gas (wei/gas)         gas=2
	OP_EXTCODESIZE    Opcode = 0x3B // addr → size       Taille bytecode externe       gas=700
	OP_EXTCODECOPY    Opcode = 0x3C // addr,dst,src,size → Copie bytecode externe      gas=700+dyn
	OP_RETURNDATASIZE Opcode = 0x3D // → size            Taille du dernier returndata  gas=2
	OP_RETURNDATACOPY Opcode = 0x3E // dst,src,size →    Copie returndata en mémoire   gas=3+dyn
	OP_EXTCODEHASH    Opcode = 0x3F // addr → hash       Keccak du bytecode externe    gas=700

	// Opcodes de contexte de bloc
	OP_BLOCKHASH  Opcode = 0x40 // n → hash        Hash du bloc n (256 derniers) gas=20
	OP_COINBASE   Opcode = 0x41 // → addr          Adresse du validateur         gas=2
	OP_TIMESTAMP  Opcode = 0x42 // → t             Timestamp Unix du bloc        gas=2
	OP_NUMBER     Opcode = 0x43 // → n             Numéro du bloc courant        gas=2
	OP_PREVRANDAO Opcode = 0x44 // → r             Valeur aléatoire du bloc      gas=2
	OP_GASLIMIT   Opcode = 0x45 // → gl            Gas limit du bloc             gas=2
	OP_CHAINID    Opcode = 0x46 // → id            ID de la chaîne               gas=2
	OP_SELFBALANCE Opcode = 0x47 // → bal          Solde du contrat courant      gas=5
	OP_BASEFEE    Opcode = 0x48 // → fee           Base fee du bloc (EIP-1559)   gas=2

	// Opcodes de pile et mémoire
	OP_POP      Opcode = 0x50 // a →               Supprime le sommet            gas=2
	OP_MLOAD    Opcode = 0x51 // addr → val        Charge 8 octets LE            gas=3
	OP_MSTORE   Opcode = 0x52 // addr, val →       Stocke 8 octets LE            gas=3
	OP_MSTORE8  Opcode = 0x53 // addr, val →       Stocke 1 octet                gas=3
	OP_SLOAD    Opcode = 0x54 // key → val         Stockage persistant lecture   gas=800
	OP_SSTORE   Opcode = 0x55 // key, val →        Stockage persistant écriture  gas=dyn
	OP_JUMP     Opcode = 0x56 // dest →            Saut inconditionnel           gas=8
	OP_JUMPI    Opcode = 0x57 // dest, cond →      Saut conditionnel             gas=10
	OP_PC       Opcode = 0x58 // → pc              Compteur ordinal courant      gas=2
	OP_MSIZE    Opcode = 0x59 // → size            Taille mémoire allouée        gas=2
	OP_GAS      Opcode = 0x5A // → gas             Gas restant                   gas=2
	OP_JUMPDEST Opcode = 0x5B // →                 Destination de saut valide    gas=1
	OP_TLOAD    Opcode = 0x5C // key → val         Stockage transitoire lecture  gas=100
	OP_TSTORE   Opcode = 0x5D // key, val →        Stockage transitoire écriture gas=100
	OP_MCOPY    Opcode = 0x5E // dst, src, size →  Copie mémoire→mémoire         gas=3+dyn
	OP_PUSH0    Opcode = 0x5F // → 0               Pousse 0                      gas=2

	// Opcodes PUSH : poussent N octets (little-endian, étendu à 64 bits)
	OP_PUSH1 Opcode = 0x60 // Pousse 1 octet
	OP_PUSH2 Opcode = 0x61 // Pousse 2 octets
	OP_PUSH3 Opcode = 0x62 // Pousse 3 octets
	OP_PUSH4 Opcode = 0x63 // Pousse 4 octets
	OP_PUSH5 Opcode = 0x64 // Pousse 5 octets
	OP_PUSH6 Opcode = 0x65 // Pousse 6 octets
	OP_PUSH7 Opcode = 0x66 // Pousse 7 octets
	OP_PUSH8 Opcode = 0x67 // Pousse 8 octets (I64 complet)            gas=2

	// DUP et SWAP (1-8 seulement dans HolyCVM)
	OP_DUP1  Opcode = 0x80
	OP_DUP2  Opcode = 0x81
	OP_SWAP1 Opcode = 0x90
	OP_SWAP2 Opcode = 0x91

	// Contrôle
	OP_RETURN Opcode = 0xF3 // offset, size → retourne données
	OP_REVERT Opcode = 0xFD // offset, size → revert
)

var opcodeInfo = map[Opcode]struct {
	Name    string
	Gas     int
	Args    int // nombre d'opérandes consommées sur la pile
	Results int // nombre de résultats poussés sur la pile
}{
	OP_STOP:       {"STOP", 0, 0, 0},
	OP_ADD:        {"ADD", 3, 2, 1},
	OP_MUL:        {"MUL", 5, 2, 1},
	OP_SUB:        {"SUB", 3, 2, 1},
	OP_DIV:        {"DIV", 5, 2, 1},
	OP_SDIV:       {"SDIV", 5, 2, 1},
	OP_MOD:        {"MOD", 5, 2, 1},
	OP_SMOD:       {"SMOD", 5, 2, 1},
	OP_ADDMOD:     {"ADDMOD", 8, 3, 1},
	OP_MULMOD:     {"MULMOD", 8, 3, 1},
	OP_EXP:        {"EXP", 8, 2, 1},
	OP_SIGNEXTEND: {"SIGNEXTEND", 5, 2, 1},
	OP_MULHI:      {"MULHI", 5, 2, 1},
	OP_MODEXP:     {"MODEXP", 20, 3, 1},
	OP_ADDCARRY:   {"ADDCARRY", 5, 3, 2},
	OP_FIXMUL18:   {"FIXMUL18", 5, 2, 1},
	// Comparaison et logique
	OP_LT:     {"LT", 3, 2, 1},
	OP_GT:     {"GT", 3, 2, 1},
	OP_SLT:    {"SLT", 3, 2, 1},
	OP_SGT:    {"SGT", 3, 2, 1},
	OP_EQ:     {"EQ", 3, 2, 1},
	OP_ISZERO: {"ISZERO", 3, 1, 1},
	OP_AND:    {"AND", 3, 2, 1},
	OP_OR:     {"OR", 3, 2, 1},
	OP_XOR:    {"XOR", 3, 2, 1},
	OP_NOT:    {"NOT", 3, 1, 1},
	OP_SHL:    {"SHL", 3, 2, 1},
	OP_SHR:    {"SHR", 3, 2, 1},
	OP_SAR:      {"SAR", 3, 2, 1},
	OP_CLZ:      {"CLZ", 3, 1, 1},
	OP_FIXDIV18: {"FIXDIV18", 5, 2, 1},
	// Hash et bits
	OP_HASH:   {"HASH", 30, 2, 1},
	OP_ROL:    {"ROL", 3, 2, 1},
	OP_ROR:    {"ROR", 3, 2, 1},
	OP_POPCNT: {"POPCNT", 3, 1, 1},
	OP_BSWAP:  {"BSWAP", 3, 1, 1},

	// État du contrat
	OP_ADDRESS:        {"ADDRESS", 2, 0, 1},
	OP_BALANCE:        {"BALANCE", 700, 1, 1},
	OP_ORIGIN:         {"ORIGIN", 2, 0, 1},
	OP_CALLER:         {"CALLER", 2, 0, 1},
	OP_CALLVALUE:      {"CALLVALUE", 2, 0, 1},
	OP_CALLDATALOAD:   {"CALLDATALOAD", 3, 1, 1},
	OP_CALLDATASIZE:   {"CALLDATASIZE", 2, 0, 1},
	OP_CALLDATACOPY:   {"CALLDATACOPY", 3, 3, 0},
	OP_CODESIZE:       {"CODESIZE", 2, 0, 1},
	OP_CODECOPY:       {"CODECOPY", 3, 3, 0},
	OP_GASPRICE:       {"GASPRICE", 2, 0, 1},
	OP_EXTCODESIZE:    {"EXTCODESIZE", 700, 1, 1},
	OP_EXTCODECOPY:    {"EXTCODECOPY", 700, 4, 0},
	OP_RETURNDATASIZE: {"RETURNDATASIZE", 2, 0, 1},
	OP_RETURNDATACOPY: {"RETURNDATACOPY", 3, 3, 0},
	OP_EXTCODEHASH:    {"EXTCODEHASH", 700, 1, 1},

	// Contexte de bloc
	OP_BLOCKHASH:   {"BLOCKHASH", 20, 1, 1},
	OP_COINBASE:    {"COINBASE", 2, 0, 1},
	OP_TIMESTAMP:   {"TIMESTAMP", 2, 0, 1},
	OP_NUMBER:      {"NUMBER", 2, 0, 1},
	OP_PREVRANDAO:  {"PREVRANDAO", 2, 0, 1},
	OP_GASLIMIT:    {"GASLIMIT", 2, 0, 1},
	OP_CHAINID:     {"CHAINID", 2, 0, 1},
	OP_SELFBALANCE: {"SELFBALANCE", 5, 0, 1},
	OP_BASEFEE:     {"BASEFEE", 2, 0, 1},

	// Pile et mémoire
	OP_POP:      {"POP", 2, 1, 0},
	OP_MLOAD:    {"MLOAD", 3, 1, 1},
	OP_MSTORE:   {"MSTORE", 3, 2, 0},
	OP_MSTORE8:  {"MSTORE8", 3, 2, 0},
	OP_SLOAD:    {"SLOAD", 800, 1, 1},
	OP_SSTORE:   {"SSTORE", 0, 2, 0}, // gas dynamique (cold/warm/refund)
	OP_JUMP:     {"JUMP", 8, 1, 0},
	OP_JUMPI:    {"JUMPI", 10, 2, 0},
	OP_PC:       {"PC", 2, 0, 1},
	OP_MSIZE:    {"MSIZE", 2, 0, 1},
	OP_GAS:      {"GAS", 2, 0, 1},
	OP_JUMPDEST: {"JUMPDEST", 1, 0, 0},
	OP_TLOAD:    {"TLOAD", 100, 1, 1},
	OP_TSTORE:   {"TSTORE", 100, 2, 0},
	OP_MCOPY:    {"MCOPY", 3, 3, 0},
	OP_PUSH0:    {"PUSH0", 2, 0, 1},

	// PUSH
	OP_PUSH1: {"PUSH1", 2, 0, 1},
	OP_PUSH2: {"PUSH2", 2, 0, 1},
	OP_PUSH3: {"PUSH3", 2, 0, 1},
	OP_PUSH4: {"PUSH4", 2, 0, 1},
	OP_PUSH5: {"PUSH5", 2, 0, 1},
	OP_PUSH6: {"PUSH6", 2, 0, 1},
	OP_PUSH7: {"PUSH7", 2, 0, 1},
	OP_PUSH8: {"PUSH8", 2, 0, 1},

	// DUP/SWAP
	OP_DUP1:  {"DUP1", 3, 1, 2},
	OP_DUP2:  {"DUP2", 3, 2, 3},
	OP_SWAP1: {"SWAP1", 3, 2, 2},
	OP_SWAP2: {"SWAP2", 3, 3, 3},

	// Contrôle
	OP_RETURN: {"RETURN", 0, 2, 0},
	OP_REVERT: {"REVERT", 0, 2, 0},
}

func (op Opcode) String() string {
	if info, ok := opcodeInfo[op]; ok {
		return info.Name
	}
	return fmt.Sprintf("UNKNOWN(0x%02X)", byte(op))
}

// IsPush retourne true si l'opcode est un PUSH1-PUSH8.
func (op Opcode) IsPush() bool {
	return op >= OP_PUSH1 && op <= OP_PUSH8
}

// PushSize retourne le nombre d'octets immédiats pour un PUSH.
func (op Opcode) PushSize() int {
	if op.IsPush() {
		return int(op-OP_PUSH1) + 1
	}
	return 0
}

// Instruction représente une instruction bytecode avec opérande optionnel.
type Instruction struct {
	Op      Opcode
	Operand int64 // utilisé uniquement par PUSH
}

// Gas retourne le coût en gas de l'instruction.
func (inst Instruction) Gas() int {
	if info, ok := opcodeInfo[inst.Op]; ok {
		return info.Gas
	}
	return 0
}

func (inst Instruction) String() string {
	if inst.Op.IsPush() {
		return fmt.Sprintf("%s 0x%X", inst.Op.String(), uint64(inst.Operand))
	}
	if inst.Op == OP_PUSH0 {
		return "PUSH0"
	}
	return inst.Op.String()
}
