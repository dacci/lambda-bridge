package sqs

import (
	"context"
	"log"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/dacci/lambda-bridge/util"
)

type handler struct {
	queueUrl string
	arn      string
	region   string
}

func newHandler(q string) (*handler, error) {
	u, err := url.Parse(q)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		resp, err := svc.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
			QueueName: aws.String(q),
		})
		if err != nil {
			return nil, err
		}

		n := q
		q = aws.ToString(resp.QueueUrl)
		log.Printf("queue name %s resolved to %s", n, q)
	}

	resp, err := svc.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(q),
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameQueueArn,
		},
	})
	if err != nil {
		return nil, err
	}

	a, err := arn.Parse(resp.Attributes[string(types.QueueAttributeNameQueueArn)])
	if err != nil {
		return nil, err
	}

	return &handler{
		queueUrl: q,
		arn:      a.String(),
		region:   a.Region,
	}, nil
}

func (h *handler) mapToEvent(msgs []types.Message) *events.SQSEvent {
	r := make([]events.SQSMessage, 0, len(msgs))

	for _, msg := range msgs {
		m := events.SQSMessage{
			MessageId:              aws.ToString(msg.MessageId),
			ReceiptHandle:          aws.ToString(msg.ReceiptHandle),
			Body:                   aws.ToString(msg.Body),
			Md5OfBody:              aws.ToString(msg.MD5OfBody),
			Md5OfMessageAttributes: aws.ToString(msg.MD5OfMessageAttributes),
			Attributes:             msg.Attributes,
			MessageAttributes:      make(map[string]events.SQSMessageAttribute),
			EventSourceARN:         h.arn,
			EventSource:            "aws:sqs",
			AWSRegion:              h.region,
		}

		for k, v := range msg.MessageAttributes {
			m.MessageAttributes[k] = events.SQSMessageAttribute{
				StringValue:      v.StringValue,
				BinaryValue:      v.BinaryValue,
				StringListValues: v.StringListValues,
				BinaryListValues: v.BinaryListValues,
				DataType:         aws.ToString(v.DataType),
			}
		}

		r = append(r, m)
	}

	return &events.SQSEvent{
		Records: r,
	}
}

func (h *handler) deleteMessageBatch(msgs []events.SQSMessage) (*sqs.DeleteMessageBatchOutput, error) {
	e := make([]types.DeleteMessageBatchRequestEntry, 0, len(msgs))
	for _, m := range msgs {
		e = append(e, types.DeleteMessageBatchRequestEntry{
			Id:            aws.String(m.MessageId),
			ReceiptHandle: aws.String(m.ReceiptHandle),
		})
	}

	return svc.DeleteMessageBatch(context.TODO(), &sqs.DeleteMessageBatchInput{
		QueueUrl: aws.String(h.queueUrl),
		Entries:  e,
	})
}

func (h *handler) run() error {
	for util.Running {
		resp, err := svc.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl: aws.String(h.queueUrl),
			AttributeNames: []types.QueueAttributeName{
				types.QueueAttributeNameAll,
			},
			MessageAttributeNames: []string{
				"All",
			},
			MaxNumberOfMessages: int32(batchSize),
			WaitTimeSeconds:     1,
		})
		if err != nil {
			return err
		}

		if len(resp.Messages) == 0 {
			continue
		}

		event := h.mapToEvent(resp.Messages)
		err = util.InvokeLambda(event)

		if _, err := h.deleteMessageBatch(event.Records); err != nil {
			log.Print(err)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
