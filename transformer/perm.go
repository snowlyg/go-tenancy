package transformer

type MenuTable struct {
	Id          int64  `json:"id"`
	OrderNumber int64  `json:"order_number"`
	Method      string `json:"method"`
	Checked     int8   `json:"checked"`
	IsMenu      int8   `json:"is_menu"`
	Title       string `json:"title"`
	Href        string `json:"href"`
	Icon        string `json:"icon"`
	Target      string `json:"target"`
	ParentId    int64  `json:"parent_id"`
	CreatedAt   string `json:"created_at"`
}

type PermSelect struct {
	List      []List `json:"list"`
	CheckedId []uint `json:"checkedId"`
}

type List struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
	Pid  int64  `json:"pid"`
}
