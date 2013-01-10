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

func (n *Neighbor) FindAsn(ip *net.IP) ASN {
	DEBUG = true
	log.Println("Looking for ASN for", ip)
	i := ipToUint(ip)
	node := n.trie.Find(i, 32)
	return ASN(node.Value)
}
