```bash
make build
SIMD_BIN=./build/simd ./scripts/init-simapp.sh
./build/simd start

./build/simd keys list --keyring-backend test

# Check Alice
./build/simd q bank balances alice

# Check Bob
./build/simd q bank balances bob

# Send 1000 stake from Alice to Bob
./build/simd tx bank send alice cosmos1fx489al9pmhcju5sekkwsdg2kfym7ean74gqn8 1000stake --chain-id demo --keyring-backend test -y

# Check Alice
./build/simd q bank balances alice

# Check Bob
./build/simd q bank balances bob
```

# Generate and add deposit vk to genesis
```bash
./zk-cli setup
./build/simd genesis add-shielded-vk deposit_vk.bin --home ./test-home
```