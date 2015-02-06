package chord

func CreateNNodes(n int) ([]*Node, error) {
	if n == 0 {
		return nil, nil
	}
	nodes := make([]*Node, n)

	id := []byte{byte(0)}
	curr, err := CreateDefinedNode(nil, id)
	nodes[0] = curr
	if err != nil {
		return nil, err
	}

	for i := 1; i < n; i++ {
		id := []byte{byte(i * 10)}
		curr, err := CreateDefinedNode(nodes[0].RemoteSelf, id)
		nodes[i] = curr
		if err != nil {
			return nil, err
		}
	}

	return nodes, nil
}

func CreateNNodesRandom(n int) ([]*Node, error) {
	if n == 0 {
		return nil, nil
	}
	nodes := make([]*Node, n)

	curr, err := CreateNode(nil)
	nodes[0] = curr
	if err != nil {
		return nil, err
	}

	for i := 1; i < n; i++ {
		curr, err := CreateNode(nodes[0].RemoteSelf)
		nodes[i] = curr
		if err != nil {
			return nil, err
		}
	}

	return nodes, nil
}
