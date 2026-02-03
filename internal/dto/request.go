package dto

type CreateHashRequest struct {
	MD5Hash string `json:"md5_hash" binding:"required,len=32"`
}

type ListHashRequest struct {
	Page       int    `form:"page" binding:"min=1"`
	Size       int    `form:"size" binding:"min=1,max=100"`
	SourceFile string `form:"source_file"`
}
