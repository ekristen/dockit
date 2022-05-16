package db

import (
	"time"

	"gorm.io/gorm"
)

type PKI struct {
	ID        int64 `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Type      string
	Bits      int
	Private   string
	Public    string
	X509      string
	NotBefore *time.Time
	ExpiresAt *time.Time
	Active    bool
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func (p *PKI) AfterCreate(tx *gorm.DB) (err error) {
	if p.Active {
		tx.Model(&PKI{}).Where("id != ?", p.ID).Update("active", false)
	}
	return
}

func (p *PKI) AfterUpdate(tx *gorm.DB) (err error) {
	if p.Active {
		tx.Model(&PKI{}).Where("id != ?", p.ID).Update("active", false)
	}
	return
}
