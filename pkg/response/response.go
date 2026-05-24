package response

// Standart Response Struct
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors"`
	Data    interface{} `json:"data"`
}

// PaginatedData represents paginated response data
type PaginatedData struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// Func to Build a Successfull Response
func BuildSuccessResponse(message string, data interface{}) Response {
	return Response{
		Success: true,
		Message: message,
		Errors:  nil,
		Data:    data,
	}
}

// Func to Build a Failed Response
func BuildFailedResponse(message string, errors interface{}) Response {
	return Response{
		Success: false,
		Message: message,
		Errors:  errors,
		Data:    nil,
	}
}

// BuildPaginatedResponse builds a standardized paginated response
func BuildPaginatedResponse(message string, items interface{}, total int64, page, limit int) Response {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return Response{
		Success: true,
		Message: message,
		Errors:  nil,
		Data: PaginatedData{
			Items:      items,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}
}
