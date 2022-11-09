package experiment

import "strings"

type node struct {
	value string
	next  map[string]*node
}

var rootNode = &node{
	value: "/",
}

func insert(path string) {
	routeBlock := strings.Split(path, "/")

	p := rootNode
	for _, v := range routeBlock {
		//未找到
		if _, ok := p.next[v]; !ok {
			//所有:lang\:day等形式都视为:，作为动态区域的标记
			if strings.HasPrefix(v, ":") {
				v = ":"
			}
			//判断是否为动态区域
			p.next[v] = &node{
				value: v,
			}
		}
		p = p.next[v]
	}
}

func match(path string) *node {
	routeBlock := strings.Split(path, "/")

	p := rootNode
	for _, part := range routeBlock {
		//未找到
		if nextNode, ok := p.next[part]; !ok {
			if nextNode.value == ":" {
				part = ":"
			}
			return nil
		}
		p = p.next[part]
	}
	return p
}
