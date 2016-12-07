package mirango

import (
	"sort"
	"strings"
)

type node struct {
	route  *Route
	parent *node

	order int

	index       int
	hasWildcard bool

	nodes nodes

	text      string
	lowerText string
	param     string

	paramsCount   int
	caseSensitive bool
}

func (n *node) String() string {
	str := ""
	for _, cn := range n.nodes {
		str += "\n" + cn.String()
	}
	str = strings.Replace(str, "\n", "\n\t", -1)
	return n.text + str
}

func newNode(text string) *node {
	return &node{
		text:      text,
		lowerText: strings.ToLower(text),
		index:     -1,
	}
}

type nodes []*node

func (n nodes) Len() int {
	return len(n)
}

func (n nodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n nodes) Less(i, j int) bool {
	if n[j].index == -1 {
		return true
	} else if n[i].caseSensitive && !n[j].caseSensitive {
		return true
	} else if n[i].text[0] < n[j].text[0] {
		return true
	} else if !n[i].hasWildcard && n[j].hasWildcard {
		return true
	}
	return false
}

type result struct {
	node *node
	path string
}

func (n *node) matches(part string) bool {
	if len(n.text) < len(part) && n.index != -1 {
		if (n.caseSensitive && part[:len(n.text)] == n.text) || (!n.caseSensitive && strings.ToLower(part[:len(n.text)]) == n.lowerText) {
			return true
		}
	} else if (n.caseSensitive && part == n.text) || (!n.caseSensitive && strings.ToLower(part) == n.lowerText) {
		return true
	}

	return false
}

func (n *node) match(path string) (res result, found bool, alt *node) {

	var part string

	i := 0

	ps := 0
	cs := 0

	res.path = path

walk:
	for i = cs + 1; i < len(path)+1; i++ {
		if i == len(path) {
			ps = cs
			cs = i
			if i != ps+1 {
				if found && len(n.nodes) == 0 {
					if n.hasWildcard {
						found = true
						res.node = n
						return
					}
					found = false
					res.node = n
					return
				}
				part = path[ps+1 : cs]
			}
			break
		}
		if path[i] == '/' {
			ps = cs
			cs = i
			if i != ps+1 {
				if found && len(n.nodes) == 0 {
					if n.hasWildcard {
						found = true
						res.node = n
						return
					}
					found = false
					res.node = n
					return
				}
				part = path[ps+1 : cs]
				break
			}
		}
	}

	if !found {
		if n.parent == nil && n.text == "" {
			goto find
		}
		if ((len(n.text) < len(part) && n.index != -1) && (!n.caseSensitive && strings.ToLower(part[:len(n.text)]) == n.lowerText) ||
			(n.caseSensitive && part[:len(n.text)] == n.text)) ||
			((!n.caseSensitive && strings.ToLower(part) == n.lowerText) || (n.caseSensitive && part == n.text)) {
			found = true
			res.node = n

			if cs < len(path)-1 {
				goto walk
			}

			if res.node == nil {
				res.node = n
			}
			return
		}
		found = false

		return
	}

find:
	for _, cn := range n.nodes {
		if ((len(cn.text) < len(part) && cn.index != -1) && (!cn.caseSensitive && strings.ToLower(part[:len(cn.text)]) == cn.lowerText) ||
			(cn.caseSensitive && part[:len(cn.text)] == cn.text)) ||
			((!cn.caseSensitive && strings.ToLower(part) == cn.lowerText) || (cn.caseSensitive && part == cn.text)) {
			found = true
			res.node = cn

			if cs < len(path)-1 {
				n = cn
				goto walk
			}
			if res.node == nil {
				res.node = cn
			}
			return
		}

		found = false
	}

	return
}

func (n *node) subtract(on *node) *node {
	var topNode *node
	for n != on {
		topNode = on
		on = on.parent
	}
	return topNode
}

func (n *node) getRoot() *node {
	for n.parent != nil {
		n = n.parent
	}
	return n
}

func (n *node) getRoute() *Route {
	if n.route != nil {
		return n.route
	}
	if n.parent != nil {
		return n.parent.getRoute()
	}
	return nil
}

func (n *node) setParam(param string, index int, wildcard bool) {
	if index < 0 {
		return
	}
	if n.index == -1 {
		n.paramsCount = n.paramsCount + 1
	}
	n.index = index
	n.param = param
	n.hasWildcard = wildcard
}

func (n *node) setOrder() {
	n.order = -1
	node := n
	for node != nil {
		if (node.text != "" || node.index != -1) && !(node.text != "" && node.index != -1) {
			n.order = n.order + 1
		}
		node = node.parent
	}
}

func (n *node) finalize() {
	n.setOrder()
	sort.Sort(n.nodes)
	for _, cn := range n.nodes {
		cn.finalize()
	}
}

func (n *node) getRoutedNode() *node {
	if n.route != nil {
		return n
	}
	nn := n
	for len(n.nodes) > 0 && n.nodes[0].route == nil {
		n = n.nodes[0]
	}
	if n.route != nil {
		return n
	}
	for nn.parent != nil {
		nn = nn.parent
		if nn.route != nil {
			return nn
		}
	}
	return nil
}

func (n *node) getChildRoutes() (routes []*Route) {
	n = n.getRoutedNode()
	if n == nil || n.route == nil {
		return
	}
	for _, cn := range n.nodes {
		for len(cn.nodes) > 0 && cn.nodes[0].route == nil {
			cn = cn.nodes[0]
		}
		if cn.route != nil {
			routes = append(routes, cn.route)
		}
	}
	return
}

func (n *node) addNode(on *node) {

	if n.hasWildcard {
		panic("cannot add a node to a node that has wildcard")
	}

	if on == nil {
		panic("node is nil")
	}

	if on.parent != nil {
		on = on.clone()
	}
	on.parent = nil

	found := false
	for _, cn := range n.nodes {
		if sameNodes(cn, on) {
			found = true
			recursiveCompare(cn, on)
			break
		}
	}
	if !found {
		n.nodes = append(n.nodes, on)
		on.parent = n
		if on.index != -1 {
			on.paramsCount = on.parent.paramsCount + 1
		} else {
			on.paramsCount = on.parent.paramsCount
		}
		if on.route != nil {
			on.route.node = on
		}
	}
}

func (r *result) paramIndex(name string) int {
	node := r.node
	for i := r.node.paramsCount - 1; i >= 0; {
		if node.index != -1 {
			if name == node.param {
				return i
			}
			i--
		}
		node = node.parent
	}
	return -1
}

func (r *result) paramByName(name string) string {
	if name == "" {
		return ""
	}
	_, value := r.paramByIndex(r.paramIndex(name))
	return value
}

func (r *result) paramByIndex(idx int) (string, string) {

	if idx >= r.node.paramsCount || idx < 0 {
		return "", ""
	}

	node := r.node

	steps := node.paramsCount - idx - 1
	if steps < 0 {
		return "", ""
	}

	for i := 0; i < steps; {
		if node.index != -1 {
			i++
		}
		node = node.parent
	}

	ps := 0
	cs := ps
	j := 0
	for i := 0; i <= node.order; i++ {
		for j = cs + 1; j < len(r.path)+1; j++ {
			if j == len(r.path) {
				ps = cs
				cs = j
				break
			}
			if r.path[j] == '/' {
				ps = cs
				cs = j
				if j != ps+1 {
					break
				}
			}

		}
	}

	if !node.hasWildcard {
		return node.param, r.path[ps+1+node.index : cs]
	}
	return node.param, r.path[ps+1+node.index:]
}

func sameNodes(cn *node, on *node) bool {
	return cn.text == on.text && cn.index == on.index && cn.hasWildcard == on.hasWildcard && cn.caseSensitive == on.caseSensitive
}

func recursiveCompare(cn *node, on *node) {
walk:
	for _, con := range on.nodes {
		for _, ccn := range cn.nodes {
			if sameNodes(ccn, con) {
				recursiveCompare(ccn, con)
				continue walk
			}
		}
		cn.nodes = append(cn.nodes, con)
		con.parent = cn
		if con.index != -1 {
			con.paramsCount = con.parent.paramsCount + 1
		} else {
			con.paramsCount = con.parent.paramsCount
		}
		if con.route != nil {
			con.route.node = con
		}
	}
}

func (n *node) clone() *node {
	return nil
}
