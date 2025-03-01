package lrparser

import (
	"github.com/friedelschoen/glake/util/parseutil/pscan"
)

type PState struct {
	Sc   *pscan.Scanner
	Pos  int
	Node PNode
}

//func (ps *PState) NodeBytes(node PNode) []byte {
//	pos, end := node.Pos(), node.End()
//	if pos > end {
//		pos, end = end, pos
//	}
//	return ps.Sc.SrcFromTo(pos, end)
//}
//func (ps *PState) NodeString(node PNode) string {
//	return string(ps.NodeBytes(node))
//}

type RuneRange = pscan.RuneRange

//type RuneRanges = parseutil.RuneRanges
//type PosError = parseutil.ScError

type PosError struct {
	Err error
	Pos int
	//Callstack string
	Fatal bool
}

// parse node
type PNode interface {
	Pos() int
	End() int
}

//func pnodeSrc2(node PNode, fset *FileSet) string {
//	return string(PNodeBytes(node, fset.Src))
//}

// basic parse node implementation
type BasicPNode struct {
	pos int // can have pos>end when in reverse
	end int
}

func (n *BasicPNode) Pos() int {
	return n.pos
}
func (n *BasicPNode) End() int {
	return n.end
}
func (n *BasicPNode) SetPos(pos, end int) {
	n.pos = pos
	n.end = end
}
func (n *BasicPNode) PosEmpty() bool {
	return n.pos == n.end
}
func (n *BasicPNode) SrcString(src []byte) string {
	return string(src[n.pos:n.end])
}

// content parser node
type CPNode struct {
	BasicPNode
	rule      Rule // can be nil in state0
	childs    []*CPNode
	data      any
	simulated bool
}
