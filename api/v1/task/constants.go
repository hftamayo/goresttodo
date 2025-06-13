package task

const (
	errTaskNotFoundFmt = "task with id %d not found"
	errTaskReference = "task:%d"
	errInvalidateCacheFmt = "Failed to invalidate cache: %v\n"
	taskServiceListRef = "tasks:list"
	headerLastModified = "Last-Modified"
	headerCacheControl = "Cache-Control"
	taskCacheKey = "task_%d"
	taskPageCacheName = "task_page_*"
	eTagCharacterFmt = "W/\"%x\""
)