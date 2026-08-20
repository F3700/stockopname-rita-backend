package model

type Department struct {
	DepartmentID   int    `json:"department_id"`
	DepartmentCode string `json:"department_code"`
	DepartmentName string `json:"department_name"`
	DepartmentDesc string `json:"department_desc"`
}
