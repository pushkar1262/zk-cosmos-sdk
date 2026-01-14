# Trusted Setup Tool

This tool generates Groth16 proving and verification keys for the shielded payment circuits.

## Overview

The trusted setup is a critical security component. It generates:
- **Proving Keys (PK)**: Used off-chain to generate proofs (must be kept secret)
- **Verification Keys (VK)**: Used on-chain to verify proofs (public, stored on-chain)
- **Constraint Systems (CCS)**: Compiled circuit representations (for proof generation)

## Security Considerations

⚠️ **CRITICAL**: The trusted setup must be performed securely:

1. **Multi-party ceremony**: Use a powers-of-tau ceremony with multiple participants
2. **Secure environment**: Run on air-gapped machines
3. **Verify keys**: All participants should verify the generated keys
4. **Destroy toxic waste**: Securely delete any intermediate values
5. **Audit**: Have the ceremony audited by security experts

## Usage

### Build the Tool

```bash
cd x/shielded/cmd/trusted-setup
go build -o trusted-setup main.go
```

### Generate Keys for a Single Circuit

```bash
# Deposit circuit
./trusted-setup -circuit deposit -output ./keys

# Private send circuit (specify Merkle depth)
./trusted-setup -circuit private_send -merkle-depth 10 -output ./keys

# Withdrawal circuit
./trusted-setup -circuit withdrawal -output ./keys
```

### Generate Keys for All Circuits

```bash
# Generate keys for all circuits and common Merkle depths (1-20)
./trusted-setup -all -output ./keys -max-merkle-depth 20
```

## Output Files

The tool generates the following files:

### Deposit Circuit
- `deposit_pk.bin` - Proving key (keep secret!)
- `deposit_vk.bin` - Verification key (submit to governance)
- `deposit_ccs.bin` - Constraint system (for proof generation)

### Private Send Circuit (per depth)
- `private_send_depth_N_pk.bin` - Proving key for depth N
- `private_send_depth_N_vk.bin` - Verification key for depth N
- `private_send_depth_N_ccs.bin` - Constraint system for depth N

### Withdrawal Circuit
- `withdrawal_pk.bin` - Proving key (keep secret!)
- `withdrawal_vk.bin` - Verification key (submit to governance)
- `withdrawal_ccs.bin` - Constraint system (for proof generation)

## Storing Verification Keys On-Chain

After generating keys, store verification keys on-chain via governance:

```bash
# Submit governance proposal to set deposit verification key
evmosd tx gov submit-proposal \
  shielded set-verification-key-proposal \
  deposit 0 keys/deposit_vk.bin \
  "Set Deposit Verification Key" \
  "Setting verification key for deposit circuit" \
  --from validator \
  --fees 1000aevmos

# Submit governance proposal to set private send verification key (depth 10)
evmosd tx gov submit-proposal \
  shielded set-verification-key-proposal \
  private_send 10 keys/private_send_depth_10_vk.bin \
  "Set Private Send Verification Key (Depth 10)" \
  "Setting verification key for private send circuit with Merkle depth 10" \
  --from validator \
  --fees 1000aevmos

# Submit governance proposal to set withdrawal verification key
evmosd tx gov submit-proposal \
  shielded set-verification-key-proposal \
  withdrawal 0 keys/withdrawal_vk.bin \
  "Set Withdrawal Verification Key" \
  "Setting verification key for withdrawal circuit" \
  --from validator \
  --fees 1000aevmos
```

## Direct Submission (for testing only)

For testing purposes, you can submit directly (will fail if not governance):

```bash
evmosd tx shielded set-verification-key deposit 0 keys/deposit_vk.bin --from validator --fees 1000aevmos
```

## Verification Key Hash

Each verification key has a hash printed during generation. Use this to verify:
1. The key matches across different participants
2. The key stored on-chain matches the generated key

## Multi-Party Ceremony

For production, use a multi-party ceremony:

1. **Phase 1**: Each participant generates their contribution
2. **Phase 2**: Contributions are combined
3. **Phase 3**: Final keys are generated
4. **Verification**: All participants verify the final keys
5. **Storage**: Verification keys are stored on-chain via governance

## Next Steps

1. ✅ Generate keys using this tool
2. ✅ Verify key hashes match across participants
3. ✅ Submit verification keys via governance proposals
4. ✅ Distribute proving keys securely to users
5. ✅ Users can now generate proofs using the proving keys

## References

- [Groth16 Paper](https://eprint.iacr.org/2016/260)
- [Powers of Tau Ceremony](https://github.com/iden3/snarkjs#7-prepare-phase-2)
- [Gnark Documentation](https://docs.gnark.consensys.net/)

