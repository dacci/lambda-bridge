package sqs

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/urfave/cli/v2"
)

var batchSize = 10
var endpointUrl string

var Command = &cli.Command{
	Name:      "sqs",
	Usage:     "receive message from a queue",
	ArgsUsage: "[queue name or URL]",
	Action:    action,
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:        "batch-size",
			Usage:       "the maximum number of items to retrieve in a single batch",
			Value:       batchSize,
			Destination: &batchSize,
		},
		&cli.StringFlag{
			Name:        "endpoint-url",
			Usage:       "override default URL with the given URL",
			Destination: &endpointUrl,
		},
	},
}

var svc *sqs.Client

func action(ctx *cli.Context) error {
	q := ctx.Args().First()
	if q == "" {
		q = os.Getenv("QUEUE_NAME")
	}
	if q == "" {
		return fmt.Errorf("queue name or url must be specified")
	}

	optFns := make([]func(*config.LoadOptions) error, 0)
	if endpointUrl != "" {
		optFn := config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           endpointUrl,
					SigningRegion: region,
				}, nil
			}))
		optFns = append(optFns, optFn)

		log.Printf("endpoint url override: %s", endpointUrl)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), optFns...)
	if err != nil {
		return err
	}

	svc = sqs.NewFromConfig(cfg)

	h, err := newHandler(q)
	if err != nil {
		return err
	}

	return h.run()
}
