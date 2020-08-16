package cgraph

import (
	"fmt"
	"encoding/json"
	"encoding/binary"
    	"crypto/md5"
	"log"
	"gonum.org/v1/gonum/graph"
)


type CNode struct {
    Id int64 `json:"id"`
    Name string `json:"label"`
    Type string `json:"type"`
    Attributes map[string]interface{}
}

func (n *CNode) ID() int64 {
   return n.Id
}

func (n *CNode) String() string {
    return n.Name
}

func NewUser(login string, url string) graph.Node {
    sum := md5.Sum([]byte(login))
    num, _ := binary.Varint(sum[:])
    attrs := make(map[string]interface{})
    attrs["login"] = login
    attrs["url"] = url
    return &CNode{
        Id: num,
        Name: login,
        Type: "User",
        Attributes: attrs,
    }
}

func NewRepo(owner string, name string) graph.Node {
    repo := new(CNode)
    repo.Type = "Repo"
    attrs := make(map[string]interface{})
    attrs["owner"] = owner
    attrs["name"] = name
    repo.Name = fmt.Sprintf("%s/%s", owner, name) 
    sum := md5.Sum([]byte(repo.String()))
    num, _ := binary.Varint(sum[:])
    repo.Id = num
    return repo
}


func NewContribution(source_node graph.Node, target_node graph.Node , total int, first int, last int) graph.Edge {
    attrs := make(map[string]interface{})
    attrs["total"] = total
    attrs["first"] = first
    attrs["last"] = last
    name := fmt.Sprintf("%s_%s", source_node, target_node)
    sum := md5.Sum([]byte(name))
    num, _ := binary.Varint(sum[:])
    return CEdge{
        num,
        source_node,
        target_node,
        attrs,
    }
}

type CNodes struct{
    Curr int
    Nodes []graph.Node
}


// ==== NEW GRAPH ==== 
type CGraph struct {
    CNodes CGraphNodes `json:"nodes"`
    CEdges CGraphEdges `json:"edges"`
}

func (c CEdge) From() graph.Node {
    return c.Source
}

func (c CEdge) To() graph.Node {
    return c.Target
}

func (c CEdge) ReversedEdge() graph.Edge {
    c.Target = c.Source
    c.Source = c.Target
    return c
}

func (n *CNodes) Len() int {
    return len(n.Nodes)
}

func (n *CNodes) Next() bool {
    if n.Len() >= n.Curr + 1 {
        return false
    }
    return true
}

func (n *CNodes) Node() graph.Node {
    return n.Nodes[n.Curr]
}

func (n *CNodes) Reset() {
    n.Curr = 0
}

func (g *CGraph) AddNode(n graph.Node) {
    id := n.ID()
    g.CNodes[id] = n
}


func (g *CGraph) SetEdge(e graph.Edge) {
    uid := e.From().ID()
    vid := e.To().ID()
    g.CEdges[uid] = append(g.CEdges[uid], e)
    g.CEdges[vid] = append(g.CEdges[vid], e)
}

func (g *CGraph) Edge(uid, vid int64) graph.Edge {
    // u ->  v node 
    edges := g.CEdges[uid]
    for _, e := range edges {
        if e.To().ID() == vid {
            return e
        }
    }
    return nil
}

type CGraphNodes map[int64]graph.Node
type CGraphEdges map[int64][]graph.Edge

func NewCGraph() *CGraph{
    nodes := make(CGraphNodes)
    edges := make(CGraphEdges)
    return &CGraph{
        nodes,
        edges,
    }
}

func (g *CGraph) Node(id int64) graph.Node {
    return g.CNodes[id]
}


func (g *CGraph) Nodes() graph.Nodes {
    n := new(CNodes)
    for _, v := range g.CNodes {
       n.Nodes = append(n.Nodes, v)
    }
    return n
}

func (n CGraphNodes) MarshalJSON() ([]byte, error) {
    nodes := *new([]graph.Node)
    for _, v := range n {
        nodes = append(nodes, v)
    }
    return json.Marshal(nodes)
}

func (e CGraphEdges) MarshalJSON() ([]byte, error) {
    edges := *new([]CEdge)
    for _, v := range e {
        for _, edge := range v {
            edges = append(edges, edge.(CEdge))
        }
    }
    return json.Marshal(edges)
}

func (g *CGraph) From(id int64) graph.Nodes {
    edges := g.CEdges[id]
    n := new(CNodes)
    for _, e := range edges {
        node_id := e.To().ID()
        if  node_id != id {
            n.Nodes = append(n.Nodes, g.CNodes[node_id])
        }
    }
    return n
}

func (g *CGraph) HasEdgeBetween(xid, yid int64) bool {
    edges := g.CEdges[xid]
    for _, e :=  range edges {
        if e.To().ID() == yid || e.From().ID() == yid {
            return true
        }
    }
    return false
}


type CEdge struct {
    Id int64 `json:"id"`
    Source graph.Node `json:"source"`
    Target graph.Node `json:"target"`
    Attributes map[string]interface{}
}

func (e CEdge) marshalJSON() ([]byte, error) {
    log.Println("marshal cedges")
    source := *(e.Source.(*CNode))
    target := *(e.Target.(*CNode))
    id := fmt.Sprintf("%s_%s", source.Name, target.Name)
    return json.Marshal(struct{
        Id string
        From CNode
        To CNode
        Attributes map[string]interface{}
    }{
        id,
        source,
        target,
        e.Attributes,
    })
}
