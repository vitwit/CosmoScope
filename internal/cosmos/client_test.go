package cosmos

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolveSymbolForDenom(t *testing.T) {
	// Create a mock HTTP server for assetlist.json
	assetList := AssetList{
		Assets: []Asset{
			{
				Base:    "uatom",
				Display: "atom",
				Symbol:  "ATOM",
				DenomUnits: []DenomUnit{
					{Denom: "uatom", Exponent: 0},
					{Denom: "atom", Exponent: 6},
				},
			},
			{
				Base:    "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
				Display: "osmo",
				Symbol:  "OSMO",
				DenomUnits: []DenomUnit{
					{Denom: "uosmo", Exponent: 0},
					{Denom: "osmo", Exponent: 6},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(assetList)
	}))
	defer server.Close()

	// Override the registry base URL for testing
	originalURL := registryBaseURL
	registryBaseURL = server.URL
	defer func() { registryBaseURL = originalURL }()

	tests := []struct {
		name         string
		denom        string
		wantSymbol   string
		wantDecimals int
	}{
		{
			name:         "native token",
			denom:        "uatom",
			wantSymbol:   "ATOM",
			wantDecimals: 6,
		},
		{
			name:         "ibc token",
			denom:        "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
			wantSymbol:   "OSMO",
			wantDecimals: 6,
		},
		{
			name:         "unknown token",
			denom:        "unknown",
			wantSymbol:   "unknown",
			wantDecimals: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbol, decimals := resolveSymbolForDenom("cosmoshub", tt.denom)
			if symbol != tt.wantSymbol {
				t.Errorf("resolveSymbolForDenom() symbol = %v, want %v", symbol, tt.wantSymbol)
			}
			if decimals != tt.wantDecimals {
				t.Errorf("resolveSymbolForDenom() decimals = %v, want %v", decimals, tt.wantDecimals)
			}
		})
	}
}

func TestGetActiveEndpoint(t *testing.T) {
	// Create multiple test servers
	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"node_info": map[string]interface{}{
				"network": "test-chain",
			},
		})
	}))
	defer goodServer.Close()

	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer badServer.Close()

	tests := []struct {
		name      string
		endpoints []RestEndpoint
		want      string
	}{
		{
			name: "first endpoint good",
			endpoints: []RestEndpoint{
				{Address: goodServer.URL},
				{Address: badServer.URL},
			},
			want: goodServer.URL,
		},
		{
			name: "second endpoint good",
			endpoints: []RestEndpoint{
				{Address: badServer.URL},
				{Address: goodServer.URL},
			},
			want: goodServer.URL,
		},
		{
			name: "no good endpoints",
			endpoints: []RestEndpoint{
				{Address: badServer.URL},
				{Address: badServer.URL},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getActiveEndpoint(tt.endpoints)
			if got != tt.want {
				t.Errorf("getActiveEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}
