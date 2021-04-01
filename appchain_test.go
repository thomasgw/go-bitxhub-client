package rpcx

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/meshplus/bitxhub-model/constant"

	"github.com/meshplus/bitxhub-kit/crypto"
	"github.com/meshplus/bitxhub-kit/crypto/asym"
	"github.com/meshplus/bitxhub-model/pb"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	keyPassword = "bitxhub"
)

var AppChainID string

func TestAppChain_Register_Audit(t *testing.T) {
	privKey, err := asym.GenerateKeyPair(crypto.Secp256k1)
	require.Nil(t, err)
	var cfg = &config{
		nodesInfo: []*NodeInfo{
			{Addr: "localhost:60011"},
		},
		logger:     logrus.New(),
		privateKey: privKey,
	}
	cli, err := New(
		WithNodesInfo(cfg.nodesInfo...),
		WithLogger(cfg.logger),
		WithPrivateKey(cfg.privateKey),
	)
	require.Nil(t, err)
	pubKey, err := privKey.PublicKey().Address()
	var pubKeyStr string = hex.EncodeToString(pubKey.Bytes())
	args := []*pb.Arg{
		String(""),                 //validators
		String("rbft"),             //consensus_type
		String("hyperchain"),       //chain_type
		String("AppChain1"),        //name
		String("Appchain for tax"), //desc
		String("1.8"),              //version
		String(pubKeyStr),          //public key
	}
	res, err := cli.InvokeBVMContract(constant.AppchainMgrContractAddr.Address(), "Register", nil, args...)
	require.Nil(t, err)
	appChain := &Appchain{}
	err = json.Unmarshal(res.Ret, appChain)
	require.Nil(t, err)
	require.NotNil(t, appChain.ID)
	AppChainID = appChain.ID
}

// getAdminCli returns client with admin account.
func getAdminCli(t *testing.T, keyPath string) *ChainClient {
	// you should put your bitxhub/scripts/build/node1/key.json to testdata/key.json.
	k, err := asym.RestorePrivateKey(keyPath, keyPassword)
	require.Nil(t, err)
	var cfg = &config{
		nodesInfo: []*NodeInfo{
			{Addr: "localhost:60011"},
		},
		logger:     logrus.New(),
		privateKey: k,
	}
	cli, err := New(
		WithNodesInfo(cfg.nodesInfo...),
		WithLogger(cfg.logger),
		WithPrivateKey(cfg.privateKey),
	)
	return cli
}
