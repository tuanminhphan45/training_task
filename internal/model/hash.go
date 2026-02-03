package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Hash struct {
	bun.BaseModel `bun:"table:hashes"`
	ID            int64     `bun:"id,pk,autoincrement" json:"id"`
	MD5Hash       string    `bun:"md5_hash,notnull,unique" json:"md5_hash"`
	SourceFile    string    `bun:"source_file" json:"source_file"`
	CreatedAt     time.Time `bun:"created_at,default:current_timestamp" json:"created_at"`
}
