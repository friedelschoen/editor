package lrparser

// passes rules (ruleindex) to vertices data
type VerticesData struct {
	verts   []*Vertex
	rFirst  *RuleFirstT
	reverse bool
}

type Vertex struct {
	id       VertexId
	rdslasK  RuleDotsLaSet    // kernels
	rdslasC  RuleDotsLaSet    // closure
	gotoVert map[Rule]*Vertex // goto vertex
}

type VertexId int
