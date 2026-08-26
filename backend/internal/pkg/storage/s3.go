package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const presignExpiry = 5 * time.Minute

// Presigner men-generate presigned PUT URL supaya client upload langsung ke S3
// tanpa lewat server (specs.md §7 — hindari beban upload di server backend).
type Presigner struct {
	client   *s3.PresignClient
	bucket   string
	endpoint string
}

func NewPresigner(ctx context.Context, endpoint, region, accessKey, secretKey, bucket string) (*Presigner, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true
	})

	return &Presigner{
		client:   s3.NewPresignClient(s3Client),
		bucket:   bucket,
		endpoint: endpoint,
	}, nil
}

// PresignUpload mengembalikan (uploadURL, fileURL) untuk sebuah filename+content-type.
// fileURL dipakai sebagai attachment_url final setelah upload sukses.
func (p *Presigner) PresignUpload(ctx context.Context, filename, contentType string) (string, string, error) {
	key := fmt.Sprintf("attachments/%s-%s", uuid.NewString(), filename)

	req, err := p.client.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(presignExpiry))
	if err != nil {
		return "", "", err
	}

	fileURL := fmt.Sprintf("%s/%s/%s", p.endpoint, p.bucket, key)
	return req.URL, fileURL, nil
}
