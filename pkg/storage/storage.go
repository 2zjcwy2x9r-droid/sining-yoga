package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Storage 存储接口
type Storage interface {
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) error
	GetObject(ctx context.Context, bucketName, objectName string) (io.Reader, error)
	RemoveObject(ctx context.Context, bucketName, objectName string) error
	GetObjectURL(bucketName, objectName string) string
}

// MinIOStorage MinIO存储实现
type MinIOStorage struct {
	client   *minio.Client
	endpoint string
	useSSL   bool
}

// NewMinIOStorage 创建MinIO存储
func NewMinIOStorage(endpoint, accessKeyID, secretAccessKey string, useSSL bool) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("创建MinIO客户端失败: %w", err)
	}

	return &MinIOStorage{
		client:   client,
		endpoint: endpoint,
		useSSL:   useSSL,
	}, nil
}

// EnsureBucket 确保存储桶存在
func (s *MinIOStorage) EnsureBucket(ctx context.Context, bucketName string) error {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("检查存储桶是否存在失败: %w", err)
	}

	if !exists {
		if err := s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("创建存储桶失败: %w", err)
		}
	}

	return nil
}

// PutObject 上传对象
func (s *MinIOStorage) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	if err := s.EnsureBucket(ctx, bucketName); err != nil {
		return err
	}

	_, err := s.client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("上传对象失败: %w", err)
	}

	return nil
}

// GetObject 获取对象
func (s *MinIOStorage) GetObject(ctx context.Context, bucketName, objectName string) (io.Reader, error) {
	obj, err := s.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取对象失败: %w", err)
	}
	return obj, nil
}

// RemoveObject 删除对象
func (s *MinIOStorage) RemoveObject(ctx context.Context, bucketName, objectName string) error {
	if err := s.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("删除对象失败: %w", err)
	}
	return nil
}

// GetObjectURL 获取对象URL
func (s *MinIOStorage) GetObjectURL(bucketName, objectName string) string {
	protocol := "http"
	if s.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, s.endpoint, bucketName, objectName)
}

