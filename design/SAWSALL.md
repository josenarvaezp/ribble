# Sawsall

Sawsall is a procedural programming language developed by Google used to process large datasets that often do not fit in traditional relational databases. Sawsall process data stored on GFS (Google File System) that is formatted using protocol buffers. It runs on top of MapReduce which allows Sawsall to execute operations for different data items in parallel abstracting the distributed aspects such as scheduling and fault tolerance. Sawsall offers less expressivness than SQL queries but it reiterates many data analytic applications do not need the sophistication that SQL queries provide. While it recognizes procedural programming languagues can achieve the same as Sawsall, languages like Python do not have the facility to distribute computation to thousands of computers without adding a burden to programmers. Additionally, Sawsall assures that compared to C++ code running MapReduce, it can be much simpler and shorter. [TODO: add number here] However, Sawsall can only be used to do embarassingly parallel data processing with a restricted set of operations.  

Data analytics in Sawsall is compromised of a query and an aggregation stage, where the query runs in the map phase of MapReduce and the aggregation in the reduce phase. The filter is defined by users with the procedural programming language and acts on one record at a time in isolation while the aggregation can be chosen from a set of predefined aggregators. These characteristics abstact the fact that there are potentially thousands of records being processed at the same time. 

Sawsall focuses on commutative and associative query and aggregation operations which allows the execution to process data and group intermediate values in any order. Some of the aggregators used in Sawsall are the following:
- Collection: offers a list of emmited values
- Sample: offers a sample from a collection of emmited values
- Sum: adds up the emmited values
- Maximum: returns the maximum value emmited
- Top: returns the most frequent values emmited

TODO: add example of Sawsall program

The execution of such a program is not complicated. First a job request is received by the system, after validating the Sawsall source code, it is sent to the machines for execution. These machines compile the Sawsall source code and run it to process individual data items. The values emmited are then aggregated locally, similar to the combine pahse in MapReduce, before sending them to the corresponding machines. The aggregator machines collect the intermediate output and aggregates the values accordingly and sorts the output before writting it back to GFS. At the end of the job there is a file in GFS for each aggregator machine.

In Domain-specific Properties of the Language






