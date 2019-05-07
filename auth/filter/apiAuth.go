package filter

import (
	"fmt"
	"strconv"
)

type ApiAuthService interface {
	GetClientById(id int) (*Client, error)
	GetClientByUser(userId, roleType string) ([]*UserClient, error)
	UpdateClient(fullname, redirectUri string) (*ClientInfo, error)

	GetAllResources() ([]*ApiResource, error)
	GetUserResources(userId string) ([]*ApiResource, error)
	AddResource(resources []ResourceInfo) ([]int, error)
	UpdateResource(rId int, rName, rDescription, rData string) (*ApiResource, error)
	DeleteResources(resourceIds []int) (*DeleteResInfo, error)

	GetRoleTree(relatedResource, relatedUser bool) ([]*RoleTree, error)
	GetUserRoleTree(userId string, relatedResource, relatedUser bool) ([]*UserRoleTree, error)
	GetAllRole(relatedResource, relatedUser bool) ([]*Role, error)
	GetUserRoles(userId string, isAll, relatedResource, relatedUser bool) ([]*UserRole, error)
	AddRole(name, description string, parentId int) (int, error)
	UpdateRole(roleId int, name, description string, parentId int) (*Role, error)
	DeleteRole(roleId int) (*DeleteRoleInfo, error)

	GetUsersOfRole(roleId int) ([]*RoleUser, error)
	AddUserToRole(roleId int, infos []UserInfo) (int, error)
	UpdateUserOfRole(roleId int, info UserInfo) (*RoleUser, error)
	DeleteUserFromRole(roleId int, names []string) (int, error)

	GetAllRelatedInfo() ([]*RelatedInfo, error)
	GetRelatedInfo(roleId int) ([]*RelatedInfo, error)
	AddRelations(roleId int, resIds []int) (int, error)
	UpdateRelations(roleId int, resIds []int) (int, error)
	DeleteRelations(roleId int, resIds []int) (int, error)
}

type ApiAuth struct {
	ClientId     int64  // client id
	ClientSecret string // client secret
	RedirectUri  string // 回调uri
	ApiHost      string // host
}

type ApiConfig struct {
	ClientId     string
	ClientSecret string
	RedirectUri  string
	ApiHost      string
}

func NewApiAuth(config *ApiConfig) ApiAuthService {
	if config == nil {
		panic("sso service init failed: config counld not be nil")
	}

	var clientIdFormat int64

	if number, err := strconv.ParseInt(config.ClientId, 10, 64); err != nil {
		panic(fmt.Sprintf("sso service init failed: appId is invalid %s", config.ClientId))
	} else {
		clientIdFormat = number
	}

	apiAuth := &ApiAuth{
		ClientId:     clientIdFormat,
		ClientSecret: config.ClientSecret,
		RedirectUri:  config.RedirectUri,
		ApiHost:      config.ApiHost,
	}
	return apiAuth
}