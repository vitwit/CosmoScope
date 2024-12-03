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
go build
./cosmoscope
```
### Sample output
```
...

Portfolio Summary:
+------------+-------------+-------------+---------+
| TOKEN NAME |   BALANCE   |  USD VALUE  | SHARE % |
+------------+-------------+-------------+---------+
| AKT        |    10002.97 | $48333.22   | 34.09%  |
| MATIC      |   140411.33 | $139087.56  | 24.88%  |
| ATOM       |    90000.95 | $900001.24  | 11.95%  |
| ETH        |      100.00 | $365092.06  | 6.30%   |
| PASG       | 10000000.00 | $120000.49  | 3.20%   |
| OSMO       |   200000.56 | $132000.00  | 2.47%   |
| STARS      |  5000000.05 | $35000.70   | 0.66%   |
| REGEN      |   100000.10 | $19000.35   | 0.30%   |
| TIA        |     9999.01 | $9203.85    | 0.17%   |
| JUNO       |    28000.97 | $8103.57    | 0.15%   |
| NTRN       |    12345.41 | $6991.06    | 0.13%   |
| USDC       |    635.6119 | $635.37     | 0.01%   |
| STRD       |    500.5300 | $267.17     | 0.00%   |
| ALT        |    100.0000 | $149.94     | 0.00%   |
| SOMM       |     4011.58 | $96.51      | 0.00%   |
| USDT       |      3.0000 | $3.00       | 0.00%   |
+------------+-------------+-------------+---------+
|   TOTAL    |               $2383206.73 | 100.00% |
+------------+-------------+-------------+---------+
```


## Dependencies

- Go 1.21+
- Moralis API key for EVM token data
- RPC endpoints for EVM networks
- API endpoints for Cosmos networks

## License

[PRIVATE]
