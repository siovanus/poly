package cross_chain_manager

import (
	"fmt"
	"github.com/ontio/multi-chain/native/service/cross_chain_manager/neo"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"

	"github.com/ontio/multi-chain/common/log"
	"github.com/ontio/multi-chain/native/service/cross_chain_manager/btc"
	crosscommon "github.com/ontio/multi-chain/native/service/cross_chain_manager/common"
	"github.com/ontio/multi-chain/native/service/cross_chain_manager/eth"
	"github.com/ontio/multi-chain/native/service/cross_chain_manager/ont"
	"github.com/ontio/multi-chain/native/service/side_chain_manager"
)

const (
	IMPORT_OUTER_TRANSFER_NAME = "ImportOuterTransfer"
	VOTE_NAME                  = "Vote"
	MULTI_SIGN                 = "MultiSign"
	INIT_REDEEM_SCRIPT         = "initRedeemScript"
)

func RegisterCrossChainManagerContract(native *native.NativeService) {
	native.Register(INIT_REDEEM_SCRIPT, InitRedeemScript)

	native.Register(IMPORT_OUTER_TRANSFER_NAME, ImportExTransfer)
	native.Register(VOTE_NAME, Vote)
	native.Register(MULTI_SIGN, MultiSign)
}

func GetChainHandler(chainid uint64) (crosscommon.ChainHandler, error) {
	switch chainid {
	case utils.BTC_CHAIN_ID:
		return btc.NewBTCHandler(), nil
	case utils.ETH_CHAIN_ID:
		return eth.NewETHHandler(), nil
	case utils.ONT_CHAIN_ID:
		return ont.NewONTHandler(), nil
	case utils.NEO_CHAIN_ID:
		return neo.NewNEOHandler(), nil
	default:
		return nil, fmt.Errorf("not a supported chainid:%d", chainid)
	}
}

func ImportExTransfer(native *native.NativeService) ([]byte, error) {
	params := new(crosscommon.EntranceParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, contract params deserialize error: %v", err)
	}
	log.Infof("SourceChainID:%v\n", params.SourceChainID)
	log.Infof("TargetChainID:%v\n", params.TargetChainID)
	log.Infof("Proof:%v\n", params.Proof)
	log.Infof("TxData:%v\n", params.TxData)
	log.Infof("Height:%v\n", params.Height)
	log.Infof("RelayerAddress:%v\n", params.RelayerAddress)
	log.Infof("value:%v\n", params.Value)

	chainID := params.SourceChainID

	//check if chainid exist
	sideChain, err := side_chain_manager.GetSideChain(native, chainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain.ChainId != chainID {
		return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, side chain is not registered")
	}

	handler, err := GetChainHandler(chainID)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	//1. verify tx
	if chainID == 2 || chainID == 3 {
		txParam, err := handler.MakeDepositProposal(native)
		if err != nil {
			return utils.BYTE_FALSE, err
		}

		//2. make target chain tx
		targetid := txParam.ToChainID

		//check if chainid exist
		sideChain, err = side_chain_manager.GetSideChain(native, targetid)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, side_chain_manager.GetSideChain error: %v", err)
		}
		if sideChain.ChainId != targetid {
			return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, targetid chain is not registered")
		}

		targetHandler, err := GetChainHandler(targetid)
		if err != nil {
			return utils.BYTE_FALSE, err
		}
		//NOTE, you need to store the tx in this
		err = targetHandler.MakeTransaction(native, txParam)
		if err != nil {
			return utils.BYTE_FALSE, err
		}
		return utils.BYTE_TRUE, nil
	}
	_, err = handler.MakeDepositProposal(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}

func Vote(native *native.NativeService) ([]byte, error) {
	handler := btc.NewBTCHandler()

	//1. vote
	ok, txParam, err := handler.Vote(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	if ok {
		//2. make target chain tx
		targetid := txParam.ToChainID

		//check if chainid exist
		sideChain, err := side_chain_manager.GetSideChain(native, targetid)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, side_chain_manager.GetSideChain error: %v", err)
		}
		if sideChain.ChainId != targetid {
			return utils.BYTE_FALSE, fmt.Errorf("ImportExTransfer, targetid chain is not registered")
		}

		targetHandler, err := GetChainHandler(targetid)
		if err != nil {
			return utils.BYTE_FALSE, err
		}
		//NOTE, you need to store the tx in this
		err = targetHandler.MakeTransaction(native, txParam)
		if err != nil {
			return utils.BYTE_FALSE, err
		}
		return utils.BYTE_TRUE, nil
	}
	return utils.BYTE_TRUE, nil
}

func MultiSign(native *native.NativeService) ([]byte, error) {
	handler := btc.NewBTCHandler()

	//1. multi sign
	err := handler.MultiSign(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}

func InitRedeemScript(native *native.NativeService) ([]byte, error) {
	handler := btc.NewBTCHandler()

	//1. multi sign
	err := handler.InitRedeemScript(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}
