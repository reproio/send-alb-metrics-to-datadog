package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"os"
)

func main() {
	if os.Getenv("LOCAL_INVOKE_GZ_PATH") == "" {
		lambda.Start(handler)
	} else {
		if err := devHandler(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}

func devHandler() error {
	fmt.Println("start download from s3")
	processor, err := NewProcessor()
	if err != nil {
		return err
	}

	f, err := os.Open(os.Getenv("LOCAL_INVOKE_GZ_PATH"))
	if err != nil {
		return err
	}
	defer f.Close()

	err = processor.ProcessLogfile(f, f.Name())
	if err != nil {
		return err
	}

	return nil
}

func handler(s3Event events.S3Event) error {
	fmt.Println("start handler")

	processor, err := NewProcessor()
	if err != nil {
		return err
	}

	sess := session.Must(session.NewSession(aws.NewConfig()))

	for _, record := range s3Event.Records {
		fmt.Printf("%+v\n", record)

		svc := s3.New(sess)
		obj, err := svc.GetObject(&s3.GetObjectInput{Bucket: aws.String(record.S3.Bucket.Name), Key: aws.String(record.S3.Object.Key)})
		if err != nil {
			return err
		}

		fmt.Println("finish download from s3")

		err = processor.ProcessLogfile(obj.Body, record.S3.Object.Key)
		if err != nil {
			return err
		}
		fmt.Println("finish processLogFile")
	}

	fmt.Println("finish handler")
	return nil
}
