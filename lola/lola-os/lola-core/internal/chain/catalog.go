package chain

import "strings"

// Info is the built-in metadata LOLA OS knows about a blockchain: enough to
// dial it and interpret balances without the user having to look up a chain
// ID, native-asset symbol, or a public RPC endpoint. Users can always
// override any of this in config.yaml or via LOLA_* environment variables —
// the catalog only provides sensible, ready-to-run defaults.
type Info struct {
	Name           string // canonical lowercase key, e.g. "arbitrum"
	Display        string // human label, e.g. "Arbitrum One"
	Kind           string // "evm" or "solana"
	ChainID        int64  // EVM chain id; 0 for non-EVM
	NativeSymbol   string // e.g. "ETH", "MATIC", "SOL"
	NativeDecimals int    // 18 for EVM, 9 for Solana
	DefaultRPC     string // a public RPC endpoint that works out of the box
	ExplorerAPI    string // Etherscan-compatible getabi endpoint base, if any
	Testnet        bool
}

// Catalog is LOLA OS's built-in registry of well-known chains. It is
// deliberately broad so the very first release can serve almost anyone
// without custom configuration: every major EVM L1/L2, the common testnets,
// and Solana. Enable any of these by name in config.yaml (RPC/chain-id/symbol
// are filled in from here automatically) or point them at your own RPC.
//
// Public RPCs are best-effort defaults for getting started; for production
// throughput, set a dedicated endpoint (Alchemy, Infura, QuickNode, your own
// node) via `rpc_url` in config.yaml or `LOLA_RPC_URL_<CHAIN>` in the
// environment.
var Catalog = buildCatalog(
	// --- EVM mainnets -----------------------------------------------------
	Info{Name: "ethereum", Display: "Ethereum", Kind: "evm", ChainID: 1, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://ethereum-rpc.publicnode.com", ExplorerAPI: "https://api.etherscan.io/api"},
	Info{Name: "polygon", Display: "Polygon PoS", Kind: "evm", ChainID: 137, NativeSymbol: "POL", NativeDecimals: 18, DefaultRPC: "https://polygon-bor-rpc.publicnode.com", ExplorerAPI: "https://api.polygonscan.com/api"},
	Info{Name: "arbitrum", Display: "Arbitrum One", Kind: "evm", ChainID: 42161, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://arb1.arbitrum.io/rpc", ExplorerAPI: "https://api.arbiscan.io/api"},
	Info{Name: "optimism", Display: "OP Mainnet", Kind: "evm", ChainID: 10, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://optimism-rpc.publicnode.com", ExplorerAPI: "https://api-optimistic.etherscan.io/api"},
	Info{Name: "base", Display: "Base", Kind: "evm", ChainID: 8453, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://mainnet.base.org", ExplorerAPI: "https://api.basescan.org/api"},
	Info{Name: "bnb", Display: "BNB Smart Chain", Kind: "evm", ChainID: 56, NativeSymbol: "BNB", NativeDecimals: 18, DefaultRPC: "https://bsc-rpc.publicnode.com", ExplorerAPI: "https://api.bscscan.com/api"},
	Info{Name: "avalanche", Display: "Avalanche C-Chain", Kind: "evm", ChainID: 43114, NativeSymbol: "AVAX", NativeDecimals: 18, DefaultRPC: "https://api.avax.network/ext/bc/C/rpc", ExplorerAPI: "https://api.snowtrace.io/api"},
	Info{Name: "gnosis", Display: "Gnosis Chain", Kind: "evm", ChainID: 100, NativeSymbol: "XDAI", NativeDecimals: 18, DefaultRPC: "https://rpc.gnosischain.com", ExplorerAPI: "https://api.gnosisscan.io/api"},
	Info{Name: "fantom", Display: "Fantom Opera", Kind: "evm", ChainID: 250, NativeSymbol: "FTM", NativeDecimals: 18, DefaultRPC: "https://rpc.ftm.tools", ExplorerAPI: "https://api.ftmscan.com/api"},
	Info{Name: "celo", Display: "Celo", Kind: "evm", ChainID: 42220, NativeSymbol: "CELO", NativeDecimals: 18, DefaultRPC: "https://forno.celo.org", ExplorerAPI: "https://api.celoscan.io/api"},
	Info{Name: "linea", Display: "Linea", Kind: "evm", ChainID: 59144, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://rpc.linea.build", ExplorerAPI: "https://api.lineascan.build/api"},
	Info{Name: "scroll", Display: "Scroll", Kind: "evm", ChainID: 534352, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://rpc.scroll.io", ExplorerAPI: "https://api.scrollscan.com/api"},
	Info{Name: "zksync", Display: "zkSync Era", Kind: "evm", ChainID: 324, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://mainnet.era.zksync.io"},
	Info{Name: "polygon-zkevm", Display: "Polygon zkEVM", Kind: "evm", ChainID: 1101, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://zkevm-rpc.com", ExplorerAPI: "https://api-zkevm.polygonscan.com/api"},
	Info{Name: "mantle", Display: "Mantle", Kind: "evm", ChainID: 5000, NativeSymbol: "MNT", NativeDecimals: 18, DefaultRPC: "https://rpc.mantle.xyz", ExplorerAPI: "https://api.mantlescan.xyz/api"},
	Info{Name: "blast", Display: "Blast", Kind: "evm", ChainID: 81457, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://rpc.blast.io", ExplorerAPI: "https://api.blastscan.io/api"},
	Info{Name: "mode", Display: "Mode", Kind: "evm", ChainID: 34443, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://mainnet.mode.network"},
	Info{Name: "manta", Display: "Manta Pacific", Kind: "evm", ChainID: 169, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://pacific-rpc.manta.network/http"},
	Info{Name: "metis", Display: "Metis Andromeda", Kind: "evm", ChainID: 1088, NativeSymbol: "METIS", NativeDecimals: 18, DefaultRPC: "https://andromeda.metis.io/?owner=1088"},
	Info{Name: "moonbeam", Display: "Moonbeam", Kind: "evm", ChainID: 1284, NativeSymbol: "GLMR", NativeDecimals: 18, DefaultRPC: "https://rpc.api.moonbeam.network", ExplorerAPI: "https://api-moonbeam.moonscan.io/api"},
	Info{Name: "cronos", Display: "Cronos", Kind: "evm", ChainID: 25, NativeSymbol: "CRO", NativeDecimals: 18, DefaultRPC: "https://evm.cronos.org", ExplorerAPI: "https://api.cronoscan.com/api"},
	Info{Name: "aurora", Display: "Aurora", Kind: "evm", ChainID: 1313161554, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://mainnet.aurora.dev"},
	Info{Name: "opbnb", Display: "opBNB", Kind: "evm", ChainID: 204, NativeSymbol: "BNB", NativeDecimals: 18, DefaultRPC: "https://opbnb-mainnet-rpc.bnbchain.org"},
	Info{Name: "zora", Display: "Zora", Kind: "evm", ChainID: 7777777, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://rpc.zora.energy"},
	Info{Name: "fraxtal", Display: "Fraxtal", Kind: "evm", ChainID: 252, NativeSymbol: "frxETH", NativeDecimals: 18, DefaultRPC: "https://rpc.frax.com"},
	Info{Name: "sei", Display: "Sei (EVM)", Kind: "evm", ChainID: 1329, NativeSymbol: "SEI", NativeDecimals: 18, DefaultRPC: "https://evm-rpc.sei-apis.com"},
	Info{Name: "kava", Display: "Kava EVM", Kind: "evm", ChainID: 2222, NativeSymbol: "KAVA", NativeDecimals: 18, DefaultRPC: "https://evm.kava.io"},
	Info{Name: "rootstock", Display: "Rootstock", Kind: "evm", ChainID: 30, NativeSymbol: "RBTC", NativeDecimals: 18, DefaultRPC: "https://public-node.rsk.co"},
	Info{Name: "boba", Display: "Boba Network", Kind: "evm", ChainID: 288, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://mainnet.boba.network"},

	// --- EVM testnets -----------------------------------------------------
	Info{Name: "sepolia", Display: "Ethereum Sepolia", Kind: "evm", ChainID: 11155111, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://ethereum-sepolia-rpc.publicnode.com", ExplorerAPI: "https://api-sepolia.etherscan.io/api", Testnet: true},
	Info{Name: "holesky", Display: "Ethereum Holesky", Kind: "evm", ChainID: 17000, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://ethereum-holesky-rpc.publicnode.com", Testnet: true},
	Info{Name: "base-sepolia", Display: "Base Sepolia", Kind: "evm", ChainID: 84532, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://sepolia.base.org", ExplorerAPI: "https://api-sepolia.basescan.org/api", Testnet: true},
	Info{Name: "arbitrum-sepolia", Display: "Arbitrum Sepolia", Kind: "evm", ChainID: 421614, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://sepolia-rollup.arbitrum.io/rpc", Testnet: true},
	Info{Name: "optimism-sepolia", Display: "OP Sepolia", Kind: "evm", ChainID: 11155420, NativeSymbol: "ETH", NativeDecimals: 18, DefaultRPC: "https://optimism-sepolia-rpc.publicnode.com", Testnet: true},
	Info{Name: "polygon-amoy", Display: "Polygon Amoy", Kind: "evm", ChainID: 80002, NativeSymbol: "POL", NativeDecimals: 18, DefaultRPC: "https://rpc-amoy.polygon.technology", Testnet: true},
	Info{Name: "bnb-testnet", Display: "BNB Smart Chain Testnet", Kind: "evm", ChainID: 97, NativeSymbol: "tBNB", NativeDecimals: 18, DefaultRPC: "https://bsc-testnet-rpc.publicnode.com", Testnet: true},
	Info{Name: "avalanche-fuji", Display: "Avalanche Fuji", Kind: "evm", ChainID: 43113, NativeSymbol: "AVAX", NativeDecimals: 18, DefaultRPC: "https://api.avax-test.network/ext/bc/C/rpc", Testnet: true},

	// --- Solana -----------------------------------------------------------
	Info{Name: "solana", Display: "Solana", Kind: "solana", NativeSymbol: "SOL", NativeDecimals: 9, DefaultRPC: "https://api.mainnet-beta.solana.com"},
	Info{Name: "solana-devnet", Display: "Solana Devnet", Kind: "solana", NativeSymbol: "SOL", NativeDecimals: 9, DefaultRPC: "https://api.devnet.solana.com", Testnet: true},
	Info{Name: "solana-testnet", Display: "Solana Testnet", Kind: "solana", NativeSymbol: "SOL", NativeDecimals: 9, DefaultRPC: "https://api.testnet.solana.com", Testnet: true},
)

// DefaultEnabled lists the chains LOLA OS turns on out of the box. It is a
// broad-but-fast subset of the catalog (the most-used mainnets plus the
// common Ethereum/Base/Solana testnets); every other catalog chain is one
// line of config away. Dialing these is cheap — EVM/Solana clients connect
// lazily on first use, not at startup.
var DefaultEnabled = []string{
	"ethereum", "polygon", "arbitrum", "optimism", "base", "bnb", "avalanche",
	"sepolia", "base-sepolia",
	"solana", "solana-devnet",
}

func buildCatalog(chains ...Info) map[string]Info {
	m := make(map[string]Info, len(chains))
	for _, c := range chains {
		m[c.Name] = c
	}
	return m
}

// Lookup returns catalog metadata for a chain by name (case-insensitive),
// and whether it was found.
func Lookup(name string) (Info, bool) {
	info, ok := Catalog[strings.ToLower(name)]
	return info, ok
}

// CatalogNames returns every known chain name, unsorted. Useful for docs,
// `lola doctor`, and shell completion.
func CatalogNames() []string {
	names := make([]string, 0, len(Catalog))
	for name := range Catalog {
		names = append(names, name)
	}
	return names
}
