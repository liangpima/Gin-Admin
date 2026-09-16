package common

// BuildTree 把扁平列表按「父指针」组装成树，时间复杂度 O(n)。
//
// 为什么需要它：菜单与部门原本各自实现了一份「递归扫描整个切片找子节点」的
// 树构建，每个节点都要重新遍历一遍全表，整体是 O(n²)。
// 这两张表当前只有几百行，平方级开销可以忽略；但它是随数据量放大的，
// 而每次菜单/部门列表请求都会走一遍 —— 没有理由留着。
//
// 关于「环」：如果数据里出现互相引用的父子关系，递归会无限展开。
// 这里不额外做访问标记，而是依赖调用方从根（rootID）出发这一点：
// 环上的节点其祖先链必定不包含 rootID（否则它就不在环上），
// 因此从根出发永远走不到环里，递归自然不会进入。
// 换句话说，**环只会让那批节点不可达（从界面上消失），不会导致爆栈**。
// 该性质由 TestBuildTreeIgnoresCycle 与 TestBuildTreeSelfReference 固定住。
//
// 类型参数 T 是节点类型（如 model.SysMenu），三个访问器负责解耦字段名差异，
// 避免为菜单和部门各写一份几乎相同的代码。
func BuildTree[T any](
	nodes []T,
	rootID uint,
	idOf func(T) uint,
	parentOf func(T) uint,
	setChildren func(*T, []T),
) []T {
	byParent := make(map[uint][]T, len(nodes))
	for _, n := range nodes {
		pid := parentOf(n)
		byParent[pid] = append(byParent[pid], n)
	}

	var build func(pid uint) []T
	build = func(pid uint) []T {
		children := byParent[pid]
		// 与原实现保持一致：返回空切片而非 nil，
		// 前端 JSON 序列化后是 []，不需要额外判空。
		out := make([]T, 0, len(children))
		for _, c := range children {
			node := c // 先复制再挂子节点：直接改 c 会写进 byParent 的元素，
			// 同一节点被多处引用时会互相串味
			setChildren(&node, build(idOf(node)))
			out = append(out, node)
		}
		return out
	}

	return build(rootID)
}
