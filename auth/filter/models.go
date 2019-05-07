package filter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/logs"
	"gitlab.aibee.cn/platform/golang_common/beego/controllers"
	"time"
)

type User struct {
	// user
	Id       string `json:"id"`
	Fullname string `json:"fullname"`
	Dn       string `json:"dn"`

	// resource
	Resources   []*Resource          `json:"resource"`
	ResourceMap map[string]*Resource `json:"resourceMap"`
	CacheTime   int64                `json:"cacheTime"`

	// token
	Token Token `json:"-"`
}

type Resource struct {
	Id          int64  `json:"id"`
	Description string `json:"description"`
	Data        string `json:"data"`
}

type Token struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	TokenType        string `json:"token_type"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (u *User) Init(auth *Auth) error {
	res := controllers.ResponseBody{}
	if err := httplib.Get(fmt.Sprintf("%s/api/user", auth.Host)).
		Header("Authorization", fmt.Sprintf("%s %s", u.Token.TokenType, u.Token.AccessToken)).
		ToJSON(&res); err != nil {
		return err
	} else if res.ResCode != controllers.OK {
		return errors.New(res.ResMsg)
	} else {
		userJson, _ := json.Marshal(res.Data)
		if err := json.Unmarshal(userJson, u); err != nil {
			logs.Error("res data error : %+v", res.Data)
			return errors.New("res data error")
		} else {
			return nil
		}
	}
}

func (u *User) LoadResource(auth *Auth) error {
	res := controllers.ResponseBody{}
	if err := httplib.Get(fmt.Sprintf("%s/api/userResources", auth.Host)).
		Header("Authorization", fmt.Sprintf("%s %s", u.Token.TokenType, u.Token.AccessToken)).
		ToJSON(&res); err != nil {
		return err
	} else if res.ResCode != controllers.OK {
		return errors.New(res.ResMsg)
	} else {
		resourcesJson, _ := json.Marshal(res.Data)
		if err := json.Unmarshal(resourcesJson, &u.Resources); err != nil {
			logs.Error("res data error : %+v", res.Data)
			return errors.New("res data error")
		} else {
			u.ResourceMap = make(map[string]*Resource)
			for _, v := range u.Resources {
				u.ResourceMap[v.Data] = v
			}
			u.CacheTime = time.Now().Unix()
			return nil
		}
	}
}
