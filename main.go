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
    Id int64
    Edge []int64
}

type Repo struct {
    CNode
    Owner string
    Name string
}

func (r *Repo) String() string {
    return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

type User struct {
    CNode
    Login string
    Url string
}

func NewUser(login string, url string) *User {
    sum := md5.Sum([]byte(u.Login))
    num, _ := binary.Varint(sum[:])
    edges := new([]int64)
    return &User{
        num
        edges
        login
        url
    }
}

func NewRepo(owner string, name string) *User {
    repo := new(Repo)
    repo.Owner = Owner
    repo.Name = name
    repo.Edges = new([]int64)
    sum := md5.Sum([]byte(repo.String()))
    num, _ := binary.Varint(sum[:])
    repo.Id = num
    return repo
}
func (u *User) ID() int64 {
    return u.Id
}

func (r *Repo) ID() int64 {
    return r.Id
}

func (u *User) String() string {
    return fmt.Sprintf("%s", u.Login)
}

// ==== NEW GRAPH ==== 
type CGraph struct {
    Nodes map[int64]graph.Node
}

type CEdges struct {
    from graph.Node
    to graph.Node
}

type NodeIds []int64
type NodeList []Node

func (nodes NodeList) Len() int {
    return len(nodes)
}

func (nodes NodeList) Next() int64 {
    n := 0
    return func() Node {
        n += 1
        return nodes[n-1]
    }
}

func (g *CGraph) Edge(uid, vid int64) graph.Edge {
    // u ->  v node 
    u_node := g.CNodes[uid]
    neighbor_nodes := g.CEdges[uid]
    for _, node_id := range neighbor_nodes{
        if node_id == vid {
            v_node := g.CNodes[vid]
            return CEdge{
               u_node,
               v_node,
            }
        }
    }
    return nil
}

func NewCGraph() *CGraph{
    nodes := make(map[int64]Node)
    edges := make(map[int64]NodeIds)
    return &CGraph{
        nodes,
        edges,
    }
}

func (g *CGraph) Node(id int64) graph.Node {
    return g.CNodes[id]
}


func (g *CGraph) Nodes(id int64) graph.Nodes {
    var nodes []Node
    for _, v := range g.CNodes {
       nodes = append(nodes, v)
    }
    return nodes
}

func (g *CGraph) From(id int64) graph.Nodes {
   return g.CEdges[id]
}

func (g *CGraph) HasEdgeBetween(xid, yid int64) bool {
    neighbor_nodes := g.CEdges[xid]
    for _, k :=  range neighbor_nodes {
        if k == yid {
            return true
        }
    }
    return false
}


// ==== END GRAPH ====




var repos = []Repo{
    Repo{"ethereum","go-ethereum"},
    Repo{"smartcontractkit", "chainlink"},
    Repo{"blockchainsllc", "DAO"},
    Repo{"paritytech", "polkadot"},
    Repo{"cosmos", "cosmos"},
    Repo{"neo-project", "neo"},
    Repo{"icon-project", "loopchain"},
    //Repo{"icon-project", "loopchain"},
}

// ==== EDGES ====
type Contribution struct {
    Repo Repo
    Total int
    First int
    Last int
}

var edges_repo_user = make(map[string][]*User)
var edges_user_repo = make(map[string][]*Repo)
var edges_user_contribution = make(map[string][]*Contribution)

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


func fetch_contributors(repo *Repo) {
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
        user := User{res.Author["login"].(string), res.Author["url"].(string)}
        contr := Contribution{*repo, total, first, last}
        edges_repo_user[repo.String()] = append(edges_repo_user[repo.String()], &user)
        edges_user_repo[(&user).String()] = append(edges_user_repo[(&user).String()], repo)
        edges_user_contribution[(&user).String()] = append(edges_user_contribution[(&user).String()], &contr)
    }
    defer resp.Body.Close()
}

func main() {
    log.Println("Listing for request at 8001")
    for _, repo :=  range repos{
        fetch_contributors(&repo)
    }
    for k, v := range edges_user_contribution {
        if len(v) >= 2 {
            p, err := json.MarshalIndent(v, "", " ")
            if err != nil {
                log.Fatal(err)
            }
            log.Println(fmt.Sprintf("%s : %s", k, p))
       }
    }
    p_edges_user_repo, err := json.MarshalIndent(edges_user_contribution, "", " ")
    if err != nil {
        log.Fatal(err)
    }
    //log.Println(fmt.Sprintf("%s", p_edges_user_repo))
    helloHandler := func(w http.ResponseWriter, req *http.Request) {
        io.WriteString(w, fmt.Sprintf("Blockkraken! %s\n", p_edges_user_repo))
    }
    http.HandleFunc("/blockkraken", helloHandler)
    log.Fatal(http.ListenAndServe(":8001", nil))
}
