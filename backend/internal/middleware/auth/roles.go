package auth

import "github.com/ilham/presensi-online/backend/internal/domain"

// RoleMahasiswaOrDosen returns roles allowed to submit attendance.
func RoleMahasiswaOrDosen() []domain.UserRole {
	return []domain.UserRole{domain.RoleMahasiswa, domain.RoleDosen}
}

// RoleAdminAndAbove returns roles with administrative privileges.
func RoleAdminAndAbove() []domain.UserRole {
	return []domain.UserRole{domain.RoleAdminProdi, domain.RolePimpinan, domain.RoleSuperadmin}
}

// AllRoles returns all valid roles.
func AllRoles() []domain.UserRole {
	return []domain.UserRole{
		domain.RoleMahasiswa,
		domain.RoleDosen,
		domain.RoleAdminProdi,
		domain.RolePimpinan,
		domain.RoleSuperadmin,
	}
}
