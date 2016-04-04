package tree

import (
	"strings"

	"github.com/wlMalk/mirango/framework"
)

type nodeType uint8

const (
	static nodeType = iota // default
	root
	param
	catchAll
)

type Param struct {
	Key   string
	Value string
}

type Params []*Param

type Node struct {
	parent           *Node
	route            framework.Route
	children         []*Node
	names            []string
	containsWildCard bool
}

func New(r framework.Route) *Node {
	n := &Node{}
	n.route = r
	return n
}

func (n *Node) GetFullPath() string {
	if n.parent == nil {
		return n.route.Path()
	}
	return n.parent.GetFullPath() + n.route.Path()
}

func (n *Node) GetRoot() framework.Route {
	if n.parent != nil {
		return n.parent.GetRoot()
	}
	return n.route
}

func (n *Node) GetRoute() framework.Route {
	return n.route
}

func (n *Node) GetParent() *Node {
	return n.parent
}

func (n *Node) SetRoute(r framework.Route) {
	n.route = r
}

func (n *Node) ProcessPath() {
	//path := n.route.Path()
	n.names = nil
	slices := strings.Split(n.route.Path()[1:], "/")

	if len(slices) == 0 {
		panic("path is empty")
	}

	// check that every var name has length more than 0

	for i, s := range slices {
		param := strings.LastIndex(s, ":")
		wildcardParam := strings.LastIndex(s, "*")
		if param > wildcardParam {
			n.names = append(n.names, s[param+1:])
		} else if param < wildcardParam && i == len(slices)-1 {
			n.names = append(n.names, s[wildcardParam+1:])
			n.containsWildCard = true
			n.children = nil
		} else if param == -1 && wildcardParam == -1 {
			continue
		}
	}

	// check all paths
}

func (n *Node) Branch(r framework.Route) *Node {
	return n.BranchWith(r, nil)
}

func (n *Node) BranchWith(r framework.Route, nn *Node) *Node {
	if r == nil {
		panic("route is nil")
	}

	if n.containsWildCard {
		panic("wildcard routes can not have sub-routes")
	}

	var child *Node

	if nn != nil {
		child = nn
	} else {
		child = &Node{}
	}

	child.parent = n
	child.route = r

	child.ProcessPath()

	// check path

	n.children = append(n.children, child)

	return child
}

func (n *Node) GetNotFoundHandler() interface{} {
	return nil
}

func (n *Node) Match(path string) (r framework.Route, p Params) {
	node := n

	params := 0

	if n.parent == nil {
		nPath := n.route.Path()
		if !strings.HasPrefix(path, nPath) {
			r = nil
			p = nil
			return
		}
		path = path[len(nPath):]
	}

	slices := strings.Split(path[1:], "/")
look:
	for _, c := range node.children {
		cPath := c.route.Path()
		cSlices := strings.Split(cPath[1:], "/")
	walk:
		for j := range cSlices {
			param := strings.LastIndex(cSlices[j], ":")
			wildcardParam := strings.LastIndex(cSlices[j], "*")

			if param == -1 && wildcardParam == -1 {
				if cSlices[j] != slices[j] {
					if j == len(slices)-1 && j == len(cSlices)-1 {
						n = nil
						p = nil
						return
					}
					continue look
				} else {
					if j == len(slices)-1 && j == len(cSlices)-1 {
						r = c.route
						return
					} else if j == len(cSlices)-1 {
						node = c
						slices = slices[j+1:]
						params = 0
						goto look
					} else if j == len(slices)-1 {
						n = nil
						p = nil
						return
					}
					continue walk
				}
			} else if (param > wildcardParam && (len(slices[j]) <= param || slices[j][:param] != cSlices[j][:param])) ||
				((param < wildcardParam) && (!c.containsWildCard || len(slices[j]) <= wildcardParam || slices[j][:wildcardParam] != cSlices[j][:wildcardParam])) {
				continue look
			} else if param > wildcardParam {
				p = append(p, &Param{c.names[params], slices[j][param:]})
				params++
				if j == len(slices)-1 && j == len(cSlices)-1 {
					r = c.route
					return
				} else if j == len(cSlices)-1 {
					node = c
					slices = slices[j+1:]
					params = 0
					goto look
				} else if j == len(slices)-1 {
					n = nil
					p = nil
					return
				}
				continue walk
			} else if param < wildcardParam {

				p = append(p, &Param{c.names[params], strings.TrimSuffix(slices[j][wildcardParam:]+"/"+strings.Join(slices[j+1:], "/"), "/")})
				r = c.route
				return
			}
		}
	}
	return
}
