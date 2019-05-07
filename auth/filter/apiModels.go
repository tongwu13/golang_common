package filter

import (
	"encoding/json"
	"errors"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	SUCC    = 0
	FAILED  = 1
	UNKNOWN = 2
)

type RespBody struct {
	ResCode int         `json:"res_code"` // auth接口返回状态码，SUCC-0-成功  FAILED-1-失败  UNKNOWN-2-其他
	ResMsg  string      `json:"res_msg"`  // auth接口返回状态描述
	Data    interface{} `json:"data"`     // auth接口返回数据
}

// 通过ClientId所查询到的Client
type Client struct {
	Id          int    `json:"id"`           // clientId
	Fullname    string `json:"fullname"`     // client全名
	Secret      string `json:"secret"`       // ClientSecret
	RedirectUri string `json:"redirect_uri"` // 重定向uri
	UserId      string `json:"user_id"`      // 用户Id
	Created     string `json:"created_at"`   // 创建时间
	Updated     string `json:"updated_at"`   // 更新时间
}

// 通过ClientId查询Client
func (a *ApiAuth) GetClientById(id int) (*Client, error) {
	url := a.ApiHost + "/api/client"
	resp := httplib.Get(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	var client Client
	if err = json.Unmarshal(data, &client); err != nil {
		return nil, err
	}
	logs.Info(client)
	return &client, nil
}

// 用户所在的Client
type UserClient struct {
	Id       int     `json:"id"`       // clientId
	Fullname string  `json:"fullname"` // client全名
	Roles    []*Role `json:"roles"`    // 用户在client下的角色
}

// 查询某用户在指定类型角色下所在的Client
func (a *ApiAuth) GetClientByUser(userId, roleType string) ([]*UserClient, error) {
	url := a.ApiHost + "/api/userClients?user_id=" + userId + "&role_type=" + roleType
	resp := httplib.Get(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	var userClients []*UserClient
	if err = json.Unmarshal(data, &userClients); err != nil {
		return nil, err
	}
	return userClients, nil
}

// 更新Client后，接口返回的新Client信息
type ClientInfo struct {
	Id          int    `json:"id"`           // clientId
	Fullname    string `json:"fullname"`     // client全名
	RedirectUri string `json:"redirect_uri"` // 重定向uri
}

// 更新Client
func (a *ApiAuth) UpdateClient(fullname, redirectUri string) (*ClientInfo, error) {
	url := a.ApiHost + "/api/client"
	resp := httplib.Put(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	body := make(map[string]interface{})
	body["fullname"] = fullname
	body["redirect_uri"] = redirectUri
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	response, err := resp.Body(b).Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	clientInfo := ClientInfo{}
	if err = json.Unmarshal(data, &clientInfo); err != nil {
		return nil, err
	}
	return &clientInfo, nil
}

// 接口获取的完整resource
type ApiResource struct {
	Id          int    `json:"id"`          // 资源id
	Name        string `json:"name"`        // 资源名
	Description string `json:"description"` // 资源描述
	ClientId    int    `json:"client_id"`   // ClientId
	Data        string `json:"data"`        // 资源内容
	CreatedBy   string `json:"created_by"`  // 创建者
	UpdatedBy   string `json:"updated_by"`  // 更新者
	Created     string `json:"created"`     // 创建时间
	Updated     string `json:"updated"`     // 更新时间
}

// 查看Client下全部资源
func (a *ApiAuth) GetAllResources() ([]*ApiResource, error) {
	url := a.ApiHost + "/api/resources"
	resp := httplib.Get(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Response()
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	var resources []*ApiResource
	if err = json.Unmarshal(data, &resources); err != nil {
		return nil, err
	}
	return resources, nil
}

// 查看用户在Client下的全部资源
func (a *ApiAuth) GetUserResources(userId string) ([]*ApiResource, error) {
	url := a.ApiHost + "/api/resources?user_id=" + userId
	resp := httplib.Get(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Response()
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	var resources []*ApiResource
	if err = json.Unmarshal(data, &resources); err != nil {
		return nil, err
	}
	return resources, nil
}

// 添加或修改Resource后，接口返回的新Resource信息
type ResourceInfo struct {
	Name        string `json:"name"`        // 资源名称
	Description string `json:"description"` // 资源描述
	Data        string `json:"data"`        // 资源内容
}

// 批量新增资源
func (a *ApiAuth) AddResource(resources []ResourceInfo) ([]int, error) {
	url := a.ApiHost + "/api/resources"
	resp := httplib.Post(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	b, err := json.Marshal(resources)
	if err != nil {
		return nil, err
	}
	response, err := resp.Body(b).Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	var ids []int
	if err = json.Unmarshal(data, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// 修改单个资源内容
func (a *ApiAuth) UpdateResource(rId int, rName, rDescription, rData string) (*ApiResource, error) {
	url := a.ApiHost + "/api/resources"
	var body = make(map[string]interface{})
	body["id"] = rId
	body["name"] = rName
	body["description"] = rDescription
	body["data"] = rData
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp := httplib.Put(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Body(b).Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	var resource ApiResource
	if err = json.Unmarshal(data, &resource); err != nil {
		return nil, err
	}
	return &resource, nil
}

// 删除Resource后，接口返回的删除信息
type DeleteResInfo struct {
	DelResNum     int `json:"del_resource_num"`
	DelRoleResNum int `json:"del_role_resource_num"`
}

// 批量删除资源
func (a *ApiAuth) DeleteResources(resourceIds []int) (*DeleteResInfo, error) {
	url := a.ApiHost + "/api/resources/"
	for idx, id := range resourceIds {
		if idx != 0 {
			url += ","
		}
		url += strconv.Itoa(id)
	}
	resp := httplib.Delete(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	delInfo := DeleteResInfo{}
	if err = json.Unmarshal(data, &delInfo); err != nil {
		return nil, err
	}
	return &delInfo, nil
}

// 角色
type Role struct {
	Id          int            `json:"id"`          // 角色id
	Name        string         `json:"name"`        // 角色名
	Description string         `json:"description"` // 角色类型
	ParentId    int            `json:"parent_id"`   // 父角色id
	CreatedBy   string         `json:"created_by"`  // 创建者
	UpdatedBy   string         `json:"updated_by"`  // 更新者
	Created     string         `json:"created"`     // 创建时间
	Updated     string         `json:"updated"`     // 更新时间
	Resources   []*ApiResource `json:"resources"`   // 角色相关资源
	Users       []*UserOfRole  `json:"users"`       // 角色相关用户
}

// 用户角色
type UserRole struct {
	Id          int           `json:"id"`          // 角色id
	Name        string        `json:"name"`        // 角色名
	Description string        `json:"description"` // 角色类型
	ParentId    int           `json:"parent_id"`   // 父角色id
	CreatedBy   string        `json:"created_by"`  // 创建者
	UpdatedBy   string        `json:"updated_by"`  // 更新者
	Created     string        `json:"created"`     // 创建时间
	Updated     string        `json:"updated"`     // 更新时间
	RoleType    string        `json:"role_type"`   // 角色类型
	Resources   []*Resource   `json:"resources"`   // 角色相关资源
	Users       []*UserOfRole `json:"users"`       // 角色相关用户
}

// 角色树
type RoleTree struct {
	Id          int           `json:"id"`          // 角色id
	Name        string        `json:"name"`        // 角色名
	Description string        `json:"description"` // 角色类型
	ParentId    int           `json:"parent_id"`   // 父角色id
	Created     string        `json:"created"`     // 创建时间
	Updated     string        `json:"updated"`     // 更新时间
	Resources   []*Resource   `json:"resources"`   // 角色相关资源
	Users       []*UserOfRole `json:"users"`       // 角色相关用户
	Children    []*RoleTree   `json:"children"`    // 拥有该角色的用户
}

// 用户角色树
type UserRoleTree struct {
	Id          int             `json:"id"`          // 角色id
	Name        string          `json:"name"`        // 角色名
	Description string          `json:"description"` // 角色类型
	ParentId    int             `json:"parent_id"`   // 父角色id
	Created     string          `json:"created"`     // 创建时间
	Updated     string          `json:"updated"`     // 更新时间
	RoleType    string          `json:"role_type"`   // 角色类型
	Resources   []*ApiResource  `json:"resources"`   // 角色相关资源
	Users       []*UserOfRole   `json:"users"`       // 角色相关用户
	Children    []*UserRoleTree `json:"children"`    // 拥有该角色的用户
}

// 角色相关用户
type UserOfRole struct {
	Id       string `json:"id"`         // 用户id
	Fullname string `json:"fullname"`   // 用户全名
	Dn       string `json:"dn"`         // dn
	Created  string `json:"created_at"` // 创建时间
	Updated  string `json:"updated_at"` // 更新时间
}

// 查询角色树
func (a *ApiAuth) GetRoleTree(relatedResource, relatedUser bool) ([]*RoleTree, error) {
	url := a.ApiHost + "/api/roles?is_tree=true&relate_user=" + strconv.FormatBool(relatedUser) + "relate_resource=" + strconv.FormatBool(relatedResource)
	resp := httplib.Get(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	var roleTree []*RoleTree
	if err = json.Unmarshal(data, &roleTree); err != nil {
		return nil, err
	}
	return roleTree, err
}

// 查询指定用户的角色树
func (a *ApiAuth) GetUserRoleTree(userId string, relatedResource, relatedUser bool) ([]*UserRoleTree, error) {
	url := a.ApiHost + "/api/userRoles?is_tree=true&is_all=true&user_id=" + userId + "&relate_user=" + strconv.FormatBool(relatedUser) + "relate_resource=" + strconv.FormatBool(relatedResource)
	resp := httplib.Get(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	var userRoleTree []*UserRoleTree
	if err = json.Unmarshal(data, &userRoleTree); err != nil {
		return nil, err
	}
	return userRoleTree, err
}

// 查询全部角色
func (a *ApiAuth) GetAllRole(relatedResource, relatedUser bool) ([]*Role, error) {
	url := a.ApiHost + "/api/roles?is_tree=true&relate_user=" + strconv.FormatBool(relatedUser) + "relate_resource=" + strconv.FormatBool(relatedResource)
	resp := httplib.Get(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	var roles []*Role
	if err = json.Unmarshal(data, &roles); err != nil {
		return nil, err
	}
	return roles, err
}

// 查询指定用户角色（直接关联的或全部）
func (a *ApiAuth) GetUserRoles(userId string, isAll, relatedResource, relatedUser bool) ([]*UserRole, error) {
	url := a.ApiHost + "/api/userRoles?is_all=" + strconv.FormatBool(isAll) + "&user_id=" + userId + "&relate_user=" + strconv.FormatBool(relatedUser) + "relate_resource=" + strconv.FormatBool(relatedResource)
	resp := httplib.Get(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	var userRoles []*UserRole
	if err = json.Unmarshal(data, &userRoles); err != nil {
		return nil, err
	}
	return userRoles, err
}

// 新增子角色
func (a *ApiAuth) AddRole(name string, description string, parentId int) (int, error) {
	url := a.ApiHost + "/api/roles"
	var body = make(map[string]interface{})
	body["name"] = name
	body["description"] = description
	body["parent_id"] = parentId
	b, err := json.Marshal(body)
	if err != nil {
		return -1, err
	}
	resp := httplib.Post(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Body(b).Response()
	if err != nil {
		return -1, err
	}
	data, err := processResp(response)
	if err != nil {
		return -1, err
	}
	var id int
	if err = json.Unmarshal(data, &id); err != nil {
		return -1, err
	}
	return id, nil
}

// 修改角色信息
func (a *ApiAuth) UpdateRole(roleId int, name, description string, parentId int) (*Role, error) {
	url := a.ApiHost + "/api/roles"
	var body = make(map[string]interface{})
	body["id"] = roleId
	body["name"] = name
	body["description"] = description
	body["parent_id"] = parentId
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp := httplib.Put(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Body(b).Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	newRole := Role{}
	if err = json.Unmarshal(data, &newRole); err != nil {
		return nil, err
	}
	return &newRole, nil
}

// 删除角色后，接口返回的删除信息
type DeleteRoleInfo struct {
	DelRoleNum         int `json:"del_role_num"`          // 删除的角色数量
	DelRoleResourceNum int `json:"del_role_resource_num"` // 删除的相关资源数量
	DelRoleUserNum     int `json:"del_role_user_num"`     // 删除的相关用户数量
}

// 删除单个角色
func (a *ApiAuth) DeleteRole(roleId int) (*DeleteRoleInfo, error) {
	url := a.ApiHost + "/api/roles/" + strconv.Itoa(roleId)
	resp := httplib.Delete(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	deleteRole := DeleteRoleInfo{}
	if err = json.Unmarshal(data, &deleteRole); err != nil {
		return nil, err
	}
	return &deleteRole, err
}

// 角色中的用户
type RoleUser struct {
	RoleId   int    `json:"role_id"`   // 角色Id
	UserId   string `json:"user_id"`   // 用户Id
	RoleType string `json:"role_type"` // 角色类型
}

// 用户基本信息
type UserInfo struct {
	UserId   string `json:"user_id"`   // 用户Id
	RoleType string `json:"role_type"` // 角色类型
}

// 查询角色中的用户
func (a *ApiAuth) GetUsersOfRole(roleId int) ([]*RoleUser, error) {
	url := a.ApiHost + "/api/roleUsers?role_id=" + strconv.Itoa(roleId)
	resp := httplib.Get(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	var users []*RoleUser
	if err = json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, err
}

// 向某角色内批量添加用户，返回添加数量
func (a *ApiAuth) AddUserToRole(roleId int, infos []UserInfo) (int, error) {
	url := a.ApiHost + "/api/roleUsers/" + strconv.Itoa(roleId)
	resp := httplib.Post(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	b, err := json.Marshal(infos)
	if err != nil {
		return -1, err
	}
	response, err := resp.Body(b).Response()
	if err != nil {
		return -1, err
	}
	data, err := processResp(response)
	if err != nil {
		return -1, err
	}
	var num int
	if err = json.Unmarshal(data, &num); err != nil {
		return -1, err
	}
	return num, err
}

// 修改单个用户信息，返回修改后的用户
func (a *ApiAuth) UpdateUserOfRole(roleId int, info UserInfo) (*RoleUser, error) {
	url := a.ApiHost + "/api/roleUsers/" + strconv.Itoa(roleId)
	resp := httplib.Put(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	b, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}
	response, err := resp.Body(b).Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	user := RoleUser{}
	if err = json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, err
}

// 批量删除某角色内用户，返回删除人数
func (a *ApiAuth) DeleteUserFromRole(roleId int, names []string) (int, error) {
	url := a.ApiHost + "/api/roleUsers/" + strconv.Itoa(roleId)
	resp := httplib.Delete(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	b, err := json.Marshal(names)
	if err != nil {
		return -1, err
	}
	response, err := resp.Body(b).Response()
	if err != nil {
		return -1, nil
	}
	data, err := processResp(response)
	if err != nil {
		return -1, err
	}
	var num int
	if err = json.Unmarshal(data, &num); err != nil {
		return -1, err
	}
	return num, err
}

// 角色资源关联信息
type RelatedInfo struct {
	RoleId     int `json:"role_id"`     // 角色id
	ResourceId int `json:"resource_id"` // 角色关联资源id
}

// 查看Client下全部角色资源关联
func (a *ApiAuth) GetAllRelatedInfo() ([]*RelatedInfo, error) {
	url := a.ApiHost + "/api/roleResources"
	resp := httplib.Get(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	var info []*RelatedInfo
	if err = json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return info, err
}

// 查看指定角色关联的所有资源
func (a *ApiAuth) GetRelatedInfo(roleId int) ([]*RelatedInfo, error) {
	url := a.ApiHost + "/api/roleResources/" + strconv.Itoa(roleId)
	resp := httplib.Get(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Response()
	if err != nil {
		return nil, err
	}
	data, err := processResp(response)
	if err != nil {
		return nil, err
	}
	var info []*RelatedInfo
	if err = json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return info, err
}

// 批量添加某角色和资源关联关系，返回新增关联数目
func (a *ApiAuth) AddRelations(roleId int, resIds []int) (int, error) {
	url := a.ApiHost + "/api/roleResources/" + strconv.Itoa(roleId)
	b, err := json.Marshal(resIds)
	if err != nil {
		return -1, err
	}
	resp := httplib.Post(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Body(b).Response()
	if err != nil {
		return -1, err
	}
	data, err := processResp(response)
	if err != nil {
		return -1, err
	}
	var num int
	if err = json.Unmarshal(data, &num); err != nil {
		return -1, err
	}
	return num, err
}

// 批量修改某角色和资源关联关系，返回当前全部关联数目
func (a *ApiAuth) UpdateRelations(roleId int, resIds []int) (int, error) {
	url := a.ApiHost + "/api/roleResources/" + strconv.Itoa(roleId)
	b, err := json.Marshal(resIds)
	if err != nil {
		return -1, err
	}
	resp := httplib.Put(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Body(b).Response()
	if err != nil {
		return -1, err
	}
	data, err := processResp(response)
	if err != nil {
		return -1, err
	}
	var num int
	if err = json.Unmarshal(data, &num); err != nil {
		return -1, err
	}
	return num, err
}

// 批量删除某角色和资源关联关系，返回删除的关联数目
func (a *ApiAuth) DeleteRelations(roleId int, resIds []int) (int, error) {
	url := a.ApiHost + "/api/roleResources/" + strconv.Itoa(roleId)
	b, err := json.Marshal(resIds)
	if err != nil {
		return -1, err
	}
	resp := httplib.Delete(url)
	resp.Header("client-secret", a.ClientSecret)
	resp.Header("client-id", strconv.FormatInt(a.ClientId, 10))
	response, err := resp.Body(b).Response()
	if err != nil {
		return -1, err
	}
	data, err := processResp(response)
	if err != nil {
		return -1, err
	}
	var num int
	if err = json.Unmarshal(data, &num); err != nil {
		return -1, err
	}
	return num, err
}

// 统一对接口返回结果进行处理，将有效数据部分序列化后返回
func processResp(response *http.Response) (data []byte, err error) {
	logs.Info(response.StatusCode)
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status code of " + strconv.Itoa(response.StatusCode))
	}
	rawBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var body RespBody
	if err = json.Unmarshal(rawBody, &body); err != nil {
		return nil, err
	}
	if body.ResCode != SUCC && body.Data == nil {
		return nil, errors.New(body.ResMsg)
	}
	if data, err = json.Marshal(body.Data); err != nil {
		return nil, err
	}
	return
}
