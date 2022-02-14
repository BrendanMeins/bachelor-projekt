package consensus

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/taurusgroup/frost-ed25519/pkg/frost"
	"github.com/taurusgroup/frost-ed25519/pkg/frost/party"
	"github.com/taurusgroup/frost-ed25519/pkg/frost/sign"
	"github.com/taurusgroup/frost-ed25519/pkg/messages"
	"github.com/taurusgroup/frost-ed25519/pkg/state"
)

const signProtocolId = protocol.ID("/frost/sign/0/1")
const signInitMessage = protocol.ID("/frost/init/init/1")

var signInChannel chan []byte
var signOutChannel chan messages.Message
var signInitChannel chan bool
var signState *state.State
var signOutput *sign.Output
var signers party.IDSlice

type SignInitMessage struct {
	Message string        `json:"message"`
	Signers party.IDSlice `json:"signers"`
}

//Funktion zum lesen der Nachrichten des Signaturprotokolls
func readSign() {
	for {
		select {
		case msg := <-signInChannel:
			if len(msg) != 0 {
				var msgTmp messages.Message
				if err := msgTmp.UnmarshalBinary(msg); err != nil {
					fmt.Errorf("failed to unmarshal message: %w", err)
				}
				fmt.Println("------------Message for Sign--------------")
				fmt.Println(msgTmp)
				if err := signState.HandleMessage(&msgTmp); err != nil {
					fmt.Errorf("failed to handle message: %w", err)
				}
			}
			msgsOut := signState.ProcessAll()
			//if len(msgsOut) == 0 {
			//	fmt.Println("Message channel was empty")
			//}
			for _, msgOut := range msgsOut {
				signOutChannel <- *msgOut
			}
		}
	}
}

//Funktion zum schreiben von Nachrichten, die das Signaturprotokoll betreffen
func writeSign() {
	for {
		select {
		case msg := <-signOutChannel:
			time.Sleep(time.Second * 2)
			fmt.Println("----------Sending Sign Message-----------")
			fmt.Println(msg)
			if msg.IsBroadcast() {
				for _, signer := range signers {
					sends(msg, signer)
				}
			} else {
				sends(msg, msg.To)
			}
		}
	}
}
func signInit(msg SignInitMessage) {
	//Signers sind die Nodes, die am Signaturprozess aktiv Teilnehmen,

	signers = msg.Signers
	fmt.Println("Starting Sign Protocol for message: ", msg)
	fmt.Println("My Public Key: ", keyGenOutput.Public.Shares[party.ID(selfId)])

	signSt, signOut, err := frost.NewSignState(msg.Signers, keyGenOutput.SecretKey, keyGenOutput.Public, []byte(msg.Message), 0)
	if err != nil {
		panic(err)
	}
	signOutput = signOut
	signState = signSt
	//Goroutinen zum senden und empfangen der Nachrichten, die das
	//Signaturprotokoll betreffen. Mit den Goroutinen kann über Channels
	//Kommuniziert werden
	go readSign()
	go writeSign()
	msgsOut := signState.ProcessAll()
	for _, msgOut := range msgsOut {
		signOutChannel <- *msgOut
	}
	err = signState.WaitForError()
	if err != nil {
		panic(err)
	}
	fmt.Println("\n\n\n")
	fmt.Println("Result for Sign: ")
	fmt.Println("\n")
	fmt.Println(signOutput.Signature.ToEd25519())
	fmt.Println("\n\n\n")
}
func sends(message messages.Message, id party.ID) {
	if id == party.ID(selfId) {
		//fmt.Println("send to self")
		//data, err := message.MarshalBinary()
		//if err != nil{
		//	panic(err)
		//}
		//signInChannel <- data
		return
	}
	stream, err := Host.NewStream(context.Background(), partyIDMap[int(id)], signProtocolId)
	if err != nil {
		panic(err)
	}
	data, err := message.MarshalBinary()
	stream.Write(data)
	defer stream.Close()
}

//Handler Funktion für den Libp2p Host, welche die Nachrichten des Signaturprotokolls
//empfängt und an die Goroutinen weiterleitet
func handleSignMessage(s network.Stream) {
	if signState != nil {
		data, err := ioutil.ReadAll(s)
		if err != nil {
			panic(err)
		}
		var tmp messages.Message
		err = tmp.UnmarshalBinary(data)
		if err != nil {
			panic(err)
		}
		signInChannel <- data
		err = s.Close()
		if err != nil {
			panic(err)
		}
	}
}

//Zur initialisierung des Signaturprotokolls wird eine Init Nachricht
//an das Netzwerk gesendet. Der Libp2p Host empfängt sie in dieser Funktion
//und startet das Signaturprotokoll. Durch die Nachricht weiß er, was signiert werden
//soll.
func handleSignInit(s network.Stream) {
	data, err := ioutil.ReadAll(s)
	if err != nil {
		panic(err)
	}
	var msg SignInitMessage
	json.Unmarshal(data, &msg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("received msg to sign %s\n", msg)
	if signState == nil {
		signInit(msg)
	}
}

func sendSignatureMessage(msg SignInitMessage) {
	for _, signer := range signers {
		if signer == party.ID(selfId) {
			continue
		}
		stream, err := Host.NewStream(context.Background(), partyIDMap[int(signer)], signInitMessage)
		if err != nil {
			panic(err)
		}
		var data []byte
		data, err = json.Marshal(msg)
		if err != nil {
			panic(err)
		}
		_, err = stream.Write(data)
		if err != nil {
			panic(err)
		}
		fmt.Println("Sending sign init")
		err = stream.Close()
		if err != nil {
			panic(err)
		}
	}
}

func GetInitMessage(msg string) *SignInitMessage {
	signers = establishSigners()
	return &SignInitMessage{
		Signers: signers,
		Message: msg,
	}
}

func establishSigners() party.IDSlice {
	//r := rand.Intn(int(sortedPartySlice.N()))
	//s := removeIndex(sortedPartySlice, r)
	return sortedPartySlice
}

func removeIndex(s []party.ID, index int) []party.ID {
	if s[index] == party.ID(selfId) {
		return removeIndex(s, index)
	}
	return append(s[:index], s[index+1:]...)
}
