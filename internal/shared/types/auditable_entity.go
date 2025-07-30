package types

// AuditableEntity extends BaseEntity with audit trail fields.
// This is optional and can be embedded instead of BaseEntity when audit functionality is needed.
type AuditableEntity struct {
	BaseEntity        // Embedded base entity with all common fields
	CreatedBy  int    `json:"createdBy"` // ID of the user who created the entity
	UpdatedBy  int    `json:"updatedBy"` // ID of the user who last updated the entity
	AuditNote  string `json:"auditNote"` // Optional note for audit trail
}

// GetCreatedBy returns the ID of the user who created the entity
func (a *AuditableEntity) GetCreatedBy() int {
	return a.CreatedBy
}

// GetUpdatedBy returns the ID of the user who last updated the entity
func (a *AuditableEntity) GetUpdatedBy() int {
	return a.UpdatedBy
}

// GetAuditNote returns the audit note
func (a *AuditableEntity) GetAuditNote() string {
	return a.AuditNote
}

// SetCreatedBy sets the ID of the user who created the entity
func (a *AuditableEntity) SetCreatedBy(userID int) {
	a.CreatedBy = userID
}

// SetUpdatedBy sets the ID of the user who last updated the entity
func (a *AuditableEntity) SetUpdatedBy(userID int) {
	a.UpdatedBy = userID
}

// SetAuditNote sets the audit note
func (a *AuditableEntity) SetAuditNote(note string) {
	a.AuditNote = note
}
