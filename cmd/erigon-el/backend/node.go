package backend

import (
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/ledgerwatch/erigon/params"
	"github.com/ledgerwatch/erigon/params/networkname"
	erigoncli "github.com/ledgerwatch/erigon/turbo/cli"
	"github.com/ledgerwatch/log/v3"
	"github.com/urfave/cli/v2"
)

// ErigonNode represents a single node, that runs sync and p2p network.
// it also can export the private endpoint for RPC daemon, etc.
type ErigonNode struct {
	stack   *node.Node
	backend *Ethereum
}

// Serve runs the node and blocks the execution. It returns when the node is existed.
func (eri *ErigonNode) Serve() error {
	defer eri.stack.Close()

	eri.run()

	eri.stack.Wait()

	return nil
}

func (eri *ErigonNode) run() {
	node.StartNode(eri.stack)
	// we don't have accounts locally and we don't do mining
	// so these parts are ignored
	// see cmd/geth/daemon.go#startNode for full implementation
}

// Params contains optional parameters for creating a node.
// * GitCommit is a commit from which then node was built.
// * CustomBuckets is a `map[string]dbutils.TableCfgItem`, that contains bucket name and its properties.
//
// NB: You have to declare your custom buckets here to be able to use them in the app.
type Params struct {
	CustomBuckets kv.TableCfg
}

func NewNodConfigUrfave(ctx *cli.Context, logger log.Logger) *nodecfg.Config {
	// If we're running a known preset, log it for convenience.
	chain := ctx.String(utils.ChainFlag.Name)
	switch chain {
	case networkname.SepoliaChainName:
		logger.Info("Starting Erigon on Sepolia testnet...")
	case networkname.GoerliChainName:
		logger.Info("Starting Erigon on Görli testnet...")
	case networkname.DevChainName:
		logger.Info("Starting Erigon in ephemeral dev mode...")
	case networkname.MumbaiChainName:
		logger.Info("Starting Erigon on Mumbai testnet...")
	case networkname.BorMainnetChainName:
		logger.Info("Starting Erigon on Bor Mainnet...")
	case networkname.BorDevnetChainName:
		logger.Info("Starting Erigon on Bor Devnet...")
	case networkname.OptimismMainnetChainName:
		logger.Info("Starting Erigon on Optimism Mainnet...")
	case networkname.OptimismDevnetChainName:
		logger.Info("Starting Erigon on Optimism Devnet...")
	case networkname.OptimismGoerliChainName:
		logger.Info("Starting Erigon on Optimism Görli testnet...")
	case "", networkname.MainnetChainName:
		if !ctx.IsSet(utils.NetworkIdFlag.Name) {
			log.Info("Starting Erigon on Ethereum mainnet...")
		}
	default:
		logger.Info("Starting Erigon on", "devnet", chain)
	}

	nodeConfig := NewNodeConfig()
	utils.SetNodeConfig(ctx, nodeConfig, logger)
	erigoncli.ApplyFlagsForNodeConfig(ctx, nodeConfig, logger)
	return nodeConfig
}
func NewEthConfigUrfave(ctx *cli.Context, nodeConfig *nodecfg.Config, logger log.Logger) *ethconfig.Config {
	ethConfig := &ethconfig.Defaults
	utils.SetEthConfig(ctx, nodeConfig, ethConfig, logger)
	erigoncli.ApplyFlagsForEthConfig(ctx, ethConfig, logger)

	return ethConfig
}

// New creates a new `ErigonNode`.
// * ctx - `*cli.Context` from the main function. Necessary to be able to configure the node based on the command-line flags
// * sync - `stagedsync.StagedSync`, an instance of staged sync, setup just as needed.
// * optionalParams - additional parameters for running a node.
func NewNode(
	nodeConfig *nodecfg.Config,
	ethConfig *ethconfig.Config,
	logger log.Logger,
) (*ErigonNode, error) {
	//prepareBuckets(optionalParams.CustomBuckets)
	node, err := node.New(nodeConfig, logger)
	if err != nil {
		utils.Fatalf("Failed to create Erigon node: %v", err)
	}

	ethereum, err := NewBackend(node, ethConfig, logger)
	if err != nil {
		return nil, err
	}
	return &ErigonNode{stack: node, backend: ethereum}, nil
}

func NewNodeConfig() *nodecfg.Config {
	nodeConfig := nodecfg.DefaultConfig
	// see simiar changes in `cmd/geth/config.go#defaultNodeConfig`
	if commit := params.GitCommit; commit != "" {
		nodeConfig.Version = params.VersionWithCommit(commit)
	} else {
		nodeConfig.Version = params.Version
	}
	nodeConfig.IPCPath = "" // force-disable IPC endpoint
	nodeConfig.Name = "erigon"
	return &nodeConfig
}