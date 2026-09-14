package dto

type CategoryResponse struct {
	CategoryID          int    `json:"id"`
	CategoryName        string `json:"name"`
	CategoryDescription string `json:"description"`
}

type CategoryCreateRequest struct {
	CategoryName        string `json:"name" validate:"required"`
	CategoryDescription string `json:"description"`
}

type CategoryUpdateRequest struct {
	CategoryName        *string `json:"name"`
	CategoryDescription *string `json:"description"`
}
