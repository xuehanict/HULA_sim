package main

import (
	"io/ioutil"
	"fmt"
	"os"
	"encoding/json"
)

const (
	twoNodesGraphFile = "testdata/two_nodes.json"

	basicGraphFilePath = "testdata/basic_graph.json"
)


func main()  {

	var (
		nodes = make(map[RouterID]*HulaRouter,0)
		edges = make(map[string]*HulaLink, 0)
	)
	graphJson, err := ioutil.ReadFile(twoNodesGraphFile)
	if err != nil {
		fmt.Printf("can't open the json file: %v", err)
		os.Exit(1)
	}

	var g testGraph
	if err := json.Unmarshal(graphJson, &g); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

 	for _, node := range g.Nodes {
 		router := newRouter(node.Id)
 		router.RouterBase = nodes
 		router.LinkBase = edges
 		nodes[node.Id] = router
	}

	for _, edge := range g.Edges {
		err := addLink(edge.Node1, edge.Node2, edge.Capacity, nodes, edges)
		if err != nil {
			fmt.Printf("failed: %v", err)
		}
	}





	fmt.Printf("%v\n", g)


}



