# Coordinator

// TODO: move to coordinator
Starting the reducers: 
- to start the reducers we use an event source mapping with the SQS fifo queue to the lambda coordinator

- to avoid processing every event with a new lambda we set the batch window to max (5 min) and the maximum batch size to the number of mappers of a mod of that (only makes sense to set the mappers as the max if we can do the reading in 5 mins with a payload min of 256KB)