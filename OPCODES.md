# HolyCVM — Référence complète des opcodes

VM 64 bits, little-endian. Mot natif = U64/I64 (8 octets).
Convention pile : le **sommet** (stack[0]) est l'élément le plus récemment poussé.
`pop()` retire le sommet. Les opérandes sont listés **de gauche à droite = du bas vers le haut** (le dernier listé est au sommet).

---

## Légende

| Symbole | Signification |
|---------|---------------|
| `a, b → r` | consomme a (bas), b (sommet) ; pousse r |
| `→ r` | ne consomme rien, pousse r |
| `a →` | consomme a, ne pousse rien |
| gas† | gas dynamique supplémentaire |

Constantes gas : Quick=2, Fastest=3, Fast=5, Mid=8, Slow=10, Ext=20

---

## 0x00 — Arithmétique

| Opcode | Nom | Pile | Gas | Description |
|--------|-----|------|-----|-------------|
| `0x00` | `STOP` | `→` | 0 | Arrête l'exécution |
| `0x01` | `ADD` | `a, b → a+b` | 3 | Addition 64 bits (wrapping) |
| `0x02` | `MUL` | `a, b → a*b` | 5 | Multiplication (64 bits bas) |
| `0x03` | `SUB` | `a, b → a-b` | 3 | Soustraction (wrapping) |
| `0x04` | `DIV` | `a, b → a/b` | 5 | Division non signée (0 si b=0) |
| `0x05` | `SDIV` | `a, b → a/b` | 5 | Division signée I64 (0 si b=0) |
| `0x06` | `MOD` | `a, b → a%b` | 5 | Modulo non signé (0 si b=0) |
| `0x07` | `SMOD` | `a, b → a%b` | 5 | Modulo signé I64 (0 si b=0) |
| `0x08` | `ADDMOD` | `a, b, m → (a+b)%m` | 8 | Addition modulaire (intermédiaire 65 bits) |
| `0x09` | `MULMOD` | `a, b, m → (a*b)%m` | 8 | Multiplication modulaire (intermédiaire 128 bits) |
| `0x0A` | `EXP` | `base, exp → base^exp` | 8† | Exponentiation (wrapping) |
| `0x0B` | `SIGNEXTEND` | `b, x → sext(x,b)` | 5 | Extension de signe depuis l'octet b |
| `0x0C` | `MULHI` | `a, b → hi64(a*b)` | 5 | 64 bits hauts du produit 128 bits |
| `0x0D` | `MODEXP` | `base, exp, mod → base^exp%mod` | 20 | Exponentiation modulaire overflow-safe |
| `0x0E` | `ADDCARRY` | `a, b, cin → sum, cout` | 5 | Addition avec retenue (pousse 2 valeurs) |
| `0x0F` | `FIXMUL18` | `a, b → (a*b)/10^18` | 5 | Multiplication virgule fixe 18 décimales |

> **EXP gas dynamique** : +50 par octet non nul de l'exposant
> **ADDCARRY** : pousse d'abord `cout` (carry), puis `sum` (sum au sommet)

---

## 0x10 — Comparaison & Logique binaire

| Opcode | Nom | Pile | Gas | Description |
|--------|-----|------|-----|-------------|
| `0x10` | `LT` | `a, b → a<b` | 3 | Inférieur non signé (0 ou 1) |
| `0x11` | `GT` | `a, b → a>b` | 3 | Supérieur non signé (0 ou 1) |
| `0x12` | `SLT` | `a, b → a<b` | 3 | Inférieur signé I64 (0 ou 1) |
| `0x13` | `SGT` | `a, b → a>b` | 3 | Supérieur signé I64 (0 ou 1) |
| `0x14` | `EQ` | `a, b → a==b` | 3 | Égalité (0 ou 1) |
| `0x15` | `ISZERO` | `a → a==0` | 3 | Test zéro (0 ou 1) |
| `0x16` | `AND` | `a, b → a&b` | 3 | ET bit à bit |
| `0x17` | `OR` | `a, b → a\|b` | 3 | OU bit à bit |
| `0x18` | `XOR` | `a, b → a^b` | 3 | XOR bit à bit |
| `0x19` | `NOT` | `a → ~a` | 3 | NON bit à bit |
| `0x1A` | `BYTE` | `i, x → byte[i]` | 3 | Octet i (0=MSB) de x |
| `0x1B` | `SHL` | `shift, val → val<<shift` | 3 | Décalage gauche logique |
| `0x1C` | `SHR` | `shift, val → val>>shift` | 3 | Décalage droite logique (non signé) |
| `0x1D` | `SAR` | `shift, val → val>>>shift` | 3 | Décalage droite arithmétique (signé) |
| `0x1E` | `CLZ` | `a → clz(a)` | 3 | Count Leading Zeros (0..64) |
| `0x1F` | `FIXDIV18` | `a, b → (a*10^18)/b` | 5 | Division virgule fixe 18 décimales |

---

## 0x20 — Hash & Bits étendus

| Opcode | Nom | Pile | Gas | Description |
|--------|-----|------|-----|-------------|
| `0x20` | `HASH` | `offset, size → h` | 30† | Keccak-256 de mem[offset:offset+size] |
| `0x21` | `ROL` | `shift, val → val rotl shift` | 3 | Rotation gauche 64 bits |
| `0x22` | `ROR` | `shift, val → val rotr shift` | 3 | Rotation droite 64 bits |
| `0x23` | `POPCNT` | `a → popcount(a)` | 3 | Nombre de bits à 1 |
| `0x24` | `BSWAP` | `a → bswap(a)` | 3 | Inversion des octets (LE↔BE) |

> **HASH gas dynamique** : +6 par mot de 32 octets

---

## 0x30 — État du contrat

| Opcode | Nom | Pile | Gas | Description |
|--------|-----|------|-----|-------------|
| `0x30` | `ADDRESS` | `→ addr` | 2 | Adresse du contrat courant |
| `0x31` | `BALANCE` | `addr → bal` | 700 | Solde d'une adresse |
| `0x32` | `ORIGIN` | `→ addr` | 2 | Adresse origine de la transaction |
| `0x33` | `CALLER` | `→ addr` | 2 | Adresse de l'appelant direct |
| `0x34` | `CALLVALUE` | `→ val` | 2 | Valeur en wei envoyée avec l'appel |
| `0x35` | `CALLDATALOAD` | `i → data` | 3 | 8 octets de calldata à l'offset i (LE) |
| `0x36` | `CALLDATASIZE` | `→ size` | 2 | Taille des calldata en octets |
| `0x37` | `CALLDATACOPY` | `dst, src, size →` | 3† | Copie calldata en mémoire |
| `0x38` | `CODESIZE` | `→ size` | 2 | Taille du bytecode courant |
| `0x39` | `CODECOPY` | `dst, src, size →` | 3† | Copie le bytecode en mémoire |
| `0x3A` | `GASPRICE` | `→ price` | 2 | Prix du gas (wei/gas) |
| `0x3B` | `EXTCODESIZE` | `addr → size` | 700 | Taille du bytecode d'un contrat externe |
| `0x3C` | `EXTCODECOPY` | `addr, dst, src, size →` | 700† | Copie bytecode externe en mémoire |
| `0x3D` | `RETURNDATASIZE` | `→ size` | 2 | Taille du dernier returndata |
| `0x3E` | `RETURNDATACOPY` | `dst, src, size →` | 3† | Copie returndata en mémoire |
| `0x3F` | `EXTCODEHASH` | `addr → hash` | 700 | Hash Keccak du bytecode externe |

---

## 0x40 — Contexte de bloc

| Opcode | Nom | Pile | Gas | Description |
|--------|-----|------|-----|-------------|
| `0x40` | `BLOCKHASH` | `n → hash` | 20 | Hash du bloc numéro n (256 derniers) |
| `0x41` | `COINBASE` | `→ addr` | 2 | Adresse du mineur/validateur |
| `0x42` | `TIMESTAMP` | `→ t` | 2 | Timestamp du bloc (secondes Unix) |
| `0x43` | `NUMBER` | `→ n` | 2 | Numéro du bloc courant |
| `0x44` | `PREVRANDAO` | `→ r` | 2 | Valeur aléatoire du bloc (ex-DIFFICULTY) |
| `0x45` | `GASLIMIT` | `→ gl` | 2 | Gas limit du bloc |
| `0x46` | `CHAINID` | `→ id` | 2 | ID de la chaîne |
| `0x47` | `SELFBALANCE` | `→ bal` | 5 | Solde du contrat courant |
| `0x48` | `BASEFEE` | `→ fee` | 2 | Base fee du bloc (EIP-1559) |

> ~~`0x49` BLOBHASH~~ — **supprimé** dans HolyCVM
> ~~`0x4A` BLOBBASEFEE~~ — **supprimé** dans HolyCVM

---

## 0x50 — Mémoire, stockage, contrôle

| Opcode | Nom | Pile | Gas | Description |
|--------|-----|------|-----|-------------|
| `0x50` | `POP` | `a →` | 2 | Supprime le sommet de pile |
| `0x51` | `MLOAD` | `addr → val` | 3† | Charge 8 octets little-endian de la mémoire |
| `0x52` | `MSTORE` | `addr, val →` | 3† | Stocke 8 octets little-endian en mémoire |
| `0x53` | `MSTORE8` | `addr, val →` | 3† | Stocke 1 octet (bits 0-7) en mémoire |
| `0x54` | `SLOAD` | `key → val` | 800 | Charge depuis le stockage persistant |
| `0x55` | `SSTORE` | `key, val →` | dyn | Écrit dans le stockage persistant |
| `0x56` | `JUMP` | `dest →` | 8 | Saut inconditionnel (dest doit être JUMPDEST) |
| `0x57` | `JUMPI` | `dest, cond →` | 10 | Saut conditionnel si cond ≠ 0 |
| `0x58` | `PC` | `→ pc` | 2 | Valeur courante du compteur ordinal |
| `0x59` | `MSIZE` | `→ size` | 2 | Taille de la mémoire allouée (octets) |
| `0x5A` | `GAS` | `→ gas` | 2 | Gas restant après cette instruction |
| `0x5B` | `JUMPDEST` | `→` | 1 | Marque une destination de saut valide |
| `0x5C` | `TLOAD` | `key → val` | 100 | Charge depuis le stockage transitoire |
| `0x5D` | `TSTORE` | `key, val →` | 100 | Écrit dans le stockage transitoire |
| `0x5E` | `MCOPY` | `dst, src, size →` | 3† | Copie mémoire → mémoire |
| `0x5F` | `PUSH0` | `→ 0` | 2 | Pousse la constante 0 |

> **MLOAD/MSTORE** : la mémoire s'étend par mots de 32 octets ; gas dynamique si extension

---

## 0x60–0x67 — PUSH (immédiat little-endian)

Les N octets suivant l'opcode sont lus en little-endian et zero-étendus à 64 bits.

| Opcode | Nom | Octets immédiats | Gas |
|--------|-----|-----------------|-----|
| `0x60` | `PUSH1` | 1 | 3 |
| `0x61` | `PUSH2` | 2 | 3 |
| `0x62` | `PUSH3` | 3 | 3 |
| `0x63` | `PUSH4` | 4 | 3 |
| `0x64` | `PUSH5` | 5 | 3 |
| `0x65` | `PUSH6` | 6 | 3 |
| `0x66` | `PUSH7` | 7 | 3 |
| `0x67` | `PUSH8` | 8 (I64 complet) | 3 |

> ~~PUSH9–PUSH32~~ : **supprimés** (valeurs >64 bits incompatibles avec HolyCVM)

---

## 0x68–0x6F — Accès mémoire typé (HolyCVM)

| Opcode | Nom | Pile | Gas | Description |
|--------|-----|------|-----|-------------|
| `0x68` | `MLOAD16` | `addr → val` | 3† | Charge 2 octets LE, zero-étendu à U64 |
| `0x69` | `MLOAD16S` | `addr → val` | 3† | Charge 2 octets LE, sign-étendu à I64 |
| `0x6A` | `MLOAD32` | `addr → val` | 3† | Charge 4 octets LE, zero-étendu à U64 |
| `0x6B` | `MLOAD32S` | `addr → val` | 3† | Charge 4 octets LE, sign-étendu à I64 |
| `0x6C` | `MSTORE16` | `addr, val →` | 3† | Stocke les 2 octets bas de val |
| `0x6D` | `MSTORE32` | `addr, val →` | 3† | Stocke les 4 octets bas de val |
| `0x6E` | `SEXT8` | `a → sext8(a)` | 3 | Sign-extend octet (bit 7) vers I64 |
| `0x6F` | `SEXT16` | `a → sext16(a)` | 3 | Sign-extend 16 bits (bit 15) vers I64 |

---

## 0x70–0x73 — Conversions de type (HolyCVM)

| Opcode | Nom | Pile | Gas | Description |
|--------|-----|------|-----|-------------|
| `0x70` | `SEXT32` | `a → sext32(a)` | 3 | Sign-extend 32 bits (bit 31) vers I64 |
| `0x71` | `TRUNC8` | `a → a & 0xFF` | 3 | Troncature à U8 |
| `0x72` | `TRUNC16` | `a → a & 0xFFFF` | 3 | Troncature à U16 |
| `0x73` | `TRUNC32` | `a → a & 0xFFFFFFFF` | 3 | Troncature à U32 |

---

## 0x80–0x87 — DUP (duplication de pile)

| Opcode | Nom | Effet | Gas |
|--------|-----|-------|-----|
| `0x80` | `DUP1` | Copie stack[0] (sommet) | 3 |
| `0x81` | `DUP2` | Copie stack[1] | 3 |
| `0x82` | `DUP3` | Copie stack[2] | 3 |
| `0x83` | `DUP4` | Copie stack[3] | 3 |
| `0x84` | `DUP5` | Copie stack[4] | 3 |
| `0x85` | `DUP6` | Copie stack[5] | 3 |
| `0x86` | `DUP7` | Copie stack[6] | 3 |
| `0x87` | `DUP8` | Copie stack[7] | 3 |

> ~~DUP9–DUP16~~ : **supprimés** dans HolyCVM

---

## 0x90–0x97 — SWAP (échange de pile)

| Opcode | Nom | Effet | Gas |
|--------|-----|-------|-----|
| `0x90` | `SWAP1` | Échange stack[0] ↔ stack[1] | 3 |
| `0x91` | `SWAP2` | Échange stack[0] ↔ stack[2] | 3 |
| `0x92` | `SWAP3` | Échange stack[0] ↔ stack[3] | 3 |
| `0x93` | `SWAP4` | Échange stack[0] ↔ stack[4] | 3 |
| `0x94` | `SWAP5` | Échange stack[0] ↔ stack[5] | 3 |
| `0x95` | `SWAP6` | Échange stack[0] ↔ stack[6] | 3 |
| `0x96` | `SWAP7` | Échange stack[0] ↔ stack[7] | 3 |
| `0x97` | `SWAP8` | Échange stack[0] ↔ stack[8] | 3 |

> ~~SWAP9–SWAP16~~ : **supprimés** dans HolyCVM

---

## 0xA0–0xA4 — Logs (événements)

| Opcode | Nom | Pile | Gas | Description |
|--------|-----|------|-----|-------------|
| `0xA0` | `LOG0` | `offset, size →` | dyn | Émet un log sans topic |
| `0xA1` | `LOG1` | `offset, size, t1 →` | dyn | Émet un log avec 1 topic |
| `0xA2` | `LOG2` | `offset, size, t1, t2 →` | dyn | Émet un log avec 2 topics |
| `0xA3` | `LOG3` | `offset, size, t1, t2, t3 →` | dyn | Émet un log avec 3 topics |
| `0xA4` | `LOG4` | `offset, size, t1, t2, t3, t4 →` | dyn | Émet un log avec 4 topics |

---

## 0xF0+ — Appels inter-contrats & contrôle

| Opcode | Nom | Pile | Gas | Description |
|--------|-----|------|-----|-------------|
| `0xF0` | `CREATE` | `val, offset, size → addr` | 32000† | Crée un nouveau contrat |
| `0xF1` | `CALL` | `gas, addr, val, aOff, aSize, rOff, rSize → success` | dyn | Appel de contrat externe |
| `0xF3` | `RETURN` | `offset, size →` | dyn | Termine l'exécution, retourne mem[offset:offset+size] |
| `0xF4` | `DELEGATECALL` | `gas, addr, aOff, aSize, rOff, rSize → success` | dyn | Appel délégué (contexte du caller) |
| `0xF5` | `CREATE2` | `val, offset, size, salt → addr` | 32000† | Crée un contrat à adresse déterministe |
| `0xFA` | `STATICCALL` | `gas, addr, aOff, aSize, rOff, rSize → success` | dyn | Appel en lecture seule |
| `0xFD` | `REVERT` | `offset, size →` | dyn | Annule l'exécution, retourne les données d'erreur |
| `0xFE` | `INVALID` | — | — | Instruction invalide (abort) |

> ~~`0xF2` CALLCODE~~ : **supprimé** (dangereux, déprécié)
> ~~`0xFF` SELFDESTRUCT~~ : **supprimé** (dangereux)

---

## Notes importantes pour le compilateur

### Ordre des arguments sur la pile

La pile est **LIFO**. L'opcode pop le **sommet** en premier. Exemple pour `SUB` :

```
PUSH1(10)   → pile: [10]
PUSH1(3)    → pile: [10, 3]   ← 3 au sommet
SUB         → a=3, b=10 → 3-10 = -7 ⚠️
```

Pour calculer `10 - 3`, pousser dans cet ordre : `PUSH(3)` puis `PUSH(10)`, SUB → 10-3=7.

### RETURN

```
RETURN : offset = pop() (sommet), size = pop() (deuxième)
```

Pour retourner 8 octets depuis mem[0] :
```
PUSH1(8)   ← size
PUSH0      ← offset (au sommet)
RETURN
```

### MSTORE / MLOAD

```
MSTORE : offset = pop() (sommet), value = pop() (deuxième)
MLOAD  : offset = pop() (sommet) → value
```

Pour stocker une valeur en mémoire à l'offset 0 :
```
<calculer valeur>   → pile: [val]
PUSH0               → pile: [val, 0]   ← offset au sommet
MSTORE
```

### JUMPDEST obligatoire

Toute destination de `JUMP` ou `JUMPI` **doit** être marquée par `JUMPDEST` (0x5B).

### Encodage PUSH little-endian

`PUSH8 0xDE0B6B3A7640000` (= 1×10¹⁸) s'encode :
```
67 00 00 64 A7 B3 B6 E0 0D
```
Les octets sont dans l'ordre LE (octet de poids faible en premier).

---

## Résumé : total 99 opcodes actifs

| Groupe | Plage | Nombre |
|--------|-------|--------|
| Arithmétique | 0x00–0x0F | 16 |
| Comparaison/bits | 0x10–0x1F | 16 |
| Hash/bits étendus | 0x20–0x24 | 5 |
| État contrat | 0x30–0x3F | 16 |
| Contexte bloc | 0x40–0x48 | 9 |
| Mémoire/stockage | 0x50–0x5F | 16 |
| PUSH1–PUSH8 | 0x60–0x67 | 8 |
| Mémoire typée | 0x68–0x6F | 8 |
| Conversions type | 0x70–0x73 | 4 |
| DUP1–DUP8 | 0x80–0x87 | 8 |
| SWAP1–SWAP8 | 0x90–0x97 | 8 |
| Logs | 0xA0–0xA4 | 5 |
| Appels/contrôle | 0xF0–0xFE | 8 |
