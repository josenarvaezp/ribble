
## Quotes
- A significant feature we do not get is the co-location of our data and the processing of that data
- For each file in S3, the driver triggers an SNS event that ultimately triggers a mapper function.
- Downstream, the reducers will need to know how many total mappers were executed so that they know when to begin their work. 
    N: And this is why the lenght of the mappers is added into the payload
- I prefer using SNS in these cases since the behavior is asynchronous by default. 
    N: if we can invoke lamdas async then we don't really need an SNS


## Most important takeaway
Serverless mapreduce is mostly limited by the size of the mappers output since all needs to fit in the memory of the final aggregator reducer. This last reducer is needed because there was no shuffling or sorting involved.

## Source
Serverless Design Patterns and Best Practices.pdf