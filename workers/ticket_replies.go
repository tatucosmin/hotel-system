package workers

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/tatucosmin/hotel-system/config"
	"github.com/tatucosmin/hotel-system/store"
)

func SaveTicketToS3(ctx context.Context, ticketId uuid.UUID, ticketReplyStore *store.TicketReplyStore, cfg *config.Config) error {
	ticketReplies, err := ticketReplyStore.ByTicketId(ctx, ticketId)
	if err != nil {
		return fmt.Errorf("failed to get ticket replies: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	for _, reply := range *ticketReplies {
		buf.WriteString(fmt.Sprintf("Creator: %s\nMessage: %s\n\n", reply.Creator, reply.Message))
	}

	s3FilePath := fmt.Sprintf("tickets/ticket_%s.txt", ticketId)

	fmt.Println(buf.String())

	_, err = cfg.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(cfg.S3Bucket),
		Key:    aws.String(s3FilePath),
		Body:   bytes.NewReader(buf.Bytes()),
	})

	if err != nil {
		return fmt.Errorf("failed to upload file to s3: %w", err)
	}

	return nil

}
