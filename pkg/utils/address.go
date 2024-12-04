package utils

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func ConvertCosmosAddress(address, toPrefix string) (string, error) {
	_, bz, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return "", err
	}

	converted, err := bech32.ConvertAndEncode(toPrefix, bz)
	if err != nil {
		return "", err
	}

	return converted, nil
}

func ShortenAddress(address string) string {
	if len(address) <= 12 {
		return address
	}
	return fmt.Sprintf("%s...%s", address[:6], address[len(address)-6:])
}
