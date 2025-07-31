package query

func NormalizePagination(offset, limit, page, pageSize int) (normalizedPage int, normalizedPageSize int) {

	if offset >= 0 && limit > 0 {

		normalizedPage = (offset / limit) + 1
		normalizedPageSize = limit
		return normalizedPage, normalizedPageSize
	}

	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 50
	}

	return page, pageSize
}

func CalculateOffsetLimit(page, pageSize int) (offset, limit int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}

	offset = (page - 1) * pageSize
	limit = pageSize
	return offset, limit
}

func ValidatePaginationBounds(page, pageSize, maxPageSize int) (validatedPage, validatedPageSize int) {
	if maxPageSize <= 0 {
		maxPageSize = 200
	}

	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	return page, pageSize
}
