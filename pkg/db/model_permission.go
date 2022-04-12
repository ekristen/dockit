package db

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/ekristen/dockit/pkg/common"
	"gorm.io/gorm"
)

type PermissionType string

const (
	Registry   PermissionType = "registry"
	Catalog    PermissionType = "catalog"
	Namespace  PermissionType = "namespace"
	Repository PermissionType = "repository"
)

type PermissionAction string

const (
	Pull  PermissionAction = "pull"
	Push  PermissionAction = "push"
	Admin PermissionAction = "admin"
)

type Permission struct {
	ID        int64            `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Type      PermissionType   `gorm:"index:idx_permissions_unique,unique" json:"type"`
	Class     string           `gorm:"index:idx_permissions_unique,unique" json:"class,omitempty"`
	Name      string           `gorm:"index:idx_permissions_unique,unique" json:"name"`
	Action    PermissionAction `json:"action"`
	EntityID  int64            `json:"-"`
	User      *User            `gorm:"foreignKey:EntityID" json:"user,omitempty"`
	Group     *Group           `gorm:"foreignKey:EntityID" json:"group,omitempty"`
	CreatedAt *time.Time       `json:"created_at"`
	UpdatedAt *time.Time       `json:"updated_at"`
}

func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == 0 {
		node := tx.Statement.Context.Value(common.ContextKeyNode).(*snowflake.Node)
		p.ID = node.Generate().Int64()
	}

	return nil
}
