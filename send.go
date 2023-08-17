package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Dataform struct {
	Content string `json:"content"`
}

type TextDataform struct {
	Content             string   `json:"content"`
	MentionedMobileList []string `json:"mentioned_mobile_list"`
}

type Markdowndata struct {
	Msgtype   string `json:"msgtype"`
	*Dataform `json:"markdown"`
}
type Textdata struct {
	Msgtype       string `json:"msgtype"`
	*TextDataform `json:"text"`
}

// 找到需要@的人员对应手机号
func phonelist(s string) string {
	Someone := map[string]string{
		"ns":         "1812***86881",
		"tristanwqy": "130***170984",
		"ssh415":     "130***259541",
		"fonshin":    "130***889455",
		"fanqidi":    "185***19002",
		"yumo":       "135***60160",
	}
	if v, ok := Someone[s]; ok {
		fmt.Println(v)
		return v
	} else {
		fmt.Println("Key Not Found")
		return ""
	}
}

// 找到需要被@的开发人员
func getuser(g *Bodydata) string {
	if g.ObjectAttributes.Action == "approved" && g.ObjectAttributes.LastCommit.Author.Name != "" {
		return fmt.Sprintf("%v", g.ObjectAttributes.LastCommit.Author.Name)
	} else if g.ObjectAttributes.Action == "open" && g.ObjectAttributes.Assignee.Name != "" {
		return fmt.Sprintf("%v", g.ObjectAttributes.Assignee.Name)
	} else {
		return ""
	}
}

// 序列化text模板
func textjson(da string, at string) string {
	a := phonelist(at)
	t := &Textdata{
		Msgtype: "text",
		TextDataform: &TextDataform{
			da,
			[]string{fmt.Sprintf("%v", a)},
		},
	}
	d, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	return string(d)
}

// 序列化md模板
func mdjson(da string) string {
	t := &Markdowndata{
		Msgtype:  "markdown",
		Dataform: &Dataform{da},
	}
	d, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	return string(d)
}

// 触发企业微信机器人
func WetchatWebhook(s string) {
	uri := os.Getenv("URL")
	req, _ := http.NewRequest("POST", uri, strings.NewReader(s))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	fmt.Println(res.StatusCode)
	fmt.Println(body)
}

// md类型企业微信模板
func template(g *Bodydata) string {
	kind := g.ObjectKind
	switch kind {
	case "note1":
		return fmt.Sprintf(`<font color="warning">Gitlab事件通知</font>。
			>事件类型: <font color="red">%v</font>
			>源分支: <font color="green">%v</font>
			>目的分支: <font color="green">%v</font>
			>Title: <font color="green">%v</font>
			>描述: <font color="green">%v</font>
			>更新时间: <font color="green">%v</font>
			>MR地址: <font color="warning">%v</font>
			>评论内容: <font color="comment">%v</font>
			>评论人: <font color="comment">%v</font>
			>评论时间: <font color="green">%v</font>
			>提交人:<font color="comment">%v</font>
			`, g.ObjectKind, g.MergeRequest.SourceBranch, g.MergeRequest.TargetBranch, g.MergeRequest.Title, g.MergeRequest.Description, g.MergeRequest.UpdatedAt, g.ObjectAttributes.URL, g.ObjectAttributes.Note, g.ObjectAttributes.AuthorID, g.ObjectAttributes.UpdatedAt, g.MergeRequest.LastCommit.Author.Name)
	case "merge_request":
		if g.ObjectAttributes.State != "merged" && !g.ObjectAttributes.WorkInProgress && g.ObjectAttributes.Action != "update" {
			return fmt.Sprintf(
				`%v <font color="warning">%v</font> [Merge Request](%V)
			>分支: <font color="green">%v ---> %v</font>
			>Title: <font color="green">%v</font>
			>描述: <font color="green">%v</font>
			>Merge状态: <font color="green">%v</font>`, g.User.Name, g.ObjectAttributes.Action, g.ObjectAttributes.URL, g.ObjectAttributes.SourceBranch, g.ObjectAttributes.TargetBranch, g.ObjectAttributes.Title, g.ObjectAttributes.Description, g.ObjectAttributes.MergeStatus)
		}
		return ""
	case "build1":
		if g.BuildStatus == "failed" {
			return fmt.Sprintf(`<font color="warning">Gitlab事件通知</font>。
			>事件类型: <font color="red">%v</font>
			>构建项目: <font color="green">%v</font>
			>构建状态: <font color="warning">%v</font>
			>构建开始时间: <font color="green">%v</font>
			>构建结束时间: <font color="green">%v</font>
			>commIt_id: <font color="green">%v</font>
			>提交人: <font color="green">%v</font>
			`, g.ObjectKind, g.ProjectName, g.BuildStatus, g.Commit.StartedAt, g.Commit.FinishedAt, g.Commit.ID, g.Commit.AuthorName)
		}
		return ""
	default:
		return ""
	}
}

// text 类型模板
func templatetext(g *Bodydata) string {
	kind := g.ObjectKind
	switch kind {
	case "merge_request":
		if g.ObjectAttributes.State != "merged" && !g.ObjectAttributes.WorkInProgress && g.ObjectAttributes.Action != "update" {
			return g.ObjectAttributes.URL
		}
		return ""
	default:
		return ""
	}
}

// 请求函数
func gitPush(c *gin.Context) {
	var bodydata = &Bodydata{}
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
	err = json.Unmarshal(body, bodydata) // 解析完request body 数据
	if err != nil {
		fmt.Println("error:", err)
		return
	} else {
		md := template(bodydata)
		text := templatetext(bodydata)
		user := getuser(bodydata)
		fmt.Println("user=", user)
		fmt.Println("md=", md)
		fmt.Println("text=", text)
		tj := textjson(text, user)
		mj := mdjson(md)
		fmt.Println("md:", mj)
		fmt.Println("text:", tj)
		WetchatWebhook(mj) // 调用企业微信机器人接口
		WetchatWebhook(tj) // 单独发送url
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
