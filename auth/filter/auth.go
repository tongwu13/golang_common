package filter

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/logs"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	SESSION_KEY_USER = "user"
)

type AuthService interface {
	CheckLoginFilter(ctx *context.Context)
	CheckAuthorityFilter(ctx *context.Context, routerPattern string)
	Login(code string, ctx *context.Context)
	RedirectToLogin(ctx *context.Context)
	Logout(ctx *context.Context, state string)
	CurrentUser(ctx *context.Context) User
}

type Auth struct {
	ClientId         int64
	ClientSecret     string
	RedirectUri      string
	Host             string
	AutoLoadResource bool
	Scope            string
	CacheExpire      int64
	UrlControl       map[string][]string
	/*
		UrlControl:
		url - [resource1,resource2] mapping or method:url - [resource1,resource2]
	*/
}

type Config struct {
	ClientId         string
	ClientSecret     string
	RedirectUri      string
	Host             string
	AutoLoadResource string
	Scope            string
	CacheExpire      string
	UrlControl       map[string]string
}

func NewAuthService(config *Config) AuthService {
	if config == nil {
		panic("auth service init failed: config counld not be nil")
	}

	auth := &Auth{}

	if clientId, err := strconv.ParseInt(config.ClientId, 10, 64); err != nil {
		panic(fmt.Sprintf("auth service init failed: clientId is invalid %s", config.ClientId))
	} else {
		auth.ClientId = clientId
	}
	if config.AutoLoadResource == "true" {
		auth.AutoLoadResource = true
	}
	if auth.AutoLoadResource {
		if cacheExpire, err := strconv.ParseInt(config.CacheExpire, 10, 64); err != nil && cacheExpire > 0 {
			panic(fmt.Sprintf("auth service init failed: cacheExpire is invalid %s", config.CacheExpire))
		} else {
			auth.CacheExpire = cacheExpire
		}
	}
	auth.ClientSecret = config.ClientSecret
	auth.RedirectUri = config.RedirectUri
	auth.Host = config.Host
	if config.Scope == "" {
		auth.Scope = "all:all"
	} else {
		auth.Scope = config.Scope
	}
	auth.UrlControl = make(map[string][]string)
	for k, v := range config.UrlControl {
		auth.UrlControl[strings.ToLower(k)] = strings.Split(strings.ToLower(v), "|")
	}
	return auth
}

func (a *Auth) CheckLoginFilter(ctx *context.Context) {
	code := ctx.Input.Query("code")
	if code == "" {
		//没有code，判断session是否有效
		user, ok := ctx.Input.CruSession.Get(SESSION_KEY_USER).(User)
		if !ok {
			a.RedirectToLogin(ctx)
		} else if a.AutoLoadResource {
			if time.Now().Unix()-user.CacheTime > a.CacheExpire {
				if err := user.LoadResource(a); err == nil {
					ctx.Input.CruSession.Set(SESSION_KEY_USER, user)
				} else {
					a.RedirectToLogin(ctx)
				}
			}
		}
	} else {
		//有code，本次请求来自于sso的回调
		if token, err := a.queryTokenFromOauth2(code, ctx); err != nil {
			logs.Error(err)
			a.RedirectToLogin(ctx)
		} else {
			user := User{ResourceMap: make(map[string]*Resource)}
			user.Token = token
			if err := user.Init(a); err != nil {
				logs.Error(err)
			} else if a.AutoLoadResource {
				if err := user.LoadResource(a); err != nil {
					logs.Error(err)
				}
			}
			logs.Info(user)
			ctx.Input.CruSession.Set(SESSION_KEY_USER, user)
			if ctx.Input.Header("x-requested-with") == "XMLHttpRequest" {
				ctx.ResponseWriter.WriteHeader(200)
				ctx.WriteString("登录成功")
			} else {
				if originUrl := ctx.Input.Query("state"); originUrl != "" {
					ctx.Redirect(http.StatusFound, originUrl)
				} else {
					ctx.Redirect(http.StatusFound, "/")
				}
			}
		}
	}
}

/**
authority filter 校验对应url是否有权限
*/
func (a *Auth) CheckAuthorityFilter(ctx *context.Context, routerPattern string) {
	urlPattern := strings.ToLower(routerPattern)
	method := strings.ToLower(ctx.Request.Method)
	key := fmt.Sprintf("%s:%s", method, urlPattern)
	user := a.CurrentUser(ctx)
	isPass := true

	// 先查看该url是否需要权限控制
	if isPass {
		if resources, ok := a.UrlControl[urlPattern]; ok {
			for _, v := range resources {
				if _, ok := user.ResourceMap[v]; !ok {
					isPass = false
					break
				}
			}
		}
	}

	// 接着，查看该url的某种request method是否控制权限，key的格式为method:urlPattern
	if isPass {
		if resources, ok := a.UrlControl[key]; ok {
			for _, v := range resources {
				if _, ok := user.ResourceMap[v]; !ok {
					isPass = false
					break
				}
			}
		}
	}

	if !isPass {
		if ctx.Input.Header("x-requested-with") == "XMLHttpRequest" {
			ctx.ResponseWriter.WriteHeader(403)
			ctx.WriteString("已授权，访问被拒绝，当前用户没有权限访问该内容")
		} else {
			beego.Exception(403, ctx)
		}
	}

}

/**
重定向到登录页或返回未登录状态
*/
func (a *Auth) RedirectToLogin(ctx *context.Context) {
	if ctx.Input.Header("x-requested-with") == "XMLHttpRequest" {
		ctx.ResponseWriter.WriteHeader(401)
		ctx.WriteString("未授权或获取授权失败，访问被拒绝")
	} else {
		params := url.Values{}
		params.Add("client_id", strconv.FormatInt(a.ClientId, 10))
		params.Add("redirect_uri", a.RedirectUri)
		params.Add("response_type", "code")
		params.Add("state", ctx.Input.URI())
		params.Add("scope", a.Scope)
		url := fmt.Sprintf("%s/oauth2/authorize?%s", a.Host, params.Encode())
		logs.Info(url)
		ctx.Redirect(http.StatusFound, url)
	}
}

/**
登录（用于前后端分离的项目，sso登录授权后重定向到前端页面，前端页面将code传给后端获取sso token）
*/

func (a *Auth) Login(code string, ctx *context.Context) {
	if user, err := a.queryTokenFromOauth2(code, ctx); err != nil {
		a.RedirectToLogin(ctx)
	} else {
		ctx.Input.CruSession.Set(SESSION_KEY_USER, user)
		if ctx.Input.Header("x-requested-with") == "XMLHttpRequest" {
			ctx.ResponseWriter.WriteHeader(200)
			ctx.WriteString("登录成功")
		} else {
			if originUrl := ctx.Input.Query("state"); originUrl != "" {
				ctx.Redirect(http.StatusFound, originUrl)
			} else {
				ctx.Redirect(http.StatusFound, "/")
			}
		}
	}
}

/**
登出
*/
func (a *Auth) Logout(ctx *context.Context, state string) {
	if user, ok := ctx.Input.CruSession.Get(SESSION_KEY_USER).(User); ok {
		url := "https://auth.yxapp.in/oauth2/token?access_token=" + user.Token.AccessToken
		res, err := httplib.Delete(url).Response()
		if err != nil {
			logs.Error(err)
		} else if res.StatusCode != http.StatusOK {
			logs.Error(res)
		}
	}
	ctx.Input.CruSession.Flush()
	params := url.Values{}
	params.Add("client_id", strconv.FormatInt(a.ClientId, 10))
	params.Add("redirect_uri", a.RedirectUri)
	params.Add("response_type", "code")
	params.Add("state", state)
	params.Add("scope", a.Scope)
	url := fmt.Sprintf("%s/oauth2/authorize?%s", a.Host, params.Encode())
	ctx.Redirect(http.StatusFound, url)
}

/**
查询当前session的用户信息（未登录会返回默认用户信息）
*/
func (a *Auth) CurrentUser(ctx *context.Context) User {
	if user, ok := ctx.Input.CruSession.Get(SESSION_KEY_USER).(User); ok {
		return user
	} else {
		return User{Fullname: "未登录用户", Id: " AnonymousUser", ResourceMap: map[string]*Resource{}}
	}
}

/**
用code向sso请求获取access token
*/
func (a *Auth) queryTokenFromOauth2(code string, ctx *context.Context) (Token, error) {
	params := url.Values{}
	params.Add("client_id", strconv.FormatInt(a.ClientId, 10))
	params.Add("client_secret", a.ClientSecret)
	params.Add("redirect_uri", a.RedirectUri)
	params.Add("grant_type", "authorization_code")
	params.Add("code", code)

	url := fmt.Sprintf("%s/oauth2/token?%s", a.Host, params.Encode())
	var token Token
	if err := httplib.Post(url).ToJSON(&token); err != nil {
		return token, err
	} else if token.Error != "" {
		return token, errors.New(token.Error + ":" + token.ErrorDescription)
	} else {
		return token, nil
	}
}
