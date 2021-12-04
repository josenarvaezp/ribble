# Driver

The driver is responsible of starting the framework. This driver is the only component of the system that is not serverless, meaning that user will need to run this from either their local computers or a vm hosted in the cloud. However, this driver is not responsible of coordinating the different components after the startup, and for that reason the user does not need to rely on the computer running the driver. 

One of the responsabilities of the driver is to setup the cloud resources needed to run the processing framework. This includes setting up the following resources:
1. The kinesis stream used by the mappers to send its output to the reducers.
2. A Dynamo DB table used by the coordinator to orchestrate the phases and resources of the framework.
3. An S3 bucket for the job. This will be used to store the final results and it will also be used by the reducers and coordinators to save their state if needed.
4. Upload the images of the mapper function to ECR. This image contains the function specified by the user which will be used to process the data. Note that the images for the rest of the serverless functions needed by the framework such as the coordinator and reducers are already publicly available in ECR. 

Once this resources are created the framwork is ready to start processing data. The driver executes the following steps:

1. From the configuration file, it gets the buckets and keys of the data to process. 
2. It splits the data into chunks of similar size (I thinnk default is 128 MB but it could probably be changed)
3. Each split will be processed by a different map process. The driver is responsible of starting the lambda functions that will procces the input splits. Lambda will receive the invokations and will start provision functions. It is important to notice that lambda will first invoke a maximum of 1000 lambdas (depending on the region) to run concurrently. This means that at the startup the framework will be able to handle approximatly 128 GB concurrently and will be able to scale to (TODO: how much it scales per minute?) process additional x GB per minute.
4. After the mapper functions have been invoked, the driver starts up the coordinator, a serverless function that will be resonsible of orchestrating the framework.
5. At this point the driver finishes up the execution and the machine used to run it becomes free.

TODO:
-> When we speak about the implementation of our driver, explain that the driver creates and invokes the processes and then it invokes a coordinator responsible of orchestrating the job for the rest of the duration. Given that the coordintaro runs in serveleress funciton, there exists a mechanism tosave its state one it reaches a timeout maximum. This allows the framewor to run in a completly serveleress manner and a job can last for more than 15 minutes. 