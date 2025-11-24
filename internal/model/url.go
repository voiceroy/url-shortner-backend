package model

// URL Model
type URL struct {
	URL        string `json:"url" binding:"required"`
	Days       int    `json:"days" binding:"gt=0,lte=7"`
	CustomCode string `json:"custom_code" binding:"omitempty"`
}

// Code Model
type Code struct {
	Code string `uri:"code" binding:"required"`
}
