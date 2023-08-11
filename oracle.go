package main

import (
	"fmt"
	"regexp"
	"log"
	"strings"
	"github.com/axgle/mahonia"
)

func fixChineseEncoding(input string) string {
    decoder := mahonia.NewDecoder("gbk")
    input = decoder.ConvertString(input)
    return input
}

// 解析TCP负载中的Oracle SQL语句
func ParseOracleSQL(client, server string, payloadStr  []byte) string {
    
	// 使用正则表达式匹配SQL语句
	re := regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE)\b.*`)
// 	re := regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE)\b.*?;`)

	resss := regexp.MustCompile(`[\s\r\n]+`)
	sql := resss.ReplaceAllString(string(payloadStr), " ")
// 	fmt.Println(sql)
	matches := re.FindAllString(sql, -1)
	for _, match := range matches {
	    // 去除空白字符和注释
        sql = strings.TrimSpace(match)
        //中文乱码
        sql = fixChineseEncoding(sql)
        sql = strings.ReplaceAll(sql, "@", "")
        
		verboseStr := fmt.Sprintf("From %s To %s;  %s \n\n", client, server, sql)
// 		Log.Infof("sql: \n", verboseStr)
		log.Print(verboseStr)
	}

	return ""
}
