package output

import (
	"github.com/Jeffail/benthos/v3/internal/docs"
	"github.com/Jeffail/benthos/v3/lib/log"
	"github.com/Jeffail/benthos/v3/lib/message/batch"
	"github.com/Jeffail/benthos/v3/lib/metrics"
	"github.com/Jeffail/benthos/v3/lib/output/writer"
	"github.com/Jeffail/benthos/v3/lib/types"
	"github.com/Jeffail/benthos/v3/lib/util/aws/session"
	"github.com/Jeffail/benthos/v3/lib/util/retries"
)

//------------------------------------------------------------------------------

func init() {
	Constructors[TypeAWSKinesis] = TypeSpec{
		constructor: fromSimpleConstructor(NewAWSKinesis),
		Version:     "3.36.0",
		Summary: `
Sends messages to a Kinesis stream.`,
		Description: `
Both the ` + "`partition_key`" + `(required) and ` + "`hash_key`" + ` (optional)
fields can be dynamically set using function interpolations described
[here](/docs/configuration/interpolation#bloblang-queries). When sending batched messages the
interpolations are performed per message part.

### Credentials

By default Benthos will use a shared credentials file when connecting to AWS
services. It's also possible to set them explicitly at the component level,
allowing you to transfer data across accounts. You can find out more
[in this document](/docs/guides/cloud/aws).`,
		Async:   true,
		Batches: true,
		FieldSpecs: docs.FieldSpecs{
			docs.FieldCommon("stream", "The stream to publish messages to."),
			docs.FieldCommon("partition_key", "A required key for partitioning messages.").IsInterpolated(),
			docs.FieldAdvanced("hash_key", "A optional hash key for partitioning messages.").IsInterpolated(),
			docs.FieldCommon("max_in_flight", "The maximum number of messages to have in flight at a given time. Increase this to improve throughput."),
			batch.FieldSpec(),
		}.Merge(session.FieldSpecs()).Merge(retries.FieldSpecs()),
		Categories: []Category{
			CategoryServices,
			CategoryAWS,
		},
	}

	Constructors[TypeKinesis] = TypeSpec{
		constructor: fromSimpleConstructor(NewKinesis),
		Status:      docs.StatusDeprecated,
		Summary: `
Sends messages to a Kinesis stream.`,
		Description: `
## Alternatives

This output has been renamed to ` + "[`aws_kinesis`](/docs/components/outputs/aws_kinesis)" + `.

Both the ` + "`partition_key`" + `(required) and ` + "`hash_key`" + ` (optional)
fields can be dynamically set using function interpolations described
[here](/docs/configuration/interpolation#bloblang-queries). When sending batched messages the
interpolations are performed per message part.

### Credentials

By default Benthos will use a shared credentials file when connecting to AWS
services. It's also possible to set them explicitly at the component level,
allowing you to transfer data across accounts. You can find out more
[in this document](/docs/guides/cloud/aws).`,
		Async:   true,
		Batches: true,
		FieldSpecs: docs.FieldSpecs{
			docs.FieldCommon("stream", "The stream to publish messages to."),
			docs.FieldCommon("partition_key", "A required key for partitioning messages.").IsInterpolated(),
			docs.FieldAdvanced("hash_key", "A optional hash key for partitioning messages.").IsInterpolated(),
			docs.FieldCommon("max_in_flight", "The maximum number of messages to have in flight at a given time. Increase this to improve throughput."),
			batch.FieldSpec(),
		}.Merge(session.FieldSpecs()).Merge(retries.FieldSpecs()),
		Categories: []Category{
			CategoryServices,
			CategoryAWS,
		},
	}
}

//------------------------------------------------------------------------------

// NewAWSKinesis creates a new Kinesis output type.
func NewAWSKinesis(conf Config, mgr types.Manager, log log.Modular, stats metrics.Type) (Type, error) {
	return newKinesis(TypeAWSKinesis, conf.AWSKinesis, mgr, log, stats)
}

// NewKinesis creates a new Kinesis output type.
func NewKinesis(conf Config, mgr types.Manager, log log.Modular, stats metrics.Type) (Type, error) {
	return newKinesis(TypeKinesis, conf.Kinesis, mgr, log, stats)
}

func newKinesis(name string, conf writer.KinesisConfig, mgr types.Manager, log log.Modular, stats metrics.Type) (Type, error) {
	kin, err := writer.NewKinesis(conf, log, stats)
	if err != nil {
		return nil, err
	}
	var w Type
	if conf.MaxInFlight == 1 {
		w, err = NewWriter(name, kin, log, stats)
	} else {
		w, err = NewAsyncWriter(name, conf.MaxInFlight, kin, log, stats)
	}
	if err != nil {
		return w, err
	}
	return NewBatcherFromConfig(conf.Batching, w, mgr, log, stats)
}

//------------------------------------------------------------------------------
