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

## CI/CD

The project uses GitHub Actions for continuous integration and deployment:

- Automated testing and code coverage reporting
- Linting with golangci-lint
- Multi-platform builds (Linux, macOS, Windows)
- Automatic releases on tags
- Coverage reporting with Codecov

### Release Process

To create a new release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

This will trigger the CI pipeline to:
1. Run tests and coverage
2. Build binaries for all platforms
3. Create a GitHub release
4. Upload binaries to the release

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests and linting (`make test lint`)
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details