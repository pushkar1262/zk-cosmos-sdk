# Complete Guide: From Starting Chain to Private Transfer

## Step 0: Build and Start Local Chain

**IMPORTANT:** The chain must be initialized with accounts in genesis. Use one of these methods:

### Method 1: Use local_node.sh (Recommended)

```bash
# Navigate to project directory
cd /data/evmos

# Build evmosd binary
make install

# Start local node (this will initialize and start the chain with dev0, dev1, dev2, dev3 accounts)
./local_node.sh

# This creates accounts: dev0, dev1, dev2, dev3 with funds in genesis
# The script will automatically use one of these accounts if validator has no funds
```

### Method 2: Manual Setup

```bash
# Navigate to project directory
cd /data/evmos

# Build evmosd binary
make install

# Initialize chain
evmosd init mynode --chain-id evmos_9000-4 --home ~/.tmp-evmosd

# Add validator key
evmosd keys add validator --keyring-backend test --home ~/.tmp-evmosd

# Add genesis account (MUST do this before starting chain!)
evmosd add-genesis-account $(evmosd keys show validator -a --keyring-backend test --home ~/.tmp-evmosd) 100000000000000000000000000aevmos --keyring-backend test --home ~/.tmp-evmosd

# Create gentx
evmosd gentx validator 1000000000000000000000aevmos --chain-id evmos_9000-4 --keyring-backend test --home ~/.tmp-evmosd

# Collect gentx
evmosd collect-gentxs --home ~/.tmp-evmosd

# Validate genesis
evmosd validate-genesis --home ~/.tmp-evmosd

# Start the node (in a separate terminal or background)
evmosd start --keyring-backend test --chain-id evmos_9000-4 --home ~/.tmp-evmosd
```

**Note:** 
- Keep the chain running in one terminal
- Use another terminal for commands below
- If validator has no funds, the script will try to use dev0, dev1, dev2, or dev3 accounts

---

## Step 1: Create Wallets

```bash
# Create Alice's wallet
evmosd keys add alice --keyring-backend test

# Create Bob's wallet  
evmosd keys add bob --keyring-backend test

# View addresses
ALICE=$(evmosd keys show alice -a --keyring-backend test)
BOB=$(evmosd keys show bob -a --keyring-backend test)

echo "Alice: $ALICE"
echo "Bob: $BOB"
```

---

## Step 2: Fund Wallets

```bash
# Get addresses
ALICE=$(evmosd keys show alice -a --keyring-backend test)
BOB=$(evmosd keys show bob -a --keyring-backend test)

# Get validator address (or use dev0, dev1, dev2, dev3 if validator doesn't exist)
VALIDATOR=$(evmosd keys show validator -a --keyring-backend test 2>/dev/null || echo "")
# OR use dev0 if validator doesn't exist:
# VALIDATOR=$(evmosd keys show dev0 -a --keyring-backend test)

# Find an account with funds (try dev0, dev1, dev2, dev3, or validator)
# Check which accounts exist and have funds
evmosd keys list --keyring-backend test

# IMPORTANT: The command requires 3 positional arguments:
# evmosd tx bank send [from_address] [to_address] [amount] --from [key_name]
# The --from flag specifies which key to use for signing, but you still need to provide the from_address

# Send funds to Alice
evmosd tx bank send $VALIDATOR $ALICE 1000000000000aevmos \
  --from validator \
  --chain-id evmos_9000-4 \
  --keyring-backend test \
  --fees 1000aevmos \
  -y

# OR if validator doesn't exist, use dev0, dev1, dev2, or dev3:
# DEV0=$(evmosd keys show dev0 -a --keyring-backend test)
# evmosd tx bank send $DEV0 $ALICE 1000000000000aevmos \
#   --from dev0 \
#   --chain-id evmos_9000-4 \
#   --keyring-backend test \
#   --fees 1000aevmos \
#   -y

# Send funds to Bob
evmosd tx bank send $VALIDATOR $BOB 1000000000000aevmos \
  --from validator \
  --chain-id evmos_9000-4 \
  --keyring-backend test \
  --fees 1000aevmos \
  -y

# Check balances
evmosd query bank balances $ALICE --chain-id evmos_9000-4
evmosd query bank balances $BOB --chain-id evmos_9000-4
```

**Common Error:** If you see "accepts 3 arg(s), received 2", the command requires 3 positional arguments:
- `[from_address]` - sender's account address
- `[to_address]` - recipient's account address  
- `[amount]` - amount to send

**Correct format:** `evmosd tx bank send [from_address] [to_address] [amount] --from [key_name]`

The `--from` flag specifies which key to use for signing, but you still need to provide the `from_address` as the first argument.

---

## Step 3: Store Verification Keys (One-Time Setup)

**Navigate to keys directory:**
```bash
cd /data/evmos/x/shielded/cmd/trusted-setup/keys
```

**Store verification keys:**
```bash
# Get validator address
VALIDATOR=$(evmosd keys show validator -a --keyring-backend test)

# Store deposit verification key
evmosd tx shielded set-verification-key \
  deposit 0 deposit_vk.bin \
  --from validator \
  --chain-id evmos_9000-4 \
  --keyring-backend test \
  --fees 1000aevmos \
  -y

# Store private send verification key (depth 10)
evmosd tx shielded set-verification-key \
  private_send 10 private_send_depth_10_vk.bin \
  --from validator \
  --chain-id evmos_9000-4 \
  --keyring-backend test \
  --fees 1000aevmos \
  -y

# Store withdrawal verification key
evmosd tx shielded set-verification-key \
  withdrawal 0 withdrawal_vk.bin \
  --from validator \
  --chain-id evmos_9000-4 \
  --keyring-backend test \
  --fees 1000aevmos \
  -y
```

---

## Step 4: Generate Deposit Proof (Off-Chain)

**Note:** This requires a Go client. For now, you'll need to implement this using the `gnark` library. See `COMPLETE_GUIDE.md` for implementation details.

```bash
# Example (you need to implement this):
# go run generate_deposit_proof.go
# 
# Expected output:
# Commitment: 0x1234abcd...
# Proof: 0x5678ef01...
```

**For testing, you can use placeholder values:**
```bash
COMMITMENT="0x1234abcd5678ef0123456789abcdef0123456789abcdef0123456789abcdef01"
PROOF="0x5678ef0123456789abcdef0123456789abcdef0123456789abcdef0123456789ab"
```

---

## Step 5: Make Deposit

```bash
# Make deposit transaction
evmosd tx shielded deposit \
  1 \
  300 \
  $COMMITMENT \
  $PROOF \
  --from alice \
  --chain-id evmos_9000-4 \
  --keyring-backend test \
  --fees 1000aevmos \
  --gas auto \
  -y

# Check transaction status
evmosd query tx <tx-hash> --chain-id evmos_9000-4
```

---

## Step 6: Generate Private Send Proof (Off-Chain)

**Note:** This also requires a Go client implementation.

```bash
# Example (you need to implement this):
# go run generate_private_send_proof.go
#
# Expected output:
# Nullifier: 0xabcd...
# Sender Identity: 0xef01...
# Recipient Commitment: 0x2345...
# Change Commitment: 0x6789...
# Merkle Root: 0xdef0...
# Proof: 0x1111...
```

**For testing, use placeholder values:**
```bash
NULLIFIER="0xabcd1234ef5678901234567890abcdef1234567890abcdef1234567890abcdef"
SENDER_IDENTITY="0xef0123456789abcdef0123456789abcdef0123456789abcdef0123456789ab"
RECIPIENT_COMMITMENT="0x23456789abcdef0123456789abcdef0123456789abcdef0123456789abcd"
CHANGE_COMMITMENT="0x6789abcdef0123456789abcdef0123456789abcdef0123456789abcdef01"
MERKLE_ROOT="0xdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789ab"
PROOF_PRIVATE_SEND="0x1111222233334444555566667777888899990000aaaabbbbccccddddeeeeffff"
```

---

## Step 7: Make Private Send (Amount Hidden)

```bash
evmosd tx shielded private-send \
  1 \
  $NULLIFIER \
  $SENDER_IDENTITY \
  $RECIPIENT_COMMITMENT \
  $CHANGE_COMMITMENT \
  $MERKLE_ROOT \
  $PROOF_PRIVATE_SEND \
  10 \
  --from alice \
  --chain-id evmos_9000-4 \
  --keyring-backend test \
  --fees 1000aevmos \
  --gas auto \
  -y
```

---

## Step 8: Generate Withdrawal Proof (Off-Chain)

**Note:** This also requires a Go client implementation.

```bash
# Example (you need to implement this):
# go run generate_withdrawal_proof.go
#
# Expected output:
# Commitment: 0xaaaa...
# Recipient Identity: 0xbbbb...
# Withdrawal Amount: 300
# Nullifier: 0xcccc...
# Salt: 0xdddd...
# Proof: 0xeeee...
```

**For testing, use placeholder values:**
```bash
WITHDRAWAL_COMMITMENT="0xaaaabbbbccccddddeeeeffff000011112222333344445555666677778888"
RECIPIENT_IDENTITY="0xbbbbccccddddeeeeffff0000111122223333444455556666777788889999"
WITHDRAWAL_NULLIFIER="0xccccddddeeeeffff0000111122223333444455556666777788889999aaaa"
SALT="0xddddeeeeffff0000111122223333444455556666777788889999aaaabbbbcccc"
PROOF_WITHDRAWAL="0xeeeeffff0000111122223333444455556666777788889999aaaabbbbccccdddd"
```

---

## Step 9: Make Withdrawal

```bash
evmosd tx shielded withdraw \
  1 \
  $WITHDRAWAL_COMMITMENT \
  $BOB \
  300 \
  $SALT \
  $RECIPIENT_IDENTITY \
  $PROOF_WITHDRAWAL \
  $WITHDRAWAL_NULLIFIER \
  --from alice \
  --chain-id evmos_9000-4 \
  --keyring-backend test \
  --fees 1000aevmos \
  --gas auto \
  -y

# Check Bob's balance
evmosd query bank balances $BOB --chain-id evmos_9000-4
```

---

## Complete Automated Script

A complete script with logging and error handling is available:

```bash
# Run the complete test script
cd /data/evmos/x/shielded
./run_complete_test.sh
```

This script will:
- Check prerequisites
- Verify chain is running
- Create wallets (Alice and Bob)
- Fund wallets from validator
- Store verification keys
- Attempt deposit transaction (with placeholder proof)
- Provide detailed logging throughout

**Log file:** The script creates `shielded_test.log` with all output.

**Manual Commands (if you prefer step-by-step):**

```bash
#!/bin/bash
set -e

# Step 0: Start chain (do this first in a separate terminal)
# ./local_node.sh
# OR manually start: evmosd start --keyring-backend test --chain-id evmos_9000-4

# Step 1: Create wallets
evmosd keys add alice --keyring-backend test
evmosd keys add bob --keyring-backend test
ALICE=$(evmosd keys show alice -a --keyring-backend test)
BOB=$(evmosd keys show bob -a --keyring-backend test)

# Step 2: Fund wallets
evmosd tx bank send $ALICE 1000000000000aevmos --from validator --chain-id evmos_9000-4 --keyring-backend test --fees 1000aevmos -y
evmosd tx bank send $BOB 1000000000000aevmos --from validator --chain-id evmos_9000-4 --keyring-backend test --fees 1000aevmos -y

# Step 3: Store verification keys
cd /data/evmos/x/shielded/cmd/trusted-setup/keys
evmosd tx shielded set-verification-key deposit 0 deposit_vk.bin --from validator --chain-id evmos_9000-4 --keyring-backend test --fees 1000aevmos -y
evmosd tx shielded set-verification-key private_send 10 private_send_depth_10_vk.bin --from validator --chain-id evmos_9000-4 --keyring-backend test --fees 1000aevmos -y
evmosd tx shielded set-verification-key withdrawal 0 withdrawal_vk.bin --from validator --chain-id evmos_9000-4 --keyring-backend test --fees 1000aevmos -y

# Step 4-5: Deposit (requires proof generation - use placeholders for now)
COMMITMENT="0x1234abcd5678ef0123456789abcdef0123456789abcdef0123456789abcdef01"
PROOF="0x5678ef0123456789abcdef0123456789abcdef0123456789abcdef0123456789ab"
evmosd tx shielded deposit 1 300 $COMMITMENT $PROOF --from alice --chain-id evmos_9000-4 --keyring-backend test --fees 1000aevmos --gas auto -y

echo "Done! Check balances:"
evmosd query bank balances $ALICE --chain-id evmos_9000-4
evmosd query bank balances $BOB --chain-id evmos_9000-4
```

---

## Important Notes

1. **Chain must be running** - Start it first with `./local_node.sh` or `evmosd start`
2. **Use `--keyring-backend test`** - Required for local testnet
3. **Use `--chain-id evmos_9000-4`** - Match your chain ID
4. **Proof generation** - Requires Go client implementation (see `COMPLETE_GUIDE.md`)
5. **`--from alice`** - Uses wallet key name, NOT a hash!

---

## Troubleshooting

```bash
# Check if chain is running
curl http://localhost:26657/status

# Check if evmosd is built
which evmosd

# Rebuild if needed
cd /data/evmos && make install

# Check if shielded module is registered
evmosd tx shielded --help

# Check wallet keys
evmosd keys list --keyring-backend test

# Check balances
evmosd query bank balances $ALICE --chain-id evmos_9000-4
```
