package chord

// NOTE:
func CreateNNodes(n int) ([]*Node, error) {
	if n == 0 {
		return nil, nil
	}
	nodes := make([]*Node, n)

	id := make([]byte, KEY_LENGTH)
	id[0] = byte(0)
	curr, err := CreateDefinedNode(nil, id)
	nodes[0] = curr
	if err != nil {
		return nil, err
	}

	for i := 1; i < n; i++ {
		id := make([]byte, KEY_LENGTH)
		id[0] = byte(i * 10)
		curr, err := CreateDefinedNode(nodes[0].RemoteSelf, id)
		nodes[i] = curr
		if err != nil {
			return nil, err
		}
	}

	return nodes, nil
}
