package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Markdownform struct {
	content string
}

type Webhookdata struct {
	msgtype string
	*Markdownform
}

// 触发企业微信机器人
func WetchatWebhook(dat string) {
	uri := os.Getenv("URL")
	client := &http.Client{}
	urlmap := url.Values{}
	w := &Webhookdata{
		msgtype: "markdown",
		content: dat,
	}
	urlmap.Add("msgtype", w.msgtype)
	urlmap.Add("markdown", w.content)
	parms := ioutil.NopCloser(strings.NewReader(urlmap.Encode())) //把form数据编下码
	req, err := http.NewRequest("POST", uri, parms)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	fmt.Println(string(body))
}

// // 定义Gitdata 数据结构,封装请求体解码数据
// type Gitdata struct{
// 	*Bodydata.ObjectKind
// 	*Bodydata.Project
// 	*Bodydata.ObjectAttributes
// 	*Bodydata.LastCommit
// 	*Bodydata.MergeRequest
// 	*Bodydata.MergeRequest
// 	*Bodydata.Author
// 	Flag string
// }

// 企业微信模板
func template(g, *Bodydata) string {
	kind := g.Objectkind
	switch kind {
	case "note":
		return fmt.Sprintf(`<font color="warning">Gitlab事件通知</font>。
			>事件类型: <font color="red">%v</font>
			>项目名称: <font color="green">%v</font>
			>评论内容: <font color="green">%v</font>
			>评论地址: <font color="warning">%v</font>
			>评论时间: <font color="comment">%v</font>
			>提交人:<font color="comment">%v</font>
			`, g.Objectkind, g.Projectname, g.Commit, g.Projecturl, g.Data, g.User)
	case "merge_request":
		return fmt.Sprintf(`<font color="warning">Gitlab事件通知</font>。
			>事件类型: <font color="red">%v</font>
			>项目名称: <font color="green">%v</font>
			>源分支: <font color="green">%v</font>
			>目的分支: <font color="warning">%v</font>
			>最后commit: <font color="comment">%v</font>
			>评论时间: <font color="comment">%v</font>
			>提交人:<font color="comment">%v</font>
			`, g.Objectkind, g.Projectname, g.Sourcebranch, g.Targetbranch, g.Commit, g.Data, g.User)
	default:
		return ""
	}
}

// 请求函数
func gitPush(c *gin.Context) {
	matched, _ := VerifySignature(c)
	if !matched {
		err := "Token did not match"
		c.String(http.StatusForbidden, err)
		fmt.Println(err)
		return
	}
	fmt.Println("Token is matched ~")
	// >>>>>>>>>>>>>>>>>
	body, err := c.GetRawData()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	err := json.Unmarshal(body, &Gitdata) // 解析完request body 数据
	if err != nil {
		fmt.Println("error:", err)
		t := template(&Gitdata)
		go WetchatWebhook(t) // 调用企业微信机器人接口
		return
	} else {
		c.String(http.StatusOK, "ok")
	}
}

// 验证token
func VerifySignature(c *gin.Context) (bool, error) {
	// Get Header with X-Hub-Signature
	XLibToken := c.GetHeader("X-Gitlab-Token")
	signature := GetToken("TOKEN_KEY")
	fmt.Println(signature)
	return XLibToken == signature, nil
}

// 从环境变量获取token
func GetToken(e string) string {
	env := os.Getenv(e)
	if env != "" {
		fmt.Printf("from os get successfully env!, %s=%s", e, env)
	} else {
		fmt.Printf("%s env not found", e)
	}
	return env
}

// 首页返回提示语
func defaultPage(g *gin.Context) {
	firstName := g.DefaultQuery("firstName", "Xkooler")
	lastName := g.Query("lastName")
	g.String(http.StatusOK, "Hello %s %s, This is Gitlab Events Server~", firstName, lastName)
}

func main() {
	router := gin.Default()
	router.GET("/", defaultPage)
	router.POST("/send", gitPush)
	_ = router.Run(":8079")
}
