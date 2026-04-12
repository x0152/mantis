package mappers

import (
	"encoding/json"

	"mantis/core/types"
	"mantis/infrastructure/models"
)

func PlanToRow(p types.Plan) models.PlanRow {
	graph, _ := json.Marshal(p.Graph)
	return models.PlanRow{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Schedule:    p.Schedule,
		Enabled:     p.Enabled,
		Graph:       graph,
	}
}

func PlanFromRow(r models.PlanRow) types.Plan {
	var graph types.PlanGraph
	_ = json.Unmarshal(r.Graph, &graph)
	if graph.Nodes == nil {
		graph.Nodes = []types.PlanNode{}
	}
	if graph.Edges == nil {
		graph.Edges = []types.PlanEdge{}
	}
	return types.Plan{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Schedule:    r.Schedule,
		Enabled:     r.Enabled,
		Graph:       graph,
	}
}
