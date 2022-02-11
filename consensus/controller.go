package consensus

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/taurusgroup/frost-ed25519/pkg/messages"
)

var Host host.Host

func Init(h host.Host) {
	Host = h
	Host.SetStreamHandler(KeyGenProtocolId, handleKeyGenMessage)
	Host.SetStreamHandler(signProtocolId, handleSignMessage)
	Host.SetStreamHandler(signInitMessage, handleSignInit)
	keygenInChannel = make(chan []byte)
	keygenOutChannel = make(chan messages.Message)

	signInChannel = make(chan []byte, 5)
	signOutChannel = make(chan messages.Message)

}

func StartKeyGen() {
	go keyGenInit()
}

func StartSign(msg SignInitMessage) {

	if keyGenState == nil {
		fmt.Errorf("KeyGenState is nil")
	} else if !keyGenState.IsFinished() {
		fmt.Errorf("KeyGen did not finish")
	} else if keyGenOutput.SecretKey == nil {
		fmt.Errorf("KeyGenOutput is nil")
	} else {
		sendSignatureMessage(msg)
		go signInit(msg)
	}

}
