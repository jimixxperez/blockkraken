package main

import (
    _ "io"
    "io/ioutil"
    "log"
    "net/http"
    "fmt"
    _ "encoding/binary"
    "encoding/json"
    "html/template"
//    "github.com/google/jsonapi"
    _ "crypto/md5"
    "gonum.org/v1/gonum/graph"
    cgraph "github.com/jimixxperez/blockkraken/internal"
    _ "gonum.org/v1/gonum/graph/encoding/dot"
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


// ==== END GRAPH ====




//var repos = new([]Repo)
var repos = []graph.Node{
    cgraph.NewRepo("ethereum", "go-ethereum"),
    cgraph.NewRepo("smartcontractkit", "chainlink"),
    cgraph.NewRepo("blockchainsllc", "DAO"),
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
//
// 

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


func fetch_contributors(g *cgraph.CGraph, repo graph.Node) {
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
        user := cgraph.NewUser(res.Author["login"].(string), res.Author["url"].(string))
        if g.Node(user.ID()) == nil {
            g.AddNode(user)
        }
        contr := cgraph.NewContribution(user, repo, total, first, last)
        log.Println(repo)
        g.SetEdge(contr)
    }
    defer resp.Body.Close()
}

func main() {
    tmpl := template.Must(template.ParseFiles("templates/layout.html"))
    g := cgraph.NewCGraph()
    log.Println("Listing for request at 8001")
    for _, repo :=  range repos{
        g.AddNode(repo)
        fetch_contributors(g, repo)
    }
    //log.Println(fmt.Sprintf("%s", p_edges_user_repo))
    helloHandler := func(w http.ResponseWriter, req *http.Request) {
        data := map[string]interface{}{"PageTitle": "Blockkraken!!!"}
        tmpl.Execute(w, data)
    }
    http.HandleFunc("/blockkraken", helloHandler)
    http.HandleFunc("/graph", func(w http.ResponseWriter, req *http.Request) {
        gr, err := json.Marshal(*g) 
        if err != nil {
            log.Fatal(err) 
        }
        w.Header().Set("Content-Type", "application/json")
        w.Write(gr)
    })
    log.Fatal(http.ListenAndServe(":8001", nil))
}
