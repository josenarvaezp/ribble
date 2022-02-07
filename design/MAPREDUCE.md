## Limitations

From [1]: The file system can become a bottle neck
"Recall that each of the N map instances produces M output files — each destined for a different reduce instance. These files are written to a disk local to the computer used to run the map instance. If N is 1,000 and M is 500, the map phase produces 500,000 local files. When the reduce phase starts, each of the 500 reduce instances needs to read its 1,000 input files and must use a protocol like FTP to “pull” each of its input files from the nodes on which the map instances were run. With 100s of reduce instances running simultaneously, it is inevitable that two or more reduce instances will attempt to read their input files from the same map node simultaneously — inducing large numbers of disk seeks and slowing the effective disk transfer rate by more than a factor of 20."

From[1]: Some reducers might take a lot more time than others because of the skew of partition keys
"As described in “Parallel Database System: The Future of High Performance Database Systems,” [6] skew is a huge impediment to achieving successful scale-up in parallel query systems. The problem occurs in the map phase when there is wide variance in the distribution of records with the same key. This variance, in turn, causes some reduce instances to take much longer to run than others, resulting in the execution time for the computation being the running time of the slowest reduce instance."

Sources:
- [1] http://craig-henderson.blogspot.com/2009/11/dewitt-and-stonebrakers-mapreduce-major.html