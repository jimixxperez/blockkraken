package internal

import(
	"fmt"	
	"encoding/binary"
	"crypto/md5"
	"gonum.org/v1/gonum/graph"
)

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
