package main

import (
	"net"
)

type ASN uint

type ASPath []ASN

type Prefixes map[string]ASN

type Neighbor struct {
	State     string
	AsnPrefix map[ASN]Prefixes
	PrefixAsn Prefixes
	Updates   int
}

type Route struct {
	Options    map[string]string
	Prefix     *net.IPNet
	ASPath     ASPath
	PrimaryASN ASN
}

type Neighbors map[string]*Neighbor

const (
	parseKey = iota
	parseValue
	parseList
	parseSkip
)

var DEBUG bool

func (n *Neighbor) PrefixCount() int {
	return len(n.PrefixAsn)
}

func (n *Neighbor) AsnCount() int {
	return len(n.AsnPrefix)
}
