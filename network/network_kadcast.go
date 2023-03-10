package network

import (
	"IBS/network/num_set"
	"IBS/node"
	"IBS/node/hash"
	"IBS/node/routing"
	"log"
)

const K = 5

func NewKadcastNode(index int64, downloadBandwidth, uploadBandwidth int, region string) *node.Node {
	nodeID := hash.Hash64(uint64(index))
	//nodeID := uint64(index)
	return node.NewNode(
		nodeID,
		downloadBandwidth,
		uploadBandwidth,
		region,
		routing.NewKadcastTable(nodeID, K),
	)
}

type KadcastNet struct {
	*Network
	idSet *num_set.Set
}

func NewKadcastNet(size int) *KadcastNet {
	// bootNode is used for message generation (from node) only here
	bootNode := node.NewNode(BootNodeID, 0, 0, "", routing.NewKadcastTable(BootNodeID, K))
	net := NewNetwork(bootNode)
	net.generateNodes(size, NewKadcastNode)
	kNet := &KadcastNet{
		net,
		num_set.NewSet(),
	}
	kNet.initConnections()
	return kNet
}

// Introduce : return n nodes
func (kNet *KadcastNet) Introduce(id uint64, n int) []*node.Node {
	var nodes []*node.Node
	for i := 0; i < routing.KeySpaceBits; i++ {
		fakeID, err := routing.FakeIDForBucket(id, i)
		if err != nil {
			log.Fatal(err)
		}
		// 2*n+1 make sures the right keys for the bucket are covered
		// a crude node discovery implementation
		ids := kNet.idSet.Around(fakeID, 2*n+1)
		for _, u := range ids {
			nodes = append(nodes, kNet.Node(u))
		}
		//fmt.Println("introduce to:", id, "peerIDS:", ids)
	}
	return nodes
}

func (kNet *KadcastNet) initConnections() {
	for _, node := range kNet.nodes {
		kNet.idSet.Insert(node.Id())
	}
	for _, node := range kNet.nodes {
		peers := kNet.Introduce(node.Id(), K)
		for _, peer := range peers {
			kNet.Connect(node, peer, NewBasicPeerInfo)
		}
	}
	//for _, node := range kNet.nodes {
	//	fmt.Println("id=", node.Id())
	//	node.PrintTable()
	//}
}
