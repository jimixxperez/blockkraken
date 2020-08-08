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

func fetch_contributors(repo string) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/stats/contributors", repo)
    resp, err := http.Get(url)
    if err != nil {
        log.Println("An error occured!")
    }
    body, err := ioutil.ReadAll(resp.Body)
    //log.Println(fmt.Sprintf("%s",body))
    var result []Contributor
    err = json.Unmarshal(body, &result)
    if err != nil {
        log.Fatal(err)
        return
    }
    login := (result[0].Author)["login"].(string)
    log.Println(fmt.Sprintf("%s", login))
    defer resp.Body.Close()
}

func main() {
    helloHandler := func(w http.ResponseWriter, req *http.Request) {
        io.WriteString(w, "Hello, world!\n")
    }
    http.HandleFunc("/hello", helloHandler)
    log.Println("Listing for request at 8001")
    fetch_contributors("ethereum/go-ethereum")
    log.Fatal(http.ListenAndServe(":8001", nil))
}
