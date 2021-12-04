# SQS (Simple Queue Service)

Benefits:
- SQS uses multiple servers to host your data. Standard queues supports at least once message delivery while FIFO supports exactly once. 
- SQS is a highly available that supports high concurrency
- SQS is auto-scaling and transparent. It can handle load increases or spikes automatically. Given that we are creating an ad-hoc framweork, this sounds ideal.
- SQS can hold a pointer to a s3 bucket. (this could be useful)

Standard queues vs FIFO:
- Standard queues have (nerly) unlimited thoughput for sendin, receiving and deleting messages
- At-Least-Once Delivery for standard queues means we can have the data multiple times in the queue, hence, the fraework needs to be idempotent (or figure out the case when this happpens to avoid processing things twice)
- We don't care about order so it doesn't matter if we don't have FIFO behavior
- If we think 300 api calls per second are enough in terms of throughput limit, of if we think we can use batching, we can consider using FIFO

Queues configuration:
- Visibility timeout: "The length of time that a message received from a queue (by one consumer) won't be visible to the other message consumers."  0 seconds to 12 hours.
- Message retention period: amount of time messages can stay in the queue. (MAx 14 days so can't run a job that lasts longer than 14 days)
- Maximum message size 1 KB to 256 KB
- Receive message wait time – "The maximum amount of time that Amazon SQS waits for messages to become available after the queue gets a receive request" (short or long polling -> for the mappers we want short and reducers want longer)
- Delivery delay – "The amount of time that Amazon SQS will delay before delivering a message that is added to the queue." (this could be useful for sending jobs once mappers are done -> as long as we have an estimate of how long mappers will take?)

Dead-letter queues:
A dead letter queue is a queue where messages that are not consumed successfully can go to. 

## Using AWS with lambda: (https://docs.aws.amazon.com/lambda/latest/dg/with-sqs.html)
- Lambda polls the queue for messages and sends messages to lambda functions as an event. Lambda can read messages in batches and invoke a lamda function per message. When lamda finishes with a message, it deletes the message from the queue. 
- batch window: To avoid invoking a lamda with few records, the event source mapping can be configured to buffer records up to a time limit. If the time limit happens before the buffer is full, the batch is processed by lamda.
- Lambda polls message from the queue until the batch window time limit is reached, there are enough records in the buffer or the payload limit is reached. (6 MB)
- When a batch is sent to lamda, the batch remain in the queue, but is hidden according to the queue's visibility timeout. Bear in mind that if lambda takes longer to process messages than the visibility timeout, the messaages will be read by another lambda so special care is needed for that case. 
- A message becomes visible again (i.e other consumers can read the message again) if the function is throttled, it returns an error or doesn't respond. 

## Scaling and processing
- Lambda starts reading up to 5 batches and sends them to the function
- If more messages are present, lambda can scale to up to 60 function a minute until it reaches 60
- Q: Can reserve concurrency be used to deal with the scaling behavior? I think lambdas will be provisioned quicker but still at the rate of 60

## Quotas
- unlimited number of messages (backlog)
- 120,000 inflight messages (messages recieved by consumers but not yet deleted)
- nearly unlimited api requests for standard queues
- message size: min of 1kb max of 256kb

Sources:
https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/welcome.html
