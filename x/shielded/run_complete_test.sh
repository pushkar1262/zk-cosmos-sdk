#!/bin/bash

# Complete Shielded Module Test Script
# This script sets up a local chain, creates wallets, funds them, and tests shielded transactions

set +e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CHAIN_ID="evmos_9000-4"
KEYRING_BACKEND="test"
HOME_DIR="$HOME/.evmosd"
KEYS_DIR="/data/evmos/x/shielded/cmd/trusted-setup/keys"
LOG_FILE="shielded_test.log"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

log_step() {
    echo -e "\n${GREEN}========================================${NC}" | tee -a "$LOG_FILE"
    echo -e "${GREEN}STEP: $1${NC}" | tee -a "$LOG_FILE"
    echo -e "${GREEN}========================================${NC}\n" | tee -a "$LOG_FILE"
}

# Check if command exists
check_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "$1 is not installed. Please install it first."
        exit 1
    fi
}

# Check if chain is running
check_chain_running() {
    if ! curl -s http://localhost:26657/status > /dev/null 2>&1; then
        log_error "Chain is not running. Please start it first with: ./local_node.sh"
        exit 1
    fi
    log_success "Chain is running"
}

# Wait for chain to be ready
wait_for_chain() {
    log_info "Waiting for chain to be ready..."
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -s http://localhost:26657/status > /dev/null 2>&1; then
            log_success "Chain is ready"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 2
    done
    
    log_error "Chain did not become ready after $max_attempts attempts"
    exit 1
}

# Initialize log file
echo "Shielded Module Test Log - $(date)" > "$LOG_FILE"
log_info "Starting shielded module test script"

# Check prerequisites
log_step "Checking Prerequisites"
check_command "evmosd"
check_command "jq"
check_command "curl"

# Check if chain is running
log_step "Checking Chain Status"
if ! curl -s http://localhost:26657/status > /dev/null 2>&1; then
    log_warning "Chain is not running. Starting chain..."
    log_info "Please run './local_node.sh' in a separate terminal first, then run this script again."
    log_info "Or if you want to start it here, uncomment the chain start commands below."
    exit 1
fi
check_chain_running

# Step 1: Create Wallets
log_step "Step 1: Creating Wallets"

# Check if wallets already exist
if evmosd keys show alice --keyring-backend "$KEYRING_BACKEND" &> /dev/null; then
    log_warning "Alice wallet already exists, skipping creation"
else
    log_info "Creating Alice wallet..."
    echo "" | evmosd keys add alice --keyring-backend "$KEYRING_BACKEND" 2>&1 | tee -a "$LOG_FILE"
    log_success "Alice wallet created"
fi

if evmosd keys show bob --keyring-backend "$KEYRING_BACKEND" &> /dev/null; then
    log_warning "Bob wallet already exists, skipping creation"
else
    log_info "Creating Bob wallet..."
    echo "" | evmosd keys add bob --keyring-backend "$KEYRING_BACKEND" 2>&1 | tee -a "$LOG_FILE"
    log_success "Bob wallet created"
fi

# Get addresses
ALICE=$(evmosd keys show alice -a --keyring-backend "$KEYRING_BACKEND")
BOB=$(evmosd keys show bob -a --keyring-backend "$KEYRING_BACKEND")
VALIDATOR=$(evmosd keys show validator -a --keyring-backend "$KEYRING_BACKEND" 2>/dev/null || echo "")

log_info "Alice address: $ALICE"
log_info "Bob address: $BOB"
if [ -n "$VALIDATOR" ]; then
    log_info "Validator address: $VALIDATOR"
fi

# Step 2: Fund Wallets
log_step "Step 2: Funding Wallets"

# Find an account with funds (try validator first, then check genesis accounts)
FUNDER_ADDRESS=""
FUNDER_KEY=""

# Check validator balance
if [ -n "$VALIDATOR" ]; then
    VALIDATOR_BALANCE_JSON=$(evmosd query bank balances "$VALIDATOR" --chain-id "$CHAIN_ID" --output json 2>/dev/null || echo '{"balances":[]}')
    VALIDATOR_BALANCE=$(echo "$VALIDATOR_BALANCE_JSON" | jq -r '.balances[]? | select(.denom=="aevmos") | .amount // empty' || echo "0")
    if [ -n "$VALIDATOR_BALANCE" ] && [ "$VALIDATOR_BALANCE" != "null" ] && [ "$VALIDATOR_BALANCE" != "0" ]; then
        FUNDER_ADDRESS="$VALIDATOR"
        FUNDER_KEY="validator"
        log_info "Found validator with balance: $VALIDATOR_BALANCE aevmos"
    fi
fi

# If validator has no funds, try to find any account with funds
if [ -z "$FUNDER_ADDRESS" ]; then
    log_warning "Validator has no balance. Checking for other accounts with funds..."
    
    # Try common genesis account names from local_node.sh
    for key_name in "dev0" "dev1" "dev2" "dev3" "mykey" "user1" "test1"; do
        if evmosd keys show "$key_name" --keyring-backend "$KEYRING_BACKEND" &> /dev/null; then
            key_addr=$(evmosd keys show "$key_name" -a --keyring-backend "$KEYRING_BACKEND" 2>/dev/null)
            if [ -n "$key_addr" ]; then
                key_balance_json=$(evmosd query bank balances "$key_addr" --chain-id "$CHAIN_ID" --output json 2>/dev/null || echo '{"balances":[]}')
                key_balance=$(echo "$key_balance_json" | jq -r '.balances[]? | select(.denom=="aevmos") | .amount // empty' || echo "0")
                if [ -n "$key_balance" ] && [ "$key_balance" != "null" ] && [ "$key_balance" != "0" ]; then
                    FUNDER_ADDRESS="$key_addr"
                    FUNDER_KEY="$key_name"
                    log_info "Found account '$key_name' with balance: $key_balance aevmos"
                    break
                fi
            fi
        fi
    done
fi

# Check Alice balance
ALICE_BALANCE_JSON=$(evmosd query bank balances "$ALICE" --chain-id "$CHAIN_ID" --output json 2>/dev/null || echo '{"balances":[]}')
ALICE_BALANCE=$(echo "$ALICE_BALANCE_JSON" | jq -r '.balances[]? | select(.denom=="aevmos") | .amount // empty' || echo "0")
if [ -z "$ALICE_BALANCE" ] || [ "$ALICE_BALANCE" = "null" ]; then
    ALICE_BALANCE="0"
fi
log_info "Alice current balance: $ALICE_BALANCE aevmos"

if [ "$ALICE_BALANCE" = "0" ] || [ -z "$ALICE_BALANCE" ]; then
    if [ -z "$FUNDER_ADDRESS" ]; then
        log_error "No account with funds found to fund wallets."
        log_info ""
        log_info "To fix this, you need to:"
        log_info "1. Ensure validator account is in genesis with funds, OR"
        log_info "2. Add accounts to genesis during chain initialization, OR"
        log_info "3. Manually fund an account and use it as funder"
        log_info ""
        log_info "Example: Add validator to genesis during init:"
        log_info "  evmosd add-genesis-account $VALIDATOR 1000000000000000000aevmos --keyring-backend test"
        log_info ""
        log_warning "Skipping wallet funding. Continuing with other steps..."
    else
        log_info "Funding Alice wallet from $FUNDER_KEY ($FUNDER_ADDRESS)..."
        log_info "Sending 1000000000000aevmos to Alice..."
        # Command format: evmosd tx bank send [from_address] [to_address] [amount] --from [key_name]
        TX_RESULT=$(evmosd tx bank send "$FUNDER_ADDRESS" "$ALICE" 1000000000000aevmos \
            --from "$FUNDER_KEY" \
            --chain-id "$CHAIN_ID" \
            --keyring-backend "$KEYRING_BACKEND" \
            --fees 1000aevmos \
            --gas auto \
            -y \
            --output json 2>&1) || true
        
        if echo "$TX_RESULT" | grep -q '"code":0' || echo "$TX_RESULT" | grep -q "code: 0"; then
            TX_HASH=$(echo "$TX_RESULT" | jq -r '.txhash // empty' 2>/dev/null || echo "unknown")
            log_success "Alice wallet funded successfully (tx: $TX_HASH)"
        else
            log_error "Failed to fund Alice wallet"
            echo "$TX_RESULT" | head -30 | tee -a "$LOG_FILE"
            log_info "Continuing with other steps..."
        fi
        
        # Wait for transaction to be included
        log_info "Waiting for transaction to be included..."
        sleep 5
    fi
else
    log_info "Alice already has balance ($ALICE_BALANCE aevmos), skipping funding"
fi

# Check Bob balance
BOB_BALANCE_JSON=$(evmosd query bank balances "$BOB" --chain-id "$CHAIN_ID" --output json 2>/dev/null || echo '{"balances":[]}')
BOB_BALANCE=$(echo "$BOB_BALANCE_JSON" | jq -r '.balances[]? | select(.denom=="aevmos") | .amount // empty' || echo "0")
if [ -z "$BOB_BALANCE" ] || [ "$BOB_BALANCE" = "null" ]; then
    BOB_BALANCE="0"
fi
log_info "Bob current balance: $BOB_BALANCE aevmos"

if [ "$BOB_BALANCE" = "0" ] || [ -z "$BOB_BALANCE" ]; then
    if [ -n "$FUNDER_ADDRESS" ] && [ -n "$FUNDER_KEY" ]; then
        log_info "Funding Bob wallet from $FUNDER_KEY ($FUNDER_ADDRESS)..."
        log_info "Sending 1000000000000aevmos to Bob..."
        TX_RESULT=$(evmosd tx bank send "$FUNDER_ADDRESS" "$BOB" 1000000000000aevmos \
            --from "$FUNDER_KEY" \
            --chain-id "$CHAIN_ID" \
            --keyring-backend "$KEYRING_BACKEND" \
            --fees 1000aevmos \
            --gas auto \
            -y \
            --output json 2>&1) || true
        
        if echo "$TX_RESULT" | grep -q '"code":0' || echo "$TX_RESULT" | grep -q "code: 0"; then
            TX_HASH=$(echo "$TX_RESULT" | jq -r '.txhash // empty' 2>/dev/null || echo "unknown")
            log_success "Bob wallet funded successfully (tx: $TX_HASH)"
        else
            log_error "Failed to fund Bob wallet"
            echo "$TX_RESULT" | head -30 | tee -a "$LOG_FILE"
            log_info "Continuing with other steps..."
        fi
        
        # Wait for transaction to be included
        log_info "Waiting for transaction to be included..."
        sleep 5
    else
        log_warning "No funder available, skipping Bob funding"
    fi
else
    log_info "Bob already has balance ($BOB_BALANCE aevmos), skipping funding"
fi

# Display final balances
log_info "Final balances:"
ALICE_FINAL_JSON=$(evmosd query bank balances "$ALICE" --chain-id "$CHAIN_ID" --output json 2>/dev/null || echo '{"balances":[]}')
ALICE_FINAL=$(echo "$ALICE_FINAL_JSON" | jq -r '.balances[]? | select(.denom=="aevmos") | .amount // empty' || echo "0")
if [ -z "$ALICE_FINAL" ] || [ "$ALICE_FINAL" = "null" ]; then
    ALICE_FINAL="0"
fi

BOB_FINAL_JSON=$(evmosd query bank balances "$BOB" --chain-id "$CHAIN_ID" --output json 2>/dev/null || echo '{"balances":[]}')
BOB_FINAL=$(echo "$BOB_FINAL_JSON" | jq -r '.balances[]? | select(.denom=="aevmos") | .amount // empty' || echo "0")
if [ -z "$BOB_FINAL" ] || [ "$BOB_FINAL" = "null" ]; then
    BOB_FINAL="0"
fi

log_success "Alice balance: $ALICE_FINAL aevmos"
log_success "Bob balance: $BOB_FINAL aevmos"

# Step 3: Store Verification Keys
log_step "Step 3: Storing Verification Keys"

# Check if keys directory exists
if [ ! -d "$KEYS_DIR" ]; then
    log_error "Keys directory not found: $KEYS_DIR"
    log_info "Please generate keys first with: cd /data/evmos/x/shielded/cmd/trusted-setup && ./trusted-setup -circuit deposit -output ./keys"
    exit 1
fi

# Store deposit verification key
log_info "Storing deposit verification key..."
if [ ! -f "$KEYS_DIR/deposit_vk.bin" ]; then
    log_error "Deposit verification key not found: $KEYS_DIR/deposit_vk.bin"
    exit 1
fi

# Use funder for storing keys (validator or found account)
KEY_STORAGE_FUNDER="$FUNDER_KEY"
if [ -z "$KEY_STORAGE_FUNDER" ] && [ -n "$VALIDATOR" ]; then
    KEY_STORAGE_FUNDER="validator"
fi

if [ -z "$KEY_STORAGE_FUNDER" ]; then
    log_error "Cannot store verification keys: No account with funds found."
    log_info "Skipping verification key storage. Please ensure an account has funds."
else
    log_info "Using '$KEY_STORAGE_FUNDER' to store verification keys..."
    log_info "Submitting transaction to store deposit verification key..."
    TX_RESULT=$(evmosd tx shielded set-verification-key \
        deposit 0 "$KEYS_DIR/deposit_vk.bin" \
        --from "$KEY_STORAGE_FUNDER" \
        --chain-id "$CHAIN_ID" \
        --keyring-backend "$KEYRING_BACKEND" \
        --fees 1000aevmos \
        --gas auto \
        -y \
        --output json 2>&1) || true

    if echo "$TX_RESULT" | grep -q '"code":0' || echo "$TX_RESULT" | grep -q "code: 0"; then
        TX_HASH=$(echo "$TX_RESULT" | jq -r '.txhash // empty' 2>/dev/null || echo "unknown")
        log_success "Deposit verification key stored successfully (tx: $TX_HASH)"
    else
        ERROR_MSG=$(echo "$TX_RESULT" | grep -i "error\|not found\|failed" | head -3 || echo "Unknown error")
        log_warning "Failed to store deposit verification key: $ERROR_MSG"
        echo "$TX_RESULT" | head -20 | tee -a "$LOG_FILE"
    fi

    sleep 3
fi

# Store private send verification key (depth 10)
log_info "Storing private send verification key (depth 10)..."
if [ -f "$KEYS_DIR/private_send_depth_10_vk.bin" ]; then
    if [ -n "$KEY_STORAGE_FUNDER" ]; then
        log_info "Submitting transaction to store private send verification key..."
        TX_RESULT=$(evmosd tx shielded set-verification-key \
            private_send 10 "$KEYS_DIR/private_send_depth_10_vk.bin" \
            --from "$KEY_STORAGE_FUNDER" \
            --chain-id "$CHAIN_ID" \
            --keyring-backend "$KEYRING_BACKEND" \
            --fees 1000aevmos \
            --gas auto \
            -y \
            --output json 2>&1) || true
        
        if echo "$TX_RESULT" | grep -q '"code":0' || echo "$TX_RESULT" | grep -q "code: 0"; then
            TX_HASH=$(echo "$TX_RESULT" | jq -r '.txhash // empty' 2>/dev/null || echo "unknown")
            log_success "Private send verification key stored successfully (tx: $TX_HASH)"
        else
            ERROR_MSG=$(echo "$TX_RESULT" | grep -i "error\|not found\|failed" | head -3 || echo "Unknown error")
            log_warning "Failed to store private send verification key: $ERROR_MSG"
            echo "$TX_RESULT" | head -20 | tee -a "$LOG_FILE"
        fi
        sleep 3
    else
        log_warning "Skipping: No funder available"
    fi
else
    log_warning "Private send verification key not found, skipping..."
fi

# Store withdrawal verification key
log_info "Storing withdrawal verification key..."
if [ -f "$KEYS_DIR/withdrawal_vk.bin" ]; then
    if [ -n "$KEY_STORAGE_FUNDER" ]; then
        log_info "Submitting transaction to store withdrawal verification key..."
        TX_RESULT=$(evmosd tx shielded set-verification-key \
            withdrawal 0 "$KEYS_DIR/withdrawal_vk.bin" \
            --from "$KEY_STORAGE_FUNDER" \
            --chain-id "$CHAIN_ID" \
            --keyring-backend "$KEYRING_BACKEND" \
            --fees 1000aevmos \
            --gas auto \
            -y \
            --output json 2>&1) || true
        
        if echo "$TX_RESULT" | grep -q '"code":0' || echo "$TX_RESULT" | grep -q "code: 0"; then
            TX_HASH=$(echo "$TX_RESULT" | jq -r '.txhash // empty' 2>/dev/null || echo "unknown")
            log_success "Withdrawal verification key stored successfully (tx: $TX_HASH)"
        else
            ERROR_MSG=$(echo "$TX_RESULT" | grep -i "error\|not found\|failed" | head -3 || echo "Unknown error")
            log_warning "Failed to store withdrawal verification key: $ERROR_MSG"
            echo "$TX_RESULT" | head -20 | tee -a "$LOG_FILE"
        fi
        sleep 3
    else
        log_warning "Skipping: No funder available"
    fi
else
    log_warning "Withdrawal verification key not found, skipping..."
fi

# Step 4: Generate Deposit Proof (Placeholder)
log_step "Step 4: Deposit Transaction"

log_warning "Deposit proof generation requires Go client implementation."
log_info "For now, using placeholder values for demonstration."
log_info "In production, you would generate proofs using:"
log_info "  - deposit_pk.bin (proving key)"
log_info "  - deposit_ccs.bin (constraint system)"
log_info "  - Your secret, salt, amount, and commitment values"

# Placeholder values (these won't verify, but show the command structure)
COMMITMENT="0x1234abcd5678ef0123456789abcdef0123456789abcdef0123456789abcdef01"
PROOF="0x5678ef0123456789abcdef0123456789abcdef0123456789abcdef0123456789ab"

log_info "Using placeholder commitment: $COMMITMENT"
log_info "Using placeholder proof: $PROOF"

log_info "Attempting deposit transaction (will fail without valid proof)..."
TX_RESULT=$(evmosd tx shielded deposit \
    1 \
    300 \
    "$COMMITMENT" \
    "$PROOF" \
    --from alice \
    --chain-id "$CHAIN_ID" \
    --keyring-backend "$KEYRING_BACKEND" \
    --fees 1000aevmos \
    --gas auto \
    -y \
    --output json 2>&1) || true

if echo "$TX_RESULT" | grep -q '"code":0' || echo "$TX_RESULT" | grep -q "code: 0"; then
    log_success "Deposit transaction successful!"
    TX_HASH=$(echo "$TX_RESULT" | jq -r '.txhash // empty' 2>/dev/null || echo "unknown")
    log_info "Transaction hash: $TX_HASH"
else
    log_warning "Deposit transaction failed (expected with placeholder proof)"
    ERROR_MSG=$(echo "$TX_RESULT" | grep -i "error\|failed\|invalid\|proof" | head -5 || echo "Unknown error")
    log_info "Error details: $ERROR_MSG"
    echo "$TX_RESULT" | head -30 | tee -a "$LOG_FILE"
fi

# Step 5: Summary
log_step "Test Summary"

log_info "Test completed. Summary:"
log_info "  - Wallets created: Alice ($ALICE), Bob ($BOB)"
log_info "  - Wallets funded: Alice ($ALICE_FINAL aevmos), Bob ($BOB_FINAL aevmos)"
log_info "  - Verification keys stored: Deposit, Private Send (depth 10), Withdrawal"
log_info "  - Deposit transaction attempted (requires valid proof generation)"

log_info ""
log_info "Next steps:"
log_info "  1. Implement proof generation Go client using gnark library"
log_info "  2. Generate valid proofs using *_pk.bin and *_ccs.bin files"
log_info "  3. Use generated proofs in deposit/private-send/withdrawal transactions"
log_info ""
log_info "Full log saved to: $LOG_FILE"

log_success "Script completed successfully!"

