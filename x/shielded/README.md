# Shielded Payment Module

A privacy-preserving payment module for Evmos using zk-SNARKs (Groth16) and commitment schemes, inspired by Zcash's shielded transactions.

## Overview

This module enables private transactions where:
- ✅ Amounts are hidden on-chain
- ✅ Sender and receiver identities are hidden
- ✅ Balances are hidden
- ✅ Double-spending is prevented via nullifiers
- ✅ All validations are cryptographically proven via Groth16 circuits

## Architecture

### Core Components

1. **Commitments**: Cryptographic hashes that hide amounts
   - `Commitment = MiMC(secret, salt, amount)`

2. **Nullifiers**: Prevent double-spending
   - `Nullifier = MiMC(secret, salt)`

3. **Merkle Tree**: Stores all unspent commitments

4. **Groth16 Proofs**: Prove transaction validity without revealing amounts

### Transaction Types

1. **DepositToShielded**: Initial deposit (amount revealed once)
2. **PrivateSend**: Private transfer (amount hidden)
3. **WithdrawFromCommitment**: Withdrawal (amount revealed)

## Implementation Status

See [IMPLEMENTATION_STATUS.md](./IMPLEMENTATION_STATUS.md) for detailed status.

### ✅ Completed
- Module structure
- Types and messages
- Keeper implementation
- Commitment storage
- Nullifier tracking
- Merkle tree management
- Basic proof verification structure

### ⚠️ TODO
- Actual Groth16 proof verification (currently stubbed)
- Circuit implementation
- Trusted setup
- CLI commands
- Integration into app.go
- MiMC hash implementation

## Usage

### Deposit (Initial)

```bash
evmosd tx shielded deposit <pool-id> <amount> --from <key>
```

### Private Send

```bash
evmosd tx shielded send <pool-id> <recipient-identity> <amount> --from <key>
```

### Withdraw

```bash
evmosd tx shielded withdraw <pool-id> <commitment> <amount> <recipient> --from <key>
```

## Security Considerations

⚠️ **DO NOT deploy without:**
- Actual Groth16 proof verification
- Trusted setup completion
- Security audit
- Thorough testing

## References

- [Zcash Protocol Specification](https://zips.z.cash/protocol/protocol.pdf)
- [Groth16 Paper](https://eprint.iacr.org/2016/260)
- [MiMC Hash](https://eprint.iacr.org/2016/492)
- [Documentation](../SHIELDED_PAYMENT_FLOW.md)

