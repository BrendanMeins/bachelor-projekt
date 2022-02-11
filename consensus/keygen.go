package consensus

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/taurusgroup/frost-ed25519/pkg/frost"
	"github.com/taurusgroup/frost-ed25519/pkg/frost/keygen"
	"github.com/taurusgroup/frost-ed25519/pkg/frost/party"
	"github.com/taurusgroup/frost-ed25519/pkg/messages"
	"github.com/taurusgroup/frost-ed25519/pkg/state"
	"io/ioutil"
	"time"
)

const KeyGenProtocolId = protocol.ID("/frost/keygen/0/1")

//Channels die gelesen und beschrieben werden können
var keygenInChannel chan []byte
var keygenOutChannel chan messages.Message
var keyGenState *state.State
var keyGenOutput *keygen.Output

func readKeyGen() {
	for {
		select {
		case msg := <-keygenInChannel:
			if len(msg) != 0 {
				var msgTmp messages.Message
				if err := msgTmp.UnmarshalBinary(msg); err != nil {
					fmt.Errorf("failed to unmarshal message: %w", err)
				}
				fmt.Println("------------Message for KeyGen--------------")
				fmt.Println(msgTmp)
				if err := keyGenState.HandleMessage(&msgTmp); err != nil {
					fmt.Errorf("failed to handle message: %w", err)
				}
			}
			msgsOut := keyGenState.ProcessAll()
			for _, msgOut := range msgsOut {
				keygenOutChannel <- *msgOut
			}
		}
	}
}
func writeKeyGen() {
	for {
		select {
		case msg := <-keygenOutChannel:
			time.Sleep(time.Second * 2)
			fmt.Println("----------Sending Message-----------")
			fmt.Println(msg)
			if msg.IsBroadcast() {
				for _, peer := range Host.Peerstore().Peers() {
					send(msg, peer)
				}
			} else {
				var peerId = partyIDMap[int(msg.To)]
				send(msg, peerId)
			}
		}
	}
}
func keyGenInit() {
	generateParty()
	fmt.Println("treshold: ", sortedPartySlice.N()-1)
	keygenstate, keygenoutput, err := frost.NewKeygenState(party.ID(selfId), sortedPartySlice, sortedPartySlice.N()-1, 0)

	if err != nil {
		panic(err)
	}

	keyGenState = keygenstate
	keyGenOutput = keygenoutput

	go readKeyGen()
	go writeKeyGen()

	msgsOut := keyGenState.ProcessAll()
	for _, msgOut := range msgsOut {
		keygenOutChannel <- *msgOut
	}

	err = keyGenState.WaitForError()
	if err != nil {
		panic(err)
	}
	fmt.Println("\n\n\n")
	fmt.Println("Result for KeyGen: ")
	fmt.Println("\n")
	fmt.Println("Group Key")
	fmt.Println(keyGenOutput.Public.GroupKey.ToEd25519())
	fmt.Println("\n")
	fmt.Println("Public Shares :")
	fmt.Println(keyGenOutput.Public.Shares)
	fmt.Println("\n")
	fmt.Println("Secret Share")
	fmt.Println(keyGenOutput.SecretKey)

	fmt.Println("\n\n\n")

}
func send(message messages.Message, peer peer.ID) {
	if peer == Host.ID() {
		//KeyGenState.HandleMessage(&message)
		return
	}
	stream, err := Host.NewStream(context.Background(), peer, KeyGenProtocolId)
	if err != nil {
		panic(err)
	}
	data, err := message.MarshalBinary()
	stream.Write(data)
	defer stream.Close()

}
func handleKeyGenMessage(s network.Stream) {
	if keyGenState == nil {
		StartKeyGen()
	}
	data, err := ioutil.ReadAll(s)
	if err != nil {
		panic(err)
	}
	keygenInChannel <- data
	err = s.Close()
	if err != nil {
		panic(err)
	}

}
