# CosmoScope

CosmoScope is a command-line portfolio tracker for cross-chain assets, supporting Cosmos ecosystem and EVM networks.

## Features

- Real-time portfolio tracking across multiple chains
- Support for:
  - Cosmos-based networks (with IBC token resolution)
  - EVM networks (ETH, Polygon, etc.)
- Balance types tracked:
  - Wallet balances
  - Staked assets
  - Unclaimed rewards
- USD value calculation via CoinGecko
- Spam token filtering for EVM chains
- Address grouping and custom labels

## Installation

```bash
go install github.com/yourusername/cosmoscope@latest
```

## Configuration

Create `config.json`:

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
      "rpc": "https://eth-mainnet.g.alchemy.com/v2/YOUR-API-KEY",
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
  "ibc_assets_file": "ibc_assets.json",
  "nicknames": [
    {
      "address": "0x...",
      "nickname": "Trading Wallet"
    }
  ]
}
```

## Usage

```bash
export MORALIS_API_KEY=your_key
cosmoscope
```

## Dependencies

- Go 1.21+
- Moralis API key for EVM token data
- RPC endpoints for EVM networks
- API endpoints for Cosmos networks

## License

[PRIVATE]