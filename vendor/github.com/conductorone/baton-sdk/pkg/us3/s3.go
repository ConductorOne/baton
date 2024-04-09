package us3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	_ "github.com/aws/aws-sdk-go-v2/aws/arn"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const defaultS3MaxFilesize = 256 * 1024 * 1024

type S3Client struct {
	awsConfig  awsSdk.Config
	httpClient *http.Client
	cfg        *s3Config

	_callingConfigOnce  sync.Once
	_callingConfig      awsSdk.Config
	_callingConfigError error
}

type assumeRoleConfig struct {
	roleARN    string
	externalID string
}

type s3Config struct {
	accessKeyID     string
	secretAccessKey string
	region          string
	endpointURL     string
	bucketName      string

	maxFileSize int64
	assumeRoles []*assumeRoleConfig
}

type Option func(ctx context.Context, cfg *s3Config) error

// WithCredentials configures an AWS S3 access key and secret access key for the s3config.
func WithCredentials(accessKeyID string, secretAccessKey string) Option {
	return func(ctx context.Context, cfg *s3Config) error {
		if accessKeyID == "" || secretAccessKey == "" {
			return fmt.Errorf("accessKeyID and secretAccessKey must be specified")
		}

		cfg.accessKeyID = accessKeyID
		cfg.secretAccessKey = secretAccessKey

		ctxzap.Extract(ctx).Debug("using provided credentials", zap.String("access_key_id", cfg.accessKeyID))

		return nil
	}
}

// WithRegion sets an AWS S3 region for the s3Config.
func WithRegion(region string) Option {
	return func(ctx context.Context, cfg *s3Config) error {
		if region == "" {
			return fmt.Errorf("region for s3 client must be specified")
		}

		cfg.region = region
		ctxzap.Extract(ctx).Debug("s3 region configured", zap.String("region", cfg.region))
		return nil
	}
}

// WithCustomEndpoint sets a custom AWS S3 endpoint URL for the s3Config.
func WithCustomEndpoint(endpointURL string) Option {
	return func(ctx context.Context, cfg *s3Config) error {
		if endpointURL == "" {
			return fmt.Errorf("endpointURL URL must be specified")
		}

		eURL, err := url.Parse(endpointURL)
		if err != nil {
			return fmt.Errorf("error parsing endpointURL: %w", err)
		}

		if eURL.Scheme == "" {
			return fmt.Errorf("endpointURL scheme must be set(http, https)")
		}

		cfg.endpointURL = endpointURL

		return nil
	}
}

// WithAssumeRole allows you to assume an external role to perform actions. It can be specified multiple times, and
// the specified roles will be assumed in order.
func WithAssumeRole(roleARN string, externalID string) Option {
	return func(ctx context.Context, cfg *s3Config) error {
		if !arn.IsARN(roleARN) {
			return fmt.Errorf("must provide a valid role ARN to assume")
		}

		cfg.assumeRoles = append(cfg.assumeRoles, &assumeRoleConfig{
			roleARN:    roleARN,
			externalID: externalID,
		})

		return nil
	}
}

// WithMaxDownloadFilesize limits the maximum number of bytes that a file is allowed to have in order to be uploaded or downloaded.
// Any values greater than or less than zero results in no limit.
func WithMaxDownloadFilesize(maxFileSize int64) Option {
	return func(ctx context.Context, cfg *s3Config) error {
		if maxFileSize > 0 {
			cfg.maxFileSize = maxFileSize
		}

		return nil
	}
}

type S3BucketConfig struct {
	bucket          string
	key             string
	region          string
	endpoint        string
	accessKeyID     string
	secretAccessKey string
}

func parseS3Uri(s3Uri string) (*S3BucketConfig, error) {
	parsed, err := url.Parse(strings.TrimSpace(s3Uri))
	if err != nil {
		return nil, err
	}

	ret := &S3BucketConfig{
		bucket: parsed.Hostname(),
		key:    strings.TrimPrefix(parsed.Path, "/"),
	}

	if parsed.User != nil {
		password, ok := parsed.User.Password()
		if !ok {
			return nil, fmt.Errorf("a secretAccessKey must be specified with an accessKeyID")
		}
		ret.accessKeyID = parsed.User.Username()
		ret.secretAccessKey = password
	}

	opts := parsed.Query()
	ret.region = opts.Get("region")
	ret.endpoint = opts.Get("endpoint")

	return ret, nil
}

func fetchS3MaxFilesize() int64 {
	s3MaxFilesize := os.Getenv("BATON_S3_MAX_FILESIZE_MB")
	if s3MaxFilesize == "" {
		return defaultS3MaxFilesize
	}

	maxFilesize, err := strconv.ParseInt(s3MaxFilesize, 10, 64)
	if err != nil {
		return defaultS3MaxFilesize
	}

	return maxFilesize * 1024 * 1024
}

// NewClientFromURI parses an s3://bucket/uri and creates a client. It also returns the key specified in the URI.
func NewClientFromURI(ctx context.Context, uri string) (string, *S3Client, error) {
	s3Cfg, err := parseS3Uri(uri)
	if err != nil {
		return "", nil, err
	}

	opts := []Option{
		WithMaxDownloadFilesize(fetchS3MaxFilesize()),
	}

	if s3Cfg.region != "" {
		opts = append(opts, WithRegion(s3Cfg.region))
	}

	if s3Cfg.endpoint != "" {
		opts = append(opts, WithCustomEndpoint(s3Cfg.endpoint))
	}

	if s3Cfg.accessKeyID != "" && s3Cfg.secretAccessKey != "" {
		opts = append(opts, WithCredentials(s3Cfg.accessKeyID, s3Cfg.secretAccessKey))
	}

	c, err := NewClient(ctx, s3Cfg.bucket, opts...)
	if err != nil {
		return "", nil, err
	}

	return s3Cfg.key, c, nil
}

// NewClient creates a new AWS S3 client.
func NewClient(ctx context.Context, bucketName string, opts ...Option) (*S3Client, error) {
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, nil))
	if err != nil {
		return nil, err
	}

	cfg := &s3Config{
		bucketName: bucketName,
	}
	for _, opt := range opts {
		err := opt(ctx, cfg)
		if err != nil {
			return nil, err
		}
	}

	awsOpts := []func(*awsConfig.LoadOptions) error{
		awsConfig.WithHTTPClient(httpClient),
		awsConfig.WithDefaultsMode(awsSdk.DefaultsModeInRegion),
	}

	// If the region wasn't set, set it to whatever is in the AWS_REGION envvar.
	if cfg.region == "" {
		cfg.region = os.Getenv("AWS_REGION")
	}

	if cfg.region != "" {
		awsOpts = append(awsOpts, awsConfig.WithRegion(cfg.region))
	}

	if cfg.accessKeyID != "" && cfg.secretAccessKey != "" {
		awsOpts = append(awsOpts,
			awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.accessKeyID, cfg.secretAccessKey, "")),
		)
	}

	if cfg.endpointURL != "" {
		customResolver := awsSdk.EndpointResolverWithOptionsFunc(func(service string, region string, optFns ...interface{}) (awsSdk.Endpoint, error) {
			if service == s3.ServiceID && region == cfg.region {
				return awsSdk.Endpoint{
					PartitionID:       "aws",
					URL:               cfg.endpointURL,
					SigningRegion:     cfg.region,
					HostnameImmutable: true,
				}, nil
			}

			return awsSdk.Endpoint{}, &awsSdk.EndpointNotFoundError{}
		})
		awsOpts = append(awsOpts, awsConfig.WithEndpointResolverWithOptions(customResolver))
	}

	baseConfig, err := awsConfig.LoadDefaultConfig(ctx, awsOpts...)
	if err != nil {
		return nil, fmt.Errorf("error loading aws cfg: %w", err)
	}

	ret := &S3Client{
		awsConfig:  baseConfig,
		cfg:        cfg,
		httpClient: httpClient,
	}

	return ret, nil
}

func (s *S3Client) getCallingConfig(ctx context.Context) (awsSdk.Config, error) {
	l := ctxzap.Extract(ctx)
	s._callingConfigOnce.Do(func() {
		s._callingConfig, s._callingConfigError = func() (awsSdk.Config, error) {
			if len(s.cfg.assumeRoles) == 0 {
				return s.awsConfig, nil
			}

			var creds *awsSdk.CredentialsCache
			for _, ar := range s.cfg.assumeRoles[:len(s.cfg.assumeRoles)] {
				stsConfig := s.awsConfig.Copy()
				if creds != nil {
					stsConfig.Credentials = creds
				}

				stsSvc := sts.NewFromConfig(stsConfig)
				newCreds := awsSdk.NewCredentialsCache(stscreds.NewAssumeRoleProvider(stsSvc, ar.roleARN, func(aro *stscreds.AssumeRoleOptions) {
					if ar.externalID != "" {
						aro.ExternalID = awsSdk.String(ar.externalID)
					}
				}))
				_, err := creds.Retrieve(ctx)
				if err != nil {
					l.Error("assume role error", zap.Error(err), zap.String("role_arn", ar.roleARN))
					return awsSdk.Config{}, fmt.Errorf("error while assuming rule: %s: %w", ar.roleARN, err)
				}
				creds = newCreds
			}

			stsConfig := s.awsConfig.Copy()
			if creds == nil {
				stsConfig.Credentials = creds
			}
			assumeRole := s.cfg.assumeRoles[len(s.cfg.assumeRoles)-1]
			callingStsSvc := sts.NewFromConfig(stsConfig)
			callingConfig := awsSdk.Config{
				HTTPClient:   s.httpClient,
				Region:       s.cfg.region,
				DefaultsMode: awsSdk.DefaultsModeInRegion,
				Credentials: awsSdk.NewCredentialsCache(stscreds.NewAssumeRoleProvider(callingStsSvc, assumeRole.roleARN, func(aro *stscreds.AssumeRoleOptions) {
					aro.ExternalID = awsSdk.String(assumeRole.externalID)
				})),
			}

			return callingConfig, nil
		}()
	})

	return s._callingConfig, s._callingConfigError
}

func (s *S3Client) newConfiguredS3Client(ctx context.Context) (*s3.Client, error) {
	callingConfig, err := s.getCallingConfig(ctx)
	if err != nil {
		return nil, err
	}

	s3svc := s3.NewFromConfig(callingConfig)
	location, err := s3svc.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: awsSdk.String(s.cfg.bucketName),
	})
	if err != nil {
		return nil, err
	}

	if location.LocationConstraint != "" {
		callingConfig.Region = string(location.LocationConstraint)
	} else {
		callingConfig.Region = s.cfg.region
	}

	return s3.NewFromConfig(callingConfig), nil
}

// ValidateBucket calls HeadBucket to ensure that the bucket exists and we have proper access.
func (s *S3Client) ValidateBucket(ctx context.Context) error {
	s3svc, err := s.newConfiguredS3Client(ctx)
	if err != nil {
		return err
	}

	_, err = s3svc.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket:              awsSdk.String(s.cfg.bucketName),
		ExpectedBucketOwner: nil,
	})
	if err != nil {
		return err
	}

	return nil
}

type ObjectInfo struct {
	Bucket        string
	Key           string
	ContentLength int64
	ETag          string
	Sha256Sum     string
	LastModified  time.Time
}

// ObjectInfo returns the object info from AWS S3.
func (s *S3Client) ObjectInfo(ctx context.Context, key string) (*ObjectInfo, error) {
	s3svc, err := s.newConfiguredS3Client(ctx)
	if err != nil {
		return nil, err
	}

	ret := &ObjectInfo{
		Bucket: s.cfg.bucketName,
		Key:    key,
	}

	headOut, err := s3svc.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket:       awsSdk.String(s.cfg.bucketName),
		Key:          awsSdk.String(key),
		ChecksumMode: s3Types.ChecksumModeEnabled,
	})
	if err != nil {
		return nil, err
	}

	ret.ETag = awsSdk.ToString(headOut.ETag)
	ret.ContentLength = awsSdk.ToInt64(headOut.ContentLength)
	ret.Sha256Sum = awsSdk.ToString(headOut.ChecksumSHA256)
	ret.LastModified = awsSdk.ToTime(headOut.LastModified)

	return ret, nil
}

// checkFilesize returns true if the filesize is valid.
func (s *S3Client) checkFilesize(fSize int64) bool {
	if s.cfg.maxFileSize > 0 && fSize > s.cfg.maxFileSize {
		return false
	}

	return true
}

// Get downloads a file from AWS S3.
func (s *S3Client) Get(ctx context.Context, key string) (io.Reader, error) {
	l := ctxzap.Extract(ctx)

	info, err := s.ObjectInfo(ctx, key)
	if err != nil {
		return nil, err
	}

	if !s.checkFilesize(info.ContentLength) {
		return nil, fmt.Errorf("file is too large. Max size allowed is %d bytes", s.cfg.maxFileSize)
	}

	s3svc, err := s.newConfiguredS3Client(ctx)
	if err != nil {
		return nil, err
	}

	in := &s3.GetObjectInput{
		Bucket: awsSdk.String(s.cfg.bucketName),
		Key:    awsSdk.String(key),
	}

	downloader := s3manager.NewDownloader(s3svc, func(dl *s3manager.Downloader) {
		dl.Concurrency = 1
		dl.PartSize = 5 * 1024 * 1024
	})

	b := s3manager.NewWriteAtBuffer(make([]byte, info.ContentLength))
	downloadedBytes, err := downloader.Download(ctx, b, in)
	if err != nil {
		return nil, err
	}

	l.Info(
		"downloaded file from s3",
		zap.String("bucket", s.cfg.bucketName),
		zap.String("key", key),
		zap.Int64("downloaded_bytes", downloadedBytes),
	)

	ret := bytes.NewBuffer(b.Bytes())

	return ret, nil
}

// ListObjects lists all the objects in the AWS S3 bucket.
func (s *S3Client) ListObjects(ctx context.Context) ([]string, error) {
	s3svc, err := s.newConfiguredS3Client(ctx)
	if err != nil {
		return nil, err
	}

	var keys []string
	var pageToken *string
	for {
		listOutput, err := s3svc.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            awsSdk.String(s.cfg.bucketName),
			ContinuationToken: pageToken,
		})
		if err != nil {
			return nil, err
		}

		for _, o := range listOutput.Contents {
			keys = append(keys, awsSdk.ToString(o.Key))
		}

		pageToken = listOutput.NextContinuationToken
		if pageToken == nil {
			break
		}
	}

	return keys, nil
}

type Upload struct {
	Data        []byte
	ContentType string
}

// Put uploads a file to AWS S3.
func (s *S3Client) Put(ctx context.Context, key string, r io.Reader, contentType string) error {
	l := ctxzap.Extract(ctx)

	s3svc, err := s.newConfiguredS3Client(ctx)
	if err != nil {
		return err
	}

	uploader := s3manager.NewUploader(s3svc)
	input := &s3.PutObjectInput{
		ACL:               s3Types.ObjectCannedACLPrivate,
		Bucket:            awsSdk.String(s.cfg.bucketName),
		Key:               awsSdk.String(key),
		Body:              r,
		ChecksumAlgorithm: s3Types.ChecksumAlgorithmSha256,
	}

	if contentType != "" {
		input.ContentType = awsSdk.String(contentType)
	}

	output, err := uploader.Upload(ctx, input)
	if err != nil {
		return err
	}

	l.Info(
		"uploaded file to s3",
		zap.String("calculated_checksum", awsSdk.ToString(output.ChecksumSHA256)),
		zap.String("etag", awsSdk.ToString(output.ETag)),
		zap.String("bucket", s.cfg.bucketName),
		zap.String("file_path", key),
	)

	return nil
}

func (s *S3Client) BucketName(ctx context.Context) string {
	return s.cfg.bucketName
}
