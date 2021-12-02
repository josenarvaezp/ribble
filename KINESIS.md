# Kinesis Data Streams

Kinesis can be used for rapid and continous data intake and aggregation. Kinesis streams are composed of data records which are the units of data. Thid data records are distributed into different shards of the same stream. A producer writes data into the shar while a consumer reads from it.

## Advantages
- real time processing -> don't need to wait for batches

## Creating data streams
Because of the aws limit services, we cannot create streams that fast so to imporve performance we can create streams before the framework starts processing data. To do that we need to know: number_of_shards = max(incoming_write_bandwidth_in_KiB/1024, outgoing_read_bandwidth_in_KiB/2048) READ: https://docs.aws.amazon.com/streams/latest/dev/amazon-kinesis-streams.html

"The data capacity of your stream is a function of the number of shards that you specify for the stream. The total capacity of the stream is the sum of the capacities of its shards."

## Producers
To send a record to a stream, we need to specify the stream, partition key and the data. The partition key is used to determine the shard we write to. It is important to have a large numer of partition keys to distribute them evenly across the shards of the stream. 

## Using Kinesis with lambda 
https://docs.aws.amazon.com/lambda/latest/dg/with-kinesis.html
- use batch window (up to 5 minutes) to avoid using small batches by configuing the event source
- event source mapping can be configured to retry with a smaller batch size, limit the number of retries, or discard records that are too old.
- event source mapping can be configured to send failed batches to sqs or SNS
- We can increase concurrency by processing multile batches from each shard in parallel (up to 10 batches per shard) configure -> ParallelizationFactor
    ** "Note that parallelization factor will not work if you are using Kinesis aggregation"
- Good thing is that we know Max concurrency levels before we start
- batch limit is 10,000 records or a payload of 6MB

### Error handling
- On throttling errors, lamdba retries until records are expired (24hrs). If the errors are function errors, they are processed until max level of retries are reached. And at that point, we can send the failed events to a DLQ if we configure it beforehand. For function errors we can split the data into 2 batches to localize the error or succeed if the error was because of a timeout. A failing record can stop the processing of a shar so we need to specify maximum record age and maximum number of retries in our source mapping. 

### Time windows
- Tumbling windows open and close at regular intervals. Lambdas are stateless, but this allows us to mantain the state across multiple invokations. 
- the shared state can be of 1MB and if it reaches the maximum level the window stops early. 
- "In each window, you can perform calculations, such as a sum or average, at the partition key level within a shard." 
This means that if we have all values relating a key in the same shard using the same partition key we can do aggregation by key. We only need to make sure that each key uses the same partition key and that the key is used exactly once. 
- Tumbling windows are configured with the event source mapping
- "Tumbling windows fully support the existing retry policies maxRetryAttempts and maxRecordAge."

### Reporting batch item failures
- If a batch fails, by default, lambda retries the whole batch again. This means that we need to process again events that were valid. To avoid this, we can add ReportBatchItemFailures to FunctionResponseTypes in the event source mapping. We then need to include ReportBatchItemFailures in the StreamsEventResponse of our lambda. If BisectBatchOnFunctionError is also on, lambda only retries the remaining records. Important to notice that with this technique we could be processing errors twice, to avoid this use checpoint strategy discussed above.

## Quotas 
- No upper quota on number of streams
- default shards per stream is 500
- standard stream can do 1000 record writes or 1MB per second (but up to 5GB with 5000 shards)
- Max data payload is 1MB
- Read with GetRecords can get 10 MB or 10,000 records per shard
- 5 transactions per second per shard
- 2 MB per second via GetRecords. So a GetRecords that gets 10MB it will take 5 seconds

API limits:
- Create stream limited to 5 transactions per second
- At most 5 streams in creating state


## Data Processing tutorial (https://data-processing.serverlessworkshops.io/)
- ERROR HANDLING WITH CHECKPOINT AND BISECT ON BATCH https://data-processing.serverlessworkshops.io/stream-processing/05-extra-credit/05-01-eh-cp-bb.html This basically creates a checkpoint of the processed data so in case of failure we only re-process new events
- ENHANCED FAN OUT https://data-processing.serverlessworkshops.io/stream-processing/05-extra-credit/05-02-efo.html If we don't want to share the shard connection between multiple consumers to increase data latency and thoguhput (2MB/sec), we can add a consumer to the data stream: 
```
aws kinesis register-stream-consumer --consumer-name con1 \
--stream-arn arn:aws:kinesis:us-east-2:123456789012:stream/lambda-stream
```
And use the consumer name for the event source mapping instead of the lambda name

Sources: 
- TODO: https://aws.amazon.com/streaming-data/
- https://docs.aws.amazon.com/streams/latest/dev/introduction.html
- https://data-processing.serverlessworkshops.io/