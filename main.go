package main

import (
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "fmt"
    "encoding/binary"
    "encoding/json"
//    "github.com/google/jsonapi"
    "crypto/md5"
    "gonum.org/v1/gonum/graph"
    "gonum.org/v1/gonum/graph/encoding/dot"
)

type Week struct {
    W int
    A int
    D int
    C int

}

type Contributor struct {
    Weeks []Week
    Total int
    Author map[string]interface{}
}

type CNode struct {
    Id int64 `json:"id"`
    Name string
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


func NewContribution(from_node graph.Node, to_node graph.Node , total int, first int, last int) graph.Edge {
    attrs := make(map[string]interface{})
    attrs["total"] = total
    attrs["first"] = first
    attrs["last"] = last
    return CEdge{
        from_node,
        to_node,
        attrs,
    }
}

type CNodes struct{
    Curr int
    Nodes []graph.Node
}


// ==== NEW GRAPH ==== 
type CGraph struct {
    CNodes map[int64]graph.Node `json:"nodes"`
    CEdges map[int64][]graph.Edge `json:"edges"`
}

func (c CEdge) From() graph.Node {
    return c.from
}

func (c CEdge) To() graph.Node {
    return c.to
}

func (c CEdge) ReversedEdge() graph.Edge {
    c.to = c.from
    c.from = c.to
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

func NewCGraph() *CGraph{
    nodes := make(map[int64]graph.Node)
    edges := make(map[int64][]graph.Edge)
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


// ==== END GRAPH ====




//var repos = new([]Repo)
var repos = []graph.Node{
    NewRepo("ethereum", "go-ethereum"),
    NewRepo("smartcontractkit", "chainlink"),
    NewRepo("blockchainsllc", "DAO"),
    //*NewRepo("paritytech", "polkadot"),
}
//repos = append(repos, NewRepo("ethereum","go-ethereum"))
//repos = append(repos, NewRepo("smartcontractkit", "chainlink"))
//repos = append(repos, NewRepo("blockchainsllc", "DAO"))
//repos = append(repos, NewRepo("paritytech", "polkadot"))
//repos = append(repos, NewRepo("cosmos", "cosmos"))
//repos = append(repos, NewRepo("neo-project", "neo"))
//repos = append(repos, NewRepo("icon-project", "loopchain"))
//repos =    //Repo{"icon-project", "loopchain"},

// ==== EDGES ====
type CEdge struct {
    from graph.Node `json:"from"`
    to graph.Node `json:"to"`
    attributes map[string]interface{}
}

func extract_contribution_interval(weeks *[]Week) (int, int) {
        var first, last int
        for _, week := range *weeks {
            cond := week.A + week.C + week.D
            if cond != 0 {
                if first == 0 {
                    first = week.W
                } else if last == 0 {
                    last = week.W
                    return first, last
                }
            }
        }
        return first, last
}


func fetch_contributors(g *CGraph, repo graph.Node) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/stats/contributors", repo)
    log.Println(fmt.Sprintf("fetch url: %s", url))
    resp, err := http.Get(url)
    if err != nil {
        log.Fatal(err)
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }
    var result []Contributor
    err = json.Unmarshal(body, &result)
    if err != nil {
        log.Fatal(err)
        return
    }
    for _, res := range result {
        total := res.Total
        first, last := extract_contribution_interval(&res.Weeks)
        user := NewUser(res.Author["login"].(string), res.Author["url"].(string))
        if g.Node(user.ID()) == nil {
            g.AddNode(user)
        }
        contr := NewContribution(user, repo, total, first, last)
        log.Println(repo)
        g.SetEdge(contr)
    }
    defer resp.Body.Close()
}

func main() {
    g := NewCGraph()
    log.Println("Listing for request at 8001")
    for _, repo :=  range repos{
        g.AddNode(repo)
        fetch_contributors(g, repo)
    }
    //log.Println(fmt.Sprintf("%s", p_edges_user_repo))
    helloHandler := func(w http.ResponseWriter, req *http.Request) {
        dot_graph, err := dot.Marshal(g, "blub", "prefix", "   ")
        if err != nil {
            log.Fatal(err)
        }
        graph_json, err := json.Marshal(g)
        if err != nil {
            log.Fatal(err)
        }
        io.WriteString(w, fmt.Sprintf("Blockkraken! %s\n %s", dot_graph, graph_json))
    }
    http.HandleFunc("/blockkraken", helloHandler)
    log.Fatal(http.ListenAndServe(":8001", nil))
}
