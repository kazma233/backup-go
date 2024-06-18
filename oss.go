package main

import (
	"backup-go/config"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type (
	OssClient struct {
		slowBucket *oss.Bucket
		fastBucket *oss.Bucket
	}

	UploadNoticeFunc func(string)
)

func CreateOSSClient() *OssClient {
	ossClient := &OssClient{
		slowBucket: must(getBucket(
			config.Config.OSS.Endpoint,
			config.Config.OSS.AccessKey,
			config.Config.OSS.AccessKeySecret,
			config.Config.OSS.BucketName)), // slowBucket must not nil
		fastBucket: getBucket(
			config.Config.OSS.FastEndpoint,
			config.Config.OSS.AccessKey,
			config.Config.OSS.AccessKeySecret,
			config.Config.OSS.BucketName,
		),
	}

	log.Printf("oss client init done: %v", ossClient)

	return ossClient
}

func (oc *OssClient) Upload(objKey, filePath string, noticeFunc UploadNoticeFunc) (err error) {
	if oc.slowBucket == nil && oc.fastBucket == nil {
		return errors.New("client not init")
	}

	noticeFunc("use slow bucket")
	err = oc.slowBucket.PutObjectFromFile(objKey, filePath)
	if err == nil {
		noticeFunc("use slow bucket upload success")
		return nil
	} else {
		noticeFunc(fmt.Sprintf("use slow bucket upload error: %v", err))
	}

	if oc.fastBucket != nil {
		noticeFunc("use fast bucket")
		err := oc.fastBucket.PutObjectFromFile(objKey, filePath)
		if err == nil {
			noticeFunc("use fast bucket upload success")
			return nil
		} else {
			noticeFunc(fmt.Sprintf("use fast bucket upload faild: %v", err))
		}
	}

	return
}

func (oc *OssClient) TempVisitLink(objKey string) (string, error) {
	return oc.slowBucket.SignURL(objKey, oss.HTTPGet, 60*60*24*1)
}

func (oc *OssClient) GetSlowClient() *oss.Bucket {
	return oc.slowBucket
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
