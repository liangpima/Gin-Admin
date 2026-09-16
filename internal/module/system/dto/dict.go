package dto

// CreateDictTypeRequest 新增字典类型
type CreateDictTypeRequest struct {
	Name string `json:"name" binding:"required"`
	// Type 字典类型编码，业务代码里按它取值（如 useDict('sys_user_status')），
	// 因此只允许小写字母、数字、下划线，格式校验在 Service 层。
	Type string `json:"type" binding:"required"`
}

// UpdateDictTypeRequest 修改字典类型。
//
// 刻意不含 Type：它是字典数据的外键键名（sys_dict_data.dict_type 存的就是它），
// 允许改会让已有字典数据全部变成孤儿，界面上表现为「数据还在但一个都取不到」。
// 需要换编码时正确做法是新建类型再迁数据。
type UpdateDictTypeRequest struct {
	Name   string `json:"name" binding:"required"`
	Status *int8  `json:"status"`
	Remark string `json:"remark"`
}

// CreateDictDataRequest 新增字典数据
type CreateDictDataRequest struct {
	DictType  string `json:"dictType" binding:"required"`
	Label     string `json:"label" binding:"required"`
	Value     string `json:"value" binding:"required"`
	Sort      int    `json:"sort"`
	CssClass  string `json:"cssClass"`
	ListClass string `json:"listClass"`
	Remark    string `json:"remark"`
}

// UpdateDictDataRequest 修改字典数据。
//
// 不含 DictType：归属类型不允许改。改归属相当于把选项从一组搬到另一组，
// 原类型下会少一个选项、新类型下会多一个，语义上应「删除后新建」而不是修改。
type UpdateDictDataRequest struct {
	Label     string `json:"label" binding:"required"`
	Value     string `json:"value" binding:"required"`
	Sort      int    `json:"sort"`
	CssClass  string `json:"cssClass"`
	ListClass string `json:"listClass"`
	Status    *int8  `json:"status"`
	Remark    string `json:"remark"`
}
