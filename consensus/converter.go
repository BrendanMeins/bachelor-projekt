package consensus

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/taurusgroup/frost-ed25519/pkg/frost/party"
	"sort"
)

/*Da die Frost-Bibliothek mit Integer int16 IDs arbeitet und die libp2p
Bibliothek mit Strings, muss ein Mapping erzeugt werden, sodass jeder
Node dieselbe ID für jeden Peer erzeugt, und dieselbe Liste von Teilnehmern.
(Frost erwartet eine aufsteigend sortierte Liste)
*/
var selfId int
var partyIDMap map[int]peer.ID
var partyIds []int
var sortedPartySlice party.IDSlice

func generateParty() {
	var idMap = make(map[int]peer.ID)

	for _, peer := range Host.Peerstore().Peers() {
		partyid := GetPartyId(peer)
		idMap[partyid] = peer
	}

	createSliceFromPartyMap(idMap)
	createSortedSlice()
	partyIDMap = idMap
	selfId = GetPartyId(Host.ID())
	fmt.Println(">Party Size: ", sortedPartySlice.N())
	fmt.Println(">Self ID Size: ", selfId)
}

func GetPartyId(peer peer.ID) int {
	hash := md5.Sum([]byte(peer.String()))
	var res []byte
	for _, b := range hash {
		res = append(res, b)
	}
	val := binary.BigEndian.Uint16(res)
	return int(val)
}
func createSliceFromPartyMap(partyMap map[int]peer.ID) {
	var idslice []int
	for key := range partyMap {
		idslice = append(idslice, key)
	}

	sort.Slice(idslice, func(i, j int) bool {
		return idslice[i] < idslice[j]
	})
	partyIds = idslice
}

func createSortedSlice() {
	var tmp party.IDSlice
	for _, v := range partyIds {
		tmp = append(tmp, party.ID(v))
	}
	sortedPartySlice = tmp
}
