package p2p

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/config"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/airchains-network/decentralized-sequencer/utilis"
	v1 "github.com/airchains-network/decentralized-sequencer/zk/v1"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func BatchGeneration(wg *sync.WaitGroup) {
	defer wg.Done()
	GenerateUnverifiedPods()
}

func createPOD(lds *leveldb.DB, ldt *leveldb.DB, batchStartIndex []byte) (witness []byte, unverifiedProof []byte, MRH []byte, podData *types.BatchStruct, err error) {
	limit, err := lds.Get([]byte("batchCount"), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in getting batchCount from static db : %s", err.Error()))
		return nil, nil, nil, nil, err
	}
	limitInt, _ := strconv.Atoi(strings.TrimSpace(string(limit)))
	batchStartIndexInt, _ := strconv.Atoi(strings.TrimSpace(string(batchStartIndex)))

	var batch types.BatchStruct

	var From []string
	var To []string
	var Amounts []string
	var TransactionHash []string
	var SenderBalances []string
	var ReceiverBalances []string
	var Messages []string
	var TransactionNonces []string
	var AccountNonces []string

	for i := batchStartIndexInt; i < (config.PODSize * (limitInt + 1)); i++ {
		findKey := fmt.Sprintf("txns-%d", i+1)
		txData, err := ldt.Get([]byte(findKey), nil)
		if err != nil {
			i--
			time.Sleep(1 * time.Second)
			continue
		}
		var tx types.TransactionStruct
		err = json.Unmarshal(txData, &tx)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in unmarshalling tx data : %s", err.Error()))
			os.Exit(0)
		}

		senderBalancesCheck, err := utilis.GetBalance(tx.From, (tx.BlockNumber - 1))
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting sender balance : %s", err.Error()))
			os.Exit(0)
		}

		receiverBalancesCheck, err := utilis.GetBalance(tx.To, (tx.BlockNumber - 1))
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting reciver balance : %s", err.Error()))
			os.Exit(0)
		}

		accountNouceCheck, err := utilis.GetAccountNonce(context.Background(), tx.Hash, tx.BlockNumber)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting account nonce : %s", err.Error()))
			os.Exit(0)
		}

		From = append(From, tx.From)
		To = append(To, tx.To)
		Amounts = append(Amounts, tx.Value)
		TransactionHash = append(TransactionHash, tx.Hash)
		SenderBalances = append(SenderBalances, senderBalancesCheck)
		ReceiverBalances = append(ReceiverBalances, receiverBalancesCheck)
		Messages = append(Messages, tx.Input)
		TransactionNonces = append(TransactionNonces, tx.Nonce)
		AccountNonces = append(AccountNonces, accountNouceCheck)
	}

	batch.From = From
	batch.To = To
	batch.Amounts = Amounts
	batch.TransactionHash = TransactionHash
	batch.SenderBalances = SenderBalances
	batch.ReceiverBalances = ReceiverBalances
	batch.Messages = Messages
	batch.TransactionNonces = TransactionNonces
	batch.AccountNonces = AccountNonces

	logs.Log.Warn("batch:")
	fmt.Println(batch)

	witnessVector, currentStatusHash, proofByte, pkErr := v1.GenerateProof(batch, limitInt+1)
	if pkErr != nil {
		logs.Log.Error(fmt.Sprintf("Error in generating proof : %s", pkErr.Error()))
		return nil, nil, nil, nil, pkErr
	}
	logs.Log.Warn(fmt.Sprintf("Successfully generated  Unverified proof for Batch %s in the latest phase", strconv.Itoa(limitInt+1)))

	// marshal witnessVector
	witnessVectorByte, err := json.Marshal(witnessVector)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling witness vector : %s", err.Error()))
		os.Exit(0)
	}

	// string to []byte currentStatusHash
	currentStatusHashByte, err := json.Marshal(currentStatusHash)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling current status hash : %s", err.Error()))
		os.Exit(0)
	}

	//fmt.Println("Witness Vector: ", witnessVector)
	//fmt.Println("Proof: ", proofByte)
	//fmt.Println("Current Status Hash: ", currentStatusHash)

	return witnessVectorByte, proofByte, currentStatusHashByte, &batch, nil

	// declear proof

	//proofGossip := types.ProofData{
	//	Proof:     proofByte,
	//	PodNumber: uint64(limitInt + 1),
	//}
	//proofByteGossip, err := json.Marshal(proofGossip)
	//if err != nil {
	//	logs.Log.Error(fmt.Sprintf("Error in marshalling proof data : %s", err.Error()))
	//	os.Exit(0)
	//}
	//
	//proofData := types.GossipData{
	//	Type: "proof",
	//	Data: proofByteGossip,
	//}
	//
	//// proofData to byte
	//ProofDataByte, err := json.Marshal(proofData)
	//if err != nil {
	//	logs.Log.Error(fmt.Sprintf("Error in marshalling proof data : %s", err.Error()))
	//	os.Exit(0)
	//}

}

func generatePodHash(Witness, uZKP, MRH []byte, podNumber []byte) []byte {
	//h := sha256.New()
	////h.Write(Witness)
	//h.Write(uZKP)
	//h.Write(MRH)
	//h.Write(podNumber)
	//return h.Sum(nil)
	//return uZKP

	return uZKP
}

func GenerateUnverifiedPods() {
	lds := shared.Node.NodeConnections.StaticDatabaseConnection
	ldt := shared.Node.NodeConnections.TxnDatabaseConnection

	latestBatch := shared.GetLatestBatchIndex(lds)
	//batchStartIndexInt, _ := strconv.Atoi(strings.TrimSpace(string(latestBatch)))

	currentPodNumber, err := lds.Get([]byte("batchCount"), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in getting sssssss from static db : %s", err.Error()))
		os.Exit(0)
	}
	SelectedMaster := MasterTracksSelection(Node)
	currentPodNumberInt, _ := strconv.Atoi(strings.TrimSpace(string(currentPodNumber)))
	batchNumber, _ := strconv.Atoi(strings.TrimSpace(string(currentPodNumberInt + 1)))

	var batchInput *types.BatchStruct
	Witness, uZKP, MRH, batchInput, err := createPOD(lds, ldt, latestBatch)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in creating POD : %s", err.Error()))
		os.Exit(0)
	}

	TrackAppHash := generatePodHash(Witness, uZKP, MRH, latestBatch)
	podState := shared.GetPodState()

	tempMasterTrackAppHash := podState.MasterTrackAppHash
	if podState.MasterTrackAppHash != nil {
		tempMasterTrackAppHash = podState.MasterTrackAppHash
	}

	updateNewPodState(TrackAppHash, Witness, uZKP, MRH, uint64(batchNumber+1), batchInput)

	// Here the MasterTrack Will Broadcast the uZKP in the Network
	if SelectedMaster == Node.ID() {
		// Preparing the Message that master track will gossip to the Network

		proofData := ProofData{
			PodNumber:    uint64(batchNumber + 1),
			TrackAppHash: TrackAppHash,
		}

		// Marshal the proofData
		proofDataByte, err := json.Marshal(proofData)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in marshalling proof data : %s", err.Error()))
		}

		gossipMsg := types.GossipData{
			Type: "proof",
			Data: proofDataByte,
		}

		gossipMsgByte, err := json.Marshal(gossipMsg)
		if err != nil {
			logs.Log.Error("Error marshaling gossip message")
			return
		}

		logs.Log.Info("Sending proof result: %s")
		BroadcastMessage(context.Background(), Node, gossipMsgByte)

	} else {
		currentPodData := shared.GetPodState()
		if bytes.Equal(currentPodData.TracksAppHash, tempMasterTrackAppHash) {
			SendValidProof(CTX, currentPodData.LatestPodHeight, SelectedMaster)
			return
		} else {
			SendInvalidProofError(CTX, currentPodData.LatestPodHeight, SelectedMaster)
			return
		}
	}

}

func updateNewPodState(CombinedPodHash, Witness, uZKP, MRH []byte, podNumber uint64, batchInput *types.BatchStruct) {
	var podState *shared.PodState
	// empty votes
	votes := make(map[string]shared.Votes)

	podState = &shared.PodState{
		LatestPodHeight:     podNumber,
		LatestPodHash:       MRH,
		LatestPodProof:      uZKP,
		LatestPublicWitness: Witness,
		Votes:               votes,
		TracksAppHash:       CombinedPodHash,
		Batch:               batchInput,
	}

	shared.SetPodState(podState)
}

func saveVerifiedPOD() {

	// get pod data from pod state
	podState := shared.GetPodState()

	// declear useful variables
	batchInput := podState.Batch
	currentPodNumber := podState.LatestPodHeight
	currentPodNumberInt := int(currentPodNumber + 1)

	batchJSON, err := json.Marshal(batchInput)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling batch data : %s", err.Error()))
		os.Exit(0)
	}
	ldbatch := shared.Node.NodeConnections.GetDataAvailabilityDatabaseConnection()
	lds := shared.Node.NodeConnections.GetStaticDatabaseConnection()
	batchKey := fmt.Sprintf("batch-%d", currentPodNumberInt+1)
	err = ldbatch.Put([]byte(batchKey), batchJSON, nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in writing batch data to file : %s", err.Error()))
		os.Exit(0)
	}

	// uint64 to int
	err = lds.Put([]byte("batchStartIndex"), []byte(strconv.Itoa(config.PODSize*(currentPodNumberInt+1))), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in updating batchStartIndex in static db : %s", err.Error()))
		os.Exit(0)
	}

	err = lds.Put([]byte("batchCount"), []byte(strconv.Itoa(currentPodNumberInt+1)), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in updating batchCount in static db : %s", err.Error()))
		os.Exit(0)
	}

	err = os.WriteFile("data/batchCount.txt", []byte(strconv.Itoa(currentPodNumberInt+1)), 0666)
	if err != nil {
		panic("Failed to update batch number: " + err.Error())
	}

}