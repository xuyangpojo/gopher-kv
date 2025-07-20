package data

import (
	"math"
	"container/list"
)

// GkvGraph 图结构
// @author xuyang
// @datetime 2025-7-16 21:00
type GkvGraph struct {
	// 节点坐标信息：节点名 -> (x, y)
	nodes map[string][2]float64
	// 邻接表：节点名 -> 邻居节点集合
	edges map[string]map[string]struct{}
}

// DataGkvGraph 全局图实例
var DataGkvGraph = &GkvGraph{
	nodes: make(map[string][2]float64),
	edges: make(map[string]map[string]struct{}),
}

// AddNode 添加节点及其坐标
// @param name string 节点名
// @param x, y float64 坐标
func (g *GkvGraph) AddNode(name string, x, y float64) {
	g.nodes[name] = [2]float64{x, y}
	if _, exists := g.edges[name]; !exists {
		g.edges[name] = make(map[string]struct{})
	}
}

// AddEdge 添加无向边
// @param from, to string 节点名
func (g *GkvGraph) AddEdge(from, to string) {
	if _, exists := g.edges[from]; !exists {
		g.edges[from] = make(map[string]struct{})
	}
	if _, exists := g.edges[to]; !exists {
		g.edges[to] = make(map[string]struct{})
	}
	g.edges[from][to] = struct{}{}
	g.edges[to][from] = struct{}{}
}

// EuclideanDistance 计算两节点的欧氏距离
// @param a, b string 节点名
// @return float64 距离, bool 是否存在
func (g *GkvGraph) EuclideanDistance(a, b string) (float64, bool) {
	na, oka := g.nodes[a]
	nb, okb := g.nodes[b]
	if !oka || !okb {
		return 0, false
	}
	dx := na[0] - nb[0]
	dy := na[1] - nb[1]
	return math.Sqrt(dx*dx + dy*dy), true
}

// DFS 深度优先搜索，返回遍历顺序
// @param start string 起点
// @return []string 遍历顺序
func (g *GkvGraph) DFS(start string) []string {
	visited := make(map[string]bool)
	result := []string{}
	var dfs func(string)
	dfs = func(node string) {
		if visited[node] {
			return
		}
		visited[node] = true
		result = append(result, node)
		for neighbor := range g.edges[node] {
			dfs(neighbor)
		}
	}
	dfs(start)
	return result
}

// BFS 广度优先搜索，返回遍历顺序
// @param start string 起点
// @return []string 遍历顺序
func (g *GkvGraph) BFS(start string) []string {
	visited := make(map[string]bool)
	result := []string{}
	q := list.New()
	q.PushBack(start)
	visited[start] = true
	for q.Len() > 0 {
		e := q.Front()
		node := e.Value.(string)
		q.Remove(e)
		result = append(result, node)
		for neighbor := range g.edges[node] {
			if !visited[neighbor] {
				visited[neighbor] = true
				q.PushBack(neighbor)
			}
		}
	}
	return result
}
