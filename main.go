package main

import (
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "fmt"
    "encoding/json"
//    "github.com/google/jsonapi"
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

type Repo struct {
    Owner string
    Name string
}

func (r *Repo) String() string {
    return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

type User struct {
    Login string
    Url string
}

func (u *User) String() string {
    return fmt.Sprintf("%s", u.Login)
}

var repos = []Repo{
    Repo{"ethereum","go-ethereum"},
    Repo{"ethereum","viper"},
    Repo{"smartcontractkit", "chainlink"},
    Repo{"blockchainsllc", "DAO"},
}

// ==== EDGES ====
var edges_repo_user = make(map[*Repo][]*User)
var edges_user_repo = make(map[*User][]*Repo)


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
        user := User{res.Author["login"].(string), res.Author["url"].(string)}
        edges_repo_user[repo] = append(edges_repo_user[repo], &user)
        edges_user_repo[&user] = append(edges_user_repo[&user], repo)
    }
    defer resp.Body.Close()
}

func main() {
    helloHandler := func(w http.ResponseWriter, req *http.Request) {
        io.WriteString(w, "Blockkraken!\n")
    }
    http.HandleFunc("/blockkraken", helloHandler)
    log.Println("Listing for request at 8001")
    for _, repo :=  range repos{
        fetch_contributors(&repo)
    }
    // for k, v := range edges_user_repo {
    //     log.Println(fmt.Sprintf("%i", len(v)))
    //     if len(v) >= 2 {
    //         log.Println(fmt.Sprintf("%s", k))
    //     }
    // }
    //p_edges_user_repo, err := json.MarshalIndent(edges_user_repo, "", " ")
    //if err != nil {
    //    log.Fatal(err)
    //}
    log.Println(fmt.Sprintf("%s", pedges_user_repo))
    log.Fatal(http.ListenAndServe(":8001", nil))
}
