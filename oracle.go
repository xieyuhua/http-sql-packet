package main

import (
	"fmt"
	"regexp"
// 	"log"
	"strings"
	"bytes"
	"github.com/axgle/mahonia"
)


func fixChineseEncoding(input string) string {
    decoder := mahonia.NewDecoder("gbk")
    input = decoder.ConvertString(input)
    return input
}

//BytesCombine 多个[]byte数组合并成一个[]byte
func BytesCombine(pBytes ...[]byte) []byte {
    return bytes.Join(pBytes, []byte(""))
}


// 解析TCP负载中的Oracle SQL语句
func ParseOracleSQL(client, server string,n int, payloadStr []byte) string {
    
	// 使用正则表达式匹配SQL语句
// 	re := regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE|FROM)\b.*`)
    
	resss := regexp.MustCompile(`[\s\r\n]+`)
	sql := resss.ReplaceAllString(string(payloadStr[:n]), " ")
    //中文乱码
    sql = fixChineseEncoding(sql)
    sql = strings.ReplaceAll(sql, "@", "")
    sql = strings.ReplaceAll(sql, "�", "")
    sql = strings.TrimSpace(sql)
    
    re := regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE)\b.*`)
    matches := re.FindAllString(sql, -1)
    // fmt.Println(matches)
    if len(matches)>0 {
        verboseStr := fmt.Sprintf("From %s To %s;  %v \n\n", client, server, sql)
    	fmt.Println(verboseStr)
    }

// 	re := regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE|\s+|FROM)\b.*`)
// 	matches := re.FindAllString(sql, -1)
// 	for _, match := range matches {
// 	    sql = strings.TrimSpace(match)
// 		verboseStr := fmt.Sprintf("From %s To %s;  %s \n\n", client, server, sql)
// // 		Log.Infof("sql: \n", verboseStr)
// 		fmt.Print(verboseStr)
// 	}

	return ""
}
