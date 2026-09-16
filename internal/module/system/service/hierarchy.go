package service

// maxHierarchyDepth 层级树的最大上溯步数。
// 菜单/部门的层级在实践中很浅，超过这个值只可能是数据已损坏（存在环），
// 此时保守拒绝比继续遍历更安全。
const maxHierarchyDepth = 200

// hasCycleInHierarchy 判断「把节点 id 挂到 newParentID 之下」是否会形成环。
//
// 为什么需要：菜单与部门更新时直接赋值 ParentID，若把某个节点挂到它自己或它的后代之下，
// 就会形成环。环本身不会导致递归死循环（buildMenuTree 中每个节点只有一个父节点、
// 遍历中最多出现一次），但那棵子树会**从界面上消失** ——
// 既看不见也删不掉，成为库里长期堆积的孤儿数据。
//
// 做法是沿 newParentID 的父链上溯，若途中遇到 id 则说明会成环。
// 只做 O(深度) 次主键查询，不必把整张表读进内存。
// getParent 返回给定节点的父节点 ID，ok=false 表示节点不存在。
func hasCycleInHierarchy(id, newParentID uint, getParent func(uint) (uint, bool, error)) (bool, error) {
	// 移到根级不会成环
	if newParentID == 0 {
		return false, nil
	}
	// 自己当自己的父节点
	if newParentID == id {
		return true, nil
	}

	cur := newParentID
	for i := 0; i < maxHierarchyDepth; i++ {
		if cur == id {
			return true, nil
		}

		parent, ok, err := getParent(cur)
		if err != nil {
			return false, err
		}
		// 已到根、节点不存在，或自引用（自身即为环），都说明继续上溯不会碰到 id
		if !ok || parent == 0 || parent == cur {
			return false, nil
		}
		cur = parent
	}

	// 超出最大深度仍未到根：数据里已存在环，保守拒绝本次变更
	return true, nil
}
