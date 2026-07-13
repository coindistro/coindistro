package rbac

import (
	"fmt"
	"strings"
)

// Role represents a user role with associated permissions.
type Role string

const (
	RoleSuperAdmin        Role = "super_admin"
	RoleAdmin             Role = "admin"
	RoleComplianceOfficer Role = "compliance_officer"
	RoleSupport           Role = "support"
	RoleMerchant          Role = "merchant"
	RoleInstructor        Role = "instructor"
	RoleTrader            Role = "trader"
	RoleStudent           Role = "student"
	RoleUser              Role = "user"
)

// Permission represents a granular access permission.
type Permission string

const (
	// User permissions
	PermUsersRead   Permission = "users.read"
	PermUsersWrite  Permission = "users.write"
	PermUsersDelete Permission = "users.delete"

	// Academy permissions
	PermAcademyCreate Permission = "academy.create"
	PermAcademyUpdate Permission = "academy.update"
	PermAcademyDelete Permission = "academy.delete"
	PermAcademyRead   Permission = "academy.read"

	// Signals permissions
	PermSignalsPublish Permission = "signals.publish"
	PermSignalsDelete  Permission = "signals.delete"
	PermSignalsRead    Permission = "signals.read"

	// Merchant permissions
	PermMerchantApprove Permission = "merchant.approve"
	PermMerchantRead    Permission = "merchant.read"
	PermMerchantWrite   Permission = "merchant.write"

	// Wallet permissions
	PermWalletRead   Permission = "wallet.read"
	PermWalletWrite  Permission = "wallet.write"
	PermWalletFreeze Permission = "wallet.freeze"

	// Trading bot permissions
	PermBotsManage Permission = "bots.manage"
	PermBotsRead   Permission = "bots.read"

	// Reports permissions
	PermReportsView Permission = "reports.view"

	// Admin permissions
	PermAdminAccess       Permission = "admin.access"
	PermAdminSettings     Permission = "admin.settings"
	PermAdminUsers        Permission = "admin.users"
	PermAdminAudit        Permission = "admin.audit"
	PermAdminFeatureFlags Permission = "admin.feature_flags"

	// KYC permissions
	PermKYCAccess  Permission = "kyc.access"
	PermKYCApprove Permission = "kyc.approve"

	// Payment permissions
	PermPaymentsRead    Permission = "payments.read"
	PermPaymentsWrite   Permission = "payments.write"
	PermPaymentsApprove Permission = "payments.approve"

	// Notification permissions
	PermNotificationsSend Permission = "notifications.send"
)

// RoleDefinition defines a role and its permissions.
type RoleDefinition struct {
	Role        Role
	Permissions []Permission
	Inherits    []Role
}

// RBAC manages role-based access control.
type RBAC struct {
	roles map[Role]*RoleDefinition
}

// New creates a new RBAC instance with default roles.
func New() *RBAC {
	rbac := &RBAC{
		roles: make(map[Role]*RoleDefinition),
	}
	rbac.registerDefaultRoles()
	return rbac
}

// RegisterRole registers a custom role definition.
func (r *RBAC) RegisterRole(def RoleDefinition) {
	r.roles[def.Role] = &def
}

// HasPermission checks if a role has a specific permission, considering inheritance.
func (r *RBAC) HasPermission(role Role, permission Permission) bool {
	def, ok := r.roles[role]
	if !ok {
		return false
	}

	// Check direct permissions
	for _, p := range def.Permissions {
		if p == permission {
			return true
		}
	}

	// Check inherited roles
	for _, inherited := range def.Inherits {
		if r.HasPermission(inherited, permission) {
			return true
		}
	}

	return false
}

// HasAnyPermission checks if a role has any of the given permissions.
func (r *RBAC) HasAnyPermission(role Role, permissions ...Permission) bool {
	for _, p := range permissions {
		if r.HasPermission(role, p) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a role has all of the given permissions.
func (r *RBAC) HasAllPermissions(role Role, permissions ...Permission) bool {
	for _, p := range permissions {
		if !r.HasPermission(role, p) {
			return false
		}
	}
	return true
}

// GetPermissions returns all permissions for a role (including inherited).
func (r *RBAC) GetPermissions(role Role) []Permission {
	permSet := make(map[Permission]bool)
	r.collectPermissions(role, permSet)

	perms := make([]Permission, 0, len(permSet))
	for p := range permSet {
		perms = append(perms, p)
	}
	return perms
}

// GetRoles returns all registered roles.
func (r *RBAC) GetRoles() []Role {
	roles := make([]Role, 0, len(r.roles))
	for role := range r.roles {
		roles = append(roles, role)
	}
	return roles
}

// HasRole checks if a role exists.
func (r *RBAC) HasRole(role Role) bool {
	_, ok := r.roles[role]
	return ok
}

// ValidateRoles checks if all given roles are valid.
func (r *RBAC) ValidateRoles(roles []Role) error {
	for _, role := range roles {
		if !r.HasRole(role) {
			return fmt.Errorf("invalid role: %s", role)
		}
	}
	return nil
}

// ParseRole parses a role string into a Role.
func ParseRole(s string) (Role, error) {
	role := Role(strings.ToLower(strings.TrimSpace(s)))
	switch role {
	case RoleSuperAdmin, RoleAdmin, RoleComplianceOfficer, RoleSupport,
		RoleMerchant, RoleInstructor, RoleTrader, RoleStudent, RoleUser:
		return role, nil
	default:
		return "", fmt.Errorf("unknown role: %s", s)
	}
}

// ParsePermission parses a permission string into a Permission.
func ParsePermission(s string) (Permission, error) {
	perm := Permission(strings.ToLower(strings.TrimSpace(s)))
	// Validate against known permissions
	knownPermissions := map[Permission]bool{
		PermUsersRead: true, PermUsersWrite: true, PermUsersDelete: true,
		PermAcademyCreate: true, PermAcademyUpdate: true, PermAcademyDelete: true, PermAcademyRead: true,
		PermSignalsPublish: true, PermSignalsDelete: true, PermSignalsRead: true,
		PermMerchantApprove: true, PermMerchantRead: true, PermMerchantWrite: true,
		PermWalletRead: true, PermWalletWrite: true, PermWalletFreeze: true,
		PermBotsManage: true, PermBotsRead: true,
		PermReportsView: true,
		PermAdminAccess: true, PermAdminSettings: true, PermAdminUsers: true, PermAdminAudit: true, PermAdminFeatureFlags: true,
		PermKYCAccess: true, PermKYCApprove: true,
		PermPaymentsRead: true, PermPaymentsWrite: true, PermPaymentsApprove: true,
		PermNotificationsSend: true,
	}
	if !knownPermissions[perm] {
		return "", fmt.Errorf("unknown permission: %s", s)
	}
	return perm, nil
}

func (r *RBAC) collectPermissions(role Role, permSet map[Permission]bool) {
	def, ok := r.roles[role]
	if !ok {
		return
	}

	for _, p := range def.Permissions {
		permSet[p] = true
	}

	for _, inherited := range def.Inherits {
		r.collectPermissions(inherited, permSet)
	}
}

func (r *RBAC) registerDefaultRoles() {
	// Super Admin - full access
	r.roles[RoleSuperAdmin] = &RoleDefinition{
		Role: RoleSuperAdmin,
		Permissions: []Permission{
			PermAdminAccess, PermAdminSettings, PermAdminUsers, PermAdminAudit, PermAdminFeatureFlags,
			PermUsersRead, PermUsersWrite, PermUsersDelete,
			PermAcademyCreate, PermAcademyUpdate, PermAcademyDelete, PermAcademyRead,
			PermSignalsPublish, PermSignalsDelete, PermSignalsRead,
			PermMerchantApprove, PermMerchantRead, PermMerchantWrite,
			PermWalletRead, PermWalletWrite, PermWalletFreeze,
			PermBotsManage, PermBotsRead,
			PermReportsView,
			PermKYCAccess, PermKYCApprove,
			PermPaymentsRead, PermPaymentsWrite, PermPaymentsApprove,
			PermNotificationsSend,
		},
	}

	// Admin - most access except super admin specific
	r.roles[RoleAdmin] = &RoleDefinition{
		Role: RoleAdmin,
		Permissions: []Permission{
			PermAdminAccess, PermAdminSettings, PermAdminUsers, PermAdminAudit,
			PermUsersRead, PermUsersWrite,
			PermAcademyCreate, PermAcademyUpdate, PermAcademyRead,
			PermSignalsPublish, PermSignalsRead,
			PermMerchantApprove, PermMerchantRead,
			PermWalletRead, PermWalletFreeze,
			PermBotsManage, PermBotsRead,
			PermReportsView,
			PermKYCAccess, PermKYCApprove,
			PermPaymentsRead, PermPaymentsApprove,
			PermNotificationsSend,
		},
	}

	// Compliance Officer
	r.roles[RoleComplianceOfficer] = &RoleDefinition{
		Role: RoleComplianceOfficer,
		Permissions: []Permission{
			PermKYCAccess, PermKYCApprove,
			PermUsersRead,
			PermMerchantRead, PermMerchantApprove,
			PermPaymentsRead, PermPaymentsApprove,
			PermReportsView,
			PermAdminAudit,
		},
	}

	// Support
	r.roles[RoleSupport] = &RoleDefinition{
		Role: RoleSupport,
		Permissions: []Permission{
			PermUsersRead,
			PermMerchantRead,
			PermWalletRead,
			PermPaymentsRead,
			PermKYCAccess,
		},
	}

	// Merchant
	r.roles[RoleMerchant] = &RoleDefinition{
		Role: RoleMerchant,
		Permissions: []Permission{
			PermMerchantRead, PermMerchantWrite,
			PermPaymentsRead,
			PermWalletRead,
		},
	}

	// Instructor
	r.roles[RoleInstructor] = &RoleDefinition{
		Role: RoleInstructor,
		Permissions: []Permission{
			PermAcademyCreate, PermAcademyUpdate, PermAcademyRead,
		},
	}

	// Trader
	r.roles[RoleTrader] = &RoleDefinition{
		Role: RoleTrader,
		Permissions: []Permission{
			PermSignalsRead,
			PermBotsRead, PermBotsManage,
			PermWalletRead, PermWalletWrite,
		},
	}

	// Student
	r.roles[RoleStudent] = &RoleDefinition{
		Role: RoleStudent,
		Permissions: []Permission{
			PermAcademyRead,
		},
	}

	// Default User
	r.roles[RoleUser] = &RoleDefinition{
		Role: RoleUser,
		Permissions: []Permission{
			PermUsersRead, PermUsersWrite,
			PermWalletRead, PermWalletWrite,
			PermAcademyRead,
			PermSignalsRead,
			PermBotsRead,
		},
	}
}
