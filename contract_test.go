package rpcx

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/meshplus/bitxhub-model/constant"

	"github.com/stretchr/testify/require"
)

func TestChainClient_DeployXVMContract(t *testing.T) {
	cli, err := Cli()
	require.Nil(t, err)

	contract, err := ioutil.ReadFile("./testdata/example.wasm")
	require.Nil(t, err)

	_, err = cli.DeployContract(contract, nil)
	require.Nil(t, err)
}

// TODO: Waiting for the PR from XLW
//func TestChainClient_InvokeXVMContract(t *testing.T) {
//	cli, err := Cli()
//	require.Nil(t, err)
//
//	contract, err := ioutil.ReadFile("./testdata/example.wasm")
//	require.Nil(t, err)
//
//	addr, err := cli.DeployContract(contract, nil)
//	require.Nil(t, err)
//
//	result, err := cli.InvokeXVMContract(addr, "a", nil, Int32(1), Int32(2))
//	require.Nil(t, err)
//	require.True(t, CheckReceipt(result))
//	require.Equal(t, "336", string(result.Ret))
//}

func TestChainClient_InvokeBVMContract(t *testing.T) {
	cli, err := Cli()
	require.Nil(t, err)

	result, err := cli.InvokeBVMContract(constant.StoreContractAddr.Address(), "Set", nil, String("a"), String("10"))
	require.Nil(t, err)
	require.Nil(t, result.Ret)
	res, err := cli.InvokeBVMContract(constant.StoreContractAddr.Address(), "Get", nil, String("a"))
	require.Nil(t, err)
	require.Equal(t, string(res.Ret), "10")
}

func TestSetInterchainSwap_InvokeBVMContract(t *testing.T) {
	cli, err := Cli2()
	require.Nil(t, err)
	result, err := cli.InvokeBVMContract(constant.EthHeaderMgrContractAddr.Address(), "SetInterchainSwapAddr",
		nil, String("0xc58c24Af11eEF57975b9d07429F202C3CCca5bBD"))
	require.Nil(t, err)
	fmt.Println(result)
}
func TestSetEscrow_InvokeBVMContract(t *testing.T) {
	cli, err := Cli2()
	require.Nil(t, err)
	result, err := cli.InvokeBVMContract(constant.EthHeaderMgrContractAddr.Address(), "SetEscrowAddr",
		nil, String("0xa670a4891490a732BF17B5f8022B51c989c5E5ab"))
	require.Nil(t, err)
	fmt.Println(result)
}

func TestGetEscrow_InvokeBVMContract(t *testing.T) {
	cli, err := Cli2()
	require.Nil(t, err)
	result, err := cli.InvokeBVMContract(constant.EthHeaderMgrContractAddr.Address(), "GetEscrowAddr",
		nil)
	require.Nil(t, err)
	fmt.Println(string(result.Ret))
}
