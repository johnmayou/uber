package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/uber/h3-go/v4"
)

type Node struct {
	Id    uint32
	Coord Coord
}

type Edge struct {
	To     uint32
	Weight uint32
}

type Graph struct {
	adj     map[uint32][]Edge // adj[node] = that node's outgoing edges
	nodeMap map[uint32]Node
}

func (g *Graph) ForEachEdge(node uint32, fn func(e Edge)) {
	for _, e := range g.adj[node] {
		fn(e)
	}
}

func buildGraph(in io.Reader) *Graph {
	scanner := osmpbf.New(context.Background(), in, 4)
	defer func() {
		if err := scanner.Close(); err != nil {
			log.Printf("closing scanner: %v", err)
		}
	}()

	header, err := scanner.Header()
	if err != nil {
		log.Fatalf("scanning header: %v", err)
	} else {
		b, _ := json.MarshalIndent(header, "", "  ")
		log.Printf("header (%T):\n%s", header, b)
	}

	adj := make(map[uint32][]Edge)
	nodeMap := make(map[uint32]Node)

	for scanner.Scan() {
		object := scanner.Object()
		switch obj := object.(type) {
		case *osm.Node:
			src := uint32(obj.ObjectID())
			nodeMap[src] = Node{
				Id: src,
				Coord: Coord{
					lat: obj.Lat,
					lng: obj.Lon,
				},
			}
			if _, ok := adj[src]; !ok {
				adj[src] = []Edge{}
			}
		case *osm.Way:
			for i := 0; i < len(obj.Nodes)-1; i++ {
				src := uint32(obj.Nodes[i].ID)
				dst := uint32(obj.Nodes[i+1].ID)
				adj[src] = append(adj[src], Edge{To: dst})
			}
		default:
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("scanning: %v", err)
	}

	return &Graph{adj: adj, nodeMap: nodeMap}
}

type CellIndex struct {
	index map[h3.Cell][]uint32
}

const cellIndexResolution = 8

func buildCellIndex(nodes map[uint32]Node) (*CellIndex, error) {
	ci := &CellIndex{index: make(map[h3.Cell][]uint32)}

	for _, node := range nodes {
		cell, err := h3.LatLngToCell(h3.LatLng{Lat: node.Coord.lat, Lng: node.Coord.lng}, cellIndexResolution)
		if err != nil {
			return nil, fmt.Errorf("convert lat lng to cell: %w", err)
		}
		ci.index[cell] = append(ci.index[cell], node.Id)
	}

	return ci, nil
}

func (ci *CellIndex) getClosestNode(coord Coord, nodeMap map[uint32]Node) (Node, error) {
	nodeIds, err := ci.getClosestNodeIds(coord)
	if err != nil {
		return Node{}, fmt.Errorf("get closest nodes: %w", err)
	}

	calcDist := func(candidate Coord) float64 {
		return math.Pow(candidate.lat-coord.lat, 2) + math.Pow(candidate.lng-coord.lng, 2)
	}

	firstNode, ok := nodeMap[nodeIds[0]]
	if !ok {
		return Node{}, fmt.Errorf("first node id (%d) not found in map", nodeIds[0])
	}
	closest := firstNode
	closestDist := calcDist(firstNode.Coord)

	for _, nodeId := range nodeIds[1:] {
		node, ok := nodeMap[nodeId]
		if !ok {
			return Node{}, fmt.Errorf("node id (%d) not found in map", nodeId)
		}

		if dist := calcDist(node.Coord); dist < closestDist {
			closest = node
			closestDist = dist
		}
	}

	return closest, nil
}

func (ci *CellIndex) getClosestNodeIds(coord Coord) ([]uint32, error) {
	cell, err := h3.LatLngToCell(h3.LatLng{Lat: coord.lat, Lng: coord.lng}, cellIndexResolution)
	if err != nil {
		log.Fatalf("src convert to cell: %v", err)
	}

	cells, err := h3.GridDisk(cell, 1)
	if err != nil {
		return nil, fmt.Errorf("requesting neighboring cells for %q: %w", cell, err)
	}

	var out []uint32
	for _, c := range cells {
		if nodeIds, ok := ci.index[c]; ok {
			out = append(out, nodeIds...)
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("node not found near cell %q", cell)
	}

	return out, nil
}

type Coord struct {
	lat float64
	lng float64
}

func fetchCoord(ctx context.Context, client http.Client, query string) (Coord, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.opencagedata.com/geocode/v1/json", nil)
	if err != nil {
		return Coord{}, fmt.Errorf("request create: %w", err)
	}

	queryParams := url.Values{}
	queryParams.Add("q", query)
	queryParams.Add("key", os.Getenv("OPEN_CAGE_DATA_API_KEY"))

	req.URL.RawQuery = queryParams.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return Coord{}, fmt.Errorf("client request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("close response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return Coord{}, fmt.Errorf("reading body: %w", err)
		}
		return Coord{}, fmt.Errorf("http %d: %s", resp.StatusCode, body)
	}

	type Result struct {
		Components struct {
			Type string `json:"_type"`
		} `json:"components"`
		Confidence int `json:"confidence"`
		Geometry   struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"geometry"`
	}

	type Response struct {
		Results []Result
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Coord{}, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(response.Results) == 0 {
		return Coord{}, fmt.Errorf("no places found for query: %q", query)
	}

	var buildings []Result
	for i := range len(response.Results) {
		result := response.Results[i]
		if result.Components.Type == "building" {
			buildings = append(buildings, result)
		}
	}

	if len(buildings) == 0 {
		return Coord{}, fmt.Errorf("no buildings found for query: %q", query)
	}

	maxConfidenceGeo := buildings[0].Geometry
	maxConfidence := buildings[0].Confidence
	for i := 1; i < len(buildings); i++ {
		building := buildings[i]
		if building.Confidence > maxConfidence {
			maxConfidenceGeo = building.Geometry
			maxConfidence = building.Confidence
		}
	}

	coord := Coord{
		lat: maxConfidenceGeo.Lat,
		lng: maxConfidenceGeo.Lng,
	}
	return coord, nil
}

func main() {
	godotenv.Load()

	file, err := os.Open("osm/minneapolis-260629.osm.pbf")
	if err != nil {
		log.Fatalf("opening osm file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("closing file: %v", err)
		}
	}()

	var src, dst Coord
	if false {
		client := http.Client{
			Timeout: 10 * time.Second,
		}
		ctx := context.Background()

		src, err = fetchCoord(ctx, client, "601 SE Main Street, Minneapolis, MN 55413")
		if err != nil {
			log.Fatalf("fetch src coord: %v", err)
		}
		dst, err = fetchCoord(ctx, client, "333 S 7th St, Minneapolis, MN 55402")
		if err != nil {
			log.Fatalf("fetch dst coord: %v", err)
		}
	} else {
		src = Coord{lat: 44.98193, lng: -93.24811}
		dst = Coord{lat: 44.97422, lng: -93.26758}
	}

	fmt.Printf("coord for src: %+v\n", src)
	fmt.Printf("coord for dst: %+v\n", dst)

	graph := buildGraph(file)
	fmt.Printf("node count: %d\n", len(graph.adj))

	index, err := buildCellIndex(graph.nodeMap)
	if err != nil {
		log.Fatalf("build cell index: %v", err)
	}

	srcClosest, err := index.getClosestNode(src, graph.nodeMap)
	if err != nil {
		log.Fatalf("closest node to src: %v", err)
	}
	dstClosest, err := index.getClosestNode(dst, graph.nodeMap)
	if err != nil {
		log.Fatalf("closest node to dst: %v", err)
	}

	fmt.Printf("closest node to src: %+v\n", srcClosest)
	fmt.Printf("closest node to dst: %+v\n", dstClosest)
}
