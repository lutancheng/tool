package iam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

/*type iam struct {
	dataFile *excelize.File
}

func NewIam() *iam {
	return &iam{}
}*/

type Doc struct {
	Files []file `json:"files,omitempty"`
}
type file struct {
	Name     string    `json:"name,omitempty"`
	Package  string    `json:"package,omitempty"`
	Services []service `json:"services,omitempty"`
}
type service struct {
	Methods []method `json:"methods,omitempty"`
}
type method struct {
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Options     options `json:"options,omitempty"`
}
type options struct {
	Api api `json:"google.api.http,omitempty"`
}
type api struct {
	Rules []rule `json:"rules,omitempty"`
}
type rule struct {
	Method  string `json:"method,omitempty"`
	Pattern string `json:"pattern,omitempty"`
}

type AuthProxyTag struct {
	Action string
	Ingore bool
}

const (
	AUTH_PROXY    = "auth-proxy"
	INSERT_SQL    = "INSERT INTO `service_relation` (`id`, `expr`, `action`, `resource`, `type`) VALUES ('%d', '(re2)^/%s$', '%s', 'i_tc_%s', '%s');"
	INSERT_V1_SQL = "INSERT INTO `service_relation` (`id`, `expr`, `action`, `resource`, `type`) VALUES ('1', '(re2)^/traffic-control-api/v1/download/.*', 'download', 'i_tc_traffic_control_api', 'get');\n"
)

var (
	id = 2
)

func GenIamSql(inFilePath, outFilePaht string) {
	content := getSqlFromFile(inFilePath)
	writeFile(content, outFilePaht)
}

func getSqlFromFile(inFilePath string) string {
	filePtr, err := os.Open(inFilePath)
	if err != nil {
		fmt.Println("Open file failed [Err:%s]", err.Error())
		return ""
	}
	defer filePtr.Close()

	doc := &Doc{}
	b, err := ioutil.ReadAll(filePtr)
	if err != nil {
		log.Fatalf("文件读取错误: %s", err)
	}
	err = json.Unmarshal(b, &doc)
	if err != nil {
		log.Fatal(err)
	}
	var buffer bytes.Buffer
	buffer.WriteString(INSERT_V1_SQL)
	for _, file := range doc.Files {
		buffer.WriteString(genSql(file))
	}
	//fmt.Printf(buffer.String())
	return buffer.String()
}

func genSql(file file) string {
	var sql, appName string
	packageName := file.Package
	pkgNameArr := strings.Split(packageName, ".")
	if len(pkgNameArr) == 0 {
		return ""
	}

	//	traffic-control-api的包有例外：tcgroup.traffic_control_api.v2
	if pkgNameArr[0] == "tcgroup" {
		appName = pkgNameArr[1]
	} else {
		appName = pkgNameArr[0]
	}

	services := file.Services
	if services == nil || len(services) == 0 {
		return ""
	}
	service := services[0]
	for _, method := range service.Methods {
		authProxyTag := getAuthProxyTag(method.Description)
		//	设置忽略
		if authProxyTag.Ingore {
			continue
		}
		rules := method.Options.Api.Rules
		for _, rule := range rules {
			httpType := strings.ToLower(rule.Method)
			action := httpType
			switch httpType {
			case "post":
				action = "create"
			case "get":
				action = "query"
			case "put":
				action = "update"
			}
			if authProxyTag.Action != "" {
				action = authProxyTag.Action
			}

			url := appName + rule.Pattern
			sql += fmt.Sprintf(INSERT_SQL, id, url, action, appName, httpType) + "\n"
			id++
		}
	}
	return sql
}

func writeFile(content, outFilePaht string) {
	b := []byte(content)
	file, err := os.OpenFile(outFilePaht, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)

	defer file.Close()
	if err != nil {
		panic(err)
	}

	// 写入文件
	n, err := file.Write(b)
	// 当 n != len(b) 时，返回非零错误
	if err == nil && n != len(b) {
		println(`错误代码：`, n)
		panic(err)
	}
}

func getAuthProxyTag(description string) AuthProxyTag {
	tagMap := make(map[string]string)
	authProxyTag := AuthProxyTag{}

	if description == "" {
		return authProxyTag
	}

	desc := strings.TrimSpace(description)
	tagList := strings.Split(desc, " ")
	for _, tag := range tagList {
		kv := strings.Split(tag, ":")
		if len(kv) != 2 {
			continue
		}
		tagMap[kv[0]] = kv[1]
	}
	if vList, ok := tagMap[AUTH_PROXY]; ok {
		//	去除两边的""
		vList = trimQuotes(vList)
		attList := strings.Split(vList, ";")
		for _, att := range attList {
			attKv := strings.Split(att, "=")
			if attKv[0] == "action" && attKv[1] != "" {
				authProxyTag.Action = attKv[1]
			} else if attKv[0] == "ignore" {
				authProxyTag.Ingore = true
			}
		}
	}
	return authProxyTag
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}
