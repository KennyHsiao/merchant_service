package utils

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

func GetSign(source string) string {
	data := []byte(source)
	result := fmt.Sprintf("%x", md5.Sum(data))
	return result
}

/*
JoinStringsInASCII 按照规则，参数名ASCII码从小到大排序后拼接
data 待拼接的数据
sep 连接符
onlyValues 是否只包含参数值，true则不包含参数名，否则参数名和参数值均有
includeEmpty 是否包含空值，true则包含空值，否则不包含，注意此参数不影响参数名的存在
exceptKeys 被排除的参数名，不参与排序及拼接
*/
func JoinStringsInASCII(data map[string]string, sep string, onlyValues, includeEmpty bool, key string, exceptKeys ...string) string {
	var list []string
	var keyList []string
	m := make(map[string]int)
	if len(exceptKeys) > 0 {
		for _, except := range exceptKeys {
			m[except] = 1
		}
	}
	for k := range data {
		if _, ok := m[k]; ok {
			continue
		}
		value := data[k]
		if !includeEmpty && value == "" {
			continue
		}
		if onlyValues {
			keyList = append(keyList, k)
		} else {
			list = append(list, fmt.Sprintf("%s=%s", k, value))
		}
	}
	if onlyValues {
		sort.Strings(keyList)
		keyList = append(keyList, key) //加key
		for _, v := range keyList {
			list = append(list, data[v])
		}
	} else {
		sort.Strings(list)
		list = append(list, fmt.Sprintf("%s=%s", "Key", key))
	}
	return strings.Join(list, sep)
}

// 验签
func VerifySign(reqSign string, data interface{}, screctKey string, ctx context.Context) bool {
	m := CovertToMap(data)
	source := JoinStringsInASCII(m, "&", false, false, screctKey)
	sign := GetSign(source)
	fmt.Sprintf("-------" + source)
	logx.WithContext(ctx).Info("verifySource: ", source)
	logx.WithContext(ctx).Info("verifySign: ", sign)
	logx.WithContext(ctx).Info("reqSign: ", reqSign)

	if reqSign == sign {
		return true
	}

	return false
}

func SortAndSignFromUrlValues(data url.Values, screctKey string) string {
	m := CovertUrlValuesToMap(data)
	return SortAndSign(m, screctKey)
}

// SortAndSign2 排序后加签
func SortAndSign2(data interface{}, screctKey string) string {
	m := CovertToMap(data)
	newSource := JoinStringsInASCII(m, "&", false, false, screctKey)
	newSign := GetSign(newSource)
	logx.Info("加签参数: ", newSource)
	logx.Info("签名字串: ", newSign)
	return newSign
}

// 排序后加签
func SortAndSign(newData map[string]string, screctKey string) string {
	newSource := JoinStringsInASCII(newData, "&", false, false, screctKey)
	newSign := GetSign(newSource)
	logx.Info("加签参数: ", newSource)
	logx.Info("签名字串: ", newSign)
	return newSign
}

func CovertUrlValuesToMap(values url.Values) map[string]string {
	m := make(map[string]string)
	for k := range values {
		m[k] = values.Get(k)
	}
	return m
}

// 檢查請求參數是否有空值
func CovertToMap(req interface{}) map[string]string {
	m := make(map[string]string)

	val := reflect.ValueOf(req)
	for i := 0; i < val.Type().NumField(); i++ {
		jsonTag := val.Type().Field(i).Tag.Get("json")
		parts := strings.Split(jsonTag, ",")
		name := parts[0]
		if name != "sign" {
			if val.Field(i).Type().Name() == "float64" {
				precise := GetDecimalPlaces(val.Field(i).Float())
				valTrans := strconv.FormatFloat(val.Field(i).Float(), 'f', precise, 64)
				m[name] = valTrans
			} else if val.Field(i).Type().Name() == "string" {
				m[name] = val.Field(i).String()
			} else if val.Field(i).Type().Name() == "int64" {
				m[name] = strconv.FormatInt(val.Field(i).Int(), 10)
			}

		}
	}

	return m
}

func MicroServiceEncrypt(key, publicKey string) (sing string, err error) {
	str := key + time.Now().Format("200601021504")
	src := []byte(str)
	if src, err = DesCBCEncrypt(src, []byte(publicKey)); err != nil {
		return
	}
	return base64.StdEncoding.EncodeToString(src), err
}

func MicroServiceVerification(sing, key, publicKey string) (isOk bool, err error) {
	var singByte []byte
	if singByte, err = base64.StdEncoding.DecodeString(sing); err != nil {
		return
	}
	if singByte, err = DesCBCDecrypt(singByte, []byte(publicKey)); err != nil {
		return
	}

	decryptStr := string(singByte)

	if strings.Index(decryptStr, key) > -1 {
		trimStr := strings.Replace(decryptStr, key, "", 1)
		if timeX, err := time.Parse("200601021504", trimStr); err == nil {
			if time.Now().Sub(timeX).Minutes() <= 5 {
				isOk = true
			}
		}
	}
	return
}

func DesCBCEncrypt(origData, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	origData = PKCS5Padding(origData, block.BlockSize())
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key)
	crypted := make([]byte, len(origData))
	// 根據CryptBlocks方法的說明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func DesCBCDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, key)
	//origData := make([]byte, len(crypted))
	origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	//origData = PKCS5UnPadding(origData)

	origData = PKCS5UnPadding(origData)
	return origData, nil
}

func AesCBCEncrypt(origData, key []byte) ([]byte, error) {
	iv := key[0:16]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	origData = PKCS5Padding(origData, block.BlockSize())

	blockMode := cipher.NewCBCEncrypter(block, iv)
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func AesCBCDecrypt(crypted, key []byte) ([]byte, error) {
	iv := key[0:16]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func GetDecimalPlaces(f float64) int {
	numstr := fmt.Sprint(f)
	tmp := strings.Split(numstr, ".")
	if len(tmp) <= 1 {
		return 0
	}
	return len(tmp[1])
}
