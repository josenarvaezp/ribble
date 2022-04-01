What we want to do -> Problem -> existing solutions that address some part of the problem -> my solution

## Introduction
## What we want to achieve:
- Data processing in the cloud
    - investigate cloud computing and define the trend of using more cloud services and moving away from on-prem infrastructure
    - Define that we want to achieve this with serverless computing
- Define the type of data processing we want to have 
- Easy to use for data analysts

## Background
- big data and data processing
    - Big Data
    - Batch Processing
        - Batch Processing Model
        - Ridiculous paralelizable one instruction multiple data
        - Apache Hadoop
            - components: HDFS, YARN and MapReduce as its native batch processing engine
            - MapReduce (Google's implementation)
            - Limitations: steep learning curve and having a cluster of commodity servers
            - Doesn't use in-memory to process data unlike spark -> slower
    - Stream Processing 
        - Storm (trident at least once to exactly once)
        - Samza (uses kafka and checkpointing system using a local key-value store)
    - Hybrid Processing
        - Spark
            - Spark processes everything in memory using RDDs 
            - uses micro batches in the stream
            - "Unlike MapReduce, Spark processes all data in-memory, only interacting with the storage layer to initially load the data into memory and at the end to persist the final results. All intermediate results are managed in memory."
        - Flink
            - Treats batches to be streams with finite boundaries
            - Kappa architecutre
    - cloud computing
        - serverless computing
        - SQS 
        - S3
    - Data processing in the cloud
        - write comparison of different existing solutions
        - Overall, this solved the issue of doing cluster administration and they show good performance in serverless architectures but they are still diffuclt to use as they implement map reduce interfaces -> user still needs to be aware of the map reduce paradigm

SOLUTION
- Exaplain Serverless architectures and why it is a good idea
- Explain why we will use AWS and describe services to use
- Explain why go and describe the methodology we used (clean architecture TODO)
- Explain overall solution 
- Explain MapSum data structure 
    - For the deduplication mechanism we can say we got inspiration from TCP -> duplicates are identified by package numbers (batch in our case)
    - At least once to exactly once
    - Make sure to explain how idempotency is reached
    - exaplain why we couldn't use an evet source mapper as we needed to keep track of duplicates in DynamoDB which would eventually become a bottleneck
    - explain that for the data shuffling we are doing hash-key load balancing
- Explain Sum data stucture TODO


TODOs research
- apache storm trident paper for exactly one delivery

Sources:
TheRealTimeBigDataProcessingFrameworkAdvantagesandLimitations.pdf

type Sum int

func ParseInt(n int) Sum {
	return Sum(n)
}