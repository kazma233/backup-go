package main

import (
	"errors"
	"log"
	"reflect"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type (
	OssClient struct {
		slowBucket *oss.Bucket
		fastBucket *oss.Bucket
	}
)

var (
	ossClient *OssClient
)

func InitOSS() {
	// slowBucket must not nil
	ossClient = &OssClient{
		slowBucket: must(getBucket(Config.OSS.Endpoint, Config.OSS.AccessKey, Config.OSS.AccessKeySecret, Config.OSS.BucketName)),
		fastBucket: getBucket(Config.OSS.FastEndpoint, Config.OSS.AccessKey, Config.OSS.AccessKeySecret, Config.OSS.BucketName),
	}

	log.Printf("oss client init done: %v", ossClient)
}

func must[T any](obj T) T {
	if isNil(obj) {
		panic(errors.New("obj is nil"))
	}

	return obj
}

func isNil[T any](obj T) bool {
	v := reflect.ValueOf(obj)
	kind := v.Kind()
	return canBeNil(kind) && v.IsNil()
}

func canBeNil(kind reflect.Kind) bool {
	return kind == reflect.Ptr ||
		kind == reflect.Interface ||
		kind == reflect.Slice ||
		kind == reflect.Map ||
		kind == reflect.Chan ||
		kind == reflect.Func
}

func getBucket(endpoint, ak, aks, buckatName string) *oss.Bucket {
	if endpoint == "" || ak == "" || aks == "" || buckatName == "" {
		return nil
	}

	client, err := oss.New(endpoint, ak, aks, oss.Timeout(10, 60*60*3))
	if err != nil {
		panic(err)
	}

	bucket, err := client.Bucket(buckatName)
	if err != nil {
		panic(err)
	}

	return bucket
}

func Upload(objKey, filePath string) (err error) {
	if ossClient == nil || (ossClient.slowBucket == nil && ossClient.fastBucket == nil) {
		return errors.New("client not init")
	}

	err = ossClient.slowBucket.PutObjectFromFile(objKey, filePath)
	if err == nil {
		return nil
	}

	if ossClient.fastBucket != nil {
		err := ossClient.fastBucket.PutObjectFromFile(objKey, filePath)
		if err == nil {
			return nil
		}
	}

	return
}

func TempVisitLink(objKey string) (string, error) {
	return ossClient.slowBucket.SignURL(objKey, oss.HTTPGet, 60*60*24*7)
}

func GetSlowClient() *oss.Bucket {
	return ossClient.slowBucket
}
