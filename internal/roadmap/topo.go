package roadmap

import (
	"fmt"
	"sort"
)

func (g *Graph) TopologicalSort() ([]*Problem, error) {
	inDegree := make(map[int]int, len(g.Problems))
	for _, p := range g.Problems {
		if _, ok := inDegree[p.ID]; !ok {
			inDegree[p.ID] = 0
		}
		for _, prereq := range p.Prerequisites {
			if _, ok := g.Problems[prereq]; !ok {
				return nil, fmt.Errorf("problem %d has unknown prerequisite %d", p.ID, prereq)
			}
		}
	}

	dependents := make(map[int][]int, len(g.Problems))
	for _, p := range g.Problems {
		for _, prereq := range p.Prerequisites {
			inDegree[p.ID]++
			dependents[prereq] = append(dependents[prereq], p.ID)
		}
	}

	var queue []int
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	sort.Ints(queue)

	var sorted []*Problem
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		sorted = append(sorted, g.Problems[id])

		for _, dep := range dependents[id] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
		sort.Ints(queue)
	}

	if len(sorted) != len(g.Problems) {
		return nil, fmt.Errorf("cycle detected in roadmap: sorted %d of %d problems", len(sorted), len(g.Problems))
	}

	return sorted, nil
}
