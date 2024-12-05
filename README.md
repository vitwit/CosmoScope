# CosmoScope

CosmoScope is a command-line portfolio tracker that aggregates balances across multiple blockchain networks, including Cosmos ecosystem and EVM chains.

## Features

- Multi-chain portfolio tracking
  - Cosmos ecosystem networks
  - EVM networks (Ethereum, Polygon, etc.)
- Balance types supported:
  - Wallet balances
  - Staked assets
  - Unclaimed rewards
  - Fixed balances (Exchange/Cold storage)
- IBC token resolution
- Spam token filtering
- Real-time USD value calculation
- Detailed and summary views
- Account grouping

## Installation

```bash
# Clone the repository
git clone https://github.com/anilcse/cosmoscope.git
cd cosmoscope

# Copy and configure settings
cp configs/config_example.json configs/config.json

# Edit the configuration file with your details
vim configs/config.json

# Build the project
go build ./cmd/cosmoscope

# Run
./cosmoscope
```

## Configuration

1. Copy the example configuration:
```bash
cp configs/config_example.json configs/config.json
```

2. Update configs/config.json with your details:
   - Add your network RPC endpoints
   - Configure your addresses
   - Add your Moralis API key
   - Set up fixed balances
   - Update IBC assets mapping

Example configuration:
```json
{
    "cosmos_networks": [
        {
            "name": "osmosis",
            "api": "https://api.osmosis.zone",
            "prefix": "osmo",
            "chain_id": "osmosis-1"
        }
    ],
    "evm_networks": [
        {
            "name": "ethereum",
            "rpc": "https://eth-mainnet.alchemyapi.io/v2/YOUR-API-KEY",
            "chain_id": 1,
            "native_token": {
                "symbol": "ETH",
                "name": "Ethereum",
                "decimals": 18
            }
        }
    ],
    "cosmos_addresses": ["osmo1..."],
    "evm_addresses": ["0x..."],
    "ibc_assets_file": "configs/ibc_assets.json",
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

## Required API Keys

1. Moralis API Key
   - Sign up at https://moralis.io/
   - Create an API key
   - Add to config.json

2. EVM RPC Endpoints
   - Alchemy: https://www.alchemy.com/
   - Infura: https://infura.io/
   - Or other RPC providers

## Usage

```bash
# Run with default config
./cosmoscope

# Example output:
*******************************************************************************
*                                                                             *
*                BALANCES REPORT   (2024-1-2 15:4:5)                         *
*                                                                             *
*******************************************************************************

Detailed Balance View:
+------------------+-----------+---------+----------+-----------+
|      ACCOUNT     |  NETWORK  |  TOKEN  |  AMOUNT  | USD VALUE |
+------------------+-----------+---------+----------+-----------+
| Cold Wallet      | Exchange  | BTC     |    1.000 | $42000.00 |
| osmo1...         | osmosis   | OSMO    |  100.000 |   $500.00 |
| 0x...            | ethereum  | ETH     |    1.500 |  $3000.00 |
+------------------+-----------+---------+----------+-----------+

Portfolio Summary:
+---------+----------+-----------+----------+
|  TOKEN  |  AMOUNT  | USD VALUE |  SHARE % |
+---------+----------+-----------+----------