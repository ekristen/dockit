package db

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/ekristen/dockit/pkg/common"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Group --
type Group struct {
	ID          int64         `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Name        string        `json:"name"`
	Active      bool          `json:"active"`
	CreatedAt   *time.Time    `json:"created_at"`
	UpdatedAt   *time.Time    `json:"updated_at"`
	Users       []*User       `gorm:"many2many:user_groups" json:"users,omitempty"`
	Permissions []*Permission `gorm:"foreignKey:EntityID" json:"permissions,omitempty"`
}

// BeforeCreate --
func (g *Group) BeforeCreate(tx *gorm.DB) error {
	if g.ID == 0 {
		node := tx.Statement.Context.Value(common.ContextKeyNode).(*snowflake.Node)
		g.ID = node.Generate().Int64()
	}

	return nil
}

// User --
type User struct {
	ID          int64         `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Name        string        `json:"name"`
	Username    string        `gorm:"uniqueIndex,size:255"`
	Password    string        `json:"-"`
	Admin       bool          `json:"admin"`
	Active      bool          `json:"active"`
	CreatedAt   *time.Time    `json:"created_at"`
	UpdatedAt   *time.Time    `json:"updated_at"`
	Groups      []*Group      `gorm:"many2many:user_groups" json:"groups,omitempty"`
	Permissions []*Permission `gorm:"foreignKey:EntityID" json:"permissions,omitempty"`
	Tokens      []*Token      `gorm:"foreignKey:UserID" json:"users,omitempty"`
}

// BeforeCreate --
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == 0 {
		node := tx.Statement.Context.Value(common.ContextKeyNode).(*snowflake.Node)
		u.ID = node.Generate().Int64()
	}

	return u.bcryptPassword(tx)
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("Password") {
		return u.bcryptPassword(tx)
	}

	return nil
}

func (u *User) bcryptPassword(tx *gorm.DB) error {
	var newPass string
	switch u := tx.Statement.Dest.(type) {
	case map[string]interface{}:
		newPass = u["password"].(string)
	case *User:
		newPass = u.Password
	case []*User:
		newPass = u[tx.Statement.CurDestIndex].Password
	}

	b, err := bcrypt.GenerateFromPassword([]byte(newPass), 10)
	if err != nil {
		return err
	}
	tx.Statement.SetColumn("password", b)

	return nil
}
