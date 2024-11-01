package store

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var Client *s3.Client

var CourseImageBuckerName string = "course-image"
var AvatarBucketsName string = "avatar"

func ConnectS3() {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(os.Getenv("S3_ACCESS_KEY_ID"), os.Getenv("S3_SECRET_KEY"), "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create a custom transport to disable TLS verification
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Disable SSL verification
		},
	}

	// Parse S3_PATH_STYLE environment variable
	pathStyle := os.Getenv("S3_PATH_STYLE")
	usePathStyle := false // Default to false
	if pathStyle == "true" {
		usePathStyle = true
	} else if pathStyle == "false" {
		usePathStyle = false
	} else {
		log.Printf("S3_PATH_STYLE is not set to a recognized value. Defaulting to false.")
	}

	Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("S3_ENDPOINT"))
		o.HTTPClient = &http.Client{Transport: transport}
		o.UsePathStyle = usePathStyle // Enable path-style URLs for MinIO
	})

	// Check if the bucket exists
	bucketsName := []string{CourseImageBuckerName, AvatarBucketsName}
	for _, bucketName := range bucketsName {
		_, err = Client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
			Bucket: &bucketName,
		})
		if err != nil {
			// Check if the error is due to a non-existent bucket
			var notFoundErr *types.NotFound
			if errors.As(err, &notFoundErr) {
				// If the bucket does not exist, create it
				_, createErr := Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
					Bucket: &bucketName,
				})
				if createErr != nil {
					log.Fatalf("Failed to create bucket %s: %v", bucketName, createErr)
					return
				}
				log.Printf("Bucket %s created successfully.", bucketName)

				// Set the bucket policy to make it publicly readable
				policy := fmt.Sprintf(`{
					"Version": "2012-10-17",
					"Statement": [
						{
							"Effect": "Allow",
							"Principal": "*",
							"Action": "s3:GetObject",
							"Resource": "arn:aws:s3:::%s/*"
						}
					]
				}`, bucketName)

				_, policyErr := Client.PutBucketPolicy(context.TODO(), &s3.PutBucketPolicyInput{
					Bucket: &bucketName,
					Policy: &policy,
				})
				if policyErr != nil {
					log.Fatalf("Failed to set public read policy for bucket %s: %v", bucketName, policyErr)
					return
				}
				log.Printf("Public read policy set for bucket %s.", bucketName)
			} else {
				log.Fatalf("Failed to check bucket %s: %v", bucketName, err)
				return
			}
		}
	}
}
