package main

import (
	"github.com/miekg/bitradix"
	"log"
	"net"
	"reflect"
)

func addRoute(r *bitradix.Radix32, ipnet *net.IPNet, asn uint32) {
	net, mask := ipNetToUint(ipnet)
	r.Insert(net, mask, asn)
}

func removeRoute(r *bitradix.Radix32, ipnet *net.IPNet, asn uint32) {
	net, mask := ipNetToUint(ipnet)
	r.Remove(net, mask)
}

func ipNetToUint(n *net.IPNet) (i uint32, mask int) {
	i = ipToUint(&n.IP)
	mask, _ = n.Mask.Size()
	return
}

func ipToUint(nip *net.IP) (i uint32) {
	ip := nip.To4()
	fv := reflect.ValueOf(&i).Elem()
	fv.SetUint(uint64(uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[+3])))
	return
}

func uintToIP(n uint32) net.IP {
	return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

func nodeToIPNet(node *bitradix.Radix32) *net.IPNet {
	ip := uintToIP(node.Key())

	ipnet := net.IPNet{IP: ip, Mask: net.CIDRMask(node.Bits(), 32)}
	return &ipnet
}

func (n *Neighbor) FindNode(ip *net.IP) *bitradix.Radix32 {
	log.Println("Looking for ASN for", ip)
	i := ipToUint(ip)
	node := n.trie.Find(i, 32)
	return node
}

func (n *Neighbor) FindAsn(ip *net.IP) ASN {
	node := n.FindNode(ip)
	return ASN(node.Value)
}
