package main

import "strconv"

func newLinkKey(r1, r2 RouterID) string {
	return strconv.Itoa(int(r1)) + "-" + strconv.Itoa(int(r2))
}

func min(n1, n2 int64) int64 {
	if n1 < n2 {
		return n1
	} else {
		return n2
	}
}

type testGraph struct {
	Info  []string   `json:"info"`
	Nodes []testNode `json:"nodes"`
	Edges []testEdge `json:"edges"`
}

type testNode struct {
	Id RouterID `json:"id"`
}

type testEdge struct {
	Node1    RouterID `json:"node_1"`
	Node2    RouterID `json:"node_2"`
	Capacity int64    `json:"capacity"`
}
