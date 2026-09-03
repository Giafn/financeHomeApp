package storage

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const presignExpiry = 12 * time.Hour

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
	return p.UploadURL(ctx, filename, contentType)
}
func (p *Presigner) UploadURL(ctx context.Context, filename, contentType string) (string, string, error) {
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

// storageKeyPrefix adalah prefix yang digunakan saat upload (lihat UploadURL).
const storageKeyPrefix = "attachments/"

// ReadURL mengimplementasikan Storage — membuat presigned GET URL supaya client
// bisa membaca file yang tersimpan. storedURL adalah attachment_url ter-stored
// (bisa berupa fileURL penuh maupun key relatif). Nama file di dalamnya berpotensi
// ter-URL-encode (misal spasi jadi %20), jadi key di-decode dulu agar cocok dengan
// key asli objek di bucket.
func (p *Presigner) ReadURL(ctx context.Context, storedURL string) (string, error) {
	rawKey := extractS3Key(storedURL, p.endpoint, p.bucket)
	if rawKey == "" {
		return "", fmt.Errorf("URL bukan milik bucket ini")
	}

	key, err := url.PathUnescape(rawKey)
	if err != nil {
		return "", fmt.Errorf("key tidak valid: %v", err)
	}
	if key == "" {
		return "", fmt.Errorf("key kosong")
	}

	req, err := p.client.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(presignExpiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// extractS3Key mengambil key dari fileURL (misal
// "https://s3.giafn.my.id/muslim/attachments/<uuid>-<nama>"). Bila yang diberikan
// sudah berupa key relatif (tanpa scheme/host), dikembalikan apa adanya.
func extractS3Key(storedURL, endpoint, bucket string) string {
	candidate := strings.TrimSpace(storedURL)
	if candidate == "" {
		return ""
	}

	// Bila sudah berupa key relatif (mulai "attachments/"), langsung dipakai.
	if strings.HasPrefix(candidate, storageKeyPrefix) {
		return candidate
	}

	base := strings.TrimRight(endpoint, "/") + "/" + bucket + "/"
	base = strings.Replace(base, "https://", "", 1)
	base = strings.Replace(base, "http://", "", 1)

	var stripped string
	for _, scheme := range []string{"https://", "http://"} {
		if strings.HasPrefix(candidate, scheme) {
			stripped = strings.TrimPrefix(candidate, scheme)
			break
		}
	}
	if stripped == "" {
		// Bukan URL absolut; coba treat sebagai key relatif bila valid.
		return candidate
	}

	if strings.HasPrefix(stripped, base) {
		return strings.TrimPrefix(stripped, base)
	}
	return ""
}
