package entities

import "time"

// Tenant represents a tenant in the system
type Tenant struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	About          string    `json:"about"`
	Address        string    `json:"address"`
	PhoneNumber    string    `json:"phone_number"`
	Logo           string    `json:"logo"`
	TermsOfService string    `json:"terms_of_service"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TenantRepository defines the interface for tenant data operations
type TenantRepository interface {
	Create(tenant *Tenant) error
	GetByID(id uint) (*Tenant, error)
	List() ([]*Tenant, error)
	Update(tenant *Tenant) error
	Delete(id uint) error
}

// TenantUseCase defines the interface for tenant business logic
type TenantUseCase interface {
	Create(tenant *Tenant) error
	GetByID(id uint) (*Tenant, error)
	List() ([]*Tenant, error)
	Update(tenant *Tenant) error
	Delete(id uint) error
}
