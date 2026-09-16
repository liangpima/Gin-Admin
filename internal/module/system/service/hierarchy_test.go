package service

import "testing"

// makeParentFn 用 map 构造一个 getParent 实现，模拟层级表的父指针
func makeParentFn(parentOf map[uint]uint) func(uint) (uint, bool, error) {
	return func(id uint) (uint, bool, error) {
		p, ok := parentOf[id]
		return p, ok, nil
	}
}

// TestHasCycleInHierarchy 覆盖成环判定的各类情形。
//
// 背景：菜单/部门更新时若把节点挂到它自己或它的后代之下会形成环，
// 该子树将无法从根节点遍历到（buildMenuTree 从 parentID=0 出发），
// 于是从界面消失却仍留在库里，成为不可见也不可删的孤儿数据。
func TestHasCycleInHierarchy(t *testing.T) {
	// 典型层级：1(根) ← 2 ← 3 ← 4 ← 5
	parentOf := map[uint]uint{1: 0, 2: 1, 3: 2, 4: 3, 5: 4}
	getParent := makeParentFn(parentOf)

	cases := []struct {
		name        string
		id          uint
		newParentID uint
		wantCycle   bool
	}{
		{"移到根级不成环", 3, 0, false},
		{"自己当自己的父节点成环", 3, 3, true},
		{"挂到直接子节点之下成环", 3, 4, true},
		{"挂到更深的后代之下也成环", 2, 5, true},
		{"挂到无关联的上级不成环", 5, 1, false},
		{"挂到同级兄弟之下不成环", 4, 2, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := hasCycleInHierarchy(c.id, c.newParentID, getParent)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.wantCycle {
				t.Errorf("hasCycleInHierarchy(%d, %d) = %v, want %v",
					c.id, c.newParentID, got, c.wantCycle)
			}
		})
	}
}

// TestHasCycleInHierarchyEdgeCases 边界：父节点不存在、自引用脏数据、已有环
func TestHasCycleInHierarchyEdgeCases(t *testing.T) {
	t.Run("父节点不存在不成环", func(t *testing.T) {
		getParent := makeParentFn(map[uint]uint{1: 0})
		got, err := hasCycleInHierarchy(1, 999, getParent)
		if err != nil || got {
			t.Errorf("父节点不存在时应返回 (false, nil)，实际 (%v, %v)", got, err)
		}
	})

	t.Run("父链中遇到自引用脏数据不成环", func(t *testing.T) {
		// 2 的父节点是自己（脏数据），上溯应立即停止而不是死循环
		getParent := makeParentFn(map[uint]uint{1: 0, 2: 2})
		got, err := hasCycleInHierarchy(1, 2, getParent)
		if err != nil || got {
			t.Errorf("应安全终止并返回 false，实际 (%v, %v)", got, err)
		}
	})

	t.Run("数据中已存在的长环保守拒绝", func(t *testing.T) {
		// 构造一个长度超过 maxHierarchyDepth 的链，上溯无法到达根 → 保守判定成环
		parentOf := make(map[uint]uint, maxHierarchyDepth+2)
		for i := uint(1); i <= maxHierarchyDepth+1; i++ {
			parentOf[i] = i + 1 // 一直指向下一个，永不到 0
		}
		getParent := makeParentFn(parentOf)

		got, err := hasCycleInHierarchy(99999, 1, getParent)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Error("超过最大深度仍未到根时应保守判定为成环")
		}
	})

	t.Run("查询报错时向上传递", func(t *testing.T) {
		getParent := func(uint) (uint, bool, error) {
			return 0, false, errBoom
		}
		if _, err := hasCycleInHierarchy(1, 2, getParent); err != errBoom {
			t.Errorf("应透传底层错误，实际 %v", err)
		}
	})
}

var errBoom = &boomError{}

type boomError struct{}

func (*boomError) Error() string { return "boom" }
