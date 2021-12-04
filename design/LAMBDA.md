# Functionas as a service

## What is function as a service

Benefits of lamda functions:
- Pay-as-you-go service
- scaling
Lambda scales automatically in response to invocations. Limitations on how many lambdas can run concurrently (1000 as default)
Burst concurrency is the amount of lambdas initialized with the initial burst of traffic. Initialy it can reach between 500 and 3000 depending on the region and then increase by 500 per minute. This is a limitation when using AWS Lamda functions. Given that the worloads are highly parallelizable, an ideal situation would have been to initialize as many lamdas as needed to process everything in parallel. However, AWS needs to protect the resources as they are shared across all customers using the service. When Lambda reaches its concurrency limit, new requests will fail with a throttled error so this should be handled by the framework. 
- Provisioned concurrency
When a lambda function is invoked, the lamda service needs to initialize the function before it can process events. This start up time can increase latency significantly. This is known as a cold start, and it is a limitation of using serverless functions, however, there are some solutions to decrease its impact. Reserve concurrency can be used to specify the amount of concurrency a function can use, and this limits other functions to use the concurrency reserved for the specific function. Provisioned concurrency initializes instances before time to make sure functions are ready to recieve events. 

- Event source mapping
Event source mappings gives us the ability to invoke lambdas when SQS messages arrive to a queue. This events from the queue can be processed in batches. If a whole batch fails to be processed, all the events in the batch are retried. Given that SQS have eventual consistency (define this), event source mappings can process the same item from a queue twice even when no errors occur. This means that the framework should handle this case. (See the SQS.md for more information on this)

- Asynchronous invocations 
Asynchronous invocations allows us to have a serverless system given that we do not want to have a long running system waiting for a response from the system. The frameworks starts within the local computer of a developer (or a VM in the cloud) but given the need to have a fully serverless framework, we don't want to depend on that instance appart from the job initialization phase. This however comes with challenges when it comes to handling error and responses from the functions.

Configuring asynchronous invocations This also has details on how to handle errors: https://docs.aws.amazon.com/lambda/latest/dg/invocation-async.html

For asynchronous invokations, events are sent to lamda's event queue and lambda handles the events according to the concurrency limit. On function errors, lambda retries the event up to 3 times. If the concurrency limit for a function is reached, additinal events will be throttled. When this happens, lambda sends the event back to the queue and tries to porocess the events for up to 6 hours. However, when there are many events in the queue (more than the concurrency option) lambda will reduce the rate at which it reads events from the queue and will increase the retry interval to avoid throttling. Given that lambda is working with an eventual consistent queue, the same event can be recieved by a lambda function more than once, so it is important to be able to handle repeated events. Lambda allows us to send event invokations to different AWS services including SQS. This allows us to send succesfull and failed requests to different queues if needed. A dead-letter queue can also be used to send discarded events. 

To configure error handling for asynchronous invocation we can modify the maximum age of event (the time an event is allowed to stay in the queue do that it can be processed again) and the retry attempts which specifies the number of retries a failed event can be do. Notice that when an event reches the maximum retries, the event is discared (an dead-letter queue can be configured to handle those events or a failed-queue destination). 

- Architectures
arm64 vs x86_64 – 64-bit x86 -> in general arm64 is faster which will make the platform quicker. 
There is not a native Go runtime for arm64 currently but we can run Go with the provided.al2 runtime:
https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html#golang-package-al2
https://docs.aws.amazon.com/lambda/latest/dg/go-image.html#go-image-al2


- Error Handling for asynchrous invokations
When a Lambda function has reached its concurrency limit new incoming events will be throttled. This events are returned to the queue and will be tried to run again for up to 6 hours. The maximum age of events and the number of retries can be configured. Failed invokations can also be sent to other services such as SQS. (https://docs.aws.amazon.com/lambda/latest/dg/invocation-async.html#invocation-async-destinations)


- Quotas
Concurrent executions: 1,000 (but can be increased)
Memory Allocation: 128mb to 10,240mb
Timeout 15 min
Burst concurrency: 500 - 3000 (varies per Region)
Invocation payload: 256 KB (asynchronous)
Execution processes/threads 1,024

- Running a lambda function with Go (https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/lambda-go-example-run-function.html)

Error handling: 
- When invoking async lambdas, only a status code is returned
- "Invocation errors include an error type and status code in the response that indicate the cause of the error."
https://docs.aws.amazon.com/lambda/latest/dg/golang-exceptions.html
https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html#API_Invoke_Errors
https://docs.aws.amazon.com/lambda/latest/dg/invocation-retries.html

Implementation Notes:
- Before calling a function check the state of it. It should be in active before we can call it. (this is if we call the function directly)
Sources: 
- https://docs.aws.amazon.com/lambda/latest/dg/welcome.html AWS lambda development guide
