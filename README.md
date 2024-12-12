# CosmoScope

CosmoScope is a command-line portfolio tracker that aggregates balances across multiple blockchain networks, including Cosmos ecosystem and EVM chains. It automatically fetches network configurations and IBC assets from the Cosmos Chain Registry.

## Features

- Multi-chain portfolio tracking
  - Cosmos ecosystem networks (auto-configured from Chain Registry)
  - EVM networks (Ethereum, Polygon, etc.)
- Balance types supported:
  - Wallet balances
  - Staked assets
  - Unclaimed rewards
  - Fixed balances (Exchange/Cold storage)
- Automatic IBC token resolution using Chain Registry
- Spam token filtering
- Real-time USD value calculation
- Detailed and summary views
- Account grouping

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/cosmoscope.git
cd cosmoscope

# Install development dependencies
make dev-deps

# Copy and configure settings
cp configs/config_example.json configs/config.json

# Edit the configuration file with your details
vim configs/config.json

# Build the project
make build

# Run tests
make test

# Run the application
make run
```

## Development

### Prerequisites

- Go 1.21 or later
- Make
- golangci-lint (installed via make dev-deps)

### Available Make Commands

```bash
make build           # Build the binary
make test           # Run tests
make lint           # Run linter
make coverage       # Generate coverage report
make clean          # Clean build artifacts
make dev-deps       # Install development dependencies
make deps-update    # Update dependencies
make check-tools    # Check tool versions
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# View coverage report in browser
go tool cover -html=coverage.out
```

## Configuration

1. Copy the example configuration:
```bash
cp configs/config_example.json configs/config.json
```

2. Update configs/config.json with your details:
   - Configure your addresses
   - Add your Moralis API key
   - Set up fixed balances

Example configuration:
```json
{
    "cosmos_addresses": ["cosmos1..."],
    "evm_networks": [
        {
            "name": "ethereum",
            "rpc": "https://mainnet.infura.io/v3/YOUR_KEY",
            "chain_id": 1,
            "native_token": {
                "symbol": "ETH",
                "name": "Ethereum",
                "decimals": 18
            }
        }
    ],
    "evm_addresses": ["0x..."],
    "moralis_api_key": "YOUR-MORALIS-API-KEY",
    "fixed_balances": [
        {
            "token": "BTC",
            "amount": 1,
            "label": "Cold Wallet"
        }
    ]
}
```

Note: Cosmos network configurations are now automatically fetched from the [Cosmos Chain Registry](https://github.com/cosmos/chain-registry).

## Required API Keys

1. Moralis API Key
   - Sign up at https://moralis.io/
   - Create an API key
   - Add to config.json

2. EVM RPC Endpoints
   - Alchemy: https://www.alchemy.com/
   - Infura: https://infura.io/
   - Or other RPC providers

## Sample Output

Running `cosmoscope` produces a detailed breakdown of your portfolio across different networks and asset types:

```
╔════════════════════════════════════════════════════════════╗
║                                                            ║
║              BALANCES REPORT - 2024-03-12 15:04:05         ║
║                                                            ║
╚════════════════════════════════════════════════════════════╝

Detailed Balance View:
+----------------------+-----------------+----------+---------------+--------------+
|       ACCOUNT        |     NETWORK     |  TOKEN   |    AMOUNT     |  USD VALUE  |
+----------------------+-----------------+----------+---------------+--------------+
| cosmos1abc...def     | cosmos-bank     | ATOM     |      520.4530 |  $5,204.53  |
| cosmos1abc...def     | cosmos-staking  | ATOM     |     2500.0000 | $25,000.00  |
| cosmos1abc...def     | cosmos-rewards  | ATOM     |        3.4520 |     $34.52  |
| osmo1xyz...789       | osmosis-bank    | OSMO     |     1200.0000 |  $2,400.00  |
| osmo1xyz...789       | osmosis-bank    | ATOM     |      150.0000 |  $1,500.00  |
| osmo1xyz...789       | osmosis-staking | OSMO     |     5000.0000 | $10,000.00  |
| stars1pqr...456      | stargaze-bank   | STARS    |    15000.0000 |  $1,500.00  |
| evmos1mno...123      | evmos-staking   | EVMOS    |     1000.0000 |  $2,500.00  |
| 0x123...789          | ethereum        | ETH      |        1.5000 |  $4,500.00  |
| 0x123...789          | polygon         | MATIC    |    10000.0000 | $10,000.00  |
| Cold Storage         | Fixed Balance   | BTC      |        0.7500 | $30,000.00  |
| Exchange             | Fixed Balance   | ETH      |        2.0000 |  $6,000.00  |
+----------------------+-----------------+----------+---------------+--------------+

Portfolio Summary:
+----------+---------------+--------------+----------+
|  TOKEN   |    AMOUNT     |  USD VALUE   | SHARE %  |
+----------+---------------+--------------+----------+
| ATOM     |    3173.9050  | $31,739.05  |   31.74% |
| OSMO     |    6200.0000  | $12,400.00  |   12.40% |
| ETH      |       3.5000  | $10,500.00  |   10.50% |
| MATIC    |   10000.0000  | $10,000.00  |   10.00% |
| BTC      |       0.7500  | $30,000.00  |   30.00% |
| EVMOS    |    1000.0000  |  $2,500.00  |    2.50% |
| STARS    |   15000.0000  |  $1,500.00  |    1.50% |
+----------+---------------+--------------+----------+
Total Portfolio Value: $98,639.05

Network Distribution:
+-------------------+--------------+----------+
|      NETWORK      |  USD VALUE   | SHARE %  |
+-------------------+--------------+----------+
| Cosmos Hub        | $31,739.05   |   32.18% |
| Osmosis          | $12,400.00   |   12.57% |
| Ethereum         | $10,500.00   |   10.64% |
| Polygon          | $10,000.00   |   10.14% |
| Fixed Balance    | $36,000.00   |   36.50% |
| Evmos            |  $2,500.00   |    2.53% |
| Stargaze         |  $1,500.00   |    1.52% |
+-------------------+--------------+----------+

Asset Types:
+------------+--------------+----------+
|    TYPE    |  USD VALUE   | SHARE %  |
+------------+--------------+----------+
| Bank       | $15,104.53   |   15.31% |
| Staking    | $47,500.00   |   48.15% |
| Rewards    |     $34.52   |    0.04% |
| Fixed      | $36,000.00   |   36.50% |
+------------+--------------+----------+
```

## Features & Roadmap

### Current Features ✅
- **Cosmos Ecosystem**
  - Auto-configuration using Chain Registry
  - Bank, staking, and reward balances
  - IBC token resolution
- **EVM Networks**
  - Ethereum & compatible chains
  - Native token balances
  - Custom RPC support
- **Portfolio Analytics**
  - Real-time USD values
  - Network distribution
  - Asset type breakdown

### Coming Soon 🚧
- **Solana Integration**
  - Native SOL & SPL tokens
  - Stake accounts
  - Program-owned accounts
- **Enhancements**
  - NFT tracking
  - DeFi positions

### Future Plans 📋
- Additional L1 blockchains
- CSV export
- Custom grouping
- Database support for snapshots
- Historical snapshots


## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests and linting (`make test lint`)
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request