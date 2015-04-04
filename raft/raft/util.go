package raft

type UInt64Slice []uint64

func (p UInt64Slice) Len() int {
	return len(p)
}

func (p UInt64Slice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p UInt64Slice) Less(i, j int) bool {
	return p[i] < p[j]
}
