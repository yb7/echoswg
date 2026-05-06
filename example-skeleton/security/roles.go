package security

type Role string

const (
    RoleSystemAdmin Role = "system:admin"
    RoleAnonymous  Role = "user:anonymous"
)
