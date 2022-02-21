// Copyright 2020 Coinbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configuration

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/coinbase/rosetta-sdk-go/storage/encoder"
	"github.com/coinbase/rosetta-sdk-go/types"
)

// Mode is the setting that determines if
// the implementation is "online" or "offline".
type Mode string

const (
	// Online is when the implementation is permitted
	// to make outbound connections.
	Online Mode = "ONLINE"

	// Offline is when the implementation is not permitted
	// to make outbound connections.
	Offline Mode = "OFFLINE"

	// Mainnet is the EUNO Mainnet.
	Mainnet string = "MAINNET"

	// Testnet is EUNO Testnet.
	Testnet string = "TESTNET"

	// mainnetConfigPath is the path of the EUNO
	// configuration file for mainnet.
	mainnetConfigPath = "/app/euno-mainnet.conf"

	// testnetConfigPath is the path of the eunopay
	// configuration file for testnet.
	testnetConfigPath = "/app/euno-testnet.conf"

	// Zstandard compression dictionaries
	transactionNamespace         = "transaction"
	testnetTransactionDictionary = "/app/testnet-transaction.zstd"
	mainnetTransactionDictionary = "/app/mainnet-transaction.zstd"

	mainnetRPCPort = 46461
	testnetRPCPort = 46463

	// min prune depth is 288:
	// https://github.com/eunopay/eunopay/blob/ad2952d17a2af419a04256b10b53c7377f826a27/src/validation.h#L84
	// pruneDepth = int64(10000) //nolint

	// min prune height (on mainnet):
	// https://github.com/eunopay/eunopay/blob/62d137ac3b701aae36c1aa3aa93a83fd6357fde6/src/chainparams.cpp#L102
	// minPruneHeight = int64(100000) //nolint

	// attempt to prune once an hour
	// pruneFrequency = 60 * time.Minute

	// DataDirectory is the default location for all
	// persistent data.
	DataDirectory = "/data"

	eunodPath   = "eunod"
	indexerPath = "indexer"

	// allFilePermissions specifies anyone can do anything
	// to the file.
	allFilePermissions = 0777

	// ModeEnv is the environment variable read
	// to determine mode.
	ModeEnv = "MODE"

	// NetworkEnv is the environment variable
	// read to determine network.
	NetworkEnv = "NETWORK"

	// PortEnv is the environment variable
	// read to determine the port for the Rosetta
	// implementation.
	PortEnv = "PORT"
)

type Configuration struct {
	Mode                   Mode
	Network                *types.NetworkIdentifier
	Params                 *chaincfg.Params
	Currency               *types.Currency
	GenesisBlockIdentifier *types.BlockIdentifier
	Port                   int
	RPCPort                int
	ConfigPath             string
	// Pruning                *PruningConfiguration
	IndexerPath string
	EunodPath   string
	Compressors []*encoder.CompressorEntry
}

func LoadConfiguration(baseDirectory string) (*Configuration, error) {
	config := &Configuration{}
	modeValue := Mode(os.Getenv(ModeEnv))

	switch modeValue {
	case Online:
		config.Mode = Online
		config.IndexerPath = path.Join(baseDirectory, indexerPath)
		if err := ensurePathExists(config.IndexerPath); err != nil {
			return nil, fmt.Errorf("%w: unable to create indexer path", err)
		}

		config.EunodPath = path.Join(baseDirectory, eunodPath)
		if err := ensurePathExists(config.EunodPath); err != nil {
			return nil, fmt.Errorf("%w: unable to create eunod path", err)
		}
	case Offline:
		config.Mode = Offline
	case "":
		return nil, errors.New("MODE must be populated")
	default:
		return nil, fmt.Errorf("%s is not a valid mode", modeValue)
	}

	networkValue := os.Getenv(NetworkEnv)

	switch networkValue {
	case Mainnet:
		config.Network = &types.NetworkIdentifier{
			Blockchain: eunopay.Blockchain,
			Network:    eunopay.MainnetNetwork,
		}
		config.GenesisBlockIdentifier = eunopay.MainnetGenesisBlockIdentifier
		config.Params = eunopay.MainnetParams
		config.Currency = eunopay.MainnetCurrency
		config.ConfigPath = mainnetConfigPath
		config.RPCPort = mainnetRPCPort
		config.Compressors = []*encoder.CompressorEntry{
			{
				Namespace:      transactionNamespace,
				DictionaryPath: mainnetTransactionDictionary,
			},
		}
	case Testnet:
		config.Network = &types.NetworkIdentifier{
			Blockchain: eunopay.Blockchain,
			Network:    eunopay.TestnetNetwork,
		}
		config.GenesisBlockIdentifier = eunopay.TestnetGenesisBlockIdentifier
		config.Params = eunopay.TestnetParams
		config.Currency = eunopay.TestnetCurrency
		config.ConfigPath = testnetConfigPath
		config.RPCPort = testnetRPCPort
		config.Compressors = []*encoder.CompressorEntry{
			{
				Namespace:      transactionNamespace,
				DictionaryPath: testnetTransactionDictionary,
			},
		}
	case "":
		return nil, errors.New("NETWORK must be populated")
	default:
		return nil, fmt.Errorf("%s is not a valid network", networkValue)
	}

	portValue := os.Getenv(PortEnv)
	if len(portValue) == 0 {
		return nil, errors.New("PORT must be populated")
	}

	port, err := strconv.Atoi(portValue)
	if err != nil || len(portValue) == 0 || port <= 0 {
		return nil, fmt.Errorf("%w: unable to parse port %s", err, portValue)
	}
	config.Port = port

	return config, nil
}

// ensurePathsExist directories along
// a path if they do not exist.
func ensurePathExists(path string) error {
	if err := os.MkdirAll(path, os.FileMode(allFilePermissions)); err != nil {
		return fmt.Errorf("%w: unable to create %s directory", err, path)
	}

	return nil
}
