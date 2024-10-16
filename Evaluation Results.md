# Evaluation Results

## Environment

Our experiments are conducted on a public Cloud platform using three SA5.4XLARGE32 VM nodes.

+ Each node has 16 vCPUs, 32GB DRAM, and a 20GB general-purpose SSD cloud disk. 
+ One node serves as the client, and the other two serve as the data store nodes. 
+ The communication latency between each node approximates 3ms. 
+ We use MongoDB Community Server 7.0 and Redis 7.2.3 as the underlying data stores in our experiments.



## Performance of Distributed Transactions

Setup:

+ Latency = 3ms

+ RecordCount  = 1000000

+ OperationCount = 100000

+ TxnGroup = 6

+ LeaseTime = 100ms

### Workload F

#### mongo - mongo

##### Cherry Garcia

+ 8

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 8
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 1m24.170438017s
COMMIT - Takes(s): 84.1, Count: 16236, OPS: 193.0, Avg(us): 18222, Min(us): 0, Max(us): 32991, 50th(us): 18367, 90th(us): 25311, 95th(us): 25807, 99th(us): 29215, 99.9th(us): 31727, 99.99th(us): 32751
COMMIT_ERROR - Takes(s): 84.0, Count: 428, OPS: 5.1, Avg(us): 11119, Min(us): 3544, Max(us): 24383, 50th(us): 11143, 90th(us): 16063, 95th(us): 18431, 99th(us): 20191, 99.9th(us): 24383, 99.99th(us): 24383
READ   - Takes(s): 84.2, Count: 99402, OPS: 1181.0, Avg(us): 3625, Min(us): 4, Max(us): 16343, 50th(us): 3595, 90th(us): 3875, 95th(us): 3981, 99th(us): 4147, 99.9th(us): 14351, 99.99th(us): 15047
READ_ERROR - Takes(s): 83.9, Count: 598, OPS: 7.1, Avg(us): 9476, Min(us): 6572, Max(us): 12359, 50th(us): 10751, 90th(us): 11239, 95th(us): 11439, 99th(us): 12239, 99.9th(us): 12343, 99.99th(us): 12359
Start  - Takes(s): 84.2, Count: 16672, OPS: 198.1, Avg(us): 24, Min(us): 13, Max(us): 860, 50th(us): 18, 90th(us): 30, 95th(us): 39, 99th(us): 167, 99.9th(us): 298, 99.99th(us): 657
TOTAL  - Takes(s): 84.2, Count: 214746, OPS: 2551.3, Avg(us): 9215, Min(us): 0, Max(us): 64383, 50th(us): 3563, 90th(us): 39487, 95th(us): 43423, 99th(us): 47583, 99.9th(us): 52575, 99.99th(us): 58815
TXN    - Takes(s): 84.1, Count: 16236, OPS: 193.0, Avg(us): 40288, Min(us): 18016, Max(us): 63359, 50th(us): 40031, 90th(us): 47295, 95th(us): 47839, 99th(us): 51775, 99.9th(us): 56575, 99.99th(us): 62175
TXN_ERROR - Takes(s): 84.0, Count: 428, OPS: 5.1, Avg(us): 33059, Min(us): 21664, Max(us): 54271, 50th(us): 32799, 90th(us): 39871, 95th(us): 40159, 99th(us): 43999, 99.9th(us): 54271, 99.99th(us): 54271
TxnGroup - Takes(s): 84.1, Count: 16664, OPS: 198.0, Avg(us): 40093, Min(us): 18080, Max(us): 64383, 50th(us): 39999, 90th(us): 47231, 95th(us): 47839, 99th(us): 51935, 99.9th(us): 58239, 99.99th(us): 63423
UPDATE - Takes(s): 84.2, Count: 49536, OPS: 588.5, Avg(us): 4, Min(us): 1, Max(us): 771, 50th(us): 3, 90th(us): 5, 95th(us): 6, 99th(us): 15, 99.9th(us): 183, 99.99th(us): 353
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch     428

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  348
  read failed due to unknown txn status  242
rollback failed
  version mismatch  8
```

+ 16

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 16
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 42.614824831s
COMMIT - Takes(s): 42.6, Count: 15865, OPS: 372.5, Avg(us): 18411, Min(us): 0, Max(us): 32895, 50th(us): 18751, 90th(us): 24991, 95th(us): 26351, 99th(us): 29807, 99.9th(us): 30511, 99.99th(us): 32431
COMMIT_ERROR - Takes(s): 42.6, Count: 791, OPS: 18.6, Avg(us): 10957, Min(us): 3538, Max(us): 22959, 50th(us): 11271, 90th(us): 15647, 95th(us): 18735, 99th(us): 19375, 99.9th(us): 22687, 99.99th(us): 22959
READ   - Takes(s): 42.6, Count: 98918, OPS: 2321.3, Avg(us): 3671, Min(us): 4, Max(us): 16279, 50th(us): 3657, 90th(us): 3955, 95th(us): 4045, 99th(us): 4255, 99.9th(us): 14495, 99.99th(us): 15311
READ_ERROR - Takes(s): 42.5, Count: 1082, OPS: 25.4, Avg(us): 9570, Min(us): 6456, Max(us): 13055, 50th(us): 10863, 90th(us): 11423, 95th(us): 11543, 99th(us): 11887, 99.9th(us): 12575, 99.99th(us): 13055
Start  - Takes(s): 42.6, Count: 16672, OPS: 391.2, Avg(us): 31, Min(us): 13, Max(us): 1025, 50th(us): 25, 90th(us): 41, 95th(us): 57, 99th(us): 223, 99.9th(us): 452, 99.99th(us): 880
TOTAL  - Takes(s): 42.6, Count: 213333, OPS: 5006.1, Avg(us): 9292, Min(us): 0, Max(us): 64095, 50th(us): 3611, 90th(us): 37855, 95th(us): 44159, 99th(us): 48543, 99.9th(us): 53119, 99.99th(us): 59455
TXN    - Takes(s): 42.6, Count: 15865, OPS: 372.5, Avg(us): 40977, Min(us): 15184, Max(us): 63423, 50th(us): 40959, 90th(us): 48191, 95th(us): 48735, 99th(us): 52383, 99.9th(us): 56703, 99.99th(us): 60223
TXN_ERROR - Takes(s): 42.6, Count: 791, OPS: 18.6, Avg(us): 33385, Min(us): 22144, Max(us): 48479, 50th(us): 33439, 90th(us): 40319, 95th(us): 40991, 99th(us): 44543, 99.9th(us): 47135, 99.99th(us): 48479
TxnGroup - Takes(s): 42.6, Count: 16656, OPS: 391.0, Avg(us): 40596, Min(us): 17728, Max(us): 64095, 50th(us): 40863, 90th(us): 48127, 95th(us): 48767, 99th(us): 52639, 99.9th(us): 59295, 99.99th(us): 62399
UPDATE - Takes(s): 42.6, Count: 49357, OPS: 1158.3, Avg(us): 5, Min(us): 1, Max(us): 979, 50th(us): 4, 90th(us): 6, 95th(us): 8, 99th(us): 30, 99.9th(us): 295, 99.99th(us): 531
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  612
  read failed due to unknown txn status  445
rollback failed
  version mismatch  25

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch     791
```

+ 32

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 32
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 21.542626407s
COMMIT - Takes(s): 21.5, Count: 15397, OPS: 715.5, Avg(us): 18610, Min(us): 0, Max(us): 35871, 50th(us): 18991, 90th(us): 24447, 95th(us): 27151, 99th(us): 29167, 99.9th(us): 32159, 99.99th(us): 32511
COMMIT_ERROR - Takes(s): 21.5, Count: 1243, OPS: 57.8, Avg(us): 11110, Min(us): 3544, Max(us): 24479, 50th(us): 11303, 90th(us): 16207, 95th(us): 19007, 99th(us): 20543, 99.9th(us): 24367, 99.99th(us): 24479
READ   - Takes(s): 21.5, Count: 98171, OPS: 4557.5, Avg(us): 3680, Min(us): 4, Max(us): 17007, 50th(us): 3543, 90th(us): 4191, 95th(us): 4319, 99th(us): 4591, 99.9th(us): 14063, 99.99th(us): 16071
READ_ERROR - Takes(s): 21.5, Count: 1829, OPS: 85.1, Avg(us): 9279, Min(us): 6488, Max(us): 12759, 50th(us): 10023, 90th(us): 11751, 95th(us): 11959, 99th(us): 12423, 99.9th(us): 12735, 99.99th(us): 12759
Start  - Takes(s): 21.5, Count: 16672, OPS: 773.9, Avg(us): 36, Min(us): 13, Max(us): 657, 50th(us): 27, 90th(us): 46, 95th(us): 62, 99th(us): 268, 99.9th(us): 522, 99.99th(us): 625
TOTAL  - Takes(s): 21.5, Count: 211504, OPS: 9817.3, Avg(us): 9306, Min(us): 0, Max(us): 66559, 50th(us): 3445, 90th(us): 38943, 95th(us): 43423, 99th(us): 50207, 99.9th(us): 55391, 99.99th(us): 61119
TXN    - Takes(s): 21.5, Count: 15397, OPS: 715.5, Avg(us): 41503, Min(us): 16848, Max(us): 66303, 50th(us): 41855, 90th(us): 48223, 95th(us): 50687, 99th(us): 54399, 99.9th(us): 59519, 99.99th(us): 65119
TXN_ERROR - Takes(s): 21.5, Count: 1243, OPS: 57.8, Avg(us): 33904, Min(us): 20816, Max(us): 51199, 50th(us): 33951, 90th(us): 40031, 95th(us): 42367, 99th(us): 46943, 99.9th(us): 51071, 99.99th(us): 51199
TxnGroup - Takes(s): 21.5, Count: 16640, OPS: 773.2, Avg(us): 40902, Min(us): 17088, Max(us): 66559, 50th(us): 40959, 90th(us): 48223, 95th(us): 50783, 99th(us): 55071, 99.9th(us): 60031, 99.99th(us): 63647
UPDATE - Takes(s): 21.5, Count: 49227, OPS: 2285.4, Avg(us): 6, Min(us): 1, Max(us): 883, 50th(us): 4, 90th(us): 6, 95th(us): 9, 99th(us): 33, 99.9th(us): 450, 99.99th(us): 663
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    1243

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  899
  read failed due to unknown txn status  872
rollback failed
  version mismatch  58
```

+ 64

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 10.487051092s
COMMIT - Takes(s): 10.5, Count: 14730, OPS: 1409.5, Avg(us): 17917, Min(us): 0, Max(us): 35359, 50th(us): 18175, 90th(us): 23839, 95th(us): 25919, 99th(us): 28879, 99.9th(us): 31503, 99.99th(us): 35103
COMMIT_ERROR - Takes(s): 10.5, Count: 1910, OPS: 182.7, Avg(us): 10775, Min(us): 3498, Max(us): 23615, 50th(us): 10991, 90th(us): 15799, 95th(us): 18383, 99th(us): 20687, 99.9th(us): 23519, 99.99th(us): 23615
READ   - Takes(s): 10.5, Count: 97215, OPS: 9274.2, Avg(us): 3551, Min(us): 4, Max(us): 16303, 50th(us): 3403, 90th(us): 3963, 95th(us): 4235, 99th(us): 4911, 99.9th(us): 13599, 99.99th(us): 15407
READ_ERROR - Takes(s): 10.4, Count: 2785, OPS: 267.1, Avg(us): 9015, Min(us): 6452, Max(us): 18239, 50th(us): 10007, 90th(us): 11167, 95th(us): 11583, 99th(us): 12247, 99.9th(us): 14167, 99.99th(us): 18239
Start  - Takes(s): 10.5, Count: 16704, OPS: 1592.7, Avg(us): 41, Min(us): 12, Max(us): 1558, 50th(us): 28, 90th(us): 47, 95th(us): 62, 99th(us): 394, 99.9th(us): 1180, 99.99th(us): 1504
TOTAL  - Takes(s): 10.5, Count: 208567, OPS: 19885.9, Avg(us): 8924, Min(us): 0, Max(us): 67519, 50th(us): 3367, 90th(us): 37727, 95th(us): 42271, 99th(us): 48703, 99.9th(us): 54591, 99.99th(us): 59967
TXN    - Takes(s): 10.5, Count: 14730, OPS: 1409.3, Avg(us): 40356, Min(us): 17552, Max(us): 62367, 50th(us): 40351, 90th(us): 47327, 95th(us): 49311, 99th(us): 53055, 99.9th(us): 58175, 99.99th(us): 62047
TXN_ERROR - Takes(s): 10.5, Count: 1910, OPS: 182.7, Avg(us): 33156, Min(us): 19392, Max(us): 52223, 50th(us): 32767, 90th(us): 39231, 95th(us): 41759, 99th(us): 45567, 99.9th(us): 49023, 99.99th(us): 52223
TxnGroup - Takes(s): 10.5, Count: 16640, OPS: 1590.3, Avg(us): 39462, Min(us): 17728, Max(us): 67519, 50th(us): 39455, 90th(us): 47135, 95th(us): 49375, 99th(us): 54079, 99.9th(us): 59839, 99.99th(us): 64895
UPDATE - Takes(s): 10.5, Count: 48548, OPS: 4631.8, Avg(us): 7, Min(us): 1, Max(us): 1610, 50th(us): 4, 90th(us): 6, 95th(us): 9, 99th(us): 28, 99.9th(us): 666, 99.99th(us): 1126
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    1910

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  1412
  read failed due to unknown txn status  1246
rollback failed
  version mismatch  127
```

+ 96

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 96
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 6.806830222s
COMMIT - Takes(s): 6.8, Count: 14355, OPS: 2116.0, Avg(us): 17173, Min(us): 0, Max(us): 32143, 50th(us): 17695, 90th(us): 22703, 95th(us): 25023, 99th(us): 27983, 99.9th(us): 29903, 99.99th(us): 32031
COMMIT_ERROR - Takes(s): 6.8, Count: 2253, OPS: 332.5, Avg(us): 10364, Min(us): 3502, Max(us): 24127, 50th(us): 10711, 90th(us): 15183, 95th(us): 17855, 99th(us): 19871, 99.9th(us): 23311, 99.99th(us): 24127
READ   - Takes(s): 6.8, Count: 96623, OPS: 14204.6, Avg(us): 3462, Min(us): 4, Max(us): 17631, 50th(us): 3361, 90th(us): 3673, 95th(us): 3931, 99th(us): 4691, 99.9th(us): 13567, 99.99th(us): 14943
READ_ERROR - Takes(s): 6.8, Count: 3377, OPS: 499.4, Avg(us): 8865, Min(us): 6440, Max(us): 14975, 50th(us): 9975, 90th(us): 10575, 95th(us): 10871, 99th(us): 11727, 99.9th(us): 12663, 99.99th(us): 14975
Start  - Takes(s): 6.8, Count: 16704, OPS: 2453.5, Avg(us): 37, Min(us): 14, Max(us): 2261, 50th(us): 28, 90th(us): 45, 95th(us): 54, 99th(us): 305, 99.9th(us): 865, 99.99th(us): 1615
TOTAL  - Takes(s): 6.8, Count: 206671, OPS: 30355.3, Avg(us): 8609, Min(us): 0, Max(us): 73599, 50th(us): 3337, 90th(us): 36031, 95th(us): 41503, 99th(us): 47007, 99.9th(us): 53087, 99.99th(us): 58527
TXN    - Takes(s): 6.8, Count: 14355, OPS: 2116.2, Avg(us): 39220, Min(us): 16912, Max(us): 61439, 50th(us): 38879, 90th(us): 45855, 95th(us): 48223, 99th(us): 51871, 99.9th(us): 56639, 99.99th(us): 60127
TXN_ERROR - Takes(s): 6.8, Count: 2253, OPS: 332.5, Avg(us): 32380, Min(us): 20400, Max(us): 50239, 50th(us): 31743, 90th(us): 38655, 95th(us): 41247, 99th(us): 44799, 99.9th(us): 49759, 99.99th(us): 50239
TxnGroup - Takes(s): 6.8, Count: 16608, OPS: 2447.8, Avg(us): 38196, Min(us): 17088, Max(us): 73599, 50th(us): 38303, 90th(us): 45759, 95th(us): 48447, 99th(us): 52703, 99.9th(us): 58463, 99.99th(us): 63583
UPDATE - Takes(s): 6.8, Count: 48026, OPS: 7061.2, Avg(us): 5, Min(us): 1, Max(us): 1416, 50th(us): 4, 90th(us): 6, 95th(us): 8, 99th(us): 21, 99.9th(us): 507, 99.99th(us): 1086
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    2253

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  1820
  read failed due to unknown txn status  1389
rollback failed
  version mismatch  168
```

+ 128

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 5.289067273s
COMMIT - Takes(s): 5.3, Count: 14104, OPS: 2679.0, Avg(us): 17584, Min(us): 0, Max(us): 42687, 50th(us): 17983, 90th(us): 23583, 95th(us): 25663, 99th(us): 28687, 99.9th(us): 34047, 99.99th(us): 41311
COMMIT_ERROR - Takes(s): 5.3, Count: 2536, OPS: 482.2, Avg(us): 10698, Min(us): 3514, Max(us): 24415, 50th(us): 10887, 90th(us): 15703, 95th(us): 18223, 99th(us): 20159, 99.9th(us): 23039, 99.99th(us): 24415
READ   - Takes(s): 5.3, Count: 96168, OPS: 18193.7, Avg(us): 3545, Min(us): 5, Max(us): 17311, 50th(us): 3385, 90th(us): 3927, 95th(us): 4267, 99th(us): 5371, 99.9th(us): 13911, 99.99th(us): 15383
READ_ERROR - Takes(s): 5.2, Count: 3832, OPS: 730.2, Avg(us): 9080, Min(us): 6432, Max(us): 22063, 50th(us): 10055, 90th(us): 10935, 95th(us): 11311, 99th(us): 12367, 99.9th(us): 19503, 99.99th(us): 22063
Start  - Takes(s): 5.3, Count: 16768, OPS: 3170.3, Avg(us): 50, Min(us): 14, Max(us): 10135, 50th(us): 30, 90th(us): 47, 95th(us): 60, 99th(us): 650, 99.9th(us): 1638, 99.99th(us): 2773
TOTAL  - Takes(s): 5.3, Count: 205712, OPS: 38884.4, Avg(us): 8807, Min(us): 0, Max(us): 70207, 50th(us): 3355, 90th(us): 37055, 95th(us): 42399, 99th(us): 48831, 99.9th(us): 55999, 99.99th(us): 62911
TXN    - Takes(s): 5.3, Count: 14104, OPS: 2678.9, Avg(us): 40384, Min(us): 17200, Max(us): 68159, 50th(us): 40095, 90th(us): 47423, 95th(us): 49759, 99th(us): 54207, 99.9th(us): 61439, 99.99th(us): 66751
TXN_ERROR - Takes(s): 5.3, Count: 2536, OPS: 482.2, Avg(us): 33389, Min(us): 20512, Max(us): 52223, 50th(us): 32799, 90th(us): 39903, 95th(us): 42207, 99th(us): 46527, 99.9th(us): 50271, 99.99th(us): 52223
TxnGroup - Takes(s): 5.3, Count: 16640, OPS: 3157.9, Avg(us): 39188, Min(us): 17072, Max(us): 70207, 50th(us): 39167, 90th(us): 47263, 95th(us): 49855, 99th(us): 54911, 99.9th(us): 61599, 99.99th(us): 66175
UPDATE - Takes(s): 5.3, Count: 47928, OPS: 9069.2, Avg(us): 7, Min(us): 1, Max(us): 2217, 50th(us): 4, 90th(us): 6, 95th(us): 9, 99th(us): 23, 99.9th(us): 783, 99.99th(us): 1460
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    2536

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  2044
  read failed due to unknown txn status  1538
rollback failed
  version mismatch  250
```

##### Epoxy

+ 8

```bash
{
    "name": "Workload F",
    "timeTaken": 85.854,
    "throughput": 1164.7680946723508,
    "latency": "Operation: WorkloadF,Count: 16664, Avg latency: 41209, P50 latency: 41316, P99 latency: 41790, Sum: 686709325\nOperation: SingleOperation,Count: 99984, Avg latency: 5, P50 latency: 6, P99 latency: 7, Sum: 505135\n",
    "txnPerThread": 2083
}
```

+ 16

```bash
{
    "name": "Workload F",
    "timeTaken": 42.644,
    "throughput": 2344.995779007598,
    "latency": "Operation: WorkloadF,Count: 16656, Avg latency: 40948, P50 latency: 40650, P99 latency: 45170, Sum: 682044397\nOperation: SingleOperation,Count: 99936, Avg latency: 5, P50 latency: 6, P99 latency: 7, Sum: 504463\n",
    "txnPerThread": 1041
}
```

+ 32

```bash
{
    "name": "Workload F",
    "timeTaken": 21.847,
    "throughput": 4577.287499427839,
    "latency": "Operation: WorkloadF,Count: 16640, Avg latency: 41997, P50 latency: 41894, P99 latency: 45604, Sum: 698831469\nOperation: SingleOperation,Count: 99840, Avg latency: 5, P50 latency: 6, P99 latency: 7, Sum: 510446\n",
    "txnPerThread": 520
}
```

+ 64

```bash
{
    "name": "Workload F",
    "timeTaken": 11.643,
    "throughput": 8588.851670531649,
    "latency": "Operation: WorkloadF,Count: 16640, Avg latency: 44482, P50 latency: 44228, P99 latency: 51841, Sum: 740190365\nOperation: SingleOperation,Count: 99840, Avg latency: 5, P50 latency: 6, P99 latency: 8, Sum: 505176\n",
    "txnPerThread": 260
}
```

+ 96

```bash
{
    "name": "Workload F",
    "timeTaken": 8.586,
    "throughput": 11646.866992778941,
    "latency": "Operation: WorkloadF,Count: 16608, Avg latency: 46169, P50 latency: 45504, P99 latency: 57366, Sum: 766783643\nOperation: SingleOperation,Count: 99672, Avg latency: 5, P50 latency: 6, P99 latency: 10, Sum: 511538\n",
    "txnPerThread": 173
}
```

+ 128

```bash
{
    "name": "Workload F",
    "timeTaken": 8.4,
    "throughput": 11904.761904761905,
    "latency": "Operation: WorkloadF,Count: 16640, Avg latency: 59074, P50 latency: 55765, P99 latency: 117237, Sum: 983003196\nOperation: SingleOperation,Count: 99864, Avg latency: 5, P50 latency: 6, P99 latency: 10, Sum: 515790\n",
    "txnPerThread": 130
}
```

##### Oreo

+ 8

```bash
----------------------------------
Run finished, takes 1m7.87452276s
COMMIT - Takes(s): 67.8, Count: 16061, OPS: 236.7, Avg(us): 8071, Min(us): 0, Max(us): 9423, 50th(us): 8199, 90th(us): 8551, 95th(us): 8655, 99th(us): 8871, 99.9th(us): 9111, 99.99th(us): 9255
COMMIT_ERROR - Takes(s): 67.8, Count: 603, OPS: 8.9, Avg(us): 4432, Min(us): 3744, Max(us): 5307, 50th(us): 4431, 90th(us): 4787, 95th(us): 4927, 99th(us): 5127, 99.9th(us): 5303, 99.99th(us): 5307
READ   - Takes(s): 67.9, Count: 99732, OPS: 1469.4, Avg(us): 4085, Min(us): 5, Max(us): 5699, 50th(us): 4107, 90th(us): 4411, 95th(us): 4539, 99th(us): 4811, 99.9th(us): 5175, 99.99th(us): 5471
READ_ERROR - Takes(s): 67.8, Count: 268, OPS: 4.0, Avg(us): 4568, Min(us): 3762, Max(us): 5415, 50th(us): 4591, 90th(us): 4831, 95th(us): 4955, 99th(us): 5127, 99.9th(us): 5415, 99.99th(us): 5415
Start  - Takes(s): 67.9, Count: 16672, OPS: 245.6, Avg(us): 24, Min(us): 13, Max(us): 706, 50th(us): 20, 90th(us): 30, 95th(us): 35, 99th(us): 57, 99.9th(us): 287, 99.99th(us): 502
TOTAL  - Takes(s): 67.9, Count: 214955, OPS: 3166.9, Avg(us): 7467, Min(us): 0, Max(us): 35391, 50th(us): 4055, 90th(us): 32671, 95th(us): 33119, 99th(us): 33695, 99.9th(us): 34175, 99.99th(us): 34687
TXN    - Takes(s): 67.8, Count: 16061, OPS: 236.7, Avg(us): 32686, Min(us): 20192, Max(us): 34943, 50th(us): 32895, 90th(us): 33535, 95th(us): 33727, 99th(us): 34015, 99.9th(us): 34431, 99.99th(us): 34751
TXN_ERROR - Takes(s): 67.8, Count: 603, OPS: 8.9, Avg(us): 29043, Min(us): 24624, Max(us): 30863, 50th(us): 29183, 90th(us): 29807, 95th(us): 29951, 99th(us): 30399, 99.9th(us): 30799, 99.99th(us): 30863
TxnGroup - Takes(s): 67.9, Count: 16664, OPS: 245.6, Avg(us): 32552, Min(us): 20960, Max(us): 35391, 50th(us): 32895, 90th(us): 33599, 95th(us): 33791, 99th(us): 34143, 99.9th(us): 34655, 99.99th(us): 35103
UPDATE - Takes(s): 67.9, Count: 49765, OPS: 733.2, Avg(us): 3, Min(us): 1, Max(us): 385, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 11, 99.9th(us): 161, 99.99th(us): 320
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  254
  read failed due to unknown txn status    7
rollback failed
  version mismatch  7

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  603
```

+ 16

```bash
----------------------------------
Run finished, takes 34.180934277s
COMMIT - Takes(s): 34.1, Count: 15539, OPS: 455.1, Avg(us): 8234, Min(us): 0, Max(us): 11239, 50th(us): 8391, 90th(us): 8943, 95th(us): 9079, 99th(us): 9351, 99.9th(us): 10319, 99.99th(us): 11127
COMMIT_ERROR - Takes(s): 34.1, Count: 1117, OPS: 32.8, Avg(us): 4550, Min(us): 3728, Max(us): 6231, 50th(us): 4531, 90th(us): 5079, 95th(us): 5223, 99th(us): 5619, 99.9th(us): 6115, 99.99th(us): 6231
READ   - Takes(s): 34.2, Count: 99597, OPS: 2914.2, Avg(us): 4107, Min(us): 5, Max(us): 11135, 50th(us): 4095, 90th(us): 4647, 95th(us): 4795, 99th(us): 5159, 99.9th(us): 5639, 99.99th(us): 6339
READ_ERROR - Takes(s): 34.0, Count: 403, OPS: 11.8, Avg(us): 4648, Min(us): 3622, Max(us): 5787, 50th(us): 4667, 90th(us): 5219, 95th(us): 5411, 99th(us): 5611, 99.9th(us): 5787, 99.99th(us): 5787
Start  - Takes(s): 34.2, Count: 16672, OPS: 487.8, Avg(us): 27, Min(us): 13, Max(us): 714, 50th(us): 24, 90th(us): 34, 95th(us): 41, 99th(us): 164, 99.9th(us): 530, 99.99th(us): 643
TOTAL  - Takes(s): 34.2, Count: 213841, OPS: 6256.0, Avg(us): 7463, Min(us): 0, Max(us): 39839, 50th(us): 3949, 90th(us): 32527, 95th(us): 33823, 99th(us): 34943, 99.9th(us): 35775, 99.99th(us): 37183
TXN    - Takes(s): 34.1, Count: 15539, OPS: 455.1, Avg(us): 33002, Min(us): 20128, Max(us): 38655, 50th(us): 33247, 90th(us): 34687, 95th(us): 35007, 99th(us): 35551, 99.9th(us): 37183, 99.99th(us): 37791
TXN_ERROR - Takes(s): 34.1, Count: 1117, OPS: 32.8, Avg(us): 29304, Min(us): 20528, Max(us): 35487, 50th(us): 29567, 90th(us): 30943, 95th(us): 31231, 99th(us): 31647, 99.9th(us): 32095, 99.99th(us): 35487
TxnGroup - Takes(s): 34.2, Count: 16656, OPS: 487.7, Avg(us): 32747, Min(us): 20208, Max(us): 39839, 50th(us): 33151, 90th(us): 34783, 95th(us): 35103, 99th(us): 35711, 99.9th(us): 36383, 99.99th(us): 39103
UPDATE - Takes(s): 34.2, Count: 49838, OPS: 1458.2, Avg(us): 4, Min(us): 1, Max(us): 728, 50th(us): 3, 90th(us): 5, 95th(us): 5, 99th(us): 14, 99.9th(us): 218, 99.99th(us): 510
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  1117

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  344
  read failed due to unknown txn status   44
rollback failed
  version mismatch  15
```

+ 32

```bash
----------------------------------
Run finished, takes 16.801377839s
COMMIT - Takes(s): 16.8, Count: 14731, OPS: 878.1, Avg(us): 8235, Min(us): 0, Max(us): 14711, 50th(us): 8319, 90th(us): 9071, 95th(us): 9303, 99th(us): 9783, 99.9th(us): 11127, 99.99th(us): 14071
COMMIT_ERROR - Takes(s): 16.8, Count: 1909, OPS: 113.8, Avg(us): 4594, Min(us): 3666, Max(us): 6847, 50th(us): 4527, 90th(us): 5199, 95th(us): 5419, 99th(us): 5863, 99.9th(us): 6739, 99.99th(us): 6847
READ   - Takes(s): 16.8, Count: 99406, OPS: 5917.9, Avg(us): 4028, Min(us): 5, Max(us): 9919, 50th(us): 3919, 90th(us): 4655, 95th(us): 4879, 99th(us): 5367, 99.9th(us): 5983, 99.99th(us): 6883
READ_ERROR - Takes(s): 16.7, Count: 594, OPS: 35.5, Avg(us): 4663, Min(us): 3674, Max(us): 6351, 50th(us): 4635, 90th(us): 5275, 95th(us): 5531, 99th(us): 5971, 99.9th(us): 6227, 99.99th(us): 6351
Start  - Takes(s): 16.8, Count: 16672, OPS: 992.2, Avg(us): 29, Min(us): 13, Max(us): 1261, 50th(us): 26, 90th(us): 38, 95th(us): 43, 99th(us): 208, 99.9th(us): 663, 99.99th(us): 1082
TOTAL  - Takes(s): 16.8, Count: 211699, OPS: 12599.1, Avg(us): 7257, Min(us): 0, Max(us): 39583, 50th(us): 3825, 90th(us): 32095, 95th(us): 33151, 99th(us): 34367, 99.9th(us): 35583, 99.99th(us): 38143
TXN    - Takes(s): 16.8, Count: 14731, OPS: 878.1, Avg(us): 32551, Min(us): 19184, Max(us): 39583, 50th(us): 32751, 90th(us): 34111, 95th(us): 34495, 99th(us): 35391, 99.9th(us): 38559, 99.99th(us): 39487
TXN_ERROR - Takes(s): 16.8, Count: 1909, OPS: 113.8, Avg(us): 28909, Min(us): 22928, Max(us): 33343, 50th(us): 29039, 90th(us): 30367, 95th(us): 30751, 99th(us): 31455, 99.9th(us): 33119, 99.99th(us): 33343
TxnGroup - Takes(s): 16.8, Count: 16640, OPS: 991.9, Avg(us): 32118, Min(us): 18672, Max(us): 38143, 50th(us): 32607, 90th(us): 34111, 95th(us): 34527, 99th(us): 35391, 99.9th(us): 36447, 99.99th(us): 37855
UPDATE - Takes(s): 16.8, Count: 49519, OPS: 2948.1, Avg(us): 4, Min(us): 1, Max(us): 1415, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 15, 99.9th(us): 269, 99.99th(us): 973
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  1909

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  443
  read failed due to unknown txn status  108
rollback failed
  version mismatch  43
```

+ 64

```bash
----------------------------------
Run finished, takes 7.981509822s
COMMIT - Takes(s): 8.0, Count: 13692, OPS: 1721.6, Avg(us): 8011, Min(us): 0, Max(us): 20207, 50th(us): 7995, 90th(us): 8807, 95th(us): 9199, 99th(us): 10447, 99.9th(us): 15743, 99.99th(us): 18255
COMMIT_ERROR - Takes(s): 7.9, Count: 2948, OPS: 370.9, Avg(us): 4506, Min(us): 3618, Max(us): 13111, 50th(us): 4351, 90th(us): 5147, 95th(us): 5479, 99th(us): 6895, 99.9th(us): 11255, 99.99th(us): 13111
READ   - Takes(s): 8.0, Count: 99489, OPS: 12471.8, Avg(us): 3814, Min(us): 6, Max(us): 15095, 50th(us): 3699, 90th(us): 4283, 95th(us): 4535, 99th(us): 5227, 99.9th(us): 6807, 99.99th(us): 12127
READ_ERROR - Takes(s): 7.9, Count: 511, OPS: 64.3, Avg(us): 4427, Min(us): 3608, Max(us): 10791, 50th(us): 4347, 90th(us): 5127, 95th(us): 5455, 99th(us): 6295, 99.9th(us): 7151, 99.99th(us): 10791
Start  - Takes(s): 8.0, Count: 16704, OPS: 2092.8, Avg(us): 28, Min(us): 13, Max(us): 1331, 50th(us): 26, 90th(us): 38, 95th(us): 43, 99th(us): 136, 99.9th(us): 542, 99.99th(us): 880
TOTAL  - Takes(s): 8.0, Count: 209861, OPS: 26294.5, Avg(us): 6767, Min(us): 0, Max(us): 46975, 50th(us): 3649, 90th(us): 30335, 95th(us): 31247, 99th(us): 32799, 99.9th(us): 36959, 99.99th(us): 41759
TXN    - Takes(s): 8.0, Count: 13692, OPS: 1721.5, Avg(us): 31024, Min(us): 21424, Max(us): 46975, 50th(us): 30943, 90th(us): 32527, 95th(us): 33311, 99th(us): 36511, 99.9th(us): 41375, 99.99th(us): 44671
TXN_ERROR - Takes(s): 7.9, Count: 2948, OPS: 370.8, Avg(us): 27525, Min(us): 21920, Max(us): 37983, 50th(us): 27391, 90th(us): 28975, 95th(us): 29711, 99th(us): 32799, 99.9th(us): 36831, 99.99th(us): 37983
TxnGroup - Takes(s): 8.0, Count: 16640, OPS: 2090.4, Avg(us): 30377, Min(us): 18304, Max(us): 46271, 50th(us): 30703, 90th(us): 32335, 95th(us): 32991, 99th(us): 35359, 99.9th(us): 41055, 99.99th(us): 44511
UPDATE - Takes(s): 8.0, Count: 49644, OPS: 6223.6, Avg(us): 4, Min(us): 1, Max(us): 1357, 50th(us): 3, 90th(us): 5, 95th(us): 5, 99th(us): 14, 99.9th(us): 206, 99.99th(us): 694
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2948

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  261
  read failed due to unknown txn status  217
rollback failed
  version mismatch  33
```

+ 96

```bash
----------------------------------
Run finished, takes 5.902570589s
COMMIT - Takes(s): 5.9, Count: 13086, OPS: 2226.3, Avg(us): 8912, Min(us): 0, Max(us): 24751, 50th(us): 8823, 90th(us): 10495, 95th(us): 11135, 99th(us): 12679, 99.9th(us): 15847, 99.99th(us): 17471
COMMIT_ERROR - Takes(s): 5.9, Count: 3522, OPS: 599.8, Avg(us): 5252, Min(us): 3726, Max(us): 14447, 50th(us): 4999, 90th(us): 6599, 95th(us): 7159, 99th(us): 8431, 99.9th(us): 11191, 99.99th(us): 14447
READ   - Takes(s): 5.9, Count: 99276, OPS: 16829.3, Avg(us): 4190, Min(us): 5, Max(us): 20175, 50th(us): 3957, 90th(us): 5107, 95th(us): 5583, 99th(us): 6747, 99.9th(us): 8663, 99.99th(us): 11191
READ_ERROR - Takes(s): 5.9, Count: 724, OPS: 123.5, Avg(us): 5193, Min(us): 3644, Max(us): 11279, 50th(us): 4967, 90th(us): 6539, 95th(us): 7159, 99th(us): 8791, 99.9th(us): 11207, 99.99th(us): 11279
Start  - Takes(s): 5.9, Count: 16704, OPS: 2830.0, Avg(us): 35, Min(us): 14, Max(us): 2729, 50th(us): 27, 90th(us): 40, 95th(us): 45, 99th(us): 283, 99.9th(us): 1285, 99.99th(us): 2023
TOTAL  - Takes(s): 5.9, Count: 208518, OPS: 35324.5, Avg(us): 7368, Min(us): 0, Max(us): 52223, 50th(us): 3829, 90th(us): 32767, 95th(us): 34623, 99th(us): 37471, 99.9th(us): 40991, 99.99th(us): 44447
TXN    - Takes(s): 5.9, Count: 13086, OPS: 2226.1, Avg(us): 34223, Min(us): 20576, Max(us): 48927, 50th(us): 34079, 90th(us): 37119, 95th(us): 38143, 99th(us): 40607, 99.9th(us): 43615, 99.99th(us): 46047
TXN_ERROR - Takes(s): 5.9, Count: 3522, OPS: 599.8, Avg(us): 30665, Min(us): 20496, Max(us): 47807, 50th(us): 30415, 90th(us): 33439, 95th(us): 34463, 99th(us): 36927, 99.9th(us): 40159, 99.99th(us): 47807
TxnGroup - Takes(s): 5.9, Count: 16608, OPS: 2822.7, Avg(us): 33422, Min(us): 19984, Max(us): 52223, 50th(us): 33567, 90th(us): 36703, 95th(us): 37791, 99th(us): 40319, 99.9th(us): 44191, 99.99th(us): 46751
UPDATE - Takes(s): 5.9, Count: 49758, OPS: 8435.8, Avg(us): 4, Min(us): 1, Max(us): 2413, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 16, 99.9th(us): 323, 99.99th(us): 1349
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3522

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  371
  read failed due to unknown txn status  282
rollback failed
  version mismatch  71
```

+ 128

```bash
----------------------------------
Run finished, takes 5.129370197s
COMMIT - Takes(s): 5.1, Count: 12749, OPS: 2501.2, Avg(us): 10402, Min(us): 0, Max(us): 35359, 50th(us): 10239, 90th(us): 13159, 95th(us): 14159, 99th(us): 16639, 99.9th(us): 20575, 99.99th(us): 24991
COMMIT_ERROR - Takes(s): 5.1, Count: 3891, OPS: 763.5, Avg(us): 6561, Min(us): 3754, Max(us): 18287, 50th(us): 6187, 90th(us): 8927, 95th(us): 9943, 99th(us): 12279, 99.9th(us): 15063, 99.99th(us): 18287
READ   - Takes(s): 5.1, Count: 99131, OPS: 19342.5, Avg(us): 4800, Min(us): 5, Max(us): 20159, 50th(us): 4391, 90th(us): 6443, 95th(us): 7275, 99th(us): 9383, 99.9th(us): 12543, 99.99th(us): 14839
READ_ERROR - Takes(s): 5.1, Count: 869, OPS: 170.7, Avg(us): 6423, Min(us): 3718, Max(us): 21375, 50th(us): 5887, 90th(us): 9095, 95th(us): 10631, 99th(us): 12743, 99.9th(us): 16655, 99.99th(us): 21375
Start  - Takes(s): 5.1, Count: 16768, OPS: 3268.7, Avg(us): 39, Min(us): 13, Max(us): 2657, 50th(us): 28, 90th(us): 41, 95th(us): 47, 99th(us): 306, 99.9th(us): 1999, 99.99th(us): 2549
TOTAL  - Takes(s): 5.1, Count: 207262, OPS: 40403.1, Avg(us): 8456, Min(us): 0, Max(us): 63807, 50th(us): 4127, 90th(us): 36863, 95th(us): 40191, 99th(us): 45055, 99.9th(us): 50591, 99.99th(us): 55871
TXN    - Takes(s): 5.1, Count: 12749, OPS: 2501.3, Avg(us): 39404, Min(us): 23024, Max(us): 63807, 50th(us): 39135, 90th(us): 44415, 95th(us): 46271, 99th(us): 50175, 99.9th(us): 55039, 99.99th(us): 61119
TXN_ERROR - Takes(s): 5.1, Count: 3891, OPS: 763.4, Avg(us): 35851, Min(us): 25344, Max(us): 54591, 50th(us): 35487, 90th(us): 40799, 95th(us): 42591, 99th(us): 45855, 99.9th(us): 50399, 99.99th(us): 54591
TxnGroup - Takes(s): 5.1, Count: 16640, OPS: 3258.8, Avg(us): 38513, Min(us): 20240, Max(us): 62175, 50th(us): 38463, 90th(us): 43839, 95th(us): 45663, 99th(us): 49503, 99.9th(us): 54687, 99.99th(us): 59391
UPDATE - Takes(s): 5.1, Count: 49225, OPS: 9604.3, Avg(us): 5, Min(us): 1, Max(us): 2739, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 17, 99.9th(us): 571, 99.99th(us): 2171
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3891

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  450
  read failed due to unknown txn status  314
rollback failed
  version mismatch  105
```



##### Native

+ 8

```bash
----------------------------------
Run finished, takes 1m5.127738877s
READ   - Takes(s): 65.1, Count: 100000, OPS: 1535.5, Avg(us): 3440, Min(us): 3184, Max(us): 4563, 50th(us): 3433, 90th(us): 3521, 95th(us): 3561, 99th(us): 3711, 99.9th(us): 3885, 99.99th(us): 4267
TOTAL  - Takes(s): 65.1, Count: 166587, OPS: 2558.0, Avg(us): 6222, Min(us): 3184, Max(us): 45151, 50th(us): 3457, 90th(us): 20479, 95th(us): 31119, 99th(us): 37855, 99.9th(us): 41471, 99.99th(us): 41887
TxnGroup - Takes(s): 65.1, Count: 16664, OPS: 256.0, Avg(us): 31131, Min(us): 20400, Max(us): 45151, 50th(us): 31119, 90th(us): 37855, 95th(us): 38143, 99th(us): 41471, 99.9th(us): 41887, 99.99th(us): 44735
UPDATE - Takes(s): 65.1, Count: 49923, OPS: 766.6, Avg(us): 3479, Min(us): 3216, Max(us): 4599, 50th(us): 3473, 90th(us): 3549, 95th(us): 3589, 99th(us): 3745, 99.9th(us): 3971, 99.99th(us): 4319
Error Summary:
```

+ 16

```bash
----------------------------------
Run finished, takes 33.776373975s
READ   - Takes(s): 33.8, Count: 100000, OPS: 2961.0, Avg(us): 3558, Min(us): 3178, Max(us): 5175, 50th(us): 3535, 90th(us): 3769, 95th(us): 3861, 99th(us): 4011, 99.9th(us): 4263, 99.99th(us): 4599
TOTAL  - Takes(s): 33.8, Count: 166755, OPS: 4937.5, Avg(us): 6437, Min(us): 3178, Max(us): 46815, 50th(us): 3569, 90th(us): 4531, 95th(us): 32015, 99th(us): 38911, 99.9th(us): 42719, 99.99th(us): 45887
TxnGroup - Takes(s): 33.8, Count: 16656, OPS: 493.4, Avg(us): 32258, Min(us): 20592, Max(us): 46815, 50th(us): 32015, 90th(us): 38911, 95th(us): 39263, 99th(us): 42719, 99.9th(us): 45887, 99.99th(us): 46559
UPDATE - Takes(s): 33.8, Count: 50099, OPS: 1483.6, Avg(us): 3600, Min(us): 3206, Max(us): 4743, 50th(us): 3579, 90th(us): 3803, 95th(us): 3901, 99th(us): 4031, 99.9th(us): 4291, 99.99th(us): 4543
Error Summary:
```

+ 32

```bash
----------------------------------
Run finished, takes 17.415192978s
READ   - Takes(s): 17.4, Count: 100000, OPS: 5743.1, Avg(us): 3638, Min(us): 3164, Max(us): 5967, 50th(us): 3633, 90th(us): 3915, 95th(us): 4011, 99th(us): 4223, 99.9th(us): 4675, 99.99th(us): 5591
TOTAL  - Takes(s): 17.4, Count: 166630, OPS: 9569.9, Avg(us): 6592, Min(us): 3164, Max(us): 47967, 50th(us): 3681, 90th(us): 5299, 95th(us): 33023, 99th(us): 39871, 99.9th(us): 43967, 99.99th(us): 44895
TxnGroup - Takes(s): 17.4, Count: 16640, OPS: 956.9, Avg(us): 33070, Min(us): 19696, Max(us): 47967, 50th(us): 33023, 90th(us): 39871, 95th(us): 40543, 99th(us): 43967, 99.9th(us): 44895, 99.99th(us): 47679
UPDATE - Takes(s): 17.4, Count: 49990, OPS: 2871.6, Avg(us): 3688, Min(us): 3190, Max(us): 7399, 50th(us): 3691, 90th(us): 3953, 95th(us): 4043, 99th(us): 4255, 99.9th(us): 4659, 99.99th(us): 5175
Error Summary:
```

+ 64

```bash
----------------------------------
Run finished, takes 8.941305811s
READ   - Takes(s): 8.9, Count: 100000, OPS: 11188.0, Avg(us): 3710, Min(us): 3160, Max(us): 12375, 50th(us): 3693, 90th(us): 4187, 95th(us): 4331, 99th(us): 4683, 99.9th(us): 5355, 99.99th(us): 11631
TOTAL  - Takes(s): 8.9, Count: 166726, OPS: 18653.9, Avg(us): 6733, Min(us): 3160, Max(us): 50815, 50th(us): 3777, 90th(us): 11383, 95th(us): 34079, 99th(us): 39871, 99.9th(us): 45343, 99.99th(us): 47103
TxnGroup - Takes(s): 8.9, Count: 16640, OPS: 1866.1, Avg(us): 33875, Min(us): 19760, Max(us): 50815, 50th(us): 34079, 90th(us): 39871, 95th(us): 41951, 99th(us): 45343, 99.9th(us): 47103, 99.99th(us): 47967
UPDATE - Takes(s): 8.9, Count: 50086, OPS: 5606.0, Avg(us): 3751, Min(us): 3196, Max(us): 12255, 50th(us): 3755, 90th(us): 4211, 95th(us): 4351, 99th(us): 4663, 99.9th(us): 5255, 99.99th(us): 11687
Error Summary:
```

+ 96

```bash
----------------------------------
Run finished, takes 5.934720956s
READ   - Takes(s): 5.9, Count: 100000, OPS: 16861.5, Avg(us): 3640, Min(us): 3158, Max(us): 15759, 50th(us): 3423, 90th(us): 4315, 95th(us): 4539, 99th(us): 5223, 99.9th(us): 7739, 99.99th(us): 13559
TOTAL  - Takes(s): 5.9, Count: 166667, OPS: 28102.3, Avg(us): 6592, Min(us): 3158, Max(us): 50495, 50th(us): 3491, 90th(us): 10447, 95th(us): 33119, 99th(us): 39711, 99.9th(us): 44703, 99.99th(us): 48447
TxnGroup - Takes(s): 5.9, Count: 16608, OPS: 2810.9, Avg(us): 33201, Min(us): 19760, Max(us): 50495, 50th(us): 33151, 90th(us): 39711, 95th(us): 41375, 99th(us): 44703, 99.9th(us): 48447, 99.99th(us): 49887
UPDATE - Takes(s): 5.9, Count: 50059, OPS: 8446.5, Avg(us): 3661, Min(us): 3200, Max(us): 14471, 50th(us): 3453, 90th(us): 4331, 95th(us): 4531, 99th(us): 5135, 99.9th(us): 6023, 99.99th(us): 13359
Error Summary:
```

+ 128

```bash
----------------------------------
Run finished, takes 4.401713154s
READ   - Takes(s): 4.4, Count: 100000, OPS: 22734.8, Avg(us): 3583, Min(us): 3158, Max(us): 11927, 50th(us): 3385, 90th(us): 4203, 95th(us): 4519, 99th(us): 5291, 99.9th(us): 8091, 99.99th(us): 10759
TOTAL  - Takes(s): 4.4, Count: 166767, OPS: 37917.4, Avg(us): 6493, Min(us): 3158, Max(us): 51647, 50th(us): 3437, 90th(us): 9935, 95th(us): 32607, 99th(us): 38911, 99.9th(us): 43711, 99.99th(us): 46943
TxnGroup - Takes(s): 4.4, Count: 16640, OPS: 3804.4, Avg(us): 32709, Min(us): 19632, Max(us): 51647, 50th(us): 32623, 90th(us): 38911, 95th(us): 40703, 99th(us): 43711, 99.9th(us): 46943, 99.99th(us): 50143
UPDATE - Takes(s): 4.4, Count: 50127, OPS: 11413.8, Avg(us): 3596, Min(us): 3196, Max(us): 11727, 50th(us): 3415, 90th(us): 4191, 95th(us): 4483, 99th(us): 5215, 99.9th(us): 6183, 99.99th(us): 10007
Error Summary:
```



#### redis - mongo

##### Cherry Garcia

+ 8

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 8
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 1m24.815421199s
COMMIT - Takes(s): 84.8, Count: 16225, OPS: 191.4, Avg(us): 18158, Min(us): 0, Max(us): 46047, 50th(us): 18031, 90th(us): 24703, 95th(us): 25487, 99th(us): 28591, 99.9th(us): 31503, 99.99th(us): 37727
COMMIT_ERROR - Takes(s): 84.5, Count: 439, OPS: 5.2, Avg(us): 10845, Min(us): 3634, Max(us): 23471, 50th(us): 10911, 90th(us): 15431, 95th(us): 17999, 99th(us): 19935, 99.9th(us): 23471, 99.99th(us): 23471
READ   - Takes(s): 84.8, Count: 99470, OPS: 1172.8, Avg(us): 3711, Min(us): 4, Max(us): 28015, 50th(us): 3579, 90th(us): 3925, 95th(us): 4031, 99th(us): 10591, 99.9th(us): 11815, 99.99th(us): 15527
READ_ERROR - Takes(s): 84.6, Count: 530, OPS: 6.3, Avg(us): 9168, Min(us): 6332, Max(us): 12335, 50th(us): 10367, 90th(us): 11495, 95th(us): 11695, 99th(us): 11991, 99.9th(us): 12087, 99.99th(us): 12335
Start  - Takes(s): 84.8, Count: 16672, OPS: 196.6, Avg(us): 23, Min(us): 13, Max(us): 374, 50th(us): 18, 90th(us): 28, 95th(us): 36, 99th(us): 120, 99.9th(us): 262, 99.99th(us): 318
TOTAL  - Takes(s): 84.8, Count: 214970, OPS: 2534.6, Avg(us): 9300, Min(us): 0, Max(us): 74623, 50th(us): 3531, 90th(us): 38815, 95th(us): 42655, 99th(us): 49631, 99.9th(us): 56671, 99.99th(us): 62559
TXN    - Takes(s): 84.8, Count: 16225, OPS: 191.4, Avg(us): 40692, Min(us): 17248, Max(us): 70271, 50th(us): 39583, 90th(us): 46943, 95th(us): 49983, 99th(us): 54367, 99.9th(us): 60959, 99.99th(us): 67967
TXN_ERROR - Takes(s): 84.5, Count: 439, OPS: 5.2, Avg(us): 33341, Min(us): 21216, Max(us): 46751, 50th(us): 32415, 90th(us): 39071, 95th(us): 42591, 99th(us): 46335, 99.9th(us): 46751, 99.99th(us): 46751
TxnGroup - Takes(s): 84.8, Count: 16664, OPS: 196.5, Avg(us): 40490, Min(us): 18032, Max(us): 74623, 50th(us): 39487, 90th(us): 46751, 95th(us): 49951, 99th(us): 54463, 99.9th(us): 62047, 99.99th(us): 68223
UPDATE - Takes(s): 84.8, Count: 49714, OPS: 586.2, Avg(us): 3, Min(us): 1, Max(us): 432, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 12, 99.9th(us): 137, 99.99th(us): 241
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  272
  read failed due to unknown txn status  249
rollback failed
  version mismatch  9

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch     439
```

+ 16

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 16
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 42.708591575s
COMMIT - Takes(s): 42.7, Count: 15855, OPS: 371.5, Avg(us): 18199, Min(us): 0, Max(us): 53695, 50th(us): 18495, 90th(us): 24223, 95th(us): 26063, 99th(us): 29359, 99.9th(us): 31471, 99.99th(us): 49023
COMMIT_ERROR - Takes(s): 42.7, Count: 801, OPS: 18.8, Avg(us): 11037, Min(us): 3494, Max(us): 23375, 50th(us): 11191, 90th(us): 15767, 95th(us): 18687, 99th(us): 19775, 99.9th(us): 22671, 99.99th(us): 23375
READ   - Takes(s): 42.7, Count: 99009, OPS: 2318.4, Avg(us): 3742, Min(us): 4, Max(us): 31327, 50th(us): 3633, 90th(us): 3987, 95th(us): 4115, 99th(us): 10847, 99.9th(us): 11879, 99.99th(us): 15127
READ_ERROR - Takes(s): 42.7, Count: 991, OPS: 23.2, Avg(us): 8970, Min(us): 6244, Max(us): 27455, 50th(us): 7699, 90th(us): 11319, 95th(us): 11551, 99th(us): 12023, 99.9th(us): 12599, 99.99th(us): 27455
Start  - Takes(s): 42.7, Count: 16672, OPS: 390.4, Avg(us): 28, Min(us): 13, Max(us): 857, 50th(us): 24, 90th(us): 37, 95th(us): 45, 99th(us): 197, 99.9th(us): 394, 99.99th(us): 660
TOTAL  - Takes(s): 42.7, Count: 213335, OPS: 4995.0, Avg(us): 9324, Min(us): 0, Max(us): 83519, 50th(us): 3579, 90th(us): 39231, 95th(us): 43775, 99th(us): 48799, 99.9th(us): 55743, 99.99th(us): 65247
TXN    - Takes(s): 42.7, Count: 15855, OPS: 371.5, Avg(us): 41076, Min(us): 17984, Max(us): 83519, 50th(us): 40607, 90th(us): 47935, 95th(us): 50591, 99th(us): 55007, 99.9th(us): 63327, 99.99th(us): 75967
TXN_ERROR - Takes(s): 42.7, Count: 801, OPS: 18.8, Avg(us): 33970, Min(us): 20048, Max(us): 52031, 50th(us): 33343, 90th(us): 40607, 95th(us): 43807, 99th(us): 47871, 99.9th(us): 50943, 99.99th(us): 52031
TxnGroup - Takes(s): 42.7, Count: 16656, OPS: 390.2, Avg(us): 40720, Min(us): 17984, Max(us): 75647, 50th(us): 40479, 90th(us): 47903, 95th(us): 50879, 99th(us): 55135, 99.9th(us): 62879, 99.99th(us): 72703
UPDATE - Takes(s): 42.7, Count: 49288, OPS: 1154.2, Avg(us): 4, Min(us): 1, Max(us): 485, 50th(us): 3, 90th(us): 5, 95th(us): 7, 99th(us): 19, 99.9th(us): 227, 99.99th(us): 374
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch     801

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    543
rollForward failed
  version mismatch  427
rollback failed
  version mismatch  21
```

+ 32

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 32
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 21.173039041s
COMMIT - Takes(s): 21.1, Count: 15336, OPS: 725.4, Avg(us): 17888, Min(us): 0, Max(us): 32047, 50th(us): 18127, 90th(us): 23791, 95th(us): 26111, 99th(us): 28367, 99.9th(us): 31295, 99.99th(us): 31823
COMMIT_ERROR - Takes(s): 21.1, Count: 1304, OPS: 61.7, Avg(us): 10798, Min(us): 3406, Max(us): 23439, 50th(us): 10895, 90th(us): 15935, 95th(us): 18127, 99th(us): 20063, 99.9th(us): 23375, 99.99th(us): 23439
READ   - Takes(s): 21.2, Count: 98311, OPS: 4643.9, Avg(us): 3679, Min(us): 4, Max(us): 16215, 50th(us): 3443, 90th(us): 4147, 95th(us): 4311, 99th(us): 10359, 99.9th(us): 11999, 99.99th(us): 14983
READ_ERROR - Takes(s): 21.1, Count: 1689, OPS: 79.9, Avg(us): 9048, Min(us): 6228, Max(us): 12599, 50th(us): 9735, 90th(us): 11511, 95th(us): 11783, 99th(us): 12175, 99.9th(us): 12407, 99.99th(us): 12599
Start  - Takes(s): 21.2, Count: 16672, OPS: 787.4, Avg(us): 33, Min(us): 13, Max(us): 935, 50th(us): 26, 90th(us): 42, 95th(us): 55, 99th(us): 240, 99.9th(us): 461, 99.99th(us): 690
TOTAL  - Takes(s): 21.2, Count: 211531, OPS: 9989.9, Avg(us): 9110, Min(us): 0, Max(us): 75775, 50th(us): 3367, 90th(us): 37951, 95th(us): 42559, 99th(us): 49471, 99.9th(us): 56511, 99.99th(us): 61279
TXN    - Takes(s): 21.1, Count: 15336, OPS: 725.4, Avg(us): 40652, Min(us): 17024, Max(us): 75775, 50th(us): 40735, 90th(us): 48031, 95th(us): 50015, 99th(us): 54751, 99.9th(us): 60287, 99.99th(us): 63967
TXN_ERROR - Takes(s): 21.1, Count: 1304, OPS: 61.7, Avg(us): 33651, Min(us): 22000, Max(us): 52767, 50th(us): 33311, 90th(us): 40575, 95th(us): 42463, 99th(us): 48319, 99.9th(us): 52479, 99.99th(us): 52767
TxnGroup - Takes(s): 21.2, Count: 16640, OPS: 786.5, Avg(us): 40068, Min(us): 17376, Max(us): 70911, 50th(us): 40159, 90th(us): 47839, 95th(us): 50175, 99th(us): 55615, 99.9th(us): 60991, 99.99th(us): 64863
UPDATE - Takes(s): 21.2, Count: 49236, OPS: 2325.7, Avg(us): 5, Min(us): 1, Max(us): 682, 50th(us): 4, 90th(us): 5, 95th(us): 8, 99th(us): 23, 99.9th(us): 337, 99.99th(us): 543
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  836
  read failed due to unknown txn status  802
rollback failed
  version mismatch  51

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    1304
```

+ 64

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 10.400845665s
COMMIT - Takes(s): 10.4, Count: 14758, OPS: 1423.4, Avg(us): 17365, Min(us): 0, Max(us): 53407, 50th(us): 17663, 90th(us): 23103, 95th(us): 25295, 99th(us): 28575, 99.9th(us): 44799, 99.99th(us): 52383
COMMIT_ERROR - Takes(s): 10.4, Count: 1882, OPS: 181.4, Avg(us): 10607, Min(us): 3408, Max(us): 36415, 50th(us): 10727, 90th(us): 15439, 95th(us): 18063, 99th(us): 19967, 99.9th(us): 22735, 99.99th(us): 36415
READ   - Takes(s): 10.4, Count: 97431, OPS: 9369.8, Avg(us): 3592, Min(us): 4, Max(us): 32191, 50th(us): 3377, 90th(us): 3953, 95th(us): 4239, 99th(us): 10239, 99.9th(us): 12239, 99.99th(us): 31007
READ_ERROR - Takes(s): 10.4, Count: 2569, OPS: 247.7, Avg(us): 8972, Min(us): 6204, Max(us): 36031, 50th(us): 9831, 90th(us): 11095, 95th(us): 11391, 99th(us): 12071, 99.9th(us): 13111, 99.99th(us): 36031
Start  - Takes(s): 10.4, Count: 16704, OPS: 1605.9, Avg(us): 40, Min(us): 14, Max(us): 10215, 50th(us): 28, 90th(us): 45, 95th(us): 58, 99th(us): 380, 99.9th(us): 1214, 99.99th(us): 1347
TOTAL  - Takes(s): 10.4, Count: 208889, OPS: 20083.1, Avg(us): 8844, Min(us): 0, Max(us): 82047, 50th(us): 3339, 90th(us): 36959, 95th(us): 41951, 99th(us): 48319, 99.9th(us): 57247, 99.99th(us): 70463
TXN    - Takes(s): 10.4, Count: 14758, OPS: 1423.2, Avg(us): 39956, Min(us): 16688, Max(us): 74303, 50th(us): 39775, 90th(us): 47039, 95th(us): 49471, 99th(us): 54879, 99.9th(us): 67327, 99.99th(us): 73663
TXN_ERROR - Takes(s): 10.4, Count: 1882, OPS: 181.4, Avg(us): 32961, Min(us): 19936, Max(us): 63903, 50th(us): 32607, 90th(us): 39487, 95th(us): 41471, 99th(us): 46111, 99.9th(us): 51263, 99.99th(us): 63903
TxnGroup - Takes(s): 10.4, Count: 16640, OPS: 1602.7, Avg(us): 39098, Min(us): 13168, Max(us): 82047, 50th(us): 39103, 90th(us): 46975, 95th(us): 49727, 99th(us): 55231, 99.9th(us): 68991, 99.99th(us): 78527
UPDATE - Takes(s): 10.4, Count: 48598, OPS: 4674.3, Avg(us): 6, Min(us): 1, Max(us): 1117, 50th(us): 4, 90th(us): 6, 95th(us): 8, 99th(us): 25, 99.9th(us): 502, 99.99th(us): 894
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    1882

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  1352
  read failed due to unknown txn status  1111
rollback failed
  version mismatch  106
```

+ 96

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 96
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 6.969214513s
COMMIT - Takes(s): 6.9, Count: 14321, OPS: 2065.9, Avg(us): 17012, Min(us): 0, Max(us): 31727, 50th(us): 17423, 90th(us): 22815, 95th(us): 24799, 99th(us): 27679, 99.9th(us): 29775, 99.99th(us): 31583
COMMIT_ERROR - Takes(s): 6.9, Count: 2287, OPS: 329.9, Avg(us): 10421, Min(us): 3410, Max(us): 25103, 50th(us): 10647, 90th(us): 15247, 95th(us): 17503, 99th(us): 19839, 99.9th(us): 22959, 99.99th(us): 25103
READ   - Takes(s): 7.0, Count: 96768, OPS: 13894.3, Avg(us): 3560, Min(us): 5, Max(us): 15359, 50th(us): 3359, 90th(us): 3915, 95th(us): 4239, 99th(us): 10055, 99.9th(us): 11767, 99.99th(us): 14367
READ_ERROR - Takes(s): 6.9, Count: 3232, OPS: 466.2, Avg(us): 8942, Min(us): 6204, Max(us): 13583, 50th(us): 9863, 90th(us): 10879, 95th(us): 11279, 99th(us): 12119, 99.9th(us): 12975, 99.99th(us): 13583
Start  - Takes(s): 7.0, Count: 16704, OPS: 2396.3, Avg(us): 40, Min(us): 14, Max(us): 1452, 50th(us): 28, 90th(us): 44, 95th(us): 56, 99th(us): 427, 99.9th(us): 961, 99.99th(us): 1434
TOTAL  - Takes(s): 7.0, Count: 206955, OPS: 29693.5, Avg(us): 8687, Min(us): 0, Max(us): 67135, 50th(us): 3325, 90th(us): 36767, 95th(us): 41599, 99th(us): 48159, 99.9th(us): 54687, 99.99th(us): 59743
TXN    - Takes(s): 6.9, Count: 14321, OPS: 2065.6, Avg(us): 39625, Min(us): 16784, Max(us): 65663, 50th(us): 39423, 90th(us): 46655, 95th(us): 49023, 99th(us): 53055, 99.9th(us): 58847, 99.99th(us): 64703
TXN_ERROR - Takes(s): 6.9, Count: 2287, OPS: 329.9, Avg(us): 32951, Min(us): 20416, Max(us): 53503, 50th(us): 32383, 90th(us): 39487, 95th(us): 41983, 99th(us): 46399, 99.9th(us): 51807, 99.99th(us): 53503
TxnGroup - Takes(s): 6.9, Count: 16608, OPS: 2389.9, Avg(us): 38610, Min(us): 17168, Max(us): 67135, 50th(us): 38527, 90th(us): 46463, 95th(us): 49119, 99th(us): 53919, 99.9th(us): 59359, 99.99th(us): 66687
UPDATE - Takes(s): 7.0, Count: 48233, OPS: 6925.0, Avg(us): 6, Min(us): 1, Max(us): 1376, 50th(us): 4, 90th(us): 6, 95th(us): 8, 99th(us): 22, 99.9th(us): 629, 99.99th(us): 1116
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    2287

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  1695
  read failed due to unknown txn status  1318
rollback failed
  version mismatch  219
```

+ 128

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 5.236429108s
COMMIT - Takes(s): 5.2, Count: 13998, OPS: 2683.9, Avg(us): 16971, Min(us): 0, Max(us): 54879, 50th(us): 17279, 90th(us): 22847, 95th(us): 24767, 99th(us): 27951, 99.9th(us): 47071, 99.99th(us): 53887
COMMIT_ERROR - Takes(s): 5.2, Count: 2642, OPS: 507.0, Avg(us): 10435, Min(us): 3326, Max(us): 44895, 50th(us): 10591, 90th(us): 15295, 95th(us): 17903, 99th(us): 21007, 99.9th(us): 40223, 99.99th(us): 44895
READ   - Takes(s): 5.2, Count: 96162, OPS: 18374.6, Avg(us): 3581, Min(us): 5, Max(us): 40831, 50th(us): 3349, 90th(us): 3927, 95th(us): 4287, 99th(us): 10071, 99.9th(us): 14439, 99.99th(us): 32591
READ_ERROR - Takes(s): 5.2, Count: 3838, OPS: 737.0, Avg(us): 8970, Min(us): 6204, Max(us): 40223, 50th(us): 9815, 90th(us): 10879, 95th(us): 11351, 99th(us): 12535, 99.9th(us): 35359, 99.99th(us): 40223
Start  - Takes(s): 5.2, Count: 16768, OPS: 3202.2, Avg(us): 41, Min(us): 14, Max(us): 6923, 50th(us): 29, 90th(us): 44, 95th(us): 53, 99th(us): 483, 99.9th(us): 1179, 99.99th(us): 1973
TOTAL  - Takes(s): 5.2, Count: 205594, OPS: 39252.1, Avg(us): 8690, Min(us): 0, Max(us): 84095, 50th(us): 3317, 90th(us): 36735, 95th(us): 41567, 99th(us): 48735, 99.9th(us): 62527, 99.99th(us): 76415
TXN    - Takes(s): 5.2, Count: 13998, OPS: 2683.7, Avg(us): 39935, Min(us): 16688, Max(us): 81151, 50th(us): 39679, 90th(us): 47263, 95th(us): 49791, 99th(us): 57535, 99.9th(us): 75263, 99.99th(us): 80895
TXN_ERROR - Takes(s): 5.2, Count: 2642, OPS: 506.9, Avg(us): 33212, Min(us): 20224, Max(us): 69503, 50th(us): 32623, 90th(us): 39967, 95th(us): 42431, 99th(us): 49247, 99.9th(us): 65599, 99.99th(us): 69503
TxnGroup - Takes(s): 5.2, Count: 16640, OPS: 3189.2, Avg(us): 38741, Min(us): 16344, Max(us): 84095, 50th(us): 38367, 90th(us): 47295, 95th(us): 50079, 99th(us): 58175, 99.9th(us): 75775, 99.99th(us): 79999
UPDATE - Takes(s): 5.2, Count: 48028, OPS: 9176.7, Avg(us): 6, Min(us): 1, Max(us): 1635, 50th(us): 4, 90th(us): 5, 95th(us): 8, 99th(us): 20, 99.9th(us): 622, 99.99th(us): 1047
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    2642

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  2017
  read failed due to unknown txn status  1561
rollback failed
  version mismatch  260
```

##### Oreo

+ 8

```bash
----------------------------------
Run finished, takes 1m4.492350022s
COMMIT - Takes(s): 64.5, Count: 15979, OPS: 247.9, Avg(us): 7414, Min(us): 0, Max(us): 45727, 50th(us): 7515, 90th(us): 7819, 95th(us): 7927, 99th(us): 8255, 99.9th(us): 8983, 99.99th(us): 9175
COMMIT_ERROR - Takes(s): 64.3, Count: 685, OPS: 10.7, Avg(us): 4222, Min(us): 3510, Max(us): 5251, 50th(us): 4207, 90th(us): 4531, 95th(us): 4639, 99th(us): 4923, 99.9th(us): 5163, 99.99th(us): 5251
READ   - Takes(s): 64.5, Count: 99771, OPS: 1547.1, Avg(us): 3921, Min(us): 5, Max(us): 42111, 50th(us): 3915, 90th(us): 4299, 95th(us): 4407, 99th(us): 4599, 99.9th(us): 4899, 99.99th(us): 5259
READ_ERROR - Takes(s): 64.1, Count: 229, OPS: 3.6, Avg(us): 4264, Min(us): 3404, Max(us): 5027, 50th(us): 4279, 90th(us): 4643, 95th(us): 4799, 99th(us): 4991, 99.9th(us): 5027, 99.99th(us): 5027
Start  - Takes(s): 64.5, Count: 16672, OPS: 258.5, Avg(us): 21, Min(us): 13, Max(us): 379, 50th(us): 19, 90th(us): 28, 95th(us): 30, 99th(us): 41, 99.9th(us): 232, 99.99th(us): 301
TOTAL  - Takes(s): 64.5, Count: 214878, OPS: 3331.8, Avg(us): 7078, Min(us): 0, Max(us): 69951, 50th(us): 3853, 90th(us): 31007, 95th(us): 31407, 99th(us): 31903, 99.9th(us): 32559, 99.99th(us): 38047
TXN    - Takes(s): 64.5, Count: 15979, OPS: 247.9, Avg(us): 31035, Min(us): 19744, Max(us): 69439, 50th(us): 31215, 90th(us): 31775, 95th(us): 31951, 99th(us): 32335, 99.9th(us): 33727, 99.99th(us): 69055
TXN_ERROR - Takes(s): 64.3, Count: 685, OPS: 10.7, Avg(us): 27705, Min(us): 19440, Max(us): 30143, 50th(us): 27887, 90th(us): 28527, 95th(us): 28719, 99th(us): 28975, 99.9th(us): 29487, 99.99th(us): 30143
TxnGroup - Takes(s): 64.5, Count: 16664, OPS: 258.5, Avg(us): 30895, Min(us): 19008, Max(us): 69951, 50th(us): 31215, 90th(us): 31807, 95th(us): 31967, 99th(us): 32415, 99.9th(us): 33695, 99.99th(us): 69311
UPDATE - Takes(s): 64.5, Count: 49813, OPS: 772.4, Avg(us): 3, Min(us): 1, Max(us): 386, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 10, 99.9th(us): 121, 99.99th(us): 260
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  685

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  214
  read failed due to unknown txn status   10
rollback failed
  version mismatch  5
```

+ 16

```bash
----------------------------------
Run finished, takes 32.183317853s
COMMIT - Takes(s): 32.2, Count: 15402, OPS: 478.9, Avg(us): 7464, Min(us): 0, Max(us): 43615, 50th(us): 7579, 90th(us): 8007, 95th(us): 8131, 99th(us): 8431, 99.9th(us): 9191, 99.99th(us): 40095
COMMIT_ERROR - Takes(s): 32.1, Count: 1254, OPS: 39.0, Avg(us): 4283, Min(us): 3564, Max(us): 6391, 50th(us): 4243, 90th(us): 4719, 95th(us): 4847, 99th(us): 5187, 99.9th(us): 5807, 99.99th(us): 6391
READ   - Takes(s): 32.2, Count: 99679, OPS: 3097.5, Avg(us): 3912, Min(us): 5, Max(us): 40191, 50th(us): 3885, 90th(us): 4435, 95th(us): 4579, 99th(us): 4835, 99.9th(us): 5211, 99.99th(us): 5731
READ_ERROR - Takes(s): 32.1, Count: 321, OPS: 10.0, Avg(us): 4360, Min(us): 3416, Max(us): 5507, 50th(us): 4355, 90th(us): 4955, 95th(us): 5087, 99th(us): 5263, 99.9th(us): 5507, 99.99th(us): 5507
Start  - Takes(s): 32.2, Count: 16672, OPS: 518.0, Avg(us): 23, Min(us): 13, Max(us): 488, 50th(us): 21, 90th(us): 31, 95th(us): 36, 99th(us): 53, 99.9th(us): 252, 99.99th(us): 356
TOTAL  - Takes(s): 32.2, Count: 214162, OPS: 6654.4, Avg(us): 6987, Min(us): 0, Max(us): 68799, 50th(us): 3751, 90th(us): 30687, 95th(us): 31679, 99th(us): 32511, 99.9th(us): 33247, 99.99th(us): 63519
TXN    - Takes(s): 32.2, Count: 15402, OPS: 478.9, Avg(us): 31043, Min(us): 20000, Max(us): 68799, 50th(us): 31263, 90th(us): 32319, 95th(us): 32575, 99th(us): 33023, 99.9th(us): 35551, 99.99th(us): 68287
TXN_ERROR - Takes(s): 32.1, Count: 1254, OPS: 39.0, Avg(us): 27826, Min(us): 20208, Max(us): 63999, 50th(us): 27999, 90th(us): 29023, 95th(us): 29343, 99th(us): 29823, 99.9th(us): 60095, 99.99th(us): 63999
TxnGroup - Takes(s): 32.2, Count: 16656, OPS: 517.9, Avg(us): 30794, Min(us): 18368, Max(us): 68735, 50th(us): 31183, 90th(us): 32367, 95th(us): 32639, 99th(us): 33151, 99.9th(us): 33983, 99.99th(us): 68223
UPDATE - Takes(s): 32.2, Count: 50351, OPS: 1564.7, Avg(us): 3, Min(us): 1, Max(us): 520, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 12, 99.9th(us): 170, 99.99th(us): 384
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  273
  read failed due to unknown txn status   31
rollback failed
  version mismatch  17

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  1254
```

+ 32

```bash
----------------------------------
Run finished, takes 15.787526655s
COMMIT - Takes(s): 15.8, Count: 14718, OPS: 934.1, Avg(us): 7543, Min(us): 0, Max(us): 12591, 50th(us): 7607, 90th(us): 8279, 95th(us): 8519, 99th(us): 9055, 99.9th(us): 10543, 99.99th(us): 11751
COMMIT_ERROR - Takes(s): 15.8, Count: 1922, OPS: 122.0, Avg(us): 4337, Min(us): 3508, Max(us): 6867, 50th(us): 4255, 90th(us): 4895, 95th(us): 5107, 99th(us): 5539, 99.9th(us): 6095, 99.99th(us): 6867
READ   - Takes(s): 15.8, Count: 99541, OPS: 6306.9, Avg(us): 3818, Min(us): 5, Max(us): 9111, 50th(us): 3717, 90th(us): 4363, 95th(us): 4587, 99th(us): 5047, 99.9th(us): 5627, 99.99th(us): 7635
READ_ERROR - Takes(s): 15.7, Count: 459, OPS: 29.2, Avg(us): 4362, Min(us): 3278, Max(us): 6175, 50th(us): 4339, 90th(us): 5067, 95th(us): 5267, 99th(us): 5571, 99.9th(us): 6175, 99.99th(us): 6175
Start  - Takes(s): 15.8, Count: 16672, OPS: 1056.0, Avg(us): 29, Min(us): 13, Max(us): 958, 50th(us): 26, 90th(us): 37, 95th(us): 42, 99th(us): 183, 99.9th(us): 687, 99.99th(us): 943
TOTAL  - Takes(s): 15.8, Count: 212157, OPS: 13438.5, Avg(us): 6807, Min(us): 0, Max(us): 37151, 50th(us): 3647, 90th(us): 30175, 95th(us): 31087, 99th(us): 32191, 99.9th(us): 33247, 99.99th(us): 35359
TXN    - Takes(s): 15.8, Count: 14718, OPS: 934.1, Avg(us): 30579, Min(us): 19120, Max(us): 37151, 50th(us): 30719, 90th(us): 32015, 95th(us): 32367, 99th(us): 33119, 99.9th(us): 35359, 99.99th(us): 36959
TXN_ERROR - Takes(s): 15.8, Count: 1922, OPS: 122.0, Avg(us): 27384, Min(us): 19728, Max(us): 32735, 50th(us): 27455, 90th(us): 28703, 95th(us): 29007, 99th(us): 29743, 99.9th(us): 31439, 99.99th(us): 32735
TxnGroup - Takes(s): 15.8, Count: 16640, OPS: 1055.7, Avg(us): 30197, Min(us): 18336, Max(us): 36095, 50th(us): 30607, 90th(us): 31935, 95th(us): 32319, 99th(us): 33023, 99.9th(us): 34527, 99.99th(us): 35807
UPDATE - Takes(s): 15.8, Count: 49868, OPS: 3159.6, Avg(us): 4, Min(us): 1, Max(us): 914, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 14, 99.9th(us): 258, 99.99th(us): 779
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  1922

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  327
  read failed due to unknown txn status   98
rollback failed
  version mismatch  34
```

+ 64

```bash
----------------------------------
Run finished, takes 7.555554429s
COMMIT - Takes(s): 7.5, Count: 13623, OPS: 1808.9, Avg(us): 7339, Min(us): 0, Max(us): 16943, 50th(us): 7367, 90th(us): 8099, 95th(us): 8423, 99th(us): 9327, 99.9th(us): 12303, 99.99th(us): 16927
COMMIT_ERROR - Takes(s): 7.5, Count: 3017, OPS: 401.1, Avg(us): 4288, Min(us): 3420, Max(us): 9367, 50th(us): 4171, 90th(us): 4847, 95th(us): 5139, 99th(us): 6119, 99.9th(us): 7271, 99.99th(us): 9367
READ   - Takes(s): 7.6, Count: 99641, OPS: 13195.2, Avg(us): 3648, Min(us): 5, Max(us): 9535, 50th(us): 3583, 90th(us): 4029, 95th(us): 4259, 99th(us): 4815, 99.9th(us): 5795, 99.99th(us): 6635
READ_ERROR - Takes(s): 7.5, Count: 359, OPS: 47.8, Avg(us): 4020, Min(us): 3244, Max(us): 6663, 50th(us): 3971, 90th(us): 4619, 95th(us): 4843, 99th(us): 5711, 99.9th(us): 6663, 99.99th(us): 6663
Start  - Takes(s): 7.6, Count: 16704, OPS: 2210.8, Avg(us): 28, Min(us): 13, Max(us): 1054, 50th(us): 26, 90th(us): 38, 95th(us): 43, 99th(us): 93, 99.9th(us): 458, 99.99th(us): 898
TOTAL  - Takes(s): 7.6, Count: 210079, OPS: 27802.2, Avg(us): 6390, Min(us): 0, Max(us): 43583, 50th(us): 3535, 90th(us): 28751, 95th(us): 29551, 99th(us): 31055, 99.9th(us): 33407, 99.99th(us): 38623
TXN    - Takes(s): 7.5, Count: 13623, OPS: 1809.1, Avg(us): 29340, Min(us): 17744, Max(us): 43583, 50th(us): 29295, 90th(us): 30815, 95th(us): 31535, 99th(us): 33247, 99.9th(us): 38623, 99.99th(us): 43295
TXN_ERROR - Takes(s): 7.5, Count: 3017, OPS: 401.0, Avg(us): 26291, Min(us): 18112, Max(us): 35935, 50th(us): 26143, 90th(us): 27727, 95th(us): 28671, 99th(us): 30143, 99.9th(us): 34879, 99.99th(us): 35935
TxnGroup - Takes(s): 7.5, Count: 16640, OPS: 2209.2, Avg(us): 28760, Min(us): 18624, Max(us): 40383, 50th(us): 29087, 90th(us): 30575, 95th(us): 31263, 99th(us): 32895, 99.9th(us): 34975, 99.99th(us): 40031
UPDATE - Takes(s): 7.6, Count: 49848, OPS: 6602.1, Avg(us): 3, Min(us): 1, Max(us): 1307, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 13, 99.9th(us): 192, 99.99th(us): 918
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3017

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    176
rollForward failed
  version mismatch  153
rollback failed
  version mismatch  30
```

+ 96

```bash
----------------------------------
Run finished, takes 5.538992502s
COMMIT - Takes(s): 5.5, Count: 13135, OPS: 2381.7, Avg(us): 8182, Min(us): 0, Max(us): 18959, 50th(us): 8119, 90th(us): 9527, 95th(us): 10095, 99th(us): 12007, 99.9th(us): 14967, 99.99th(us): 18511
COMMIT_ERROR - Takes(s): 5.5, Count: 3473, OPS: 630.4, Avg(us): 4924, Min(us): 3488, Max(us): 12855, 50th(us): 4727, 90th(us): 5983, 95th(us): 6487, 99th(us): 7947, 99.9th(us): 10951, 99.99th(us): 12855
READ   - Takes(s): 5.5, Count: 99375, OPS: 17957.0, Avg(us): 3959, Min(us): 5, Max(us): 12559, 50th(us): 3773, 90th(us): 4715, 95th(us): 5127, 99th(us): 6107, 99.9th(us): 8167, 99.99th(us): 10415
READ_ERROR - Takes(s): 5.5, Count: 625, OPS: 113.6, Avg(us): 4877, Min(us): 3280, Max(us): 12151, 50th(us): 4655, 90th(us): 6287, 95th(us): 6863, 99th(us): 8319, 99.9th(us): 10487, 99.99th(us): 12151
Start  - Takes(s): 5.5, Count: 16704, OPS: 3015.5, Avg(us): 37, Min(us): 13, Max(us): 2611, 50th(us): 28, 90th(us): 40, 95th(us): 46, 99th(us): 315, 99.9th(us): 1341, 99.99th(us): 2393
TOTAL  - Takes(s): 5.5, Count: 208813, OPS: 37691.1, Avg(us): 6920, Min(us): 0, Max(us): 48543, 50th(us): 3675, 90th(us): 30895, 95th(us): 32431, 99th(us): 34751, 99.9th(us): 38303, 99.99th(us): 43231
TXN    - Takes(s): 5.5, Count: 13135, OPS: 2381.2, Avg(us): 32108, Min(us): 18528, Max(us): 48543, 50th(us): 32015, 90th(us): 34431, 95th(us): 35455, 99th(us): 38047, 99.9th(us): 42783, 99.99th(us): 48223
TXN_ERROR - Takes(s): 5.5, Count: 3473, OPS: 630.4, Avg(us): 28898, Min(us): 21552, Max(us): 38943, 50th(us): 28703, 90th(us): 31215, 95th(us): 32111, 99th(us): 34367, 99.9th(us): 37631, 99.99th(us): 38943
TxnGroup - Takes(s): 5.5, Count: 16608, OPS: 3010.3, Avg(us): 31398, Min(us): 18880, Max(us): 46175, 50th(us): 31583, 90th(us): 34143, 95th(us): 35071, 99th(us): 37343, 99.9th(us): 43071, 99.99th(us): 45727
UPDATE - Takes(s): 5.5, Count: 49856, OPS: 9006.3, Avg(us): 4, Min(us): 1, Max(us): 1881, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 16, 99.9th(us): 392, 99.99th(us): 1214
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3473

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  311
  read failed due to unknown txn status  258
rollback failed
  version mismatch  56
```

+ 128

```bash
----------------------------------
Run finished, takes 4.842894558s
COMMIT - Takes(s): 4.8, Count: 12727, OPS: 2666.4, Avg(us): 9713, Min(us): 0, Max(us): 55551, 50th(us): 9487, 90th(us): 11759, 95th(us): 12591, 99th(us): 15127, 99.9th(us): 46431, 99.99th(us): 51007
COMMIT_ERROR - Takes(s): 4.8, Count: 3913, OPS: 814.6, Avg(us): 5911, Min(us): 3678, Max(us): 42303, 50th(us): 5579, 90th(us): 7575, 95th(us): 8311, 99th(us): 10447, 99.9th(us): 35711, 99.99th(us): 42303
READ   - Takes(s): 4.8, Count: 99077, OPS: 20477.0, Avg(us): 4589, Min(us): 6, Max(us): 49663, 50th(us): 4247, 90th(us): 5979, 95th(us): 6647, 99th(us): 8279, 99.9th(us): 12527, 99.99th(us): 48831
READ_ERROR - Takes(s): 4.8, Count: 923, OPS: 193.5, Avg(us): 6365, Min(us): 3274, Max(us): 17231, 50th(us): 6059, 90th(us): 8503, 95th(us): 9487, 99th(us): 12103, 99.9th(us): 15079, 99.99th(us): 17231
Start  - Takes(s): 4.8, Count: 16768, OPS: 3462.0, Avg(us): 50, Min(us): 14, Max(us): 3345, 50th(us): 29, 90th(us): 43, 95th(us): 54, 99th(us): 783, 99.9th(us): 2257, 99.99th(us): 2909
TOTAL  - Takes(s): 4.8, Count: 207719, OPS: 42885.3, Avg(us): 8015, Min(us): 0, Max(us): 103615, 50th(us): 3987, 90th(us): 34943, 95th(us): 37823, 99th(us): 42207, 99.9th(us): 70975, 99.99th(us): 93567
TXN    - Takes(s): 4.8, Count: 12727, OPS: 2666.2, Avg(us): 37444, Min(us): 21616, Max(us): 103615, 50th(us): 36959, 90th(us): 41407, 95th(us): 43103, 99th(us): 49727, 99.9th(us): 93823, 99.99th(us): 95487
TXN_ERROR - Takes(s): 4.8, Count: 3913, OPS: 814.7, Avg(us): 34047, Min(us): 22336, Max(us): 91007, 50th(us): 33471, 90th(us): 37919, 95th(us): 39519, 99th(us): 45663, 99.9th(us): 87935, 99.99th(us): 91007
TxnGroup - Takes(s): 4.8, Count: 16640, OPS: 3454.2, Avg(us): 36592, Min(us): 20192, Max(us): 90815, 50th(us): 36223, 90th(us): 41023, 95th(us): 42751, 99th(us): 50015, 99.9th(us): 79999, 99.99th(us): 86399
UPDATE - Takes(s): 4.8, Count: 49780, OPS: 10292.4, Avg(us): 5, Min(us): 1, Max(us): 2147, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 16, 99.9th(us): 416, 99.99th(us): 1854
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3913

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  504
  read failed due to unknown txn status  295
rollback failed
  version mismatch  124
```

##### Native

+ 8

```bash
----------------------------------
Run finished, takes 1m5.551554352s
READ   - Takes(s): 65.5, Count: 100000, OPS: 1525.6, Avg(us): 3460, Min(us): 3102, Max(us): 15303, 50th(us): 3431, 90th(us): 3667, 95th(us): 3753, 99th(us): 3905, 99.9th(us): 4103, 99.99th(us): 4339
TOTAL  - Takes(s): 65.5, Count: 167000, OPS: 2547.7, Avg(us): 6249, Min(us): 3102, Max(us): 51935, 50th(us): 3455, 90th(us): 4243, 95th(us): 30911, 99th(us): 37535, 99.9th(us): 41215, 99.99th(us): 44415
TxnGroup - Takes(s): 65.5, Count: 16664, OPS: 254.3, Avg(us): 31334, Min(us): 20144, Max(us): 51935, 50th(us): 30927, 90th(us): 37535, 95th(us): 37951, 99th(us): 41215, 99.9th(us): 44415, 99.99th(us): 50143
UPDATE - Takes(s): 65.5, Count: 50336, OPS: 768.0, Avg(us): 3484, Min(us): 3118, Max(us): 4403, 50th(us): 3451, 90th(us): 3693, 95th(us): 3765, 99th(us): 3905, 99.9th(us): 4085, 99.99th(us): 4335
Error Summary:
```

+ 16

```bash
----------------------------------
Run finished, takes 33.047186757s
READ   - Takes(s): 33.0, Count: 100000, OPS: 3026.2, Avg(us): 3468, Min(us): 3100, Max(us): 4647, 50th(us): 3457, 90th(us): 3623, 95th(us): 3707, 99th(us): 3895, 99.9th(us): 4163, 99.99th(us): 4303
TOTAL  - Takes(s): 33.0, Count: 166967, OPS: 5052.9, Avg(us): 6261, Min(us): 3100, Max(us): 46079, 50th(us): 3479, 90th(us): 4247, 95th(us): 31231, 99th(us): 37919, 99.9th(us): 41599, 99.99th(us): 44127
TxnGroup - Takes(s): 33.0, Count: 16656, OPS: 504.4, Avg(us): 31407, Min(us): 20368, Max(us): 46079, 50th(us): 31231, 90th(us): 37919, 95th(us): 38239, 99th(us): 41599, 99.9th(us): 44127, 99.99th(us): 45375
UPDATE - Takes(s): 33.0, Count: 50311, OPS: 1522.7, Avg(us): 3489, Min(us): 3110, Max(us): 4471, 50th(us): 3477, 90th(us): 3607, 95th(us): 3711, 99th(us): 3889, 99.9th(us): 4159, 99.99th(us): 4275
Error Summary:
```

+ 32

```bash
----------------------------------
Run finished, takes 17.108203493s
READ   - Takes(s): 17.1, Count: 100000, OPS: 5846.2, Avg(us): 3589, Min(us): 3098, Max(us): 14919, 50th(us): 3577, 90th(us): 3875, 95th(us): 3957, 99th(us): 4131, 99.9th(us): 4455, 99.99th(us): 14471
TOTAL  - Takes(s): 17.1, Count: 166586, OPS: 9738.9, Avg(us): 6483, Min(us): 3098, Max(us): 51167, 50th(us): 3615, 90th(us): 11199, 95th(us): 32143, 99th(us): 38783, 99.9th(us): 42719, 99.99th(us): 46047
TxnGroup - Takes(s): 17.1, Count: 16640, OPS: 974.1, Avg(us): 32489, Min(us): 19280, Max(us): 51167, 50th(us): 32143, 90th(us): 38783, 95th(us): 39455, 99th(us): 42719, 99.9th(us): 46047, 99.99th(us): 46815
UPDATE - Takes(s): 17.1, Count: 49946, OPS: 2920.7, Avg(us): 3614, Min(us): 3102, Max(us): 14831, 50th(us): 3605, 90th(us): 3887, 95th(us): 3959, 99th(us): 4115, 99.9th(us): 4331, 99.99th(us): 5579
Error Summary:
```

+ 64

```bash
----------------------------------
Run finished, takes 8.672961965s
READ   - Takes(s): 8.7, Count: 100000, OPS: 11534.1, Avg(us): 3586, Min(us): 3098, Max(us): 8607, 50th(us): 3563, 90th(us): 4055, 95th(us): 4183, 99th(us): 4435, 99.9th(us): 5015, 99.99th(us): 7439
TOTAL  - Takes(s): 8.7, Count: 166623, OPS: 19218.3, Avg(us): 6491, Min(us): 3096, Max(us): 47455, 50th(us): 3655, 90th(us): 7223, 95th(us): 32927, 99th(us): 38399, 99.9th(us): 43103, 99.99th(us): 45823
TxnGroup - Takes(s): 8.7, Count: 16640, OPS: 1923.7, Avg(us): 32600, Min(us): 19184, Max(us): 47455, 50th(us): 32927, 90th(us): 38399, 95th(us): 40927, 99th(us): 43103, 99.9th(us): 45823, 99.99th(us): 46687
UPDATE - Takes(s): 8.7, Count: 49983, OPS: 5767.7, Avg(us): 3611, Min(us): 3096, Max(us): 7931, 50th(us): 3627, 90th(us): 4059, 95th(us): 4179, 99th(us): 4407, 99.9th(us): 4871, 99.99th(us): 7255
Error Summary:
```

+ 96

```bash
----------------------------------
Run finished, takes 5.659554662s
READ   - Takes(s): 5.7, Count: 100000, OPS: 17677.6, Avg(us): 3492, Min(us): 3098, Max(us): 14975, 50th(us): 3311, 90th(us): 4107, 95th(us): 4287, 99th(us): 4695, 99.9th(us): 5679, 99.99th(us): 8127
TOTAL  - Takes(s): 5.7, Count: 166607, OPS: 29453.7, Avg(us): 6323, Min(us): 3098, Max(us): 50079, 50th(us): 3355, 90th(us): 7599, 95th(us): 31599, 99th(us): 38015, 99.9th(us): 43103, 99.99th(us): 46111
TxnGroup - Takes(s): 5.6, Count: 16608, OPS: 2946.7, Avg(us): 31839, Min(us): 19168, Max(us): 50079, 50th(us): 31615, 90th(us): 38015, 95th(us): 39583, 99th(us): 43103, 99.9th(us): 46111, 99.99th(us): 47423
UPDATE - Takes(s): 5.7, Count: 49999, OPS: 8844.3, Avg(us): 3510, Min(us): 3100, Max(us): 14975, 50th(us): 3337, 90th(us): 4127, 95th(us): 4287, 99th(us): 4635, 99.9th(us): 5351, 99.99th(us): 8043
Error Summary:
```

+ 128

```bash
----------------------------------
Run finished, takes 4.155971376s
READ   - Takes(s): 4.2, Count: 100000, OPS: 24080.1, Avg(us): 3424, Min(us): 3096, Max(us): 25919, 50th(us): 3289, 90th(us): 3833, 95th(us): 4065, 99th(us): 4583, 99.9th(us): 13335, 99.99th(us): 22287
TOTAL  - Takes(s): 4.2, Count: 166476, OPS: 40091.4, Avg(us): 6188, Min(us): 3096, Max(us): 78207, 50th(us): 3321, 90th(us): 19247, 95th(us): 31087, 99th(us): 36703, 99.9th(us): 42591, 99.99th(us): 58751
TxnGroup - Takes(s): 4.1, Count: 16640, OPS: 4033.8, Avg(us): 31072, Min(us): 19040, Max(us): 78207, 50th(us): 31103, 90th(us): 36703, 95th(us): 38719, 99th(us): 42591, 99.9th(us): 58751, 99.99th(us): 72575
UPDATE - Takes(s): 4.1, Count: 49836, OPS: 12011.7, Avg(us): 3426, Min(us): 3098, Max(us): 17599, 50th(us): 3313, 90th(us): 3829, 95th(us): 4051, 99th(us): 4523, 99.9th(us): 9727, 99.99th(us): 16895
Error Summary:
```

### Workload A

#### mongo - mongo

##### Cherry Garcia

+ 8

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 8
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 1m24.058499994s
COMMIT - Takes(s): 84.0, Count: 16157, OPS: 192.3, Avg(us): 29221, Min(us): 0, Max(us): 62239, 50th(us): 29327, 90th(us): 43423, 95th(us): 44063, 99th(us): 51039, 99.9th(us): 55807, 99.99th(us): 61567
COMMIT_ERROR - Takes(s): 83.8, Count: 507, OPS: 6.1, Avg(us): 21957, Min(us): 3890, Max(us): 47071, 50th(us): 21935, 90th(us): 32831, 95th(us): 36447, 99th(us): 40543, 99.9th(us): 44319, 99.99th(us): 47071
READ   - Takes(s): 84.1, Count: 49679, OPS: 591.0, Avg(us): 3609, Min(us): 4, Max(us): 16175, 50th(us): 3567, 90th(us): 3799, 95th(us): 3917, 99th(us): 4151, 99.9th(us): 14455, 99.99th(us): 15039
READ_ERROR - Takes(s): 83.8, Count: 442, OPS: 5.3, Avg(us): 8750, Min(us): 6612, Max(us): 12327, 50th(us): 7371, 90th(us): 11095, 95th(us): 11327, 99th(us): 12023, 99.9th(us): 12327, 99.99th(us): 12327
Start  - Takes(s): 84.1, Count: 16672, OPS: 198.3, Avg(us): 24, Min(us): 13, Max(us): 585, 50th(us): 18, 90th(us): 31, 95th(us): 40, 99th(us): 168, 99.9th(us): 282, 99.99th(us): 453
TOTAL  - Takes(s): 84.1, Count: 165208, OPS: 1965.4, Avg(us): 11936, Min(us): 0, Max(us): 77631, 50th(us): 3495, 90th(us): 40383, 95th(us): 44255, 99th(us): 54463, 99.9th(us): 63679, 99.99th(us): 72255
TXN    - Takes(s): 84.0, Count: 16157, OPS: 192.3, Avg(us): 40366, Min(us): 17840, Max(us): 62335, 50th(us): 40191, 90th(us): 47391, 95th(us): 47903, 99th(us): 51743, 99.9th(us): 58303, 99.99th(us): 61663
TXN_ERROR - Takes(s): 83.8, Count: 507, OPS: 6.1, Avg(us): 30999, Min(us): 14568, Max(us): 52255, 50th(us): 32255, 90th(us): 39679, 95th(us): 40319, 99th(us): 44063, 99.9th(us): 50655, 99.99th(us): 52255
TxnGroup - Takes(s): 84.1, Count: 16664, OPS: 198.2, Avg(us): 40070, Min(us): 77, Max(us): 77631, 50th(us): 40159, 90th(us): 54239, 95th(us): 57887, 99th(us): 63647, 99.9th(us): 72255, 99.99th(us): 76607
UPDATE - Takes(s): 84.1, Count: 49879, OPS: 593.4, Avg(us): 3, Min(us): 1, Max(us): 450, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 14, 99.9th(us): 150, 99.99th(us): 333
Error Summary:

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    261
rollForward failed
  version mismatch  174
rollback failed
  version mismatch  7

                                                   Operation:  COMMIT
                                                        Error   Count
                                                        -----   -----
  prepare phase failed: read failed due to unknown txn status     219
prepare phase failed: rollForward failed
                        version mismatch  161
  prepare phase failed: version mismatch  124
prepare phase failed: rollback failed
  version mismatch  3
```

+ 16

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 16
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 42.472389684s
COMMIT - Takes(s): 42.4, Count: 15590, OPS: 367.3, Avg(us): 29642, Min(us): 0, Max(us): 63391, 50th(us): 30015, 90th(us): 44223, 95th(us): 45087, 99th(us): 52095, 99.9th(us): 53759, 99.99th(us): 59711
COMMIT_ERROR - Takes(s): 42.4, Count: 1066, OPS: 25.1, Avg(us): 21424, Min(us): 3886, Max(us): 48671, 50th(us): 22271, 90th(us): 33055, 95th(us): 36799, 99th(us): 41183, 99.9th(us): 48639, 99.99th(us): 48671
READ   - Takes(s): 42.5, Count: 49413, OPS: 1163.4, Avg(us): 3656, Min(us): 4, Max(us): 16311, 50th(us): 3629, 90th(us): 3937, 95th(us): 4037, 99th(us): 4271, 99.9th(us): 14727, 99.99th(us): 15583
READ_ERROR - Takes(s): 42.4, Count: 797, OPS: 18.8, Avg(us): 8898, Min(us): 6572, Max(us): 12263, 50th(us): 7643, 90th(us): 11351, 95th(us): 11503, 99th(us): 11791, 99.9th(us): 12143, 99.99th(us): 12263
Start  - Takes(s): 42.5, Count: 16672, OPS: 392.5, Avg(us): 30, Min(us): 13, Max(us): 763, 50th(us): 25, 90th(us): 41, 95th(us): 54, 99th(us): 214, 99.9th(us): 433, 99.99th(us): 652
TOTAL  - Takes(s): 42.5, Count: 163711, OPS: 3854.4, Avg(us): 11970, Min(us): 0, Max(us): 82943, 50th(us): 3513, 90th(us): 41279, 95th(us): 45375, 99th(us): 55359, 99.9th(us): 64287, 99.99th(us): 71679
TXN    - Takes(s): 42.4, Count: 15590, OPS: 367.3, Avg(us): 41168, Min(us): 17888, Max(us): 68031, 50th(us): 41119, 90th(us): 48319, 95th(us): 49023, 99th(us): 52671, 99.9th(us): 59487, 99.99th(us): 62847
TXN_ERROR - Takes(s): 42.4, Count: 1066, OPS: 25.1, Avg(us): 30875, Min(us): 10992, Max(us): 50879, 50th(us): 30223, 90th(us): 38175, 95th(us): 40991, 99th(us): 44895, 99.9th(us): 48767, 99.99th(us): 50879
TxnGroup - Takes(s): 42.5, Count: 16656, OPS: 392.2, Avg(us): 40483, Min(us): 81, Max(us): 82943, 50th(us): 40927, 90th(us): 54527, 95th(us): 58719, 99th(us): 64191, 99.9th(us): 71423, 99.99th(us): 78591
UPDATE - Takes(s): 42.5, Count: 49790, OPS: 1172.3, Avg(us): 4, Min(us): 1, Max(us): 681, 50th(us): 3, 90th(us): 5, 95th(us): 7, 99th(us): 26, 99.9th(us): 256, 99.99th(us): 446
Error Summary:

                                                   Operation:  COMMIT
                                                        Error   Count
                                                        -----   -----
  prepare phase failed: read failed due to unknown txn status     472
                       prepare phase failed: version mismatch     298
prepare phase failed: rollForward failed
  version mismatch  282
prepare phase failed: rollback failed
  version mismatch  14

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    466
rollForward failed
  version mismatch  314
rollback failed
  version mismatch  17
```

+ 32

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 32
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 21.475377814s
COMMIT - Takes(s): 21.4, Count: 14886, OPS: 694.2, Avg(us): 30044, Min(us): 0, Max(us): 66303, 50th(us): 30735, 90th(us): 42559, 95th(us): 46527, 99th(us): 52735, 99.9th(us): 55999, 99.99th(us): 63071
COMMIT_ERROR - Takes(s): 21.4, Count: 1754, OPS: 81.8, Avg(us): 21561, Min(us): 4824, Max(us): 49535, 50th(us): 21711, 90th(us): 31951, 95th(us): 35839, 99th(us): 41535, 99.9th(us): 48863, 99.99th(us): 49535
READ   - Takes(s): 21.5, Count: 48505, OPS: 2259.0, Avg(us): 3685, Min(us): 4, Max(us): 16767, 50th(us): 3561, 90th(us): 4171, 95th(us): 4307, 99th(us): 4595, 99.9th(us): 14431, 99.99th(us): 15951
READ_ERROR - Takes(s): 21.4, Count: 1396, OPS: 65.1, Avg(us): 8802, Min(us): 6464, Max(us): 13367, 50th(us): 7907, 90th(us): 11647, 95th(us): 11951, 99th(us): 12255, 99.9th(us): 12927, 99.99th(us): 13367
Start  - Takes(s): 21.5, Count: 16672, OPS: 776.3, Avg(us): 35, Min(us): 13, Max(us): 885, 50th(us): 26, 90th(us): 45, 95th(us): 62, 99th(us): 252, 99.9th(us): 457, 99.99th(us): 807
TOTAL  - Takes(s): 21.5, Count: 161688, OPS: 7529.0, Avg(us): 11925, Min(us): 0, Max(us): 92863, 50th(us): 3373, 90th(us): 42399, 95th(us): 46879, 99th(us): 55551, 99.9th(us): 66751, 99.99th(us): 74175
TXN    - Takes(s): 21.4, Count: 14886, OPS: 694.2, Avg(us): 41893, Min(us): 18496, Max(us): 66879, 50th(us): 42239, 90th(us): 48703, 95th(us): 50847, 99th(us): 54751, 99.9th(us): 61983, 99.99th(us): 66751
TXN_ERROR - Takes(s): 21.4, Count: 1754, OPS: 81.8, Avg(us): 31344, Min(us): 13864, Max(us): 59551, 50th(us): 31263, 90th(us): 39327, 95th(us): 42335, 99th(us): 45599, 99.9th(us): 52351, 99.99th(us): 59551
TxnGroup - Takes(s): 21.5, Count: 16640, OPS: 775.0, Avg(us): 40726, Min(us): 95, Max(us): 92863, 50th(us): 40607, 90th(us): 55039, 95th(us): 59039, 99th(us): 66623, 99.9th(us): 74175, 99.99th(us): 78847
UPDATE - Takes(s): 21.5, Count: 50099, OPS: 2332.7, Avg(us): 5, Min(us): 1, Max(us): 1199, 50th(us): 3, 90th(us): 5, 95th(us): 8, 99th(us): 29, 99.9th(us): 381, 99.99th(us): 635
Error Summary:

                                                   Operation:  COMMIT
                                                        Error   Count
                                                        -----   -----
  prepare phase failed: read failed due to unknown txn status     873
                       prepare phase failed: version mismatch     434
prepare phase failed: rollForward failed
  version mismatch  407
prepare phase failed: rollback failed
  version mismatch  40

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    852
rollForward failed
  version mismatch  507
rollback failed
  version mismatch  37
```

+ 64

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 10.315162937s
COMMIT - Takes(s): 10.3, Count: 14017, OPS: 1361.9, Avg(us): 28545, Min(us): 0, Max(us): 61343, 50th(us): 28847, 90th(us): 40319, 95th(us): 43935, 99th(us): 50079, 99.9th(us): 54367, 99.99th(us): 59199
COMMIT_ERROR - Takes(s): 10.3, Count: 2623, OPS: 255.0, Avg(us): 20742, Min(us): 3628, Max(us): 48959, 50th(us): 20911, 90th(us): 31295, 95th(us): 35039, 99th(us): 41823, 99.9th(us): 47263, 99.99th(us): 48959
READ   - Takes(s): 10.3, Count: 47990, OPS: 4653.6, Avg(us): 3547, Min(us): 5, Max(us): 16431, 50th(us): 3395, 90th(us): 3931, 95th(us): 4207, 99th(us): 4947, 99.9th(us): 13951, 99.99th(us): 15743
READ_ERROR - Takes(s): 10.3, Count: 2159, OPS: 209.7, Avg(us): 8405, Min(us): 6452, Max(us): 13567, 50th(us): 7379, 90th(us): 10863, 95th(us): 11351, 99th(us): 11999, 99.9th(us): 12959, 99.99th(us): 13567
Start  - Takes(s): 10.3, Count: 16704, OPS: 1619.1, Avg(us): 40, Min(us): 14, Max(us): 1425, 50th(us): 28, 90th(us): 47, 95th(us): 60, 99th(us): 353, 99.9th(us): 1144, 99.99th(us): 1417
TOTAL  - Takes(s): 10.3, Count: 159219, OPS: 15434.5, Avg(us): 11185, Min(us): 0, Max(us): 77695, 50th(us): 3329, 90th(us): 40319, 95th(us): 44927, 99th(us): 53439, 99.9th(us): 64511, 99.99th(us): 72063
TXN    - Takes(s): 10.3, Count: 14017, OPS: 1361.8, Avg(us): 40362, Min(us): 17056, Max(us): 62655, 50th(us): 40447, 90th(us): 47071, 95th(us): 48991, 99th(us): 53247, 99.9th(us): 58143, 99.99th(us): 61471
TXN_ERROR - Takes(s): 10.3, Count: 2623, OPS: 255.0, Avg(us): 30484, Min(us): 10176, Max(us): 53983, 50th(us): 30559, 90th(us): 38623, 95th(us): 41183, 99th(us): 45247, 99.9th(us): 51679, 99.99th(us): 53983
TxnGroup - Takes(s): 10.3, Count: 16640, OPS: 1614.0, Avg(us): 38692, Min(us): 123, Max(us): 77695, 50th(us): 38591, 90th(us): 52479, 95th(us): 56607, 99th(us): 64319, 99.9th(us): 71999, 99.99th(us): 76479
UPDATE - Takes(s): 10.3, Count: 49851, OPS: 4832.1, Avg(us): 5, Min(us): 1, Max(us): 1335, 50th(us): 3, 90th(us): 6, 95th(us): 8, 99th(us): 24, 99.9th(us): 531, 99.99th(us): 1093
Error Summary:

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1322
rollForward failed
  version mismatch  754
rollback failed
  version mismatch  83

                                                   Operation:  COMMIT
                                                        Error   Count
                                                        -----   -----
  prepare phase failed: read failed due to unknown txn status    1242
prepare phase failed: rollForward failed
                        version mismatch  712
  prepare phase failed: version mismatch  589
prepare phase failed: rollback failed
  version mismatch  80
```

+ 96

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 96
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 6.790099394s
COMMIT - Takes(s): 6.8, Count: 13356, OPS: 1973.7, Avg(us): 27887, Min(us): 0, Max(us): 62719, 50th(us): 28127, 90th(us): 40127, 95th(us): 42655, 99th(us): 48895, 99.9th(us): 54463, 99.99th(us): 59647
COMMIT_ERROR - Takes(s): 6.8, Count: 3252, OPS: 480.5, Avg(us): 20412, Min(us): 3718, Max(us): 52351, 50th(us): 20783, 90th(us): 30831, 95th(us): 34559, 99th(us): 39775, 99.9th(us): 44959, 99.99th(us): 52351
READ   - Takes(s): 6.8, Count: 47029, OPS: 6929.4, Avg(us): 3495, Min(us): 5, Max(us): 16623, 50th(us): 3369, 90th(us): 3753, 95th(us): 4051, 99th(us): 5011, 99.9th(us): 13887, 99.99th(us): 15311
READ_ERROR - Takes(s): 6.8, Count: 2794, OPS: 413.4, Avg(us): 8288, Min(us): 6440, Max(us): 13815, 50th(us): 7103, 90th(us): 10511, 95th(us): 10895, 99th(us): 11911, 99.9th(us): 13247, 99.99th(us): 13815
Start  - Takes(s): 6.8, Count: 16704, OPS: 2460.1, Avg(us): 37, Min(us): 13, Max(us): 1813, 50th(us): 28, 90th(us): 44, 95th(us): 54, 99th(us): 278, 99.9th(us): 932, 99.99th(us): 1529
TOTAL  - Takes(s): 6.8, Count: 157230, OPS: 23155.0, Avg(us): 10773, Min(us): 0, Max(us): 76543, 50th(us): 3311, 90th(us): 39007, 95th(us): 44063, 99th(us): 52543, 99.9th(us): 63359, 99.99th(us): 71999
TXN    - Takes(s): 6.8, Count: 13356, OPS: 1973.7, Avg(us): 39697, Min(us): 13648, Max(us): 73023, 50th(us): 39359, 90th(us): 46207, 95th(us): 48511, 99th(us): 52415, 99.9th(us): 58047, 99.99th(us): 62847
TXN_ERROR - Takes(s): 6.8, Count: 3252, OPS: 480.5, Avg(us): 30213, Min(us): 10832, Max(us): 56767, 50th(us): 30703, 90th(us): 38303, 95th(us): 40735, 99th(us): 44863, 99.9th(us): 52543, 99.99th(us): 56767
TxnGroup - Takes(s): 6.8, Count: 16608, OPS: 2446.6, Avg(us): 37688, Min(us): 68, Max(us): 76543, 50th(us): 37951, 90th(us): 51871, 95th(us): 55775, 99th(us): 63103, 99.9th(us): 71359, 99.99th(us): 76287
UPDATE - Takes(s): 6.8, Count: 50177, OPS: 7389.0, Avg(us): 5, Min(us): 1, Max(us): 2037, 50th(us): 3, 90th(us): 5, 95th(us): 7, 99th(us): 20, 99.9th(us): 462, 99.99th(us): 1268
Error Summary:

                                                   Operation:  COMMIT
                                                        Error   Count
                                                        -----   -----
  prepare phase failed: read failed due to unknown txn status    1499
prepare phase failed: rollForward failed
                        version mismatch  889
  prepare phase failed: version mismatch  730
prepare phase failed: rollback failed
  version mismatch  134

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1649
rollForward failed
  version mismatch  1026
rollback failed
  version mismatch  119
```

+ 128

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 5.181697127s
COMMIT - Takes(s): 5.2, Count: 12925, OPS: 2506.5, Avg(us): 28205, Min(us): 0, Max(us): 62911, 50th(us): 28527, 90th(us): 41695, 95th(us): 43679, 99th(us): 50431, 99.9th(us): 56351, 99.99th(us): 62815
COMMIT_ERROR - Takes(s): 5.2, Count: 3715, OPS: 720.2, Avg(us): 20702, Min(us): 3782, Max(us): 50047, 50th(us): 21055, 90th(us): 31151, 95th(us): 34879, 99th(us): 40127, 99.9th(us): 46783, 99.99th(us): 50047
READ   - Takes(s): 5.2, Count: 46777, OPS: 9034.4, Avg(us): 3567, Min(us): 5, Max(us): 16399, 50th(us): 3385, 90th(us): 3913, 95th(us): 4263, 99th(us): 6095, 99.9th(us): 14143, 99.99th(us): 15519
READ_ERROR - Takes(s): 5.2, Count: 3096, OPS: 600.3, Avg(us): 8427, Min(us): 6440, Max(us): 22671, 50th(us): 7283, 90th(us): 10671, 95th(us): 11063, 99th(us): 12079, 99.9th(us): 20847, 99.99th(us): 22671
Start  - Takes(s): 5.2, Count: 16768, OPS: 3235.4, Avg(us): 45, Min(us): 13, Max(us): 2719, 50th(us): 30, 90th(us): 47, 95th(us): 58, 99th(us): 566, 99.9th(us): 1445, 99.99th(us): 2071
TOTAL  - Takes(s): 5.2, Count: 156162, OPS: 30131.6, Avg(us): 10821, Min(us): 0, Max(us): 81663, 50th(us): 3321, 90th(us): 39679, 95th(us): 44735, 99th(us): 53791, 99.9th(us): 64991, 99.99th(us): 74111
TXN    - Takes(s): 5.2, Count: 12925, OPS: 2506.1, Avg(us): 40493, Min(us): 19984, Max(us): 73471, 50th(us): 40159, 90th(us): 47455, 95th(us): 49791, 99th(us): 55263, 99.9th(us): 62559, 99.99th(us): 64575
TXN_ERROR - Takes(s): 5.2, Count: 3715, OPS: 720.4, Avg(us): 30777, Min(us): 7424, Max(us): 57471, 50th(us): 31151, 90th(us): 38975, 95th(us): 41631, 99th(us): 46207, 99.9th(us): 51807, 99.99th(us): 57471
TxnGroup - Takes(s): 5.2, Count: 16640, OPS: 3214.8, Avg(us): 38102, Min(us): 163, Max(us): 81663, 50th(us): 38271, 90th(us): 52671, 95th(us): 56703, 99th(us): 64671, 99.9th(us): 74047, 99.99th(us): 79999
UPDATE - Takes(s): 5.2, Count: 50127, OPS: 9673.9, Avg(us): 6, Min(us): 1, Max(us): 2001, 50th(us): 4, 90th(us): 6, 95th(us): 8, 99th(us): 22, 99.9th(us): 725, 99.99th(us): 1516
Error Summary:

                                                   Operation:  COMMIT
                                                        Error   Count
                                                        -----   -----
  prepare phase failed: read failed due to unknown txn status    1663
prepare phase failed: rollForward failed
                        version mismatch  1034
  prepare phase failed: version mismatch   848
prepare phase failed: rollback failed
  version mismatch  170

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1807
rollForward failed
  version mismatch  1105
rollback failed
  version mismatch  184
```

##### Epoxy

+ 8

```bash
{
    "name": "Workload A",
    "timeTaken": 66.255,
    "throughput": 1509.3200513168817,
    "latency": "Operation: SingleOperation,Count: 99984, Avg latency: 3, P50 latency: 4, P99 latency: 4, Sum: 352844\nOperation: WorkloadA,Count: 16664, Avg latency: 31799, P50 latency: 31845, P99 latency: 34946, Sum: 529913928\n",
    "txnPerThread": 2083
}
```

+ 16

```bash
{
    "name": "Workload A",
    "timeTaken": 33.262,
    "throughput": 3006.4337682640853,
    "latency": "Operation: SingleOperation,Count: 99942, Avg latency: 3, P50 latency: 3, P99 latency: 4, Sum: 349318\nOperation: WorkloadA,Count: 16656, Avg latency: 31481, P50 latency: 31299, P99 latency: 35111, Sum: 524357819\n",
    "txnPerThread": 1041
}
```

+ 32

```bash
{
    "name": "Workload A",
    "timeTaken": 17.02,
    "throughput": 5875.440658049354,
    "latency": "Operation: SingleOperation,Count: 99840, Avg latency: 3, P50 latency: 4, P99 latency: 5, Sum: 360044\nOperation: WorkloadA,Count: 16640, Avg latency: 32695, P50 latency: 32340, P99 latency: 39335, Sum: 544045319\n",
    "txnPerThread": 520
}
```

+ 64

```bash
{
    "name": "Workload A",
    "timeTaken": 9.36,
    "throughput": 10683.760683760684,
    "latency": "Operation: SingleOperation,Count: 99840, Avg latency: 3, P50 latency: 4, P99 latency: 6, Sum: 355563\nOperation: WorkloadA,Count: 16640, Avg latency: 35549, P50 latency: 35237, P99 latency: 44169, Sum: 591546199\n",
    "txnPerThread": 260
}
```

+ 96

```bash
{
    "name": "Workload A",
    "timeTaken": 7.77,
    "throughput": 12870.01287001287,
    "latency": "Operation: SingleOperation,Count: 99660, Avg latency: 3, P50 latency: 4, P99 latency: 8, Sum: 369529\nOperation: WorkloadA,Count: 16608, Avg latency: 41062, P50 latency: 40147, P99 latency: 66837, Sum: 681966119\n",
    "txnPerThread": 173
}
```

+ 128

```bash
{
    "name": "Workload A",
    "timeTaken": 7.321,
    "throughput": 13659.336156262807,
    "latency": "Operation: SingleOperation,Count: 99864, Avg latency: 3, P50 latency: 4, P99 latency: 9, Sum: 373029\nOperation: WorkloadA,Count: 16640, Avg latency: 51635, P50 latency: 45846, P99 latency: 89120, Sum: 859217481\n",
    "txnPerThread": 130
}
```

##### Oreo

+ 8

```bash
----------------------------------
Run finished, takes 42.888783001s
COMMIT - Takes(s): 42.9, Count: 16131, OPS: 376.2, Avg(us): 8336, Min(us): 0, Max(us): 10959, 50th(us): 8455, 90th(us): 9039, 95th(us): 9191, 99th(us): 9519, 99.9th(us): 9983, 99.99th(us): 10367
COMMIT_ERROR - Takes(s): 42.8, Count: 533, OPS: 12.5, Avg(us): 4836, Min(us): 3908, Max(us): 7719, 50th(us): 4811, 90th(us): 5451, 95th(us): 5591, 99th(us): 5907, 99.9th(us): 6299, 99.99th(us): 7719
READ   - Takes(s): 42.9, Count: 49802, OPS: 1161.3, Avg(us): 4024, Min(us): 5, Max(us): 6075, 50th(us): 3973, 90th(us): 4575, 95th(us): 4707, 99th(us): 5099, 99.9th(us): 5471, 99.99th(us): 5879
READ_ERROR - Takes(s): 42.5, Count: 227, OPS: 5.3, Avg(us): 4486, Min(us): 3614, Max(us): 5719, 50th(us): 4471, 90th(us): 5103, 95th(us): 5211, 99th(us): 5403, 99.9th(us): 5719, 99.99th(us): 5719
Start  - Takes(s): 42.9, Count: 16672, OPS: 388.7, Avg(us): 22, Min(us): 13, Max(us): 472, 50th(us): 19, 90th(us): 29, 95th(us): 32, 99th(us): 51, 99.9th(us): 246, 99.99th(us): 346
TOTAL  - Takes(s): 42.9, Count: 165371, OPS: 3855.7, Avg(us): 6085, Min(us): 0, Max(us): 35775, 50th(us): 3761, 90th(us): 20527, 95th(us): 24303, 99th(us): 28847, 99.9th(us): 32127, 99.99th(us): 34591
TXN    - Takes(s): 42.9, Count: 16131, OPS: 376.2, Avg(us): 20529, Min(us): 7624, Max(us): 31903, 50th(us): 20623, 90th(us): 26463, 95th(us): 28655, 99th(us): 30111, 99.9th(us): 30847, 99.99th(us): 31439
TXN_ERROR - Takes(s): 42.8, Count: 533, OPS: 12.5, Avg(us): 16328, Min(us): 4304, Max(us): 26495, 50th(us): 16767, 90th(us): 21647, 95th(us): 22527, 99th(us): 25599, 99.9th(us): 26447, 99.99th(us): 26495
TxnGroup - Takes(s): 42.9, Count: 16664, OPS: 388.6, Avg(us): 20391, Min(us): 78, Max(us): 35775, 50th(us): 20495, 90th(us): 27151, 95th(us): 29119, 99th(us): 32111, 99.9th(us): 34591, 99.99th(us): 35359
UPDATE - Takes(s): 42.9, Count: 49971, OPS: 1165.1, Avg(us): 2, Min(us): 1, Max(us): 322, 50th(us): 2, 90th(us): 4, 95th(us): 4, 99th(us): 9, 99.9th(us): 83, 99.99th(us): 226
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  414
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  103
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  11
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  5

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  179
  read failed due to unknown txn status   43
rollback failed
  version mismatch  4
     key not found  1
```

+ 16

```bash
----------------------------------
Run finished, takes 21.565899618s
COMMIT - Takes(s): 21.6, Count: 15702, OPS: 728.5, Avg(us): 8676, Min(us): 0, Max(us): 224383, 50th(us): 8479, 90th(us): 9231, 95th(us): 9447, 99th(us): 9983, 99.9th(us): 99263, 99.99th(us): 161023
COMMIT_ERROR - Takes(s): 21.6, Count: 954, OPS: 44.3, Avg(us): 5424, Min(us): 3860, Max(us): 149759, 50th(us): 4935, 90th(us): 5667, 95th(us): 5927, 99th(us): 6507, 99.9th(us): 136959, 99.99th(us): 149759
READ   - Takes(s): 21.6, Count: 49491, OPS: 2295.4, Avg(us): 3980, Min(us): 5, Max(us): 90623, 50th(us): 3869, 90th(us): 4579, 95th(us): 4771, 99th(us): 5263, 99.9th(us): 5791, 99.99th(us): 6391
READ_ERROR - Takes(s): 21.5, Count: 401, OPS: 18.6, Avg(us): 4454, Min(us): 3610, Max(us): 6095, 50th(us): 4387, 90th(us): 5115, 95th(us): 5311, 99th(us): 5571, 99.9th(us): 6095, 99.99th(us): 6095
Start  - Takes(s): 21.6, Count: 16672, OPS: 773.0, Avg(us): 24, Min(us): 13, Max(us): 754, 50th(us): 20, 90th(us): 30, 95th(us): 37, 99th(us): 65, 99.9th(us): 283, 99.99th(us): 464
TOTAL  - Takes(s): 21.6, Count: 164331, OPS: 7619.9, Avg(us): 6091, Min(us): 0, Max(us): 234751, 50th(us): 3709, 90th(us): 20479, 95th(us): 24175, 99th(us): 28799, 99.9th(us): 34239, 99.99th(us): 165759
TXN    - Takes(s): 21.6, Count: 15702, OPS: 728.5, Avg(us): 20742, Min(us): 7788, Max(us): 232447, 50th(us): 20607, 90th(us): 26447, 95th(us): 28559, 99th(us): 30095, 99.9th(us): 114623, 99.99th(us): 177151
TXN_ERROR - Takes(s): 21.6, Count: 954, OPS: 44.3, Avg(us): 16666, Min(us): 4524, Max(us): 153983, 50th(us): 16799, 90th(us): 21791, 95th(us): 23535, 99th(us): 26207, 99.9th(us): 153855, 99.99th(us): 153983
TxnGroup - Takes(s): 21.6, Count: 16656, OPS: 772.5, Avg(us): 20501, Min(us): 77, Max(us): 234751, 50th(us): 20447, 90th(us): 27119, 95th(us): 29039, 99th(us): 32799, 99.9th(us): 118399, 99.99th(us): 173055
UPDATE - Takes(s): 21.6, Count: 50108, OPS: 2323.5, Avg(us): 3, Min(us): 1, Max(us): 634, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 10, 99.9th(us): 134, 99.99th(us): 282
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  740
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  175
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  32
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  7

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  242
  read failed due to unknown txn status  148
rollback failed
  version mismatch  10
     key not found   1
```

+ 32

```bash
----------------------------------
Run finished, takes 10.28768547s
COMMIT - Takes(s): 10.3, Count: 15037, OPS: 1463.1, Avg(us): 8279, Min(us): 0, Max(us): 97023, 50th(us): 8163, 90th(us): 9015, 95th(us): 9327, 99th(us): 10191, 99.9th(us): 84031, 99.99th(us): 95295
COMMIT_ERROR - Takes(s): 10.3, Count: 1603, OPS: 156.0, Avg(us): 4955, Min(us): 3650, Max(us): 84991, 50th(us): 4699, 90th(us): 5491, 95th(us): 5791, 99th(us): 6487, 99.9th(us): 75967, 99.99th(us): 84991
READ   - Takes(s): 10.3, Count: 49551, OPS: 4819.0, Avg(us): 3784, Min(us): 5, Max(us): 87487, 50th(us): 3679, 90th(us): 4247, 95th(us): 4463, 99th(us): 5031, 99.9th(us): 6047, 99.99th(us): 9447
READ_ERROR - Takes(s): 10.3, Count: 413, OPS: 40.3, Avg(us): 4394, Min(us): 3592, Max(us): 86719, 50th(us): 4111, 90th(us): 4779, 95th(us): 5207, 99th(us): 5643, 99.9th(us): 86719, 99.99th(us): 86719
Start  - Takes(s): 10.3, Count: 16672, OPS: 1620.6, Avg(us): 24, Min(us): 13, Max(us): 1091, 50th(us): 20, 90th(us): 30, 95th(us): 37, 99th(us): 56, 99.9th(us): 732, 99.99th(us): 1061
TOTAL  - Takes(s): 10.3, Count: 162973, OPS: 15842.9, Avg(us): 5722, Min(us): 0, Max(us): 111743, 50th(us): 3601, 90th(us): 19423, 95th(us): 22879, 99th(us): 27007, 99.9th(us): 31807, 99.99th(us): 102847
TXN    - Takes(s): 10.3, Count: 15037, OPS: 1463.1, Avg(us): 19770, Min(us): 7832, Max(us): 110975, 50th(us): 19599, 90th(us): 25727, 95th(us): 26751, 99th(us): 28623, 99.9th(us): 97151, 99.99th(us): 109695
TXN_ERROR - Takes(s): 10.3, Count: 1603, OPS: 156.0, Avg(us): 16003, Min(us): 4396, Max(us): 87615, 50th(us): 15999, 90th(us): 20575, 95th(us): 22639, 99th(us): 24639, 99.9th(us): 83263, 99.99th(us): 87615
TxnGroup - Takes(s): 10.3, Count: 16640, OPS: 1617.6, Avg(us): 19391, Min(us): 76, Max(us): 111743, 50th(us): 19407, 90th(us): 25967, 95th(us): 27311, 99th(us): 30831, 99.9th(us): 96511, 99.99th(us): 107391
UPDATE - Takes(s): 10.3, Count: 50036, OPS: 4863.6, Avg(us): 2, Min(us): 1, Max(us): 438, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 10, 99.9th(us): 63, 99.99th(us): 248
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  1410
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  140
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  44
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  9

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    229
rollForward failed
  version mismatch  168
rollback failed
  version mismatch  10
     key not found   6
```

+ 64

```bash
----------------------------------
Run finished, takes 5.832147885s
COMMIT - Takes(s): 5.8, Count: 14148, OPS: 2430.4, Avg(us): 9551, Min(us): 0, Max(us): 18751, 50th(us): 9471, 90th(us): 11335, 95th(us): 12047, 99th(us): 13775, 99.9th(us): 15943, 99.99th(us): 18719
COMMIT_ERROR - Takes(s): 5.8, Count: 2492, OPS: 427.8, Avg(us): 6152, Min(us): 3792, Max(us): 16055, 50th(us): 5899, 90th(us): 7819, 95th(us): 8583, 99th(us): 10367, 99.9th(us): 12591, 99.99th(us): 16055
READ   - Takes(s): 5.8, Count: 49215, OPS: 8444.0, Avg(us): 4243, Min(us): 5, Max(us): 15039, 50th(us): 3995, 90th(us): 5211, 95th(us): 5779, 99th(us): 7195, 99.9th(us): 9823, 99.99th(us): 11807
READ_ERROR - Takes(s): 5.8, Count: 808, OPS: 139.2, Avg(us): 5024, Min(us): 3602, Max(us): 11519, 50th(us): 4787, 90th(us): 6231, 95th(us): 6903, 99th(us): 8535, 99.9th(us): 11343, 99.99th(us): 11519
Start  - Takes(s): 5.8, Count: 16704, OPS: 2863.4, Avg(us): 28, Min(us): 14, Max(us): 1510, 50th(us): 25, 90th(us): 36, 95th(us): 42, 99th(us): 160, 99.9th(us): 552, 99.99th(us): 999
TOTAL  - Takes(s): 5.8, Count: 160832, OPS: 27576.3, Avg(us): 6395, Min(us): 0, Max(us): 43999, 50th(us): 3771, 90th(us): 21983, 95th(us): 25743, 99th(us): 30943, 99.9th(us): 35775, 99.99th(us): 39903
TXN    - Takes(s): 5.8, Count: 14148, OPS: 2430.2, Avg(us): 22583, Min(us): 8200, Max(us): 40543, 50th(us): 22671, 90th(us): 29119, 95th(us): 30639, 99th(us): 33375, 99.9th(us): 36735, 99.99th(us): 39775
TXN_ERROR - Takes(s): 5.8, Count: 2492, OPS: 427.8, Avg(us): 18221, Min(us): 4644, Max(us): 33279, 50th(us): 18351, 90th(us): 24047, 95th(us): 25791, 99th(us): 28415, 99.9th(us): 31663, 99.99th(us): 33279
TxnGroup - Takes(s): 5.8, Count: 16640, OPS: 2853.4, Avg(us): 21899, Min(us): 78, Max(us): 43999, 50th(us): 21903, 90th(us): 29407, 95th(us): 31455, 99th(us): 35423, 99.9th(us): 39807, 99.99th(us): 41887
UPDATE - Takes(s): 5.8, Count: 49977, OPS: 8569.1, Avg(us): 3, Min(us): 1, Max(us): 1031, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 13, 99.9th(us): 230, 99.99th(us): 684
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2116
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  265
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  76
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  35

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    427
rollForward failed
  version mismatch  325
rollback failed
  version mismatch  33
     key not found  23
```

+ 96

```bash
----------------------------------
Run finished, takes 4.706961931s
COMMIT - Takes(s): 4.7, Count: 13565, OPS: 2892.3, Avg(us): 12090, Min(us): 0, Max(us): 36351, 50th(us): 11663, 90th(us): 16071, 95th(us): 18031, 99th(us): 21743, 99.9th(us): 26271, 99.99th(us): 31599
COMMIT_ERROR - Takes(s): 4.7, Count: 3043, OPS: 648.0, Avg(us): 8398, Min(us): 4042, Max(us): 27375, 50th(us): 7843, 90th(us): 11975, 95th(us): 13671, 99th(us): 17151, 99.9th(us): 22063, 99.99th(us): 27375
READ   - Takes(s): 4.7, Count: 48327, OPS: 10277.0, Avg(us): 4875, Min(us): 6, Max(us): 26447, 50th(us): 4339, 90th(us): 6759, 95th(us): 8007, 99th(us): 11223, 99.9th(us): 15567, 99.99th(us): 20031
READ_ERROR - Takes(s): 4.7, Count: 1322, OPS: 282.2, Avg(us): 6583, Min(us): 3644, Max(us): 19455, 50th(us): 5911, 90th(us): 10031, 95th(us): 11671, 99th(us): 14527, 99.9th(us): 17951, 99.99th(us): 19455
Start  - Takes(s): 4.7, Count: 16704, OPS: 3547.8, Avg(us): 31, Min(us): 13, Max(us): 2425, 50th(us): 26, 90th(us): 38, 95th(us): 43, 99th(us): 191, 99.9th(us): 903, 99.99th(us): 1797
TOTAL  - Takes(s): 4.7, Count: 159120, OPS: 33799.6, Avg(us): 7559, Min(us): 0, Max(us): 66047, 50th(us): 3915, 90th(us): 25983, 95th(us): 30831, 99th(us): 38239, 99.9th(us): 46239, 99.99th(us): 53343
TXN    - Takes(s): 4.7, Count: 13565, OPS: 2892.2, Avg(us): 27119, Min(us): 8688, Max(us): 65791, 50th(us): 26975, 90th(us): 35295, 95th(us): 38047, 99th(us): 43231, 99.9th(us): 50143, 99.99th(us): 58879
TXN_ERROR - Takes(s): 4.7, Count: 3043, OPS: 648.3, Avg(us): 22194, Min(us): 6220, Max(us): 44927, 50th(us): 22111, 90th(us): 30159, 95th(us): 32415, 99th(us): 37471, 99.9th(us): 43231, 99.99th(us): 44927
TxnGroup - Takes(s): 4.7, Count: 16608, OPS: 3531.9, Avg(us): 26168, Min(us): 77, Max(us): 66047, 50th(us): 25967, 90th(us): 35711, 95th(us): 38783, 99th(us): 44735, 99.9th(us): 52191, 99.99th(us): 60639
UPDATE - Takes(s): 4.7, Count: 50351, OPS: 10694.5, Avg(us): 3, Min(us): 1, Max(us): 5963, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 14, 99.9th(us): 220, 99.99th(us): 1076
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2433
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  384
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  153
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  73

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    729
rollForward failed
  version mismatch  472
rollback failed
  version mismatch  84
     key not found  37
```

+ 128

```bash
----------------------------------
Run finished, takes 4.799945953s
COMMIT - Takes(s): 4.8, Count: 12973, OPS: 2717.0, Avg(us): 16830, Min(us): 0, Max(us): 64095, 50th(us): 15271, 90th(us): 26095, 95th(us): 30607, 99th(us): 40735, 99.9th(us): 53215, 99.99th(us): 61727
COMMIT_ERROR - Takes(s): 4.8, Count: 3667, OPS: 768.2, Avg(us): 12888, Min(us): 3986, Max(us): 54975, 50th(us): 11335, 90th(us): 21359, 95th(us): 25295, 99th(us): 34079, 99.9th(us): 42719, 99.99th(us): 54975
READ   - Takes(s): 4.8, Count: 47790, OPS: 9965.2, Avg(us): 6290, Min(us): 5, Max(us): 65023, 50th(us): 4823, 90th(us): 10743, 95th(us): 13855, 99th(us): 22607, 99.9th(us): 33663, 99.99th(us): 48415
READ_ERROR - Takes(s): 4.8, Count: 1881, OPS: 394.1, Avg(us): 8915, Min(us): 3576, Max(us): 64159, 50th(us): 7155, 90th(us): 15599, 95th(us): 18943, 99th(us): 27311, 99.9th(us): 42271, 99.99th(us): 64159
Start  - Takes(s): 4.8, Count: 16768, OPS: 3493.4, Avg(us): 30, Min(us): 14, Max(us): 1971, 50th(us): 26, 90th(us): 37, 95th(us): 42, 99th(us): 127, 99.9th(us): 783, 99.99th(us): 1692
TOTAL  - Takes(s): 4.8, Count: 157473, OPS: 32804.2, Avg(us): 9996, Min(us): 0, Max(us): 116863, 50th(us): 4035, 90th(us): 33823, 95th(us): 41823, 99th(us): 56959, 99.9th(us): 75455, 99.99th(us): 91903
TXN    - Takes(s): 4.8, Count: 12973, OPS: 2717.6, Avg(us): 36287, Min(us): 8688, Max(us): 116863, 50th(us): 34911, 90th(us): 50911, 95th(us): 56767, 99th(us): 70975, 99.9th(us): 86207, 99.99th(us): 102783
TXN_ERROR - Takes(s): 4.8, Count: 3667, OPS: 768.0, Avg(us): 31156, Min(us): 5108, Max(us): 95743, 50th(us): 29823, 90th(us): 45471, 95th(us): 50559, 99th(us): 62591, 99.9th(us): 83199, 99.99th(us): 95743
TxnGroup - Takes(s): 4.8, Count: 16640, OPS: 3466.7, Avg(us): 35077, Min(us): 111, Max(us): 103871, 50th(us): 33727, 90th(us): 51519, 95th(us): 57887, 99th(us): 71423, 99.9th(us): 87743, 99.99th(us): 97919
UPDATE - Takes(s): 4.8, Count: 50329, OPS: 10483.3, Avg(us): 3, Min(us): 1, Max(us): 1690, 50th(us): 3, 90th(us): 5, 95th(us): 5, 99th(us): 14, 99.9th(us): 220, 99.99th(us): 1120
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2623
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  499
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  382
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  163

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1159
rollForward failed
  version mismatch  525
rollback failed
  version mismatch  133
     key not found   64
```



##### Native

+ 8

```bash
----------------------------------
Run finished, takes 43.45410252s
READ   - Takes(s): 43.5, Count: 50148, OPS: 1154.1, Avg(us): 3455, Min(us): 3190, Max(us): 4827, 50th(us): 3447, 90th(us): 3547, 95th(us): 3599, 99th(us): 3775, 99.9th(us): 3955, 99.99th(us): 4515
TOTAL  - Takes(s): 43.5, Count: 116664, OPS: 2685.0, Avg(us): 5949, Min(us): 3190, Max(us): 23375, 50th(us): 3471, 90th(us): 20751, 95th(us): 20879, 99th(us): 21087, 99.9th(us): 22815, 99.99th(us): 23055
TxnGroup - Takes(s): 43.4, Count: 16664, OPS: 383.7, Avg(us): 20857, Min(us): 20368, Max(us): 23375, 50th(us): 20831, 90th(us): 21055, 95th(us): 21135, 99th(us): 22703, 99.9th(us): 23039, 99.99th(us): 23135
UPDATE - Takes(s): 43.5, Count: 49852, OPS: 1147.3, Avg(us): 3475, Min(us): 3208, Max(us): 4823, 50th(us): 3465, 90th(us): 3571, 95th(us): 3623, 99th(us): 3801, 99.9th(us): 3969, 99.99th(us): 4295
Error Summary:
```

+ 16

```bash
----------------------------------
Run finished, takes 22.694412657s
READ   - Takes(s): 22.7, Count: 50008, OPS: 2203.8, Avg(us): 3558, Min(us): 3180, Max(us): 5375, 50th(us): 3537, 90th(us): 3761, 95th(us): 3849, 99th(us): 4005, 99.9th(us): 4271, 99.99th(us): 4563
TOTAL  - Takes(s): 22.7, Count: 116656, OPS: 5141.1, Avg(us): 6190, Min(us): 3180, Max(us): 97151, 50th(us): 3577, 90th(us): 21311, 95th(us): 21503, 99th(us): 22847, 99.9th(us): 30959, 99.99th(us): 86335
TxnGroup - Takes(s): 22.7, Count: 16656, OPS: 734.7, Avg(us): 21722, Min(us): 19632, Max(us): 97151, 50th(us): 21407, 90th(us): 21871, 95th(us): 23071, 99th(us): 23551, 99.9th(us): 83455, 99.99th(us): 93631
UPDATE - Takes(s): 22.7, Count: 49992, OPS: 2203.2, Avg(us): 3648, Min(us): 3204, Max(us): 78719, 50th(us): 3565, 90th(us): 3793, 95th(us): 3879, 99th(us): 4051, 99.9th(us): 32447, 99.99th(us): 74687
Error Summary:
```

+ 32

```bash
----------------------------------
Run finished, takes 11.589312301s
READ   - Takes(s): 11.6, Count: 49851, OPS: 4302.8, Avg(us): 3648, Min(us): 3180, Max(us): 6599, 50th(us): 3637, 90th(us): 3953, 95th(us): 4053, 99th(us): 4279, 99.9th(us): 4691, 99.99th(us): 5823
TOTAL  - Takes(s): 11.6, Count: 116640, OPS: 10067.9, Avg(us): 6309, Min(us): 3180, Max(us): 27247, 50th(us): 3691, 90th(us): 21983, 95th(us): 22287, 99th(us): 23231, 99.9th(us): 24351, 99.99th(us): 24719
TxnGroup - Takes(s): 11.6, Count: 16640, OPS: 1438.7, Avg(us): 22236, Min(us): 19696, Max(us): 27247, 50th(us): 22143, 90th(us): 22799, 95th(us): 23535, 99th(us): 24287, 99.9th(us): 24703, 99.99th(us): 25087
UPDATE - Takes(s): 11.6, Count: 50149, OPS: 4328.5, Avg(us): 3670, Min(us): 3200, Max(us): 7783, 50th(us): 3657, 90th(us): 3981, 95th(us): 4085, 99th(us): 4319, 99.9th(us): 4779, 99.99th(us): 5707
Error Summary:
```

+ 64

```bash
----------------------------------
Run finished, takes 5.856843626s
READ   - Takes(s): 5.9, Count: 50069, OPS: 8554.6, Avg(us): 3649, Min(us): 3166, Max(us): 12031, 50th(us): 3537, 90th(us): 4179, 95th(us): 4335, 99th(us): 4735, 99.9th(us): 6243, 99.99th(us): 8879
TOTAL  - Takes(s): 5.9, Count: 116640, OPS: 19927.6, Avg(us): 6340, Min(us): 3166, Max(us): 30511, 50th(us): 3675, 90th(us): 21759, 95th(us): 23087, 99th(us): 23887, 99.9th(us): 25007, 99.99th(us): 27503
TxnGroup - Takes(s): 5.8, Count: 16640, OPS: 2851.0, Avg(us): 22438, Min(us): 19520, Max(us): 30511, 50th(us): 22655, 90th(us): 23743, 95th(us): 24031, 99th(us): 24815, 99.9th(us): 26511, 99.99th(us): 29359
UPDATE - Takes(s): 5.9, Count: 49931, OPS: 8530.7, Avg(us): 3673, Min(us): 3190, Max(us): 10735, 50th(us): 3563, 90th(us): 4199, 95th(us): 4355, 99th(us): 4791, 99.9th(us): 5795, 99.99th(us): 8639
Error Summary:
```

+ 96

```bash
----------------------------------
Run finished, takes 3.844046949s
READ   - Takes(s): 3.8, Count: 50400, OPS: 13124.8, Avg(us): 3555, Min(us): 3166, Max(us): 12367, 50th(us): 3409, 90th(us): 4067, 95th(us): 4307, 99th(us): 4919, 99.9th(us): 8163, 99.99th(us): 9943
TOTAL  - Takes(s): 3.8, Count: 116608, OPS: 30368.2, Avg(us): 6192, Min(us): 3166, Max(us): 66815, 50th(us): 3485, 90th(us): 21279, 95th(us): 22127, 99th(us): 23775, 99.9th(us): 27519, 99.99th(us): 31087
TxnGroup - Takes(s): 3.8, Count: 16608, OPS: 4345.9, Avg(us): 21949, Min(us): 19536, Max(us): 66815, 50th(us): 21711, 90th(us): 23455, 95th(us): 24015, 99th(us): 26831, 99.9th(us): 30463, 99.99th(us): 48031
UPDATE - Takes(s): 3.8, Count: 49600, OPS: 12919.9, Avg(us): 3596, Min(us): 3182, Max(us): 48927, 50th(us): 3439, 90th(us): 4115, 95th(us): 4351, 99th(us): 5111, 99.9th(us): 8599, 99.99th(us): 19135
Error Summary:
```

+ 128

```bash
----------------------------------
Run finished, takes 2.896079163s
READ   - Takes(s): 2.9, Count: 50393, OPS: 17424.8, Avg(us): 3561, Min(us): 3166, Max(us): 14903, 50th(us): 3375, 90th(us): 4103, 95th(us): 4375, 99th(us): 5223, 99.9th(us): 10215, 99.99th(us): 12519
TOTAL  - Takes(s): 2.9, Count: 116640, OPS: 40326.8, Avg(us): 6214, Min(us): 3166, Max(us): 35359, 50th(us): 3451, 90th(us): 21135, 95th(us): 22271, 99th(us): 24255, 99.9th(us): 29711, 99.99th(us): 33183
TxnGroup - Takes(s): 2.9, Count: 16640, OPS: 5793.9, Avg(us): 22038, Min(us): 19440, Max(us): 35359, 50th(us): 21775, 90th(us): 23775, 95th(us): 24687, 99th(us): 29103, 99.9th(us): 32895, 99.99th(us): 33919
UPDATE - Takes(s): 2.9, Count: 49607, OPS: 17156.9, Avg(us): 3600, Min(us): 3188, Max(us): 15823, 50th(us): 3409, 90th(us): 4163, 95th(us): 4435, 99th(us): 5311, 99.9th(us): 10495, 99.99th(us): 11999
Error Summary:
```





#### redis - mongo

##### Cherry Garcia

+ 8

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 8
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 1m24.185680851s
COMMIT - Takes(s): 84.2, Count: 16148, OPS: 191.9, Avg(us): 29207, Min(us): 0, Max(us): 71423, 50th(us): 28719, 90th(us): 42623, 95th(us): 43391, 99th(us): 50239, 99.9th(us): 59039, 99.99th(us): 63519
COMMIT_ERROR - Takes(s): 84.1, Count: 516, OPS: 6.1, Avg(us): 21597, Min(us): 6980, Max(us): 50591, 50th(us): 21439, 90th(us): 32751, 95th(us): 35743, 99th(us): 42623, 99.9th(us): 49375, 99.99th(us): 50591
READ   - Takes(s): 84.2, Count: 49671, OPS: 590.0, Avg(us): 3676, Min(us): 4, Max(us): 15975, 50th(us): 3535, 90th(us): 3873, 95th(us): 3989, 99th(us): 10599, 99.9th(us): 11815, 99.99th(us): 15343
READ_ERROR - Takes(s): 83.9, Count: 393, OPS: 4.7, Avg(us): 8253, Min(us): 6380, Max(us): 12127, 50th(us): 7219, 90th(us): 10991, 95th(us): 11263, 99th(us): 11807, 99.9th(us): 12127, 99.99th(us): 12127
Start  - Takes(s): 84.2, Count: 16672, OPS: 198.0, Avg(us): 22, Min(us): 13, Max(us): 415, 50th(us): 18, 90th(us): 28, 95th(us): 35, 99th(us): 133, 99.9th(us): 273, 99.99th(us): 369
TOTAL  - Takes(s): 84.2, Count: 165239, OPS: 1962.8, Avg(us): 11974, Min(us): 0, Max(us): 82751, 50th(us): 3449, 90th(us): 42079, 95th(us): 46111, 99th(us): 54175, 99.9th(us): 65791, 99.99th(us): 76351
TXN    - Takes(s): 84.2, Count: 16148, OPS: 191.9, Avg(us): 40513, Min(us): 17776, Max(us): 74751, 50th(us): 39487, 90th(us): 46719, 95th(us): 49919, 99th(us): 54239, 99.9th(us): 61151, 99.99th(us): 67327
TXN_ERROR - Takes(s): 84.1, Count: 516, OPS: 6.1, Avg(us): 30545, Min(us): 10584, Max(us): 50783, 50th(us): 30783, 90th(us): 39007, 95th(us): 41727, 99th(us): 46271, 99.9th(us): 50687, 99.99th(us): 50783
TxnGroup - Takes(s): 84.2, Count: 16664, OPS: 198.0, Avg(us): 40190, Min(us): 77, Max(us): 82751, 50th(us): 39359, 90th(us): 53663, 95th(us): 57439, 99th(us): 65535, 99.9th(us): 76351, 99.99th(us): 78719
UPDATE - Takes(s): 84.2, Count: 49936, OPS: 593.2, Avg(us): 2, Min(us): 1, Max(us): 345, 50th(us): 2, 90th(us): 4, 95th(us): 5, 99th(us): 11, 99.9th(us): 118, 99.99th(us): 214
Error Summary:

                                                   Operation:  COMMIT
                                                        Error   Count
                                                        -----   -----
  prepare phase failed: read failed due to unknown txn status     266
                       prepare phase failed: version mismatch     137
prepare phase failed: rollForward failed
  version mismatch  106
prepare phase failed: rollback failed
  version mismatch  7

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    275
rollForward failed
  version mismatch  116
rollback failed
  version mismatch  2
```

+ 16

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 16
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 42.467656033s
COMMIT - Takes(s): 42.4, Count: 15595, OPS: 367.5, Avg(us): 29549, Min(us): 0, Max(us): 67327, 50th(us): 29391, 90th(us): 43583, 95th(us): 44479, 99th(us): 51487, 99.9th(us): 58879, 99.99th(us): 63455
COMMIT_ERROR - Takes(s): 42.4, Count: 1061, OPS: 25.0, Avg(us): 21428, Min(us): 3640, Max(us): 50367, 50th(us): 21855, 90th(us): 32063, 95th(us): 36607, 99th(us): 44223, 99.9th(us): 47679, 99.99th(us): 50367
READ   - Takes(s): 42.5, Count: 49169, OPS: 1157.9, Avg(us): 3721, Min(us): 4, Max(us): 29839, 50th(us): 3585, 90th(us): 3951, 95th(us): 4085, 99th(us): 10831, 99.9th(us): 11975, 99.99th(us): 15079
READ_ERROR - Takes(s): 42.4, Count: 757, OPS: 17.9, Avg(us): 8457, Min(us): 6228, Max(us): 33951, 50th(us): 7343, 90th(us): 11151, 95th(us): 11359, 99th(us): 12111, 99.9th(us): 12791, 99.99th(us): 33951
Start  - Takes(s): 42.5, Count: 16672, OPS: 392.6, Avg(us): 28, Min(us): 13, Max(us): 865, 50th(us): 24, 90th(us): 37, 95th(us): 46, 99th(us): 196, 99.9th(us): 336, 99.99th(us): 745
TOTAL  - Takes(s): 42.5, Count: 163761, OPS: 3856.1, Avg(us): 11971, Min(us): 0, Max(us): 82175, 50th(us): 3465, 90th(us): 41087, 95th(us): 46847, 99th(us): 55103, 99.9th(us): 66367, 99.99th(us): 73599
TXN    - Takes(s): 42.4, Count: 15595, OPS: 367.5, Avg(us): 41159, Min(us): 17632, Max(us): 74047, 50th(us): 40479, 90th(us): 47839, 95th(us): 50815, 99th(us): 55007, 99.9th(us): 63039, 99.99th(us): 71295
TXN_ERROR - Takes(s): 42.4, Count: 1061, OPS: 25.0, Avg(us): 30901, Min(us): 14280, Max(us): 59551, 50th(us): 29999, 90th(us): 39711, 95th(us): 41503, 99th(us): 47583, 99.9th(us): 52159, 99.99th(us): 59551
TxnGroup - Takes(s): 42.5, Count: 16656, OPS: 392.2, Avg(us): 40476, Min(us): 74, Max(us): 82175, 50th(us): 40223, 90th(us): 54719, 95th(us): 58655, 99th(us): 66047, 99.9th(us): 73535, 99.99th(us): 81023
UPDATE - Takes(s): 42.5, Count: 50074, OPS: 1179.1, Avg(us): 3, Min(us): 1, Max(us): 709, 50th(us): 3, 90th(us): 5, 95th(us): 6, 99th(us): 16, 99.9th(us): 202, 99.99th(us): 340
Error Summary:

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    507
rollForward failed
  version mismatch  239
rollback failed
  version mismatch  11

                                                   Operation:  COMMIT
                                                        Error   Count
                                                        -----   -----
  prepare phase failed: read failed due to unknown txn status     523
                       prepare phase failed: version mismatch     267
prepare phase failed: rollForward failed
  version mismatch  253
prepare phase failed: rollback failed
  version mismatch  18
```

+ 32

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 32
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 21.072829407s
COMMIT - Takes(s): 21.0, Count: 14883, OPS: 707.4, Avg(us): 29260, Min(us): 0, Max(us): 73279, 50th(us): 29503, 90th(us): 41887, 95th(us): 45375, 99th(us): 52447, 99.9th(us): 60447, 99.99th(us): 72767
COMMIT_ERROR - Takes(s): 21.1, Count: 1757, OPS: 83.5, Avg(us): 21396, Min(us): 4268, Max(us): 51775, 50th(us): 21183, 90th(us): 32015, 95th(us): 37119, 99th(us): 43263, 99.9th(us): 49695, 99.99th(us): 51775
READ   - Takes(s): 21.1, Count: 48464, OPS: 2299.9, Avg(us): 3692, Min(us): 4, Max(us): 31103, 50th(us): 3437, 90th(us): 4123, 95th(us): 4307, 99th(us): 10887, 99.9th(us): 12223, 99.99th(us): 26959
READ_ERROR - Takes(s): 21.0, Count: 1308, OPS: 62.2, Avg(us): 8480, Min(us): 6200, Max(us): 35487, 50th(us): 7675, 90th(us): 11263, 95th(us): 11607, 99th(us): 12119, 99.9th(us): 12567, 99.99th(us): 35487
Start  - Takes(s): 21.1, Count: 16672, OPS: 791.1, Avg(us): 33, Min(us): 13, Max(us): 721, 50th(us): 26, 90th(us): 42, 95th(us): 54, 99th(us): 237, 99.9th(us): 461, 99.99th(us): 705
TOTAL  - Takes(s): 21.1, Count: 161770, OPS: 7676.6, Avg(us): 11684, Min(us): 0, Max(us): 87807, 50th(us): 3309, 90th(us): 41311, 95th(us): 45951, 99th(us): 55519, 99.9th(us): 67071, 99.99th(us): 76351
TXN    - Takes(s): 21.0, Count: 14883, OPS: 707.4, Avg(us): 41033, Min(us): 17824, Max(us): 80319, 50th(us): 41055, 90th(us): 48479, 95th(us): 50495, 99th(us): 56415, 99.9th(us): 68351, 99.99th(us): 79551
TXN_ERROR - Takes(s): 21.1, Count: 1757, OPS: 83.5, Avg(us): 30987, Min(us): 11160, Max(us): 63167, 50th(us): 30687, 90th(us): 39551, 95th(us): 42047, 99th(us): 46527, 99.9th(us): 53439, 99.99th(us): 63167
TxnGroup - Takes(s): 21.1, Count: 16640, OPS: 789.6, Avg(us): 39916, Min(us): 209, Max(us): 87807, 50th(us): 39839, 90th(us): 54239, 95th(us): 58559, 99th(us): 66623, 99.9th(us): 74943, 99.99th(us): 83775
UPDATE - Takes(s): 21.1, Count: 50228, OPS: 2383.4, Avg(us): 4, Min(us): 1, Max(us): 605, 50th(us): 3, 90th(us): 5, 95th(us): 7, 99th(us): 22, 99.9th(us): 276, 99.99th(us): 507
Error Summary:

                                                   Operation:  COMMIT
                                                        Error   Count
                                                        -----   -----
  prepare phase failed: read failed due to unknown txn status     831
                       prepare phase failed: version mismatch     451
prepare phase failed: rollForward failed
  version mismatch  435
prepare phase failed: rollback failed
  version mismatch  40

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    829
rollForward failed
  version mismatch  434
rollback failed
  version mismatch  45
```

+ 64

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 10.29744593s
COMMIT - Takes(s): 10.3, Count: 14042, OPS: 1366.7, Avg(us): 28247, Min(us): 0, Max(us): 59935, 50th(us): 28511, 90th(us): 41407, 95th(us): 43327, 99th(us): 50239, 99.9th(us): 56735, 99.99th(us): 59103
COMMIT_ERROR - Takes(s): 10.3, Count: 2598, OPS: 253.0, Avg(us): 20769, Min(us): 3590, Max(us): 50495, 50th(us): 21071, 90th(us): 31359, 95th(us): 35327, 99th(us): 40255, 99.9th(us): 45791, 99.99th(us): 50495
READ   - Takes(s): 10.3, Count: 47834, OPS: 4646.6, Avg(us): 3602, Min(us): 5, Max(us): 15175, 50th(us): 3381, 90th(us): 3919, 95th(us): 4199, 99th(us): 10463, 99.9th(us): 11727, 99.99th(us): 14679
READ_ERROR - Takes(s): 10.3, Count: 2062, OPS: 201.1, Avg(us): 8462, Min(us): 6224, Max(us): 12767, 50th(us): 7471, 90th(us): 10887, 95th(us): 11215, 99th(us): 11711, 99.9th(us): 12479, 99.99th(us): 12767
Start  - Takes(s): 10.3, Count: 16704, OPS: 1622.0, Avg(us): 39, Min(us): 14, Max(us): 1595, 50th(us): 28, 90th(us): 44, 95th(us): 56, 99th(us): 355, 99.9th(us): 777, 99.99th(us): 1196
TOTAL  - Takes(s): 10.3, Count: 159366, OPS: 15475.5, Avg(us): 11132, Min(us): 0, Max(us): 85503, 50th(us): 3301, 90th(us): 39935, 95th(us): 44415, 99th(us): 53855, 99.9th(us): 64831, 99.99th(us): 73663
TXN    - Takes(s): 10.3, Count: 14042, OPS: 1366.5, Avg(us): 40098, Min(us): 17376, Max(us): 70207, 50th(us): 39871, 90th(us): 47071, 95th(us): 49535, 99th(us): 54047, 99.9th(us): 59903, 99.99th(us): 64511
TXN_ERROR - Takes(s): 10.3, Count: 2598, OPS: 252.9, Avg(us): 30751, Min(us): 10640, Max(us): 57439, 50th(us): 30783, 90th(us): 39039, 95th(us): 41503, 99th(us): 46783, 99.9th(us): 50879, 99.99th(us): 57439
TxnGroup - Takes(s): 10.3, Count: 16640, OPS: 1616.0, Avg(us): 38530, Min(us): 69, Max(us): 85503, 50th(us): 38687, 90th(us): 52895, 95th(us): 56831, 99th(us): 64671, 99.9th(us): 73599, 99.99th(us): 78591
UPDATE - Takes(s): 10.3, Count: 50104, OPS: 4865.1, Avg(us): 5, Min(us): 1, Max(us): 899, 50th(us): 3, 90th(us): 5, 95th(us): 7, 99th(us): 20, 99.9th(us): 408, 99.99th(us): 751
Error Summary:

                                                   Operation:  COMMIT
                                                        Error   Count
                                                        -----   -----
  prepare phase failed: read failed due to unknown txn status    1181
prepare phase failed: rollForward failed
                        version mismatch  734
  prepare phase failed: version mismatch  588
prepare phase failed: rollback failed
  version mismatch  95

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1186
rollForward failed
  version mismatch  784
rollback failed
  version mismatch  92
```

+ 96

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 96
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 6.771577099s
COMMIT - Takes(s): 6.7, Count: 13338, OPS: 1977.7, Avg(us): 27768, Min(us): 0, Max(us): 66239, 50th(us): 27887, 90th(us): 41023, 95th(us): 42783, 99th(us): 49631, 99.9th(us): 55999, 99.99th(us): 65247
COMMIT_ERROR - Takes(s): 6.8, Count: 3270, OPS: 484.4, Avg(us): 20581, Min(us): 3492, Max(us): 68799, 50th(us): 20767, 90th(us): 30911, 95th(us): 34655, 99th(us): 40767, 99.9th(us): 46303, 99.99th(us): 68799
READ   - Takes(s): 6.8, Count: 47340, OPS: 6993.6, Avg(us): 3589, Min(us): 5, Max(us): 24255, 50th(us): 3359, 90th(us): 3907, 95th(us): 4231, 99th(us): 10319, 99.9th(us): 13479, 99.99th(us): 20047
READ_ERROR - Takes(s): 6.8, Count: 2583, OPS: 382.6, Avg(us): 8404, Min(us): 6200, Max(us): 27471, 50th(us): 7427, 90th(us): 10631, 95th(us): 11135, 99th(us): 12111, 99.9th(us): 21375, 99.99th(us): 27471
Start  - Takes(s): 6.8, Count: 16704, OPS: 2466.5, Avg(us): 39, Min(us): 14, Max(us): 1824, 50th(us): 28, 90th(us): 44, 95th(us): 54, 99th(us): 387, 99.9th(us): 1002, 99.99th(us): 1372
TOTAL  - Takes(s): 6.8, Count: 157405, OPS: 23238.0, Avg(us): 10816, Min(us): 0, Max(us): 86015, 50th(us): 3279, 90th(us): 39295, 95th(us): 44415, 99th(us): 53471, 99.9th(us): 64415, 99.99th(us): 72447
TXN    - Takes(s): 6.7, Count: 13338, OPS: 1977.6, Avg(us): 39883, Min(us): 16688, Max(us): 73855, 50th(us): 39615, 90th(us): 46911, 95th(us): 49183, 99th(us): 54399, 99.9th(us): 61567, 99.99th(us): 70975
TXN_ERROR - Takes(s): 6.8, Count: 3270, OPS: 484.3, Avg(us): 30536, Min(us): 9992, Max(us): 68927, 50th(us): 30783, 90th(us): 38655, 95th(us): 41535, 99th(us): 45983, 99.9th(us): 51167, 99.99th(us): 68927
TxnGroup - Takes(s): 6.8, Count: 16608, OPS: 2453.5, Avg(us): 37892, Min(us): 92, Max(us): 86015, 50th(us): 37919, 90th(us): 52287, 95th(us): 56223, 99th(us): 63871, 99.9th(us): 71999, 99.99th(us): 80063
UPDATE - Takes(s): 6.8, Count: 50077, OPS: 7395.2, Avg(us): 5, Min(us): 1, Max(us): 1307, 50th(us): 3, 90th(us): 5, 95th(us): 7, 99th(us): 19, 99.9th(us): 485, 99.99th(us): 1035
Error Summary:

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1463
rollForward failed
  version mismatch  971
rollback failed
  version mismatch  149

                                                   Operation:  COMMIT
                                                        Error   Count
                                                        -----   -----
  prepare phase failed: read failed due to unknown txn status    1473
prepare phase failed: rollForward failed
                        version mismatch  933
  prepare phase failed: version mismatch  740
prepare phase failed: rollback failed
  version mismatch  124
```

+ 128

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 5.128855925s
COMMIT - Takes(s): 5.1, Count: 12973, OPS: 2548.8, Avg(us): 27384, Min(us): 0, Max(us): 74943, 50th(us): 27599, 90th(us): 40639, 95th(us): 42463, 99th(us): 49183, 99.9th(us): 56127, 99.99th(us): 71423
COMMIT_ERROR - Takes(s): 5.1, Count: 3667, OPS: 720.0, Avg(us): 20533, Min(us): 4078, Max(us): 57919, 50th(us): 20511, 90th(us): 30927, 95th(us): 34975, 99th(us): 41823, 99.9th(us): 49311, 99.99th(us): 57919
READ   - Takes(s): 5.1, Count: 46946, OPS: 9157.8, Avg(us): 3569, Min(us): 5, Max(us): 22415, 50th(us): 3349, 90th(us): 3867, 95th(us): 4231, 99th(us): 10151, 99.9th(us): 12479, 99.99th(us): 16007
READ_ERROR - Takes(s): 5.1, Count: 3078, OPS: 601.8, Avg(us): 8189, Min(us): 6212, Max(us): 13567, 50th(us): 7175, 90th(us): 10471, 95th(us): 10871, 99th(us): 11727, 99.9th(us): 12743, 99.99th(us): 13567
Start  - Takes(s): 5.1, Count: 16768, OPS: 3268.4, Avg(us): 39, Min(us): 14, Max(us): 1338, 50th(us): 28, 90th(us): 44, 95th(us): 53, 99th(us): 427, 99.9th(us): 942, 99.99th(us): 1176
TOTAL  - Takes(s): 5.1, Count: 156276, OPS: 30466.3, Avg(us): 10624, Min(us): 0, Max(us): 92031, 50th(us): 3271, 90th(us): 38847, 95th(us): 44063, 99th(us): 53087, 99.9th(us): 64959, 99.99th(us): 74175
TXN    - Takes(s): 5.1, Count: 12973, OPS: 2549.1, Avg(us): 39625, Min(us): 19408, Max(us): 76671, 50th(us): 39391, 90th(us): 46815, 95th(us): 49087, 99th(us): 53919, 99.9th(us): 60927, 99.99th(us): 76095
TXN_ERROR - Takes(s): 5.1, Count: 3667, OPS: 719.9, Avg(us): 30533, Min(us): 9968, Max(us): 58271, 50th(us): 30623, 90th(us): 38815, 95th(us): 41631, 99th(us): 47775, 99.9th(us): 56191, 99.99th(us): 58271
TxnGroup - Takes(s): 5.1, Count: 16640, OPS: 3244.7, Avg(us): 37414, Min(us): 125, Max(us): 92031, 50th(us): 37439, 90th(us): 51839, 95th(us): 55903, 99th(us): 64287, 99.9th(us): 73663, 99.99th(us): 83391
UPDATE - Takes(s): 5.1, Count: 49976, OPS: 9742.9, Avg(us): 5, Min(us): 1, Max(us): 1730, 50th(us): 3, 90th(us): 5, 95th(us): 7, 99th(us): 19, 99.9th(us): 579, 99.99th(us): 1379
Error Summary:

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1840
rollForward failed
  version mismatch  1047
rollback failed
  version mismatch  191

                                                   Operation:  COMMIT
                                                        Error   Count
                                                        -----   -----
  prepare phase failed: read failed due to unknown txn status    1661
prepare phase failed: rollForward failed
                        version mismatch  991
  prepare phase failed: version mismatch  807
prepare phase failed: rollback failed
  version mismatch  208
```

##### Oreo

+ 8

```bash
----------------------------------
Run finished, takes 40.910776495s
COMMIT - Takes(s): 40.9, Count: 16167, OPS: 395.3, Avg(us): 7660, Min(us): 0, Max(us): 10223, 50th(us): 7771, 90th(us): 8247, 95th(us): 8423, 99th(us): 9039, 99.9th(us): 9655, 99.99th(us): 10087
COMMIT_ERROR - Takes(s): 40.8, Count: 497, OPS: 12.2, Avg(us): 4551, Min(us): 3558, Max(us): 6095, 50th(us): 4543, 90th(us): 5095, 95th(us): 5195, 99th(us): 5515, 99.9th(us): 6095, 99.99th(us): 6095
READ   - Takes(s): 40.9, Count: 49874, OPS: 1219.2, Avg(us): 3921, Min(us): 5, Max(us): 5491, 50th(us): 3915, 90th(us): 4419, 95th(us): 4535, 99th(us): 4707, 99.9th(us): 5019, 99.99th(us): 5299
READ_ERROR - Takes(s): 40.5, Count: 212, OPS: 5.2, Avg(us): 4194, Min(us): 3320, Max(us): 5059, 50th(us): 4179, 90th(us): 4799, 95th(us): 4907, 99th(us): 5015, 99.9th(us): 5059, 99.99th(us): 5059
Start  - Takes(s): 40.9, Count: 16672, OPS: 407.5, Avg(us): 21, Min(us): 13, Max(us): 346, 50th(us): 19, 90th(us): 28, 95th(us): 31, 99th(us): 42, 99.9th(us): 212, 99.99th(us): 343
TOTAL  - Takes(s): 40.9, Count: 165458, OPS: 4044.2, Avg(us): 5798, Min(us): 0, Max(us): 33183, 50th(us): 3697, 90th(us): 19663, 95th(us): 23391, 99th(us): 27599, 99.9th(us): 31087, 99.99th(us): 32703
TXN    - Takes(s): 40.9, Count: 16167, OPS: 395.3, Avg(us): 19525, Min(us): 7248, Max(us): 29647, 50th(us): 19711, 90th(us): 24959, 95th(us): 27423, 99th(us): 28303, 99.9th(us): 28847, 99.99th(us): 29327
TXN_ERROR - Takes(s): 40.8, Count: 497, OPS: 12.2, Avg(us): 16406, Min(us): 4716, Max(us): 25775, 50th(us): 16639, 90th(us): 21087, 95th(us): 23599, 99th(us): 24943, 99.9th(us): 25775, 99.99th(us): 25775
TxnGroup - Takes(s): 40.9, Count: 16664, OPS: 407.4, Avg(us): 19429, Min(us): 77, Max(us): 33183, 50th(us): 19631, 90th(us): 26335, 95th(us): 27775, 99th(us): 31071, 99.9th(us): 32703, 99.99th(us): 33119
UPDATE - Takes(s): 40.9, Count: 49914, OPS: 1220.1, Avg(us): 2, Min(us): 1, Max(us): 355, 50th(us): 2, 90th(us): 4, 95th(us): 4, 99th(us): 9, 99.9th(us): 53, 99.99th(us): 209
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  442
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  46
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  6
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  3

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  159
  read failed due to unknown txn status   51
rollback failed
  version mismatch  2
```

+ 16

```bash
----------------------------------
Run finished, takes 20.062583779s
COMMIT - Takes(s): 20.0, Count: 15732, OPS: 784.7, Avg(us): 7637, Min(us): 0, Max(us): 10447, 50th(us): 7707, 90th(us): 8367, 95th(us): 8567, 99th(us): 8991, 99.9th(us): 9927, 99.99th(us): 10391
COMMIT_ERROR - Takes(s): 20.0, Count: 924, OPS: 46.1, Avg(us): 4581, Min(us): 3568, Max(us): 6003, 50th(us): 4515, 90th(us): 5187, 95th(us): 5359, 99th(us): 5675, 99.9th(us): 5847, 99.99th(us): 6003
READ   - Takes(s): 20.1, Count: 49684, OPS: 2477.1, Avg(us): 3827, Min(us): 5, Max(us): 6271, 50th(us): 3727, 90th(us): 4387, 95th(us): 4555, 99th(us): 4859, 99.9th(us): 5327, 99.99th(us): 5871
READ_ERROR - Takes(s): 20.0, Count: 329, OPS: 16.5, Avg(us): 4170, Min(us): 3324, Max(us): 5419, 50th(us): 4123, 90th(us): 4803, 95th(us): 5039, 99th(us): 5223, 99.9th(us): 5419, 99.99th(us): 5419
Start  - Takes(s): 20.1, Count: 16672, OPS: 831.0, Avg(us): 24, Min(us): 13, Max(us): 590, 50th(us): 21, 90th(us): 31, 95th(us): 37, 99th(us): 64, 99.9th(us): 282, 99.99th(us): 498
TOTAL  - Takes(s): 20.1, Count: 164463, OPS: 8197.2, Avg(us): 5660, Min(us): 0, Max(us): 33183, 50th(us): 3605, 90th(us): 19231, 95th(us): 22879, 99th(us): 27087, 99.9th(us): 30751, 99.99th(us): 32319
TXN    - Takes(s): 20.0, Count: 15732, OPS: 784.7, Avg(us): 19251, Min(us): 7304, Max(us): 29775, 50th(us): 19359, 90th(us): 24735, 95th(us): 26831, 99th(us): 27967, 99.9th(us): 28815, 99.99th(us): 29375
TXN_ERROR - Takes(s): 20.0, Count: 924, OPS: 46.1, Avg(us): 15676, Min(us): 4116, Max(us): 25343, 50th(us): 15999, 90th(us): 20783, 95th(us): 22831, 99th(us): 24479, 99.9th(us): 25263, 99.99th(us): 25343
TxnGroup - Takes(s): 20.1, Count: 16656, OPS: 830.4, Avg(us): 19046, Min(us): 76, Max(us): 33183, 50th(us): 19215, 90th(us): 25759, 95th(us): 27359, 99th(us): 30671, 99.9th(us): 32303, 99.99th(us): 32895
UPDATE - Takes(s): 20.1, Count: 49987, OPS: 2491.6, Avg(us): 3, Min(us): 1, Max(us): 551, 50th(us): 2, 90th(us): 4, 95th(us): 5, 99th(us): 11, 99.9th(us): 149, 99.99th(us): 298
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  784
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  113
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  17
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  10

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  209
  read failed due to unknown txn status  111
rollback failed
  version mismatch  8
     key not found  1
```

+ 32

```bash
----------------------------------
Run finished, takes 9.949075365s
COMMIT - Takes(s): 9.9, Count: 14938, OPS: 1502.7, Avg(us): 7759, Min(us): 0, Max(us): 11095, 50th(us): 7763, 90th(us): 8631, 95th(us): 8919, 99th(us): 9431, 99.9th(us): 10047, 99.99th(us): 10807
COMMIT_ERROR - Takes(s): 9.9, Count: 1702, OPS: 171.3, Avg(us): 4696, Min(us): 3472, Max(us): 7099, 50th(us): 4587, 90th(us): 5487, 95th(us): 5787, 99th(us): 6331, 99.9th(us): 6695, 99.99th(us): 7099
READ   - Takes(s): 9.9, Count: 49394, OPS: 4966.4, Avg(us): 3759, Min(us): 5, Max(us): 10687, 50th(us): 3671, 90th(us): 4255, 95th(us): 4483, 99th(us): 4935, 99.9th(us): 5691, 99.99th(us): 6251
READ_ERROR - Takes(s): 9.9, Count: 469, OPS: 47.4, Avg(us): 4173, Min(us): 3274, Max(us): 6087, 50th(us): 4095, 90th(us): 5007, 95th(us): 5295, 99th(us): 5843, 99.9th(us): 6087, 99.99th(us): 6087
Start  - Takes(s): 9.9, Count: 16672, OPS: 1675.7, Avg(us): 26, Min(us): 13, Max(us): 784, 50th(us): 23, 90th(us): 34, 95th(us): 39, 99th(us): 88, 99.9th(us): 392, 99.99th(us): 749
TOTAL  - Takes(s): 10.0, Count: 162719, OPS: 16353.5, Avg(us): 5538, Min(us): 0, Max(us): 33919, 50th(us): 3565, 90th(us): 18991, 95th(us): 22447, 99th(us): 26639, 99.9th(us): 29935, 99.99th(us): 32143
TXN    - Takes(s): 9.9, Count: 14938, OPS: 1502.7, Avg(us): 19166, Min(us): 7336, Max(us): 31679, 50th(us): 19167, 90th(us): 24975, 95th(us): 26463, 99th(us): 28143, 99.9th(us): 29679, 99.99th(us): 31279
TXN_ERROR - Takes(s): 9.9, Count: 1702, OPS: 171.3, Avg(us): 15649, Min(us): 4196, Max(us): 26207, 50th(us): 15727, 90th(us): 20751, 95th(us): 22143, 99th(us): 24495, 99.9th(us): 25919, 99.99th(us): 26207
TxnGroup - Takes(s): 9.9, Count: 16640, OPS: 1672.4, Avg(us): 18793, Min(us): 78, Max(us): 33919, 50th(us): 18991, 90th(us): 24975, 95th(us): 26895, 99th(us): 29791, 99.9th(us): 32031, 99.99th(us): 33087
UPDATE - Takes(s): 9.9, Count: 50137, OPS: 5039.3, Avg(us): 3, Min(us): 1, Max(us): 643, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 11, 99.9th(us): 145, 99.99th(us): 396
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  226
  read failed due to unknown txn status  218
rollback failed
  version mismatch  14
     key not found  11

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  1472
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  189
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  26
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  15
```

+ 64

```bash
----------------------------------
Run finished, takes 5.17511376s
COMMIT - Takes(s): 5.2, Count: 14119, OPS: 2735.3, Avg(us): 8033, Min(us): 0, Max(us): 18031, 50th(us): 8051, 90th(us): 9103, 95th(us): 9511, 99th(us): 10447, 99.9th(us): 11823, 99.99th(us): 14271
COMMIT_ERROR - Takes(s): 5.2, Count: 2521, OPS: 488.2, Avg(us): 5025, Min(us): 3524, Max(us): 9455, 50th(us): 4895, 90th(us): 6031, 95th(us): 6435, 99th(us): 7459, 99.9th(us): 8759, 99.99th(us): 9455
READ   - Takes(s): 5.2, Count: 49660, OPS: 9600.9, Avg(us): 3797, Min(us): 5, Max(us): 17007, 50th(us): 3699, 90th(us): 4287, 95th(us): 4551, 99th(us): 5291, 99.9th(us): 6395, 99.99th(us): 8559
READ_ERROR - Takes(s): 5.2, Count: 635, OPS: 123.2, Avg(us): 4258, Min(us): 3282, Max(us): 7439, 50th(us): 4093, 90th(us): 5203, 95th(us): 5495, 99th(us): 6195, 99.9th(us): 6787, 99.99th(us): 7439
Start  - Takes(s): 5.2, Count: 16704, OPS: 3227.2, Avg(us): 28, Min(us): 13, Max(us): 2041, 50th(us): 25, 90th(us): 36, 95th(us): 41, 99th(us): 174, 99.9th(us): 764, 99.99th(us): 1581
TOTAL  - Takes(s): 5.2, Count: 160947, OPS: 31093.7, Avg(us): 5589, Min(us): 0, Max(us): 37535, 50th(us): 3585, 90th(us): 19327, 95th(us): 22943, 99th(us): 27023, 99.9th(us): 30783, 99.99th(us): 33183
TXN    - Takes(s): 5.2, Count: 14119, OPS: 2735.2, Avg(us): 19718, Min(us): 7420, Max(us): 33087, 50th(us): 19663, 90th(us): 25727, 95th(us): 26831, 99th(us): 28287, 99.9th(us): 30191, 99.99th(us): 31983
TXN_ERROR - Takes(s): 5.2, Count: 2521, OPS: 488.2, Avg(us): 16095, Min(us): 4580, Max(us): 28511, 50th(us): 16135, 90th(us): 21263, 95th(us): 23055, 99th(us): 24767, 99.9th(us): 27503, 99.99th(us): 28511
TxnGroup - Takes(s): 5.2, Count: 16640, OPS: 3217.9, Avg(us): 19142, Min(us): 75, Max(us): 37535, 50th(us): 19327, 90th(us): 25759, 95th(us): 27295, 99th(us): 30639, 99.9th(us): 33183, 99.99th(us): 34975
UPDATE - Takes(s): 5.2, Count: 49705, OPS: 9603.5, Avg(us): 3, Min(us): 1, Max(us): 1064, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 12, 99.9th(us): 175, 99.99th(us): 661
Error Summary:

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    399
rollForward failed
  version mismatch  188
rollback failed
  version mismatch  29
     key not found  19

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2231
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  218
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  43
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  29
```

+ 96

```bash
----------------------------------
Run finished, takes 4.174350064s
COMMIT - Takes(s): 4.2, Count: 13439, OPS: 3227.5, Avg(us): 10336, Min(us): 0, Max(us): 57919, 50th(us): 10199, 90th(us): 12479, 95th(us): 13303, 99th(us): 15455, 99.9th(us): 53311, 99.99th(us): 55359
COMMIT_ERROR - Takes(s): 4.2, Count: 3169, OPS: 762.2, Avg(us): 7213, Min(us): 3664, Max(us): 56095, 50th(us): 6851, 90th(us): 9287, 95th(us): 10135, 99th(us): 12263, 99.9th(us): 50687, 99.99th(us): 56095
READ   - Takes(s): 4.2, Count: 48805, OPS: 11704.8, Avg(us): 4454, Min(us): 5, Max(us): 47263, 50th(us): 4203, 90th(us): 5535, 95th(us): 6087, 99th(us): 7615, 99.9th(us): 10783, 99.99th(us): 46143
READ_ERROR - Takes(s): 4.1, Count: 1075, OPS: 259.5, Avg(us): 5934, Min(us): 3300, Max(us): 43999, 50th(us): 5611, 90th(us): 7999, 95th(us): 8783, 99th(us): 10287, 99.9th(us): 12087, 99.99th(us): 43999
Start  - Takes(s): 4.2, Count: 16704, OPS: 4001.4, Avg(us): 37, Min(us): 13, Max(us): 2677, 50th(us): 27, 90th(us): 39, 95th(us): 45, 99th(us): 465, 99.9th(us): 1187, 99.99th(us): 2551
TOTAL  - Takes(s): 4.2, Count: 159115, OPS: 38105.0, Avg(us): 6711, Min(us): 0, Max(us): 79679, 50th(us): 3817, 90th(us): 23055, 95th(us): 27439, 99th(us): 33215, 99.9th(us): 50111, 99.99th(us): 73471
TXN    - Takes(s): 4.2, Count: 13439, OPS: 3228.1, Avg(us): 24130, Min(us): 7844, Max(us): 77631, 50th(us): 23919, 90th(us): 31071, 95th(us): 32719, 99th(us): 36895, 99.9th(us): 70207, 99.99th(us): 76927
TXN_ERROR - Takes(s): 4.2, Count: 3169, OPS: 762.0, Avg(us): 19940, Min(us): 5200, Max(us): 72255, 50th(us): 19727, 90th(us): 26399, 95th(us): 28255, 99th(us): 32591, 99.9th(us): 65791, 99.99th(us): 72255
TxnGroup - Takes(s): 4.2, Count: 16608, OPS: 3979.0, Avg(us): 23272, Min(us): 78, Max(us): 79679, 50th(us): 23039, 90th(us): 31295, 95th(us): 33663, 99th(us): 39231, 99.9th(us): 70207, 99.99th(us): 78399
UPDATE - Takes(s): 4.2, Count: 50120, OPS: 12005.3, Avg(us): 3, Min(us): 1, Max(us): 1511, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 14, 99.9th(us): 212, 99.99th(us): 924
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2543
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  451
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  97
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  78

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    546
rollForward failed
  version mismatch  414
rollback failed
  version mismatch  76
     key not found  39
```

+ 128

```bash
----------------------------------
Run finished, takes 4.009960334s
COMMIT - Takes(s): 4.0, Count: 13033, OPS: 3260.6, Avg(us): 14007, Min(us): 0, Max(us): 33727, 50th(us): 14175, 90th(us): 17679, 95th(us): 18943, 99th(us): 22319, 99.9th(us): 26127, 99.99th(us): 33471
COMMIT_ERROR - Takes(s): 4.0, Count: 3607, OPS: 902.3, Avg(us): 10319, Min(us): 4084, Max(us): 21663, 50th(us): 10047, 90th(us): 13983, 95th(us): 15223, 99th(us): 17919, 99.9th(us): 20495, 99.99th(us): 21663
READ   - Takes(s): 4.0, Count: 48428, OPS: 12088.3, Avg(us): 5355, Min(us): 5, Max(us): 19455, 50th(us): 5071, 90th(us): 7395, 95th(us): 8167, 99th(us): 10815, 99.9th(us): 14447, 99.99th(us): 17071
READ_ERROR - Takes(s): 4.0, Count: 1681, OPS: 420.3, Avg(us): 8614, Min(us): 3414, Max(us): 18863, 50th(us): 8019, 90th(us): 12407, 95th(us): 13631, 99th(us): 15471, 99.9th(us): 17743, 99.99th(us): 18863
Start  - Takes(s): 4.0, Count: 16768, OPS: 4180.8, Avg(us): 43, Min(us): 13, Max(us): 3101, 50th(us): 27, 90th(us): 40, 95th(us): 46, 99th(us): 591, 99.9th(us): 1880, 99.99th(us): 2851
TOTAL  - Takes(s): 4.0, Count: 157793, OPS: 39348.8, Avg(us): 8508, Min(us): 0, Max(us): 66047, 50th(us): 3987, 90th(us): 29615, 95th(us): 35327, 99th(us): 43263, 99.9th(us): 51167, 99.99th(us): 58399
TXN    - Takes(s): 4.0, Count: 13033, OPS: 3261.0, Avg(us): 31022, Min(us): 9536, Max(us): 66047, 50th(us): 30943, 90th(us): 40319, 95th(us): 42879, 99th(us): 47903, 99.9th(us): 53663, 99.99th(us): 62943
TXN_ERROR - Takes(s): 4.0, Count: 3607, OPS: 902.2, Avg(us): 25569, Min(us): 4384, Max(us): 52799, 50th(us): 25343, 90th(us): 34751, 95th(us): 37599, 99th(us): 42495, 99.9th(us): 49599, 99.99th(us): 52799
TxnGroup - Takes(s): 4.0, Count: 16640, OPS: 4152.6, Avg(us): 29775, Min(us): 69, Max(us): 65855, 50th(us): 29583, 90th(us): 40831, 95th(us): 43935, 99th(us): 50015, 99.9th(us): 57887, 99.99th(us): 65247
UPDATE - Takes(s): 4.0, Count: 49891, OPS: 12439.4, Avg(us): 4, Min(us): 1, Max(us): 2127, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 14, 99.9th(us): 217, 99.99th(us): 1557
Error Summary:

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    781
rollForward failed
  version mismatch  726
rollback failed
  version mismatch  143
     key not found   31

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2589
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  685
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  179
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  154
```

##### Native

+ 8

```bash
----------------------------------
Run finished, takes 43.618110298s
READ   - Takes(s): 43.6, Count: 49988, OPS: 1146.1, Avg(us): 3476, Min(us): 3104, Max(us): 16279, 50th(us): 3439, 90th(us): 3713, 95th(us): 3775, 99th(us): 3875, 99.9th(us): 4027, 99.99th(us): 4343
TOTAL  - Takes(s): 43.6, Count: 116664, OPS: 2674.9, Avg(us): 5974, Min(us): 3104, Max(us): 33471, 50th(us): 3475, 90th(us): 20495, 95th(us): 20687, 99th(us): 22335, 99.9th(us): 22687, 99.99th(us): 23071
TxnGroup - Takes(s): 43.6, Count: 16664, OPS: 382.2, Avg(us): 20935, Min(us): 20064, Max(us): 33471, 50th(us): 20591, 90th(us): 22255, 95th(us): 22383, 99th(us): 22639, 99.9th(us): 22991, 99.99th(us): 33311
UPDATE - Takes(s): 43.6, Count: 50012, OPS: 1146.7, Avg(us): 3485, Min(us): 3110, Max(us): 16087, 50th(us): 3449, 90th(us): 3725, 95th(us): 3787, 99th(us): 3889, 99.9th(us): 4077, 99.99th(us): 4623
Error Summary:
```

+ 16

```bash
----------------------------------
Run finished, takes 22.078317187s
READ   - Takes(s): 22.1, Count: 49811, OPS: 2256.4, Avg(us): 3491, Min(us): 3104, Max(us): 17071, 50th(us): 3477, 90th(us): 3675, 95th(us): 3773, 99th(us): 3931, 99.9th(us): 4127, 99.99th(us): 13463
TOTAL  - Takes(s): 22.1, Count: 116656, OPS: 5284.4, Avg(us): 6038, Min(us): 3102, Max(us): 137343, 50th(us): 3505, 90th(us): 20815, 95th(us): 20975, 99th(us): 22383, 99.9th(us): 23231, 99.99th(us): 130111
TxnGroup - Takes(s): 22.1, Count: 16656, OPS: 755.2, Avg(us): 21177, Min(us): 20128, Max(us): 137343, 50th(us): 20895, 90th(us): 21343, 95th(us): 22591, 99th(us): 23103, 99.9th(us): 49759, 99.99th(us): 137215
UPDATE - Takes(s): 22.1, Count: 50189, OPS: 2273.5, Avg(us): 3543, Min(us): 3102, Max(us): 119807, 50th(us): 3483, 90th(us): 3685, 95th(us): 3785, 99th(us): 3947, 99.9th(us): 7355, 99.99th(us): 115711
Error Summary:
```

+ 32

```bash
----------------------------------
Run finished, takes 11.311646385s
READ   - Takes(s): 11.3, Count: 49999, OPS: 4421.4, Avg(us): 3582, Min(us): 3102, Max(us): 7031, 50th(us): 3577, 90th(us): 3865, 95th(us): 3951, 99th(us): 4123, 99.9th(us): 4407, 99.99th(us): 4911
TOTAL  - Takes(s): 11.3, Count: 116640, OPS: 10314.4, Avg(us): 6168, Min(us): 3094, Max(us): 25055, 50th(us): 3619, 90th(us): 21375, 95th(us): 21679, 99th(us): 23007, 99.9th(us): 23647, 99.99th(us): 23967
TxnGroup - Takes(s): 11.3, Count: 16640, OPS: 1473.9, Avg(us): 21696, Min(us): 19280, Max(us): 25055, 50th(us): 21535, 90th(us): 22751, 95th(us): 23183, 99th(us): 23599, 99.9th(us): 23935, 99.99th(us): 24255
UPDATE - Takes(s): 11.3, Count: 50001, OPS: 4421.7, Avg(us): 3587, Min(us): 3094, Max(us): 5507, 50th(us): 3579, 90th(us): 3873, 95th(us): 3965, 99th(us): 4147, 99.9th(us): 4479, 99.99th(us): 5211
Error Summary:
```

+ 64

```bash
----------------------------------
Run finished, takes 5.581596011s
READ   - Takes(s): 5.6, Count: 49939, OPS: 8951.8, Avg(us): 3499, Min(us): 3098, Max(us): 16927, 50th(us): 3355, 90th(us): 4021, 95th(us): 4151, 99th(us): 4383, 99.9th(us): 5203, 99.99th(us): 11791
TOTAL  - Takes(s): 5.6, Count: 116640, OPS: 20909.3, Avg(us): 6054, Min(us): 3098, Max(us): 34239, 50th(us): 3467, 90th(us): 20239, 95th(us): 22191, 99th(us): 23295, 99.9th(us): 24799, 99.99th(us): 28367
TxnGroup - Takes(s): 5.6, Count: 16640, OPS: 2992.3, Avg(us): 21375, Min(us): 18976, Max(us): 34239, 50th(us): 21215, 90th(us): 23135, 95th(us): 23455, 99th(us): 24543, 99.9th(us): 28239, 99.99th(us): 31391
UPDATE - Takes(s): 5.6, Count: 50061, OPS: 8973.8, Avg(us): 3509, Min(us): 3098, Max(us): 13319, 50th(us): 3373, 90th(us): 4025, 95th(us): 4159, 99th(us): 4399, 99.9th(us): 5103, 99.99th(us): 10887
Error Summary:
```

+ 96

```bash
----------------------------------
Run finished, takes 3.6914955s
READ   - Takes(s): 3.7, Count: 49968, OPS: 13546.5, Avg(us): 3439, Min(us): 3096, Max(us): 15615, 50th(us): 3315, 90th(us): 3877, 95th(us): 4091, 99th(us): 4499, 99.9th(us): 8087, 99.99th(us): 14127
TOTAL  - Takes(s): 3.7, Count: 116608, OPS: 31616.8, Avg(us): 5962, Min(us): 3096, Max(us): 35935, 50th(us): 3369, 90th(us): 20495, 95th(us): 21231, 99th(us): 22687, 99.9th(us): 25759, 99.99th(us): 33919
TxnGroup - Takes(s): 3.7, Count: 16608, OPS: 4523.1, Avg(us): 21123, Min(us): 18816, Max(us): 35935, 50th(us): 20863, 90th(us): 22463, 95th(us): 22927, 99th(us): 24383, 99.9th(us): 33503, 99.99th(us): 35583
UPDATE - Takes(s): 3.7, Count: 50032, OPS: 13565.1, Avg(us): 3449, Min(us): 3100, Max(us): 14743, 50th(us): 3329, 90th(us): 3893, 95th(us): 4107, 99th(us): 4511, 99.9th(us): 8091, 99.99th(us): 14311
Error Summary:
```

+ 128

```bash
----------------------------------
Run finished, takes 2.816832397s
READ   - Takes(s): 2.8, Count: 49922, OPS: 17743.1, Avg(us): 3460, Min(us): 3096, Max(us): 22191, 50th(us): 3309, 90th(us): 3951, 95th(us): 4175, 99th(us): 4731, 99.9th(us): 8975, 99.99th(us): 14663
TOTAL  - Takes(s): 2.8, Count: 116640, OPS: 41459.2, Avg(us): 6033, Min(us): 3096, Max(us): 45439, 50th(us): 3375, 90th(us): 20639, 95th(us): 21695, 99th(us): 23407, 99.9th(us): 30335, 99.99th(us): 36447
TxnGroup - Takes(s): 2.8, Count: 16640, OPS: 5950.8, Avg(us): 21451, Min(us): 18848, Max(us): 45439, 50th(us): 21263, 90th(us): 23039, 95th(us): 23839, 99th(us): 27887, 99.9th(us): 36287, 99.99th(us): 41823
UPDATE - Takes(s): 2.8, Count: 50078, OPS: 17798.9, Avg(us): 3476, Min(us): 3096, Max(us): 20543, 50th(us): 3331, 90th(us): 3975, 95th(us): 4203, 99th(us): 4795, 99.9th(us): 9071, 99.99th(us): 12799
Error Summary:
```





## Effectiveness of Optimizations

### Performance Optimization

Setup:

+ oreo-rm
+ Record Count = 1000000
+ Operation Count = 100000
+ Txn Group = 6 (did change this time)
+ Lease Time = 100ms
+ Zipfian Constant = 0.8
+ Thread Num = 64

#### RMW=1

##### P0A1

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.8
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 13.137140436s
COMMIT - Takes(s): 13.1, Count: 15027, OPS: 1148.5, Avg(us): 28832, Min(us): 20176, Max(us): 55007, 50th(us): 28863, 90th(us): 30719, 95th(us): 31199, 99th(us): 32063, 99.9th(us): 48639, 99.99th(us): 53375
COMMIT_ERROR - Takes(s): 13.1, Count: 1613, OPS: 123.1, Avg(us): 17140, Min(us): 3880, Max(us): 29247, 50th(us): 17759, 90th(us): 22239, 95th(us): 22895, 99th(us): 23679, 99.9th(us): 24783, 99.99th(us): 29247
READ   - Takes(s): 13.1, Count: 98561, OPS: 7504.5, Avg(us): 3585, Min(us): 5, Max(us): 20783, 50th(us): 3385, 90th(us): 3985, 95th(us): 4267, 99th(us): 9927, 99.9th(us): 12119, 99.99th(us): 14591
READ_ERROR - Takes(s): 13.1, Count: 1439, OPS: 110.1, Avg(us): 8535, Min(us): 6224, Max(us): 23759, 50th(us): 7635, 90th(us): 11015, 95th(us): 11407, 99th(us): 12175, 99.9th(us): 23759, 99.99th(us): 23759
Start  - Takes(s): 13.1, Count: 16704, OPS: 1271.5, Avg(us): 37, Min(us): 14, Max(us): 1234, 50th(us): 28, 90th(us): 43, 95th(us): 52, 99th(us): 301, 99.9th(us): 941, 99.99th(us): 1225
TOTAL  - Takes(s): 13.1, Count: 260520, OPS: 19830.6, Avg(us): 9147, Min(us): 1, Max(us): 81983, 50th(us): 3255, 90th(us): 48287, 95th(us): 50975, 99th(us): 55327, 99.9th(us): 61439, 99.99th(us): 72703
TXN    - Takes(s): 13.1, Count: 15027, OPS: 1148.6, Avg(us): 51028, Min(us): 40064, Max(us): 78335, 50th(us): 50559, 90th(us): 54111, 95th(us): 56831, 99th(us): 60543, 99.9th(us): 73919, 99.99th(us): 78079
TXN_ERROR - Takes(s): 13.1, Count: 1613, OPS: 123.1, Avg(us): 39341, Min(us): 25088, Max(us): 58623, 50th(us): 39423, 90th(us): 44927, 95th(us): 46271, 99th(us): 51775, 99.9th(us): 55679, 99.99th(us): 58623
TxnGroup - Takes(s): 13.1, Count: 16640, OPS: 1268.7, Avg(us): 49789, Min(us): 22368, Max(us): 81983, 50th(us): 50207, 90th(us): 54911, 95th(us): 57375, 99th(us): 61279, 99.9th(us): 69759, 99.99th(us): 77375
UPDATE - Takes(s): 13.1, Count: 98561, OPS: 7505.1, Avg(us): 5, Min(us): 1, Max(us): 1044, 50th(us): 4, 90th(us): 5, 95th(us): 8, 99th(us): 21, 99.9th(us): 431, 99.99th(us): 820
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    1613

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    827
rollForward failed
  version mismatch  564
rollback failed
  version mismatch  48
```

##### P0A2

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.8
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 12.086578728s
COMMIT - Takes(s): 12.0, Count: 14813, OPS: 1230.4, Avg(us): 24774, Min(us): 16672, Max(us): 58783, 50th(us): 24655, 90th(us): 26431, 95th(us): 27007, 99th(us): 28239, 99.9th(us): 53727, 99.99th(us): 56511
COMMIT_ERROR - Takes(s): 12.0, Count: 1827, OPS: 151.7, Avg(us): 17332, Min(us): 3420, Max(us): 51903, 50th(us): 17759, 90th(us): 22207, 95th(us): 22895, 99th(us): 24383, 99.9th(us): 42623, 99.99th(us): 51903
READ   - Takes(s): 12.1, Count: 98535, OPS: 8154.7, Avg(us): 3574, Min(us): 5, Max(us): 31999, 50th(us): 3379, 90th(us): 3967, 95th(us): 4287, 99th(us): 9847, 99.9th(us): 11855, 99.99th(us): 20591
READ_ERROR - Takes(s): 12.0, Count: 1465, OPS: 121.8, Avg(us): 8412, Min(us): 6216, Max(us): 14095, 50th(us): 7523, 90th(us): 10855, 95th(us): 11359, 99th(us): 12239, 99.9th(us): 13583, 99.99th(us): 14095
Start  - Takes(s): 12.1, Count: 16704, OPS: 1382.0, Avg(us): 42, Min(us): 14, Max(us): 1431, 50th(us): 29, 90th(us): 46, 95th(us): 58, 99th(us): 440, 99.9th(us): 1179, 99.99th(us): 1402
TOTAL  - Takes(s): 12.1, Count: 260040, OPS: 21514.3, Avg(us): 8386, Min(us): 1, Max(us): 85311, 50th(us): 3255, 90th(us): 44575, 95th(us): 46495, 99th(us): 51391, 99.9th(us): 58495, 99.99th(us): 75775
TXN    - Takes(s): 12.0, Count: 14813, OPS: 1230.3, Avg(us): 46913, Min(us): 37184, Max(us): 85311, 50th(us): 46175, 90th(us): 50015, 95th(us): 52607, 99th(us): 57183, 99.9th(us): 75263, 99.99th(us): 84735
TXN_ERROR - Takes(s): 12.0, Count: 1827, OPS: 151.7, Avg(us): 39391, Min(us): 24048, Max(us): 72447, 50th(us): 39167, 90th(us): 44831, 95th(us): 46687, 99th(us): 51999, 99.9th(us): 63007, 99.99th(us): 72447
TxnGroup - Takes(s): 12.1, Count: 16640, OPS: 1378.9, Avg(us): 45998, Min(us): 20368, Max(us): 83263, 50th(us): 45855, 90th(us): 50879, 95th(us): 53151, 99th(us): 57791, 99.9th(us): 75647, 99.99th(us): 78911
UPDATE - Takes(s): 12.1, Count: 98535, OPS: 8154.9, Avg(us): 5, Min(us): 1, Max(us): 1452, 50th(us): 4, 90th(us): 5, 95th(us): 8, 99th(us): 20, 99.9th(us): 447, 99.99th(us): 929
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    1827

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    868
rollForward failed
  version mismatch  553
rollback failed
  version mismatch  44
```

##### P1A1

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 1
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 3ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
ZipfianConstant: 0.8
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 9.952215401s
COMMIT - Takes(s): 9.9, Count: 14089, OPS: 1421.8, Avg(us): 16445, Min(us): 14064, Max(us): 42879, 50th(us): 16167, 90th(us): 17839, 95th(us): 18671, 99th(us): 21295, 99.9th(us): 28591, 99.99th(us): 42527
COMMIT_ERROR - Takes(s): 9.9, Count: 2551, OPS: 257.3, Avg(us): 8919, Min(us): 7364, Max(us): 20063, 50th(us): 8711, 90th(us): 9983, 95th(us): 10591, 99th(us): 12319, 99.9th(us): 17119, 99.99th(us): 20063
READ   - Takes(s): 9.9, Count: 99713, OPS: 10023.9, Avg(us): 3722, Min(us): 6, Max(us): 14199, 50th(us): 3629, 90th(us): 4143, 95th(us): 4399, 99th(us): 5051, 99.9th(us): 6231, 99.99th(us): 8679
READ_ERROR - Takes(s): 9.8, Count: 287, OPS: 29.2, Avg(us): 4163, Min(us): 3232, Max(us): 6867, 50th(us): 3991, 90th(us): 5119, 95th(us): 5683, 99th(us): 6667, 99.9th(us): 6867, 99.99th(us): 6867
Start  - Takes(s): 10.0, Count: 16704, OPS: 1678.2, Avg(us): 26, Min(us): 14, Max(us): 923, 50th(us): 25, 90th(us): 37, 95th(us): 41, 99th(us): 78, 99.9th(us): 356, 99.99th(us): 667
TOTAL  - Takes(s): 10.0, Count: 260948, OPS: 26218.8, Avg(us): 6821, Min(us): 1, Max(us): 66431, 50th(us): 3435, 90th(us): 37215, 95th(us): 38783, 99th(us): 40831, 99.9th(us): 44767, 99.99th(us): 52639
TXN    - Takes(s): 9.9, Count: 14089, OPS: 1421.8, Avg(us): 38929, Min(us): 32752, Max(us): 66431, 50th(us): 38655, 90th(us): 40735, 95th(us): 41727, 99th(us): 44767, 99.9th(us): 52351, 99.99th(us): 64959
TXN_ERROR - Takes(s): 9.9, Count: 2551, OPS: 257.3, Avg(us): 31476, Min(us): 26160, Max(us): 44383, 50th(us): 31247, 90th(us): 33151, 95th(us): 33919, 99th(us): 35935, 99.9th(us): 39839, 99.99th(us): 44383
TxnGroup - Takes(s): 9.9, Count: 16640, OPS: 1675.6, Avg(us): 37729, Min(us): 22656, Max(us): 64511, 50th(us): 38399, 90th(us): 40479, 95th(us): 41407, 99th(us): 44191, 99.9th(us): 51647, 99.99th(us): 60223
UPDATE - Takes(s): 9.9, Count: 99713, OPS: 10023.9, Avg(us): 3, Min(us): 1, Max(us): 1321, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 11, 99.9th(us): 144, 99.99th(us): 449
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2551

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    209
rollForward failed
  version mismatch  68
rollback failed
  version mismatch  10
```

##### P2A1

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 3ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
ZipfianConstant: 0.8
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 9.245110579s
COMMIT - Takes(s): 9.2, Count: 13958, OPS: 1516.4, Avg(us): 13035, Min(us): 10672, Max(us): 47391, 50th(us): 12711, 90th(us): 14783, 95th(us): 15543, 99th(us): 17871, 99.9th(us): 37247, 99.99th(us): 46303
COMMIT_ERROR - Takes(s): 9.2, Count: 2682, OPS: 291.1, Avg(us): 5164, Min(us): 3906, Max(us): 37567, 50th(us): 4895, 90th(us): 6335, 95th(us): 6963, 99th(us): 8711, 99.9th(us): 11151, 99.99th(us): 37567
READ   - Takes(s): 9.2, Count: 99774, OPS: 10796.6, Avg(us): 3852, Min(us): 6, Max(us): 37311, 50th(us): 3681, 90th(us): 4447, 95th(us): 4839, 99th(us): 5815, 99.9th(us): 7815, 99.99th(us): 34815
READ_ERROR - Takes(s): 9.1, Count: 226, OPS: 24.7, Avg(us): 5064, Min(us): 3266, Max(us): 11359, 50th(us): 4667, 90th(us): 6975, 95th(us): 7583, 99th(us): 9919, 99.9th(us): 11359, 99.99th(us): 11359
Start  - Takes(s): 9.2, Count: 16704, OPS: 1806.7, Avg(us): 28, Min(us): 14, Max(us): 1109, 50th(us): 26, 90th(us): 37, 95th(us): 42, 99th(us): 100, 99.9th(us): 561, 99.99th(us): 925
TOTAL  - Takes(s): 9.2, Count: 260808, OPS: 28206.8, Avg(us): 6352, Min(us): 1, Max(us): 69567, 50th(us): 3465, 90th(us): 33983, 95th(us): 36095, 99th(us): 38623, 99.9th(us): 43231, 99.99th(us): 66943
TXN    - Takes(s): 9.2, Count: 13958, OPS: 1516.4, Avg(us): 36313, Min(us): 29392, Max(us): 69567, 50th(us): 35935, 90th(us): 38559, 95th(us): 39519, 99th(us): 43359, 99.9th(us): 66879, 99.99th(us): 69055
TXN_ERROR - Takes(s): 9.2, Count: 2682, OPS: 291.1, Avg(us): 28570, Min(us): 22768, Max(us): 62783, 50th(us): 28239, 90th(us): 30655, 95th(us): 31423, 99th(us): 33599, 99.9th(us): 59743, 99.99th(us): 62783
TxnGroup - Takes(s): 9.2, Count: 16640, OPS: 1804.4, Avg(us): 35023, Min(us): 22672, Max(us): 69567, 50th(us): 35583, 90th(us): 38271, 95th(us): 39263, 99th(us): 42047, 99.9th(us): 65983, 99.99th(us): 69119
UPDATE - Takes(s): 9.2, Count: 99774, OPS: 10796.5, Avg(us): 3, Min(us): 1, Max(us): 1097, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 12, 99.9th(us): 206, 99.99th(us): 725
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2682

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  122
  read failed due to unknown txn status   78
rollback failed
  version mismatch  26
```

##### P2A2

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 3ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
ZipfianConstant: 0.8
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 8.3047005s
COMMIT - Takes(s): 8.3, Count: 13815, OPS: 1671.1, Avg(us): 8721, Min(us): 7056, Max(us): 26895, 50th(us): 8471, 90th(us): 10039, 95th(us): 10727, 99th(us): 12967, 99.9th(us): 17215, 99.99th(us): 25247
COMMIT_ERROR - Takes(s): 8.3, Count: 2825, OPS: 341.7, Avg(us): 5238, Min(us): 3820, Max(us): 14823, 50th(us): 4967, 90th(us): 6487, 95th(us): 7195, 99th(us): 9423, 99.9th(us): 14079, 99.99th(us): 14823
READ   - Takes(s): 8.3, Count: 99778, OPS: 12019.6, Avg(us): 3887, Min(us): 5, Max(us): 13935, 50th(us): 3721, 90th(us): 4519, 95th(us): 4903, 99th(us): 5939, 99.9th(us): 8051, 99.99th(us): 11695
READ_ERROR - Takes(s): 8.3, Count: 222, OPS: 26.9, Avg(us): 4983, Min(us): 3300, Max(us): 9391, 50th(us): 4819, 90th(us): 6683, 95th(us): 7039, 99th(us): 8463, 99.9th(us): 9391, 99.99th(us): 9391
Start  - Takes(s): 8.3, Count: 16704, OPS: 2011.4, Avg(us): 30, Min(us): 13, Max(us): 1785, 50th(us): 26, 90th(us): 38, 95th(us): 43, 99th(us): 218, 99.9th(us): 839, 99.99th(us): 1160
TOTAL  - Takes(s): 8.3, Count: 260530, OPS: 31368.6, Avg(us): 5682, Min(us): 1, Max(us): 53727, 50th(us): 3487, 90th(us): 30207, 95th(us): 32031, 99th(us): 34431, 99.9th(us): 38751, 99.99th(us): 43935
TXN    - Takes(s): 8.3, Count: 13815, OPS: 1671.0, Avg(us): 32214, Min(us): 25824, Max(us): 53727, 50th(us): 31903, 90th(us): 34271, 95th(us): 35327, 99th(us): 38911, 99.9th(us): 43935, 99.99th(us): 53535
TXN_ERROR - Takes(s): 8.3, Count: 2825, OPS: 341.7, Avg(us): 28849, Min(us): 22528, Max(us): 41023, 50th(us): 28607, 90th(us): 30991, 95th(us): 31903, 99th(us): 35519, 99.9th(us): 40159, 99.99th(us): 41023
TxnGroup - Takes(s): 8.3, Count: 16640, OPS: 2009.5, Avg(us): 31614, Min(us): 22272, Max(us): 53055, 50th(us): 31599, 90th(us): 34111, 95th(us): 35199, 99th(us): 38047, 99.9th(us): 42719, 99.99th(us): 50495
UPDATE - Takes(s): 8.3, Count: 99778, OPS: 12023.2, Avg(us): 3, Min(us): 1, Max(us): 2377, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 13, 99.9th(us): 216, 99.99th(us): 706
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2825

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  110
  read failed due to unknown txn status   92
rollback failed
  version mismatch  20
```

#### Workload F

##### P0A1

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.8
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 10.476185872s
COMMIT - Takes(s): 10.5, Count: 16054, OPS: 1535.6, Avg(us): 17705, Min(us): 0, Max(us): 32367, 50th(us): 17871, 90th(us): 23839, 95th(us): 25471, 99th(us): 28351, 99.9th(us): 30447, 99.99th(us): 31695
COMMIT_ERROR - Takes(s): 10.4, Count: 586, OPS: 56.1, Avg(us): 10942, Min(us): 3484, Max(us): 22975, 50th(us): 11039, 90th(us): 16239, 95th(us): 18335, 99th(us): 19615, 99.9th(us): 22143, 99.99th(us): 22975
READ   - Takes(s): 10.5, Count: 99212, OPS: 9473.3, Avg(us): 3553, Min(us): 5, Max(us): 16055, 50th(us): 3381, 90th(us): 3931, 95th(us): 4215, 99th(us): 9487, 99.9th(us): 11599, 99.99th(us): 13823
READ_ERROR - Takes(s): 10.4, Count: 788, OPS: 75.6, Avg(us): 8905, Min(us): 6240, Max(us): 12863, 50th(us): 9831, 90th(us): 11071, 95th(us): 11391, 99th(us): 12111, 99.9th(us): 12423, 99.99th(us): 12863
Start  - Takes(s): 10.5, Count: 16704, OPS: 1594.5, Avg(us): 39, Min(us): 14, Max(us): 1897, 50th(us): 28, 90th(us): 44, 95th(us): 55, 99th(us): 358, 99.9th(us): 1369, 99.99th(us): 1802
TOTAL  - Takes(s): 10.5, Count: 214220, OPS: 20447.8, Avg(us): 8975, Min(us): 0, Max(us): 67583, 50th(us): 3345, 90th(us): 37215, 95th(us): 41759, 99th(us): 47327, 99.9th(us): 53407, 99.99th(us): 57759
TXN    - Takes(s): 10.5, Count: 16054, OPS: 1535.5, Avg(us): 39463, Min(us): 17024, Max(us): 61151, 50th(us): 39455, 90th(us): 46303, 95th(us): 48127, 99th(us): 51999, 99.9th(us): 57055, 99.99th(us): 60511
TXN_ERROR - Takes(s): 10.4, Count: 586, OPS: 56.1, Avg(us): 32570, Min(us): 23168, Max(us): 48607, 50th(us): 32447, 90th(us): 39071, 95th(us): 40479, 99th(us): 43423, 99.9th(us): 44767, 99.99th(us): 48607
TxnGroup - Takes(s): 10.5, Count: 16640, OPS: 1591.3, Avg(us): 39152, Min(us): 19120, Max(us): 67583, 50th(us): 39231, 90th(us): 46207, 95th(us): 48031, 99th(us): 52383, 99.9th(us): 57215, 99.99th(us): 59999
UPDATE - Takes(s): 10.5, Count: 49556, OPS: 4732.1, Avg(us): 6, Min(us): 1, Max(us): 1228, 50th(us): 4, 90th(us): 5, 95th(us): 8, 99th(us): 23, 99.9th(us): 527, 99.99th(us): 865
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  411
  read failed due to unknown txn status  355
rollback failed
  version mismatch  22

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch     586
```

##### P0A2

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.8
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 9.545926887s
COMMIT - Takes(s): 9.5, Count: 15982, OPS: 1677.6, Avg(us): 14121, Min(us): 0, Max(us): 38815, 50th(us): 14159, 90th(us): 20175, 95th(us): 21759, 99th(us): 24863, 99.9th(us): 27615, 99.99th(us): 32735
COMMIT_ERROR - Takes(s): 9.5, Count: 658, OPS: 69.2, Avg(us): 10835, Min(us): 3470, Max(us): 27231, 50th(us): 10863, 90th(us): 15623, 95th(us): 18191, 99th(us): 20175, 99.9th(us): 26511, 99.99th(us): 27231
READ   - Takes(s): 9.5, Count: 99125, OPS: 10386.9, Avg(us): 3587, Min(us): 5, Max(us): 28431, 50th(us): 3393, 90th(us): 4015, 95th(us): 4303, 99th(us): 9655, 99.9th(us): 11775, 99.99th(us): 14431
READ_ERROR - Takes(s): 9.5, Count: 875, OPS: 92.2, Avg(us): 9043, Min(us): 6232, Max(us): 13143, 50th(us): 9879, 90th(us): 11183, 95th(us): 11583, 99th(us): 12255, 99.9th(us): 13103, 99.99th(us): 13143
Start  - Takes(s): 9.5, Count: 16704, OPS: 1749.7, Avg(us): 41, Min(us): 14, Max(us): 1463, 50th(us): 28, 90th(us): 46, 95th(us): 59, 99th(us): 418, 99.9th(us): 896, 99.99th(us): 1280
TOTAL  - Takes(s): 9.5, Count: 213907, OPS: 22404.6, Avg(us): 8216, Min(us): 0, Max(us): 61183, 50th(us): 3353, 90th(us): 33887, 95th(us): 38175, 99th(us): 44031, 99.9th(us): 50463, 99.99th(us): 55871
TXN    - Takes(s): 9.5, Count: 15982, OPS: 1677.7, Avg(us): 36117, Min(us): 16704, Max(us): 60543, 50th(us): 36031, 90th(us): 42847, 95th(us): 44799, 99th(us): 48863, 99.9th(us): 54751, 99.99th(us): 57247
TXN_ERROR - Takes(s): 9.5, Count: 658, OPS: 69.3, Avg(us): 32869, Min(us): 23488, Max(us): 51807, 50th(us): 32719, 90th(us): 38783, 95th(us): 40959, 99th(us): 46111, 99.9th(us): 51455, 99.99th(us): 51807
TxnGroup - Takes(s): 9.5, Count: 16640, OPS: 1746.9, Avg(us): 35936, Min(us): 19440, Max(us): 61183, 50th(us): 35903, 90th(us): 42719, 95th(us): 44671, 99th(us): 49567, 99.9th(us): 54815, 99.99th(us): 58751
UPDATE - Takes(s): 9.5, Count: 49474, OPS: 5184.2, Avg(us): 6, Min(us): 1, Max(us): 1149, 50th(us): 4, 90th(us): 5, 95th(us): 8, 99th(us): 22, 99.9th(us): 484, 99.99th(us): 910
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch     658

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  491
  read failed due to unknown txn status  373
rollback failed
  version mismatch  11
```

##### P1A1

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 1
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 3ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
ZipfianConstant: 0.8
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 9.512997436s
COMMIT - Takes(s): 9.5, Count: 15783, OPS: 1666.2, Avg(us): 14607, Min(us): 0, Max(us): 23423, 50th(us): 14719, 90th(us): 15951, 95th(us): 16415, 99th(us): 17391, 99.9th(us): 20015, 99.99th(us): 22975
COMMIT_ERROR - Takes(s): 9.5, Count: 857, OPS: 90.4, Avg(us): 7875, Min(us): 6668, Max(us): 11487, 50th(us): 7759, 90th(us): 8735, 95th(us): 9039, 99th(us): 9631, 99.9th(us): 10855, 99.99th(us): 11487
READ   - Takes(s): 9.5, Count: 99838, OPS: 10500.0, Avg(us): 3636, Min(us): 5, Max(us): 6791, 50th(us): 3557, 90th(us): 4019, 95th(us): 4259, 99th(us): 4823, 99.9th(us): 5607, 99.99th(us): 6043
READ_ERROR - Takes(s): 9.5, Count: 162, OPS: 17.1, Avg(us): 3950, Min(us): 3264, Max(us): 6015, 50th(us): 3779, 90th(us): 4703, 95th(us): 5319, 99th(us): 5563, 99.9th(us): 6015, 99.99th(us): 6015
Start  - Takes(s): 9.5, Count: 16704, OPS: 1755.9, Avg(us): 29, Min(us): 13, Max(us): 1894, 50th(us): 25, 90th(us): 37, 95th(us): 42, 99th(us): 180, 99.9th(us): 771, 99.99th(us): 1182
TOTAL  - Takes(s): 9.5, Count: 214834, OPS: 22580.7, Avg(us): 8249, Min(us): 0, Max(us): 49215, 50th(us): 3519, 90th(us): 35839, 95th(us): 37311, 99th(us): 39071, 99.9th(us): 40927, 99.99th(us): 45567
TXN    - Takes(s): 9.5, Count: 15783, OPS: 1666.1, Avg(us): 36535, Min(us): 20272, Max(us): 49215, 50th(us): 36575, 90th(us): 38751, 95th(us): 39359, 99th(us): 40639, 99.9th(us): 45567, 99.99th(us): 48415
TXN_ERROR - Takes(s): 9.5, Count: 857, OPS: 90.4, Avg(us): 29860, Min(us): 27280, Max(us): 35423, 50th(us): 29695, 90th(us): 31599, 95th(us): 32143, 99th(us): 32799, 99.9th(us): 34655, 99.99th(us): 35423
TxnGroup - Takes(s): 9.5, Count: 16640, OPS: 1753.7, Avg(us): 36138, Min(us): 20336, Max(us): 47359, 50th(us): 36447, 90th(us): 38655, 95th(us): 39263, 99th(us): 40415, 99.9th(us): 42495, 99.99th(us): 46303
UPDATE - Takes(s): 9.5, Count: 50086, OPS: 5267.4, Avg(us): 3, Min(us): 1, Max(us): 861, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 13, 99.9th(us): 208, 99.99th(us): 709
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  857

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    106
rollForward failed
  version mismatch  55
rollback failed
  version mismatch  1
```

##### P2A1

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 3ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
ZipfianConstant: 0.8
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 8.623140184s
COMMIT - Takes(s): 8.6, Count: 15749, OPS: 1831.8, Avg(us): 11185, Min(us): 0, Max(us): 25103, 50th(us): 11159, 90th(us): 12447, 95th(us): 12935, 99th(us): 14015, 99.9th(us): 18191, 99.99th(us): 23679
COMMIT_ERROR - Takes(s): 8.6, Count: 891, OPS: 103.7, Avg(us): 4290, Min(us): 3396, Max(us): 10575, 50th(us): 4131, 90th(us): 4947, 95th(us): 5359, 99th(us): 6003, 99.9th(us): 10015, 99.99th(us): 10575
READ   - Takes(s): 8.6, Count: 99899, OPS: 11591.5, Avg(us): 3657, Min(us): 5, Max(us): 16495, 50th(us): 3561, 90th(us): 4085, 95th(us): 4383, 99th(us): 5091, 99.9th(us): 6059, 99.99th(us): 9007
READ_ERROR - Takes(s): 8.6, Count: 101, OPS: 11.8, Avg(us): 4183, Min(us): 3290, Max(us): 9239, 50th(us): 3995, 90th(us): 5147, 95th(us): 5403, 99th(us): 6015, 99.9th(us): 9239, 99.99th(us): 9239
Start  - Takes(s): 8.6, Count: 16704, OPS: 1937.1, Avg(us): 29, Min(us): 14, Max(us): 2037, 50th(us): 26, 90th(us): 36, 95th(us): 41, 99th(us): 170, 99.9th(us): 816, 99.99th(us): 1192
TOTAL  - Takes(s): 8.6, Count: 214778, OPS: 24905.0, Avg(us): 7506, Min(us): 0, Max(us): 49343, 50th(us): 3523, 90th(us): 32367, 95th(us): 33855, 99th(us): 36095, 99.9th(us): 38463, 99.99th(us): 45503
TXN    - Takes(s): 8.6, Count: 15749, OPS: 1831.8, Avg(us): 33243, Min(us): 20256, Max(us): 48383, 50th(us): 33055, 90th(us): 35647, 95th(us): 36479, 99th(us): 38207, 99.9th(us): 43679, 99.99th(us): 47807
TXN_ERROR - Takes(s): 8.6, Count: 891, OPS: 103.7, Avg(us): 26382, Min(us): 21024, Max(us): 35423, 50th(us): 26063, 90th(us): 28223, 95th(us): 28847, 99th(us): 30783, 99.9th(us): 35135, 99.99th(us): 35423
TxnGroup - Takes(s): 8.6, Count: 16640, OPS: 1935.1, Avg(us): 32835, Min(us): 20240, Max(us): 49343, 50th(us): 32927, 90th(us): 35551, 95th(us): 36319, 99th(us): 37823, 99.9th(us): 41087, 99.99th(us): 48031
UPDATE - Takes(s): 8.6, Count: 50037, OPS: 5805.3, Avg(us): 4, Min(us): 1, Max(us): 1566, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 13, 99.9th(us): 205, 99.99th(us): 965
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  891

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  55
  read failed due to unknown txn status  45
rollback failed
  version mismatch  1
```

##### P2A2

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 3ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
ZipfianConstant: 0.8
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 7.697631799s
COMMIT - Takes(s): 7.7, Count: 15665, OPS: 2042.8, Avg(us): 7457, Min(us): 0, Max(us): 20607, 50th(us): 7411, 90th(us): 8295, 95th(us): 8703, 99th(us): 9887, 99.9th(us): 15119, 99.99th(us): 20511
COMMIT_ERROR - Takes(s): 7.7, Count: 975, OPS: 127.3, Avg(us): 4300, Min(us): 3528, Max(us): 10239, 50th(us): 4155, 90th(us): 4971, 95th(us): 5247, 99th(us): 5919, 99.9th(us): 9231, 99.99th(us): 10239
READ   - Takes(s): 7.7, Count: 99894, OPS: 12987.4, Avg(us): 3657, Min(us): 6, Max(us): 11799, 50th(us): 3567, 90th(us): 4057, 95th(us): 4327, 99th(us): 4955, 99.9th(us): 6251, 99.99th(us): 7255
READ_ERROR - Takes(s): 7.7, Count: 106, OPS: 13.9, Avg(us): 4013, Min(us): 3262, Max(us): 5903, 50th(us): 3897, 90th(us): 4627, 95th(us): 5499, 99th(us): 5871, 99.9th(us): 5903, 99.99th(us): 5903
Start  - Takes(s): 7.7, Count: 16704, OPS: 2170.0, Avg(us): 30, Min(us): 13, Max(us): 1023, 50th(us): 26, 90th(us): 38, 95th(us): 43, 99th(us): 214, 99.9th(us): 750, 99.99th(us): 931
TOTAL  - Takes(s): 7.7, Count: 214455, OPS: 27860.8, Avg(us): 6680, Min(us): 0, Max(us): 50687, 50th(us): 3529, 90th(us): 28895, 95th(us): 29775, 99th(us): 31583, 99.9th(us): 34751, 99.99th(us): 45183
TXN    - Takes(s): 7.7, Count: 15665, OPS: 2042.6, Avg(us): 29511, Min(us): 20656, Max(us): 50687, 50th(us): 29311, 90th(us): 31215, 95th(us): 31935, 99th(us): 33855, 99.9th(us): 45599, 99.99th(us): 50399
TXN_ERROR - Takes(s): 7.7, Count: 975, OPS: 127.3, Avg(us): 26380, Min(us): 21632, Max(us): 40735, 50th(us): 26111, 90th(us): 27887, 95th(us): 28623, 99th(us): 30207, 99.9th(us): 40127, 99.99th(us): 40735
TxnGroup - Takes(s): 7.7, Count: 16640, OPS: 2169.4, Avg(us): 29299, Min(us): 20544, Max(us): 43199, 50th(us): 29247, 90th(us): 31087, 95th(us): 31823, 99th(us): 34015, 99.9th(us): 38143, 99.99th(us): 42559
UPDATE - Takes(s): 7.7, Count: 49887, OPS: 6486.8, Avg(us): 3, Min(us): 1, Max(us): 1350, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 12, 99.9th(us): 220, 99.99th(us): 833
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  975

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status     60
rollForward failed
  version mismatch  45
rollback failed
  version mismatch  1
```

Setup:

+ oreo-rm
+ Record Count = 1000000
+ Operation Count = 100000
+ Txn Group = 6
+ Lease Time = 100ms
+ Zipfian Constant = 0.99
+ Thread Num = 64

### Protocol Optimization

Setup:

+ redis-mongo
+ Record Count = 1000000
+ Operation Count = 100000
+ Txn Group = 6
+ Lease Time = 100ms

#### RMW =1

##### Cherry Garcia

+ 8

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 8
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 1m45.596869962s
COMMIT - Takes(s): 105.5, Count: 15275, OPS: 144.7, Avg(us): 29180, Min(us): 21056, Max(us): 56223, 50th(us): 29007, 90th(us): 31263, 95th(us): 31631, 99th(us): 32127, 99.9th(us): 32623, 99.99th(us): 46591
COMMIT_ERROR - Takes(s): 105.6, Count: 1389, OPS: 13.2, Avg(us): 17112, Min(us): 3856, Max(us): 24415, 50th(us): 18031, 90th(us): 22031, 95th(us): 23119, 99th(us): 23903, 99.9th(us): 24367, 99.99th(us): 24415
READ   - Takes(s): 105.6, Count: 99008, OPS: 937.6, Avg(us): 3669, Min(us): 4, Max(us): 29711, 50th(us): 3563, 90th(us): 3911, 95th(us): 4031, 99th(us): 10503, 99.9th(us): 11743, 99.99th(us): 15143
READ_ERROR - Takes(s): 105.5, Count: 992, OPS: 9.4, Avg(us): 8573, Min(us): 6376, Max(us): 12039, 50th(us): 7411, 90th(us): 11087, 95th(us): 11527, 99th(us): 11959, 99.9th(us): 12039, 99.99th(us): 12039
Start  - Takes(s): 105.6, Count: 16672, OPS: 157.9, Avg(us): 25, Min(us): 13, Max(us): 726, 50th(us): 21, 90th(us): 32, 95th(us): 41, 99th(us): 167, 99.9th(us): 328, 99.99th(us): 588
TOTAL  - Takes(s): 105.6, Count: 261902, OPS: 2480.2, Avg(us): 9323, Min(us): 1, Max(us): 84799, 50th(us): 3375, 90th(us): 49919, 95th(us): 50591, 99th(us): 55679, 99.9th(us): 62047, 99.99th(us): 68863
TXN    - Takes(s): 105.5, Count: 15275, OPS: 144.7, Avg(us): 51629, Min(us): 36288, Max(us): 78463, 50th(us): 50431, 90th(us): 55039, 95th(us): 57311, 99th(us): 61599, 99.9th(us): 66879, 99.99th(us): 76287
TXN_ERROR - Takes(s): 105.6, Count: 1389, OPS: 13.2, Avg(us): 39581, Min(us): 22144, Max(us): 63391, 50th(us): 39455, 90th(us): 46079, 95th(us): 46975, 99th(us): 53375, 99.9th(us): 58623, 99.99th(us): 63391
TxnGroup - Takes(s): 105.6, Count: 16664, OPS: 157.8, Avg(us): 50612, Min(us): 21952, Max(us): 84799, 50th(us): 50367, 90th(us): 55135, 95th(us): 57439, 99th(us): 61855, 99.9th(us): 67967, 99.99th(us): 77311
UPDATE - Takes(s): 105.6, Count: 99008, OPS: 937.6, Avg(us): 3, Min(us): 1, Max(us): 592, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 14, 99.9th(us): 152, 99.99th(us): 264
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    1389

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    623
rollForward failed
  version mismatch  344
rollback failed
  version mismatch  25
```

+ 16

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 16
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 52.979985475s
COMMIT - Takes(s): 52.9, Count: 14268, OPS: 269.6, Avg(us): 29326, Min(us): 18528, Max(us): 36383, 50th(us): 29663, 90th(us): 30479, 95th(us): 31103, 99th(us): 32319, 99.9th(us): 33247, 99.99th(us): 34719
COMMIT_ERROR - Takes(s): 52.9, Count: 2388, OPS: 45.1, Avg(us): 17237, Min(us): 3600, Max(us): 25071, 50th(us): 18431, 90th(us): 22543, 95th(us): 22831, 99th(us): 23727, 99.9th(us): 24351, 99.99th(us): 25071
READ   - Takes(s): 53.0, Count: 98131, OPS: 1852.3, Avg(us): 3716, Min(us): 4, Max(us): 15471, 50th(us): 3611, 90th(us): 3999, 95th(us): 4135, 99th(us): 10783, 99.9th(us): 11751, 99.99th(us): 14759
READ_ERROR - Takes(s): 52.9, Count: 1869, OPS: 35.4, Avg(us): 8650, Min(us): 6212, Max(us): 12671, 50th(us): 7563, 90th(us): 11223, 95th(us): 11439, 99th(us): 11999, 99.9th(us): 12487, 99.99th(us): 12671
Start  - Takes(s): 53.0, Count: 16672, OPS: 314.7, Avg(us): 31, Min(us): 13, Max(us): 896, 50th(us): 25, 90th(us): 40, 95th(us): 53, 99th(us): 214, 99.9th(us): 351, 99.99th(us): 687
TOTAL  - Takes(s): 53.0, Count: 258126, OPS: 4872.0, Avg(us): 9199, Min(us): 1, Max(us): 77055, 50th(us): 3321, 90th(us): 50207, 95th(us): 51807, 99th(us): 57695, 99.9th(us): 62335, 99.99th(us): 66815
TXN    - Takes(s): 52.9, Count: 14268, OPS: 269.6, Avg(us): 52404, Min(us): 36224, Max(us): 73407, 50th(us): 51743, 90th(us): 55551, 95th(us): 58879, 99th(us): 60927, 99.9th(us): 66367, 99.99th(us): 70079
TXN_ERROR - Takes(s): 52.9, Count: 2388, OPS: 45.1, Avg(us): 40060, Min(us): 21488, Max(us): 56063, 50th(us): 40511, 90th(us): 44959, 95th(us): 47647, 99th(us): 51647, 99.9th(us): 55839, 99.99th(us): 56063
TxnGroup - Takes(s): 53.0, Count: 16656, OPS: 314.5, Avg(us): 50609, Min(us): 22512, Max(us): 77055, 50th(us): 51583, 90th(us): 55967, 95th(us): 58975, 99th(us): 62335, 99.9th(us): 66815, 99.99th(us): 70655
UPDATE - Takes(s): 53.0, Count: 98131, OPS: 1852.3, Avg(us): 4, Min(us): 1, Max(us): 577, 50th(us): 3, 90th(us): 5, 95th(us): 7, 99th(us): 19, 99.9th(us): 243, 99.99th(us): 401
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    2388

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1177
rollForward failed
  version mismatch  601
rollback failed
  version mismatch  91
```

+ 32

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 32
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 25.460112368s
COMMIT - Takes(s): 25.4, Count: 13170, OPS: 518.4, Avg(us): 28310, Min(us): 17456, Max(us): 61119, 50th(us): 28415, 90th(us): 30495, 95th(us): 31215, 99th(us): 32351, 99.9th(us): 40703, 99.99th(us): 51359
COMMIT_ERROR - Takes(s): 25.4, Count: 3470, OPS: 136.5, Avg(us): 16592, Min(us): 3538, Max(us): 25055, 50th(us): 17455, 90th(us): 21807, 95th(us): 22511, 99th(us): 23807, 99.9th(us): 24511, 99.99th(us): 25055
READ   - Takes(s): 25.5, Count: 97224, OPS: 3819.3, Avg(us): 3587, Min(us): 5, Max(us): 21215, 50th(us): 3377, 90th(us): 4027, 95th(us): 4275, 99th(us): 10103, 99.9th(us): 11879, 99.99th(us): 13983
READ_ERROR - Takes(s): 25.4, Count: 2776, OPS: 109.4, Avg(us): 8443, Min(us): 6224, Max(us): 13031, 50th(us): 7531, 90th(us): 11023, 95th(us): 11543, 99th(us): 12167, 99.9th(us): 12727, 99.99th(us): 13031
Start  - Takes(s): 25.5, Count: 16672, OPS: 654.8, Avg(us): 33, Min(us): 14, Max(us): 1185, 50th(us): 27, 90th(us): 43, 95th(us): 54, 99th(us): 235, 99.9th(us): 462, 99.99th(us): 760
TOTAL  - Takes(s): 25.5, Count: 254100, OPS: 9979.9, Avg(us): 8650, Min(us): 1, Max(us): 82431, 50th(us): 3235, 90th(us): 46431, 95th(us): 50207, 99th(us): 55711, 99.9th(us): 61663, 99.99th(us): 68863
TXN    - Takes(s): 25.4, Count: 13170, OPS: 518.3, Avg(us): 50908, Min(us): 34016, Max(us): 81535, 50th(us): 50143, 90th(us): 54943, 95th(us): 56799, 99th(us): 60639, 99.9th(us): 68287, 99.99th(us): 73535
TXN_ERROR - Takes(s): 25.4, Count: 3470, OPS: 136.5, Avg(us): 38983, Min(us): 23424, Max(us): 62655, 50th(us): 39071, 90th(us): 44735, 95th(us): 46719, 99th(us): 51423, 99.9th(us): 55615, 99.99th(us): 62655
TxnGroup - Takes(s): 25.4, Count: 16640, OPS: 654.1, Avg(us): 48366, Min(us): 21632, Max(us): 82431, 50th(us): 49343, 90th(us): 55423, 95th(us): 57375, 99th(us): 61823, 99.9th(us): 68351, 99.99th(us): 79167
UPDATE - Takes(s): 25.5, Count: 97224, OPS: 3819.1, Avg(us): 5, Min(us): 1, Max(us): 1390, 50th(us): 4, 90th(us): 5, 95th(us): 7, 99th(us): 21, 99.9th(us): 301, 99.99th(us): 579
Error Summary:

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1624
rollForward failed
  version mismatch  933
rollback failed
  version mismatch  219

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    3470
```

+ 64

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 12.681644773s
COMMIT - Takes(s): 12.6, Count: 12037, OPS: 952.4, Avg(us): 27894, Min(us): 14904, Max(us): 46943, 50th(us): 28111, 90th(us): 30255, 95th(us): 30799, 99th(us): 31823, 99.9th(us): 43935, 99.99th(us): 46783
COMMIT_ERROR - Takes(s): 12.7, Count: 4603, OPS: 363.9, Avg(us): 16328, Min(us): 3468, Max(us): 34047, 50th(us): 17263, 90th(us): 21599, 95th(us): 22447, 99th(us): 23503, 99.9th(us): 24703, 99.99th(us): 34047
READ   - Takes(s): 12.7, Count: 95945, OPS: 7567.0, Avg(us): 3582, Min(us): 4, Max(us): 24191, 50th(us): 3367, 90th(us): 3907, 95th(us): 4219, 99th(us): 10207, 99.9th(us): 11975, 99.99th(us): 14855
READ_ERROR - Takes(s): 12.6, Count: 4055, OPS: 320.9, Avg(us): 8368, Min(us): 6192, Max(us): 28239, 50th(us): 7419, 90th(us): 10823, 95th(us): 11231, 99th(us): 12111, 99.9th(us): 17999, 99.99th(us): 28239
Start  - Takes(s): 12.7, Count: 16704, OPS: 1317.1, Avg(us): 39, Min(us): 14, Max(us): 1594, 50th(us): 28, 90th(us): 45, 95th(us): 57, 99th(us): 305, 99.9th(us): 1452, 99.99th(us): 1575
TOTAL  - Takes(s): 12.7, Count: 249308, OPS: 19657.6, Avg(us): 8356, Min(us): 1, Max(us): 74239, 50th(us): 3225, 90th(us): 42975, 95th(us): 49983, 99th(us): 55839, 99.9th(us): 62719, 99.99th(us): 68287
TXN    - Takes(s): 12.6, Count: 12037, OPS: 952.5, Avg(us): 50850, Min(us): 40544, Max(us): 70783, 50th(us): 50143, 90th(us): 55071, 95th(us): 57183, 99th(us): 61599, 99.9th(us): 67967, 99.99th(us): 70399
TXN_ERROR - Takes(s): 12.7, Count: 4603, OPS: 363.9, Avg(us): 39122, Min(us): 23200, Max(us): 64191, 50th(us): 39135, 90th(us): 45151, 95th(us): 47519, 99th(us): 51615, 99.9th(us): 57759, 99.99th(us): 64191
TxnGroup - Takes(s): 12.7, Count: 16640, OPS: 1313.7, Avg(us): 47505, Min(us): 16896, Max(us): 74239, 50th(us): 48703, 90th(us): 55551, 95th(us): 57759, 99th(us): 62623, 99.9th(us): 67519, 99.99th(us): 72511
UPDATE - Takes(s): 12.7, Count: 95945, OPS: 7567.6, Avg(us): 5, Min(us): 1, Max(us): 1528, 50th(us): 4, 90th(us): 5, 95th(us): 8, 99th(us): 20, 99.9th(us): 418, 99.99th(us): 811
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    4603

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   2389
rollForward failed
  version mismatch  1315
rollback failed
  version mismatch  351
```

+ 96

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 96
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 8.273800253s
COMMIT - Takes(s): 8.2, Count: 11332, OPS: 1378.1, Avg(us): 27251, Min(us): 10584, Max(us): 53567, 50th(us): 27679, 90th(us): 29663, 95th(us): 30367, 99th(us): 32479, 99.9th(us): 39775, 99.99th(us): 46943
COMMIT_ERROR - Takes(s): 8.2, Count: 5276, OPS: 640.8, Avg(us): 16195, Min(us): 3444, Max(us): 55135, 50th(us): 17151, 90th(us): 21423, 95th(us): 22143, 99th(us): 24175, 99.9th(us): 31327, 99.99th(us): 34655
READ   - Takes(s): 8.3, Count: 94876, OPS: 11472.9, Avg(us): 3541, Min(us): 5, Max(us): 16383, 50th(us): 3343, 90th(us): 3797, 95th(us): 4155, 99th(us): 10023, 99.9th(us): 11695, 99.99th(us): 13911
READ_ERROR - Takes(s): 8.2, Count: 5124, OPS: 623.2, Avg(us): 8215, Min(us): 6200, Max(us): 14735, 50th(us): 7163, 90th(us): 10471, 95th(us): 10855, 99th(us): 11663, 99.9th(us): 12991, 99.99th(us): 14007
Start  - Takes(s): 8.3, Count: 16704, OPS: 2018.7, Avg(us): 38, Min(us): 14, Max(us): 1648, 50th(us): 29, 90th(us): 44, 95th(us): 55, 99th(us): 303, 99.9th(us): 975, 99.99th(us): 1477
TOTAL  - Takes(s): 8.3, Count: 245728, OPS: 29696.1, Avg(us): 8094, Min(us): 1, Max(us): 80319, 50th(us): 3197, 90th(us): 40799, 95th(us): 49151, 99th(us): 55167, 99.9th(us): 62111, 99.99th(us): 69055
TXN    - Takes(s): 8.2, Count: 11332, OPS: 1377.9, Avg(us): 50299, Min(us): 39904, Max(us): 80319, 50th(us): 49439, 90th(us): 54591, 95th(us): 56479, 99th(us): 60927, 99.9th(us): 67967, 99.99th(us): 73151
TXN_ERROR - Takes(s): 8.2, Count: 5276, OPS: 640.8, Avg(us): 38942, Min(us): 19936, Max(us): 74879, 50th(us): 38911, 90th(us): 44959, 95th(us): 47359, 99th(us): 51295, 99.9th(us): 58335, 99.99th(us): 63167
TxnGroup - Takes(s): 8.3, Count: 16608, OPS: 2012.1, Avg(us): 46547, Min(us): 19968, Max(us): 78079, 50th(us): 47871, 90th(us): 54943, 95th(us): 57279, 99th(us): 62207, 99.9th(us): 69695, 99.99th(us): 73983
UPDATE - Takes(s): 8.3, Count: 94876, OPS: 11474.6, Avg(us): 6, Min(us): 1, Max(us): 1588, 50th(us): 4, 90th(us): 5, 95th(us): 8, 99th(us): 21, 99.9th(us): 597, 99.99th(us): 1001
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    5276

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   2959
rollForward failed
  version mismatch  1655
rollback failed
  version mismatch  510
```

+ 128

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 3ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 6.376523444s
COMMIT - Takes(s): 6.3, Count: 10925, OPS: 1727.7, Avg(us): 27770, Min(us): 14624, Max(us): 44607, 50th(us): 28351, 90th(us): 30335, 95th(us): 31071, 99th(us): 32863, 99.9th(us): 37951, 99.99th(us): 42527
COMMIT_ERROR - Takes(s): 6.3, Count: 5715, OPS: 901.4, Avg(us): 16205, Min(us): 3340, Max(us): 36031, 50th(us): 16911, 90th(us): 21807, 95th(us): 22607, 99th(us): 24175, 99.9th(us): 28623, 99.99th(us): 33663
READ   - Takes(s): 6.4, Count: 94233, OPS: 14786.4, Avg(us): 3642, Min(us): 5, Max(us): 24623, 50th(us): 3377, 90th(us): 4047, 95th(us): 4483, 99th(us): 10367, 99.9th(us): 12311, 99.99th(us): 15591
READ_ERROR - Takes(s): 6.3, Count: 5767, OPS: 909.5, Avg(us): 8419, Min(us): 6212, Max(us): 20335, 50th(us): 7451, 90th(us): 10783, 95th(us): 11255, 99th(us): 12359, 99.9th(us): 14207, 99.99th(us): 20063
Start  - Takes(s): 6.4, Count: 16768, OPS: 2629.2, Avg(us): 46, Min(us): 14, Max(us): 2489, 50th(us): 30, 90th(us): 47, 95th(us): 57, 99th(us): 576, 99.9th(us): 1268, 99.99th(us): 2063
TOTAL  - Takes(s): 6.4, Count: 243724, OPS: 38215.4, Avg(us): 8216, Min(us): 1, Max(us): 76927, 50th(us): 3231, 90th(us): 40255, 95th(us): 50431, 99th(us): 56895, 99.9th(us): 63583, 99.99th(us): 70015
TXN    - Takes(s): 6.3, Count: 10925, OPS: 1727.9, Avg(us): 51694, Min(us): 41248, Max(us): 74623, 50th(us): 50783, 90th(us): 56351, 95th(us): 58111, 99th(us): 62335, 99.9th(us): 68671, 99.99th(us): 73343
TXN_ERROR - Takes(s): 6.3, Count: 5715, OPS: 901.4, Avg(us): 39930, Min(us): 23088, Max(us): 63327, 50th(us): 39839, 90th(us): 46335, 95th(us): 48991, 99th(us): 53855, 99.9th(us): 59007, 99.99th(us): 61823
TxnGroup - Takes(s): 6.4, Count: 16640, OPS: 2618.1, Avg(us): 47466, Min(us): 21152, Max(us): 76927, 50th(us): 48959, 90th(us): 56671, 95th(us): 58783, 99th(us): 63807, 99.9th(us): 70015, 99.99th(us): 74239
UPDATE - Takes(s): 6.4, Count: 94233, OPS: 14784.6, Avg(us): 6, Min(us): 1, Max(us): 1768, 50th(us): 4, 90th(us): 6, 95th(us): 8, 99th(us): 22, 99.9th(us): 726, 99.99th(us): 1386
Error Summary:

                              Operation:  COMMIT
                                   Error   Count
                                   -----   -----
  prepare phase failed: version mismatch    5715

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   3370
rollForward failed
  version mismatch  1719
rollback failed
  version mismatch  678
```

##### Oreo-P using remote

+ 8

```bash
----------------------------------
Run finished, takes 1m4.887989515s
COMMIT - Takes(s): 64.9, Count: 14460, OPS: 223.0, Avg(us): 7927, Min(us): 7064, Max(us): 11983, 50th(us): 7875, 90th(us): 8431, 95th(us): 8663, 99th(us): 9271, 99.9th(us): 10207, 99.99th(us): 11775
COMMIT_ERROR - Takes(s): 64.8, Count: 2204, OPS: 34.0, Avg(us): 4591, Min(us): 3870, Max(us): 6759, 50th(us): 4547, 90th(us): 5099, 95th(us): 5255, 99th(us): 5651, 99.9th(us): 6131, 99.99th(us): 6759
READ   - Takes(s): 64.9, Count: 99598, OPS: 1535.0, Avg(us): 3912, Min(us): 5, Max(us): 6303, 50th(us): 3885, 90th(us): 4387, 95th(us): 4523, 99th(us): 4767, 99.9th(us): 5243, 99.99th(us): 5747
READ_ERROR - Takes(s): 64.7, Count: 402, OPS: 6.2, Avg(us): 4399, Min(us): 3332, Max(us): 5691, 50th(us): 4359, 90th(us): 5063, 95th(us): 5203, 99th(us): 5455, 99.9th(us): 5691, 99.99th(us): 5691
Start  - Takes(s): 64.9, Count: 16672, OPS: 256.9, Avg(us): 22, Min(us): 13, Max(us): 915, 50th(us): 20, 90th(us): 29, 95th(us): 33, 99th(us): 48, 99.9th(us): 233, 99.99th(us): 443
TOTAL  - Takes(s): 64.9, Count: 261452, OPS: 4029.3, Avg(us): 5657, Min(us): 1, Max(us): 35999, 50th(us): 3561, 90th(us): 30447, 95th(us): 31775, 99th(us): 32703, 99.9th(us): 33567, 99.99th(us): 34559
TXN    - Takes(s): 64.9, Count: 14460, OPS: 223.0, Avg(us): 31539, Min(us): 22000, Max(us): 35999, 50th(us): 31631, 90th(us): 32639, 95th(us): 32927, 99th(us): 33503, 99.9th(us): 34623, 99.99th(us): 35935
TXN_ERROR - Takes(s): 64.8, Count: 2204, OPS: 34.0, Avg(us): 28194, Min(us): 19824, Max(us): 30639, 50th(us): 28351, 90th(us): 29359, 95th(us): 29583, 99th(us): 30031, 99.9th(us): 30607, 99.99th(us): 30639
TxnGroup - Takes(s): 64.9, Count: 16664, OPS: 256.9, Avg(us): 31094, Min(us): 22336, Max(us): 35199, 50th(us): 31551, 90th(us): 32623, 95th(us): 32895, 99th(us): 33471, 99.9th(us): 34335, 99.99th(us): 34847
UPDATE - Takes(s): 64.9, Count: 99598, OPS: 1535.0, Avg(us): 2, Min(us): 1, Max(us): 656, 50th(us): 2, 90th(us): 4, 95th(us): 4, 99th(us): 9, 99.9th(us): 122, 99.99th(us): 243
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2204

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  313
  read failed due to unknown txn status   60
rollback failed
  version mismatch  29
```

+ 16

```bash
----------------------------------
Run finished, takes 32.054359659s
COMMIT - Takes(s): 32.0, Count: 12997, OPS: 405.9, Avg(us): 8143, Min(us): 7028, Max(us): 39519, 50th(us): 8047, 90th(us): 8863, 95th(us): 9127, 99th(us): 9703, 99.9th(us): 11255, 99.99th(us): 31407
COMMIT_ERROR - Takes(s): 32.0, Count: 3659, OPS: 114.3, Avg(us): 4704, Min(us): 3888, Max(us): 21519, 50th(us): 4603, 90th(us): 5311, 95th(us): 5539, 99th(us): 6027, 99.9th(us): 7823, 99.99th(us): 21519
READ   - Takes(s): 32.1, Count: 99486, OPS: 3103.9, Avg(us): 3854, Min(us): 5, Max(us): 9167, 50th(us): 3777, 90th(us): 4379, 95th(us): 4563, 99th(us): 4951, 99.9th(us): 5651, 99.99th(us): 6339
READ_ERROR - Takes(s): 32.0, Count: 514, OPS: 16.1, Avg(us): 4478, Min(us): 3336, Max(us): 6543, 50th(us): 4435, 90th(us): 5167, 95th(us): 5351, 99th(us): 5891, 99.9th(us): 6171, 99.99th(us): 6543
Start  - Takes(s): 32.1, Count: 16672, OPS: 520.1, Avg(us): 24, Min(us): 13, Max(us): 523, 50th(us): 20, 90th(us): 32, 95th(us): 40, 99th(us): 80, 99.9th(us): 289, 99.99th(us): 463
TOTAL  - Takes(s): 32.1, Count: 258294, OPS: 8058.0, Avg(us): 5455, Min(us): 1, Max(us): 62207, 50th(us): 3511, 90th(us): 28959, 95th(us): 31487, 99th(us): 32831, 99.9th(us): 34047, 99.99th(us): 35903
TXN    - Takes(s): 32.0, Count: 12997, OPS: 405.9, Avg(us): 31414, Min(us): 23008, Max(us): 62207, 50th(us): 31407, 90th(us): 32799, 95th(us): 33215, 99th(us): 34015, 99.9th(us): 35903, 99.99th(us): 55231
TXN_ERROR - Takes(s): 32.0, Count: 3659, OPS: 114.3, Avg(us): 28023, Min(us): 19472, Max(us): 44863, 50th(us): 28063, 90th(us): 29423, 95th(us): 29855, 99th(us): 30703, 99.9th(us): 32623, 99.99th(us): 44863
TxnGroup - Takes(s): 32.0, Count: 16656, OPS: 520.0, Avg(us): 30663, Min(us): 19904, Max(us): 62175, 50th(us): 31183, 90th(us): 32703, 95th(us): 33119, 99th(us): 33951, 99.9th(us): 35551, 99.99th(us): 54111
UPDATE - Takes(s): 32.1, Count: 99486, OPS: 3104.0, Avg(us): 3, Min(us): 1, Max(us): 659, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 11, 99.9th(us): 149, 99.99th(us): 327
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  331
  read failed due to unknown txn status  122
rollback failed
  version mismatch  61

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3659
```

+ 32

```bash
----------------------------------
Run finished, takes 15.660225335s
COMMIT - Takes(s): 15.6, Count: 11381, OPS: 728.3, Avg(us): 8216, Min(us): 7072, Max(us): 42239, 50th(us): 8019, 90th(us): 9247, 95th(us): 9695, 99th(us): 10607, 99.9th(us): 13863, 99.99th(us): 22399
COMMIT_ERROR - Takes(s): 15.6, Count: 5259, OPS: 336.5, Avg(us): 4769, Min(us): 3766, Max(us): 15495, 50th(us): 4591, 90th(us): 5571, 95th(us): 5991, 99th(us): 7159, 99.9th(us): 11543, 99.99th(us): 13487
READ   - Takes(s): 15.7, Count: 99461, OPS: 6353.0, Avg(us): 3769, Min(us): 5, Max(us): 13231, 50th(us): 3671, 90th(us): 4267, 95th(us): 4535, 99th(us): 5199, 99.9th(us): 6291, 99.99th(us): 7487
READ_ERROR - Takes(s): 15.6, Count: 539, OPS: 34.5, Avg(us): 4606, Min(us): 3260, Max(us): 8215, 50th(us): 4427, 90th(us): 5823, 95th(us): 6183, 99th(us): 7255, 99.9th(us): 8111, 99.99th(us): 8215
Start  - Takes(s): 15.7, Count: 16672, OPS: 1064.5, Avg(us): 24, Min(us): 13, Max(us): 1141, 50th(us): 21, 90th(us): 33, 95th(us): 40, 99th(us): 62, 99.9th(us): 272, 99.99th(us): 517
TOTAL  - Takes(s): 15.7, Count: 254996, OPS: 16282.2, Avg(us): 5173, Min(us): 1, Max(us): 66687, 50th(us): 3439, 90th(us): 27167, 95th(us): 30623, 99th(us): 32895, 99.9th(us): 35167, 99.99th(us): 38431
TXN    - Takes(s): 15.6, Count: 11381, OPS: 728.4, Avg(us): 30986, Min(us): 22848, Max(us): 63999, 50th(us): 30799, 90th(us): 32991, 95th(us): 33695, 99th(us): 35199, 99.9th(us): 38591, 99.99th(us): 46207
TXN_ERROR - Takes(s): 15.6, Count: 5259, OPS: 336.5, Avg(us): 27574, Min(us): 18240, Max(us): 40255, 50th(us): 27423, 90th(us): 29535, 95th(us): 30239, 99th(us): 32015, 99.9th(us): 35551, 99.99th(us): 37951
TxnGroup - Takes(s): 15.6, Count: 16640, OPS: 1064.0, Avg(us): 29893, Min(us): 19424, Max(us): 66687, 50th(us): 30079, 90th(us): 32623, 95th(us): 33375, 99th(us): 34911, 99.9th(us): 37311, 99.99th(us): 45567
UPDATE - Takes(s): 15.7, Count: 99461, OPS: 6353.0, Avg(us): 3, Min(us): 1, Max(us): 1574, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 11, 99.9th(us): 120, 99.99th(us): 427
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  5259

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  243
  read failed due to unknown txn status  210
rollback failed
  version mismatch  86
```

+ 64

```bash
----------------------------------
Run finished, takes 9.063932198s
COMMIT - Takes(s): 9.0, Count: 9836, OPS: 1089.6, Avg(us): 10006, Min(us): 7212, Max(us): 58783, 50th(us): 9703, 90th(us): 11951, 95th(us): 12743, 99th(us): 14991, 99.9th(us): 51327, 99.99th(us): 58655
COMMIT_ERROR - Takes(s): 9.0, Count: 6804, OPS: 753.1, Avg(us): 5949, Min(us): 3940, Max(us): 51519, 50th(us): 5679, 90th(us): 7603, 95th(us): 8239, 99th(us): 9855, 99.9th(us): 14007, 99.99th(us): 49023
READ   - Takes(s): 9.1, Count: 99064, OPS: 10935.5, Avg(us): 4313, Min(us): 5, Max(us): 51679, 50th(us): 3981, 90th(us): 5503, 95th(us): 6167, 99th(us): 7767, 99.9th(us): 10895, 99.99th(us): 49535
READ_ERROR - Takes(s): 9.0, Count: 936, OPS: 103.8, Avg(us): 6498, Min(us): 3292, Max(us): 50815, 50th(us): 6027, 90th(us): 9111, 95th(us): 10071, 99th(us): 13367, 99.9th(us): 49183, 99.99th(us): 50815
Start  - Takes(s): 9.1, Count: 16704, OPS: 1842.7, Avg(us): 30, Min(us): 14, Max(us): 1668, 50th(us): 26, 90th(us): 39, 95th(us): 45, 99th(us): 217, 99.9th(us): 708, 99.99th(us): 1578
TOTAL  - Takes(s): 9.1, Count: 251144, OPS: 27706.9, Avg(us): 5795, Min(us): 1, Max(us): 95231, 50th(us): 3579, 90th(us): 29791, 95th(us): 35007, 99th(us): 39167, 99.9th(us): 46047, 99.99th(us): 85247
TXN    - Takes(s): 9.0, Count: 9836, OPS: 1089.5, Avg(us): 36090, Min(us): 26624, Max(us): 91071, 50th(us): 35519, 90th(us): 39583, 95th(us): 41247, 99th(us): 46079, 99.9th(us): 85695, 99.99th(us): 88255
TXN_ERROR - Takes(s): 9.0, Count: 6804, OPS: 753.2, Avg(us): 32259, Min(us): 22032, Max(us): 85951, 50th(us): 31775, 90th(us): 35711, 95th(us): 37247, 99th(us): 41855, 99.9th(us): 81471, 99.99th(us): 85311
TxnGroup - Takes(s): 9.0, Count: 16640, OPS: 1840.6, Avg(us): 34494, Min(us): 21808, Max(us): 95231, 50th(us): 34239, 90th(us): 38623, 95th(us): 40191, 99th(us): 44575, 99.9th(us): 84735, 99.99th(us): 91455
UPDATE - Takes(s): 9.1, Count: 99064, OPS: 10935.3, Avg(us): 3, Min(us): 1, Max(us): 1023, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 14, 99.9th(us): 205, 99.99th(us): 647
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  6804

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  381
  read failed due to unknown txn status  309
rollback failed
  version mismatch  245
     key not found    1
```

+ 96

```bash
----------------------------------
Run finished, takes 7.850504267s
COMMIT - Takes(s): 7.8, Count: 9323, OPS: 1192.1, Avg(us): 13382, Min(us): 7172, Max(us): 38879, 50th(us): 13247, 90th(us): 15887, 95th(us): 16735, 99th(us): 19119, 99.9th(us): 28687, 99.99th(us): 34783
COMMIT_ERROR - Takes(s): 7.8, Count: 7285, OPS: 931.7, Avg(us): 7613, Min(us): 4108, Max(us): 25807, 50th(us): 7395, 90th(us): 9567, 95th(us): 10375, 99th(us): 12455, 99.9th(us): 22687, 99.99th(us): 25775
READ   - Takes(s): 7.8, Count: 98270, OPS: 12524.5, Avg(us): 5483, Min(us): 5, Max(us): 25423, 50th(us): 5031, 90th(us): 7835, 95th(us): 8719, 99th(us): 11551, 99.9th(us): 16255, 99.99th(us): 19519
READ_ERROR - Takes(s): 7.8, Count: 1730, OPS: 221.8, Avg(us): 9974, Min(us): 3554, Max(us): 22431, 50th(us): 9375, 90th(us): 14383, 95th(us): 15527, 99th(us): 18207, 99.9th(us): 21967, 99.99th(us): 22431
Start  - Takes(s): 7.9, Count: 16704, OPS: 2127.8, Avg(us): 37, Min(us): 13, Max(us): 2223, 50th(us): 27, 90th(us): 41, 95th(us): 48, 99th(us): 385, 99.9th(us): 1155, 99.99th(us): 1840
TOTAL  - Takes(s): 7.9, Count: 248498, OPS: 31651.3, Avg(us): 7391, Min(us): 1, Max(us): 88127, 50th(us): 3701, 90th(us): 35583, 95th(us): 45311, 99th(us): 52831, 99.9th(us): 60191, 99.99th(us): 67775
TXN    - Takes(s): 7.8, Count: 9323, OPS: 1191.9, Avg(us): 46590, Min(us): 29584, Max(us): 69055, 50th(us): 46239, 90th(us): 53471, 95th(us): 55775, 99th(us): 60671, 99.9th(us): 65919, 99.99th(us): 68799
TXN_ERROR - Takes(s): 7.8, Count: 7285, OPS: 931.8, Avg(us): 41758, Min(us): 24528, Max(us): 67263, 50th(us): 41439, 90th(us): 48479, 95th(us): 50687, 99th(us): 55455, 99.9th(us): 62015, 99.99th(us): 66431
TxnGroup - Takes(s): 7.8, Count: 16608, OPS: 2121.2, Avg(us): 44425, Min(us): 19584, Max(us): 88127, 50th(us): 44223, 90th(us): 52095, 95th(us): 54783, 99th(us): 59711, 99.9th(us): 68479, 99.99th(us): 76031
UPDATE - Takes(s): 7.8, Count: 98270, OPS: 12526.8, Avg(us): 4, Min(us): 1, Max(us): 1883, 50th(us): 3, 90th(us): 5, 95th(us): 5, 99th(us): 16, 99.9th(us): 290, 99.99th(us): 1259
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  7285

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  861
rollback failed
                       version mismatch  507
  read failed due to unknown txn status  358
                          key not found    4
```

+ 128

```bash
----------------------------------
Run finished, takes 8.391969946s
COMMIT - Takes(s): 8.3, Count: 9364, OPS: 1127.8, Avg(us): 19403, Min(us): 7352, Max(us): 102079, 50th(us): 19135, 90th(us): 22719, 95th(us): 23903, 99th(us): 26639, 99.9th(us): 84671, 99.99th(us): 92159
COMMIT_ERROR - Takes(s): 8.3, Count: 7276, OPS: 871.4, Avg(us): 11414, Min(us): 4058, Max(us): 77823, 50th(us): 10927, 90th(us): 14287, 95th(us): 15287, 99th(us): 18223, 99.9th(us): 59167, 99.99th(us): 76479
READ   - Takes(s): 8.4, Count: 97180, OPS: 11593.9, Avg(us): 7542, Min(us): 6, Max(us): 76095, 50th(us): 7159, 90th(us): 12095, 95th(us): 13527, 99th(us): 20895, 99.9th(us): 31743, 99.99th(us): 74687
READ_ERROR - Takes(s): 8.3, Count: 2820, OPS: 339.9, Avg(us): 17515, Min(us): 4080, Max(us): 58687, 50th(us): 16015, 90th(us): 27119, 95th(us): 29695, 99th(us): 33695, 99.9th(us): 46399, 99.99th(us): 58687
Start  - Takes(s): 8.4, Count: 16768, OPS: 1998.0, Avg(us): 41, Min(us): 13, Max(us): 9687, 50th(us): 28, 90th(us): 41, 95th(us): 48, 99th(us): 294, 99.9th(us): 1778, 99.99th(us): 9655
TOTAL  - Takes(s): 8.4, Count: 246496, OPS: 29368.3, Avg(us): 10456, Min(us): 1, Max(us): 181119, 50th(us): 3733, 90th(us): 46175, 95th(us): 63903, 99th(us): 79423, 99.9th(us): 100095, 99.99th(us): 135295
TXN    - Takes(s): 8.3, Count: 9364, OPS: 1127.7, Avg(us): 65381, Min(us): 29488, Max(us): 143871, 50th(us): 64639, 90th(us): 79743, 95th(us): 86015, 99th(us): 100095, 99.9th(us): 127295, 99.99th(us): 137087
TXN_ERROR - Takes(s): 8.3, Count: 7276, OPS: 871.4, Avg(us): 60268, Min(us): 25712, Max(us): 140543, 50th(us): 59231, 90th(us): 75263, 95th(us): 81087, 99th(us): 97087, 99.9th(us): 124415, 99.99th(us): 131583
TxnGroup - Takes(s): 8.4, Count: 16640, OPS: 1990.4, Avg(us): 63064, Min(us): 22160, Max(us): 181119, 50th(us): 62239, 90th(us): 78335, 95th(us): 84735, 99th(us): 99071, 99.9th(us): 140671, 99.99th(us): 169215
UPDATE - Takes(s): 8.4, Count: 97180, OPS: 11596.1, Avg(us): 4, Min(us): 1, Max(us): 2541, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 16, 99.9th(us): 226, 99.99th(us): 1116
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  7276

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  1386
rollback failed
                       version mismatch  904
  read failed due to unknown txn status  506
                          key not found   24
```



##### Oreo-AC using remote

+ 8

```bash
----------------------------------
Run finished, takes 1m5.039795818s
COMMIT - Takes(s): 65.0, Count: 14362, OPS: 220.9, Avg(us): 7958, Min(us): 7056, Max(us): 52287, 50th(us): 7903, 90th(us): 8447, 95th(us): 8645, 99th(us): 9287, 99.9th(us): 10151, 99.99th(us): 49919
COMMIT_ERROR - Takes(s): 64.9, Count: 2304, OPS: 35.5, Avg(us): 4608, Min(us): 3880, Max(us): 6387, 50th(us): 4571, 90th(us): 5099, 95th(us): 5271, 99th(us): 5635, 99.9th(us): 5967, 99.99th(us): 6387
READ   - Takes(s): 65.0, Count: 99628, OPS: 1531.9, Avg(us): 3926, Min(us): 5, Max(us): 41823, 50th(us): 3895, 90th(us): 4403, 95th(us): 4539, 99th(us): 4767, 99.9th(us): 5219, 99.99th(us): 5755
READ_ERROR - Takes(s): 65.0, Count: 372, OPS: 5.7, Avg(us): 4515, Min(us): 3498, Max(us): 6523, 50th(us): 4479, 90th(us): 5135, 95th(us): 5327, 99th(us): 5571, 99.9th(us): 6523, 99.99th(us): 6523
Start  - Takes(s): 65.0, Count: 16672, OPS: 256.3, Avg(us): 22, Min(us): 13, Max(us): 982, 50th(us): 20, 90th(us): 29, 95th(us): 33, 99th(us): 52, 99.9th(us): 234, 99.99th(us): 363
TOTAL  - Takes(s): 65.0, Count: 261312, OPS: 4017.7, Avg(us): 5664, Min(us): 1, Max(us): 77119, 50th(us): 3573, 90th(us): 30575, 95th(us): 31855, 99th(us): 32799, 99.9th(us): 33599, 99.99th(us): 35551
TXN    - Takes(s): 65.0, Count: 14360, OPS: 220.9, Avg(us): 31650, Min(us): 20368, Max(us): 76927, 50th(us): 31727, 90th(us): 32719, 95th(us): 33023, 99th(us): 33535, 99.9th(us): 34943, 99.99th(us): 74111
TXN_ERROR - Takes(s): 64.9, Count: 2304, OPS: 35.5, Avg(us): 28315, Min(us): 18656, Max(us): 30911, 50th(us): 28447, 90th(us): 29423, 95th(us): 29695, 99th(us): 30111, 99.9th(us): 30799, 99.99th(us): 30911
TxnGroup - Takes(s): 65.0, Count: 16664, OPS: 256.3, Avg(us): 31186, Min(us): 20272, Max(us): 77119, 50th(us): 31647, 90th(us): 32687, 95th(us): 32991, 99th(us): 33567, 99.9th(us): 35103, 99.99th(us): 73535
UPDATE - Takes(s): 65.0, Count: 99628, OPS: 1531.9, Avg(us): 2, Min(us): 1, Max(us): 404, 50th(us): 2, 90th(us): 4, 95th(us): 4, 99th(us): 9, 99.9th(us): 120, 99.99th(us): 254
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  341
rollback failed
  version mismatch  31

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2256
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  46
```

+ 16

```bash
----------------------------------
Run finished, takes 32.127303222s
COMMIT - Takes(s): 32.1, Count: 12876, OPS: 401.2, Avg(us): 8163, Min(us): 7108, Max(us): 53247, 50th(us): 8063, 90th(us): 8831, 95th(us): 9111, 99th(us): 9695, 99.9th(us): 10895, 99.99th(us): 52927
COMMIT_ERROR - Takes(s): 32.1, Count: 3780, OPS: 117.8, Avg(us): 4706, Min(us): 3850, Max(us): 6923, 50th(us): 4639, 90th(us): 5279, 95th(us): 5531, 99th(us): 6075, 99.9th(us): 6547, 99.99th(us): 6923
READ   - Takes(s): 32.1, Count: 99602, OPS: 3100.6, Avg(us): 3863, Min(us): 5, Max(us): 49087, 50th(us): 3785, 90th(us): 4383, 95th(us): 4567, 99th(us): 4935, 99.9th(us): 5579, 99.99th(us): 6519
READ_ERROR - Takes(s): 32.1, Count: 398, OPS: 12.4, Avg(us): 4611, Min(us): 3546, Max(us): 6259, 50th(us): 4551, 90th(us): 5343, 95th(us): 5471, 99th(us): 5923, 99.9th(us): 6259, 99.99th(us): 6259
Start  - Takes(s): 32.1, Count: 16672, OPS: 518.9, Avg(us): 24, Min(us): 13, Max(us): 697, 50th(us): 20, 90th(us): 32, 95th(us): 40, 99th(us): 66, 99.9th(us): 300, 99.99th(us): 542
TOTAL  - Takes(s): 32.1, Count: 258284, OPS: 8039.3, Avg(us): 5450, Min(us): 1, Max(us): 78591, 50th(us): 3523, 90th(us): 28735, 95th(us): 31519, 99th(us): 32831, 99.9th(us): 34015, 99.99th(us): 71615
TXN    - Takes(s): 32.1, Count: 12876, OPS: 401.2, Avg(us): 31505, Min(us): 22048, Max(us): 78591, 50th(us): 31455, 90th(us): 32831, 95th(us): 33215, 99th(us): 33951, 99.9th(us): 68927, 99.99th(us): 78015
TXN_ERROR - Takes(s): 32.1, Count: 3780, OPS: 117.8, Avg(us): 28035, Min(us): 19872, Max(us): 73407, 50th(us): 28063, 90th(us): 29391, 95th(us): 29759, 99th(us): 30591, 99.9th(us): 31599, 99.99th(us): 73407
TxnGroup - Takes(s): 32.1, Count: 16656, OPS: 518.8, Avg(us): 30711, Min(us): 19792, Max(us): 78207, 50th(us): 31231, 90th(us): 32703, 95th(us): 33119, 99th(us): 33887, 99.9th(us): 36415, 99.99th(us): 77119
UPDATE - Takes(s): 32.1, Count: 99602, OPS: 3100.5, Avg(us): 3, Min(us): 1, Max(us): 711, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 12, 99.9th(us): 149, 99.99th(us): 346
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  337
rollback failed
  version mismatch  61

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3655
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  95
prepare phase failed: Remote prepare failed
  validation failed due to false assumption  30
```

+ 32

```bash
----------------------------------
Run finished, takes 15.605052837s
COMMIT - Takes(s): 15.6, Count: 11167, OPS: 717.2, Avg(us): 8236, Min(us): 7072, Max(us): 25439, 50th(us): 8039, 90th(us): 9255, 95th(us): 9743, 99th(us): 11023, 99.9th(us): 14431, 99.99th(us): 15623
COMMIT_ERROR - Takes(s): 15.6, Count: 5473, OPS: 351.5, Avg(us): 4774, Min(us): 3690, Max(us): 11359, 50th(us): 4583, 90th(us): 5587, 95th(us): 6031, 99th(us): 7195, 99.9th(us): 9767, 99.99th(us): 11247
READ   - Takes(s): 15.6, Count: 99682, OPS: 6389.3, Avg(us): 3758, Min(us): 5, Max(us): 19871, 50th(us): 3673, 90th(us): 4215, 95th(us): 4479, 99th(us): 5123, 99.9th(us): 6475, 99.99th(us): 8831
READ_ERROR - Takes(s): 15.5, Count: 318, OPS: 20.5, Avg(us): 4879, Min(us): 3498, Max(us): 9727, 50th(us): 4707, 90th(us): 5975, 95th(us): 6311, 99th(us): 7147, 99.9th(us): 9727, 99.99th(us): 9727
Start  - Takes(s): 15.6, Count: 16672, OPS: 1068.2, Avg(us): 25, Min(us): 13, Max(us): 571, 50th(us): 20, 90th(us): 34, 95th(us): 42, 99th(us): 78, 99.9th(us): 299, 99.99th(us): 512
TOTAL  - Takes(s): 15.6, Count: 255010, OPS: 16339.3, Avg(us): 5131, Min(us): 1, Max(us): 48991, 50th(us): 3447, 90th(us): 26975, 95th(us): 30511, 99th(us): 32671, 99.9th(us): 35199, 99.99th(us): 38527
TXN    - Takes(s): 15.6, Count: 11167, OPS: 717.2, Avg(us): 30929, Min(us): 22080, Max(us): 48991, 50th(us): 30703, 90th(us): 32799, 95th(us): 33503, 99th(us): 35327, 99.9th(us): 38815, 99.99th(us): 43519
TXN_ERROR - Takes(s): 15.6, Count: 5473, OPS: 351.5, Avg(us): 27518, Min(us): 21664, Max(us): 42911, 50th(us): 27359, 90th(us): 29343, 95th(us): 30079, 99th(us): 31567, 99.9th(us): 35487, 99.99th(us): 38911
TxnGroup - Takes(s): 15.6, Count: 16640, OPS: 1067.8, Avg(us): 29795, Min(us): 19376, Max(us): 48991, 50th(us): 30047, 90th(us): 32399, 95th(us): 33119, 99th(us): 34943, 99.9th(us): 37791, 99.99th(us): 44159
UPDATE - Takes(s): 15.6, Count: 99682, OPS: 6388.8, Avg(us): 3, Min(us): 1, Max(us): 778, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 11, 99.9th(us): 110, 99.99th(us): 295
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  5293
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  129
prepare phase failed: Remote prepare failed
  validation failed due to false assumption  51

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  228
rollback failed
  version mismatch  90
```

+ 64

```bash
----------------------------------
Run finished, takes 8.722508744s
COMMIT - Takes(s): 8.7, Count: 9569, OPS: 1102.0, Avg(us): 9494, Min(us): 7208, Max(us): 19887, 50th(us): 9263, 90th(us): 11151, 95th(us): 11911, 99th(us): 13695, 99.9th(us): 17183, 99.99th(us): 19663
COMMIT_ERROR - Takes(s): 8.7, Count: 7071, OPS: 813.4, Avg(us): 5664, Min(us): 3822, Max(us): 14735, 50th(us): 5403, 90th(us): 7127, 95th(us): 7891, 99th(us): 9679, 99.9th(us): 12887, 99.99th(us): 14255
READ   - Takes(s): 8.7, Count: 99419, OPS: 11404.5, Avg(us): 4175, Min(us): 6, Max(us): 12967, 50th(us): 3931, 90th(us): 5147, 95th(us): 5699, 99th(us): 6987, 99.9th(us): 8903, 99.99th(us): 10999
READ_ERROR - Takes(s): 8.7, Count: 581, OPS: 67.0, Avg(us): 6332, Min(us): 3510, Max(us): 11607, 50th(us): 6003, 90th(us): 8839, 95th(us): 9631, 99th(us): 10767, 99.9th(us): 11511, 99.99th(us): 11607
Start  - Takes(s): 8.7, Count: 16704, OPS: 1914.8, Avg(us): 29, Min(us): 14, Max(us): 1010, 50th(us): 26, 90th(us): 39, 95th(us): 47, 99th(us): 153, 99.9th(us): 609, 99.99th(us): 892
TOTAL  - Takes(s): 8.7, Count: 251320, OPS: 28809.0, Avg(us): 5532, Min(us): 1, Max(us): 52383, 50th(us): 3565, 90th(us): 28847, 95th(us): 33855, 99th(us): 37119, 99.9th(us): 41279, 99.99th(us): 45599
TXN    - Takes(s): 8.7, Count: 9569, OPS: 1101.9, Avg(us): 34711, Min(us): 24448, Max(us): 49759, 50th(us): 34463, 90th(us): 37471, 95th(us): 38591, 99th(us): 41407, 99.9th(us): 45535, 99.99th(us): 48447
TXN_ERROR - Takes(s): 8.7, Count: 7071, OPS: 813.4, Avg(us): 31070, Min(us): 19504, Max(us): 42655, 50th(us): 30847, 90th(us): 33919, 95th(us): 35071, 99th(us): 37695, 99.9th(us): 41023, 99.99th(us): 42271
TxnGroup - Takes(s): 8.7, Count: 16640, OPS: 1912.4, Avg(us): 33136, Min(us): 21184, Max(us): 52383, 50th(us): 33151, 90th(us): 36735, 95th(us): 37887, 99th(us): 41055, 99.9th(us): 45535, 99.99th(us): 49407
UPDATE - Takes(s): 8.7, Count: 99419, OPS: 11405.0, Avg(us): 3, Min(us): 1, Max(us): 1579, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 14, 99.9th(us): 194, 99.99th(us): 935
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  6835
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  156
prepare phase failed: Remote prepare failed
  validation failed due to false assumption  80

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  358
rollback failed
  version mismatch  223
```

+ 96

```bash
----------------------------------
Run finished, takes 7.997812976s
COMMIT - Takes(s): 8.0, Count: 9141, OPS: 1148.8, Avg(us): 13768, Min(us): 7288, Max(us): 42847, 50th(us): 13607, 90th(us): 16463, 95th(us): 17423, 99th(us): 19727, 99.9th(us): 31039, 99.99th(us): 41535
COMMIT_ERROR - Takes(s): 8.0, Count: 7467, OPS: 937.9, Avg(us): 7897, Min(us): 4060, Max(us): 20815, 50th(us): 7687, 90th(us): 10079, 95th(us): 10903, 99th(us): 13351, 99.9th(us): 17679, 99.99th(us): 20447
READ   - Takes(s): 8.0, Count: 98644, OPS: 12340.3, Avg(us): 5601, Min(us): 5, Max(us): 22319, 50th(us): 5119, 90th(us): 8139, 95th(us): 9055, 99th(us): 11663, 99.9th(us): 16143, 99.99th(us): 19327
READ_ERROR - Takes(s): 7.9, Count: 1356, OPS: 170.8, Avg(us): 11141, Min(us): 4032, Max(us): 23071, 50th(us): 10719, 90th(us): 15743, 95th(us): 17055, 99th(us): 19711, 99.9th(us): 21727, 99.99th(us): 23071
Start  - Takes(s): 8.0, Count: 16704, OPS: 2088.3, Avg(us): 36, Min(us): 14, Max(us): 2159, 50th(us): 27, 90th(us): 42, 95th(us): 51, 99th(us): 335, 99.9th(us): 1258, 99.99th(us): 2109
TOTAL  - Takes(s): 8.0, Count: 248882, OPS: 31118.1, Avg(us): 7507, Min(us): 1, Max(us): 80127, 50th(us): 3723, 90th(us): 35679, 95th(us): 46271, 99th(us): 54015, 99.9th(us): 61759, 99.99th(us): 68735
TXN    - Takes(s): 8.0, Count: 9141, OPS: 1148.8, Avg(us): 47590, Min(us): 28928, Max(us): 73599, 50th(us): 47231, 90th(us): 54655, 95th(us): 56991, 99th(us): 62303, 99.9th(us): 69247, 99.99th(us): 73535
TXN_ERROR - Takes(s): 8.0, Count: 7467, OPS: 937.9, Avg(us): 42818, Min(us): 25808, Max(us): 68479, 50th(us): 42527, 90th(us): 49887, 95th(us): 52415, 99th(us): 57407, 99.9th(us): 63615, 99.99th(us): 67263
TxnGroup - Takes(s): 8.0, Count: 16608, OPS: 2082.3, Avg(us): 45399, Min(us): 22368, Max(us): 80127, 50th(us): 45215, 90th(us): 53375, 95th(us): 56031, 99th(us): 61151, 99.9th(us): 67903, 99.99th(us): 73535
UPDATE - Takes(s): 8.0, Count: 98644, OPS: 12340.9, Avg(us): 4, Min(us): 1, Max(us): 2291, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 19, 99.9th(us): 257, 99.99th(us): 1319
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  7127
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  215
prepare phase failed: Remote prepare failed
  validation failed due to false assumption  125

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  817
rollback failed
  version mismatch  534
     key not found    5
```

+ 128

```bash
----------------------------------
Run finished, takes 8.531482361s
COMMIT - Takes(s): 8.5, Count: 9089, OPS: 1069.7, Avg(us): 19928, Min(us): 7252, Max(us): 77887, 50th(us): 19711, 90th(us): 23567, 95th(us): 24863, 99th(us): 29215, 99.9th(us): 70975, 99.99th(us): 75903
COMMIT_ERROR - Takes(s): 8.5, Count: 7551, OPS: 888.7, Avg(us): 11653, Min(us): 4124, Max(us): 65311, 50th(us): 11383, 90th(us): 14599, 95th(us): 15631, 99th(us): 18479, 99.9th(us): 35295, 99.99th(us): 63263
READ   - Takes(s): 8.5, Count: 97811, OPS: 11470.4, Avg(us): 7729, Min(us): 5, Max(us): 87871, 50th(us): 7399, 90th(us): 12431, 95th(us): 13911, 99th(us): 20047, 99.9th(us): 49151, 99.99th(us): 62879
READ_ERROR - Takes(s): 8.5, Count: 2189, OPS: 258.8, Avg(us): 19196, Min(us): 4476, Max(us): 69439, 50th(us): 18703, 90th(us): 28991, 95th(us): 30991, 99th(us): 36031, 99.9th(us): 63647, 99.99th(us): 69439
Start  - Takes(s): 8.5, Count: 16768, OPS: 1965.2, Avg(us): 39, Min(us): 14, Max(us): 3075, 50th(us): 28, 90th(us): 43, 95th(us): 53, 99th(us): 338, 99.9th(us): 1737, 99.99th(us): 2883
TOTAL  - Takes(s): 8.5, Count: 247208, OPS: 28974.3, Avg(us): 10570, Min(us): 1, Max(us): 184831, 50th(us): 3747, 90th(us): 45663, 95th(us): 64831, 99th(us): 80639, 99.9th(us): 106495, 99.99th(us): 144639
TXN    - Takes(s): 8.5, Count: 9089, OPS: 1070.0, Avg(us): 66579, Min(us): 25520, Max(us): 161791, 50th(us): 65663, 90th(us): 81087, 95th(us): 86911, 99th(us): 104703, 99.9th(us): 133247, 99.99th(us): 158463
TXN_ERROR - Takes(s): 8.5, Count: 7551, OPS: 888.6, Avg(us): 61673, Min(us): 25744, Max(us): 156543, 50th(us): 60799, 90th(us): 76351, 95th(us): 82239, 99th(us): 101759, 99.9th(us): 132479, 99.99th(us): 154879
TxnGroup - Takes(s): 8.5, Count: 16640, OPS: 1955.6, Avg(us): 64278, Min(us): 23024, Max(us): 184831, 50th(us): 63391, 90th(us): 79871, 95th(us): 85951, 99th(us): 105407, 99.9th(us): 149119, 99.99th(us): 170111
UPDATE - Takes(s): 8.5, Count: 97811, OPS: 11470.4, Avg(us): 4, Min(us): 1, Max(us): 2891, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 21, 99.9th(us): 297, 99.99th(us): 1605
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  7129
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  307
prepare phase failed: Remote prepare failed
  validation failed due to false assumption  115

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  1301
rollback failed
  version mismatch  870
     key not found   18
```

##### Oreo-AA using remote

+ 8

```bash
----------------------------------
Run finished, takes 1m4.936738454s
COMMIT - Takes(s): 64.9, Count: 14326, OPS: 220.8, Avg(us): 7922, Min(us): 6996, Max(us): 50655, 50th(us): 7867, 90th(us): 8391, 95th(us): 8591, 99th(us): 9231, 99.9th(us): 10151, 99.99th(us): 50303
COMMIT_ERROR - Takes(s): 64.9, Count: 2338, OPS: 36.0, Avg(us): 4584, Min(us): 3796, Max(us): 6635, 50th(us): 4539, 90th(us): 5043, 95th(us): 5223, 99th(us): 5627, 99.9th(us): 6375, 99.99th(us): 6635
READ   - Takes(s): 64.9, Count: 99673, OPS: 1535.0, Avg(us): 3916, Min(us): 5, Max(us): 46943, 50th(us): 3889, 90th(us): 4387, 95th(us): 4523, 99th(us): 4759, 99.9th(us): 5215, 99.99th(us): 5783
READ_ERROR - Takes(s): 64.5, Count: 327, OPS: 5.1, Avg(us): 4482, Min(us): 3570, Max(us): 5715, 50th(us): 4459, 90th(us): 5075, 95th(us): 5223, 99th(us): 5451, 99.9th(us): 5715, 99.99th(us): 5715
Start  - Takes(s): 64.9, Count: 16672, OPS: 256.7, Avg(us): 22, Min(us): 13, Max(us): 749, 50th(us): 20, 90th(us): 29, 95th(us): 33, 99th(us): 48, 99.9th(us): 222, 99.99th(us): 342
TOTAL  - Takes(s): 64.9, Count: 261340, OPS: 4024.5, Avg(us): 5642, Min(us): 1, Max(us): 76671, 50th(us): 3563, 90th(us): 30431, 95th(us): 31759, 99th(us): 32671, 99.9th(us): 33535, 99.99th(us): 35199
TXN    - Takes(s): 64.9, Count: 14329, OPS: 220.8, Avg(us): 31555, Min(us): 21952, Max(us): 76671, 50th(us): 31631, 90th(us): 32607, 95th(us): 32895, 99th(us): 33471, 99.9th(us): 34751, 99.99th(us): 74879
TXN_ERROR - Takes(s): 64.9, Count: 2335, OPS: 36.0, Avg(us): 28212, Min(us): 21072, Max(us): 72575, 50th(us): 28367, 90th(us): 29327, 95th(us): 29535, 99th(us): 30111, 99.9th(us): 31359, 99.99th(us): 72575
TxnGroup - Takes(s): 64.9, Count: 16664, OPS: 256.7, Avg(us): 31083, Min(us): 19472, Max(us): 75263, 50th(us): 31551, 90th(us): 32575, 95th(us): 32863, 99th(us): 33439, 99.9th(us): 34847, 99.99th(us): 74879
UPDATE - Takes(s): 64.9, Count: 99673, OPS: 1535.0, Avg(us): 2, Min(us): 1, Max(us): 375, 50th(us): 2, 90th(us): 4, 95th(us): 4, 99th(us): 9, 99.9th(us): 76, 99.99th(us): 233
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  300
rollback failed
  version mismatch  27

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2288
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  27
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong state  23
```

+ 16

```bash
----------------------------------
Run finished, takes 32.069492843s
COMMIT - Takes(s): 32.0, Count: 12848, OPS: 401.1, Avg(us): 8169, Min(us): 7052, Max(us): 52191, 50th(us): 8051, 90th(us): 8871, 95th(us): 9151, 99th(us): 9799, 99.9th(us): 12207, 99.99th(us): 51327
COMMIT_ERROR - Takes(s): 32.0, Count: 3808, OPS: 118.9, Avg(us): 4704, Min(us): 3894, Max(us): 43871, 50th(us): 4607, 90th(us): 5279, 95th(us): 5531, 99th(us): 6071, 99.9th(us): 8135, 99.99th(us): 43871
READ   - Takes(s): 32.1, Count: 99600, OPS: 3106.1, Avg(us): 3856, Min(us): 5, Max(us): 45151, 50th(us): 3775, 90th(us): 4383, 95th(us): 4563, 99th(us): 4967, 99.9th(us): 5683, 99.99th(us): 6607
READ_ERROR - Takes(s): 32.0, Count: 400, OPS: 12.5, Avg(us): 4619, Min(us): 3520, Max(us): 6635, 50th(us): 4539, 90th(us): 5355, 95th(us): 5659, 99th(us): 5983, 99.9th(us): 6635, 99.99th(us): 6635
Start  - Takes(s): 32.1, Count: 16672, OPS: 519.9, Avg(us): 24, Min(us): 13, Max(us): 630, 50th(us): 20, 90th(us): 31, 95th(us): 39, 99th(us): 67, 99.9th(us): 290, 99.99th(us): 546
TOTAL  - Takes(s): 32.1, Count: 258224, OPS: 8051.9, Avg(us): 5439, Min(us): 1, Max(us): 76095, 50th(us): 3511, 90th(us): 28655, 95th(us): 31455, 99th(us): 32831, 99.9th(us): 34239, 99.99th(us): 68543
TXN    - Takes(s): 32.0, Count: 12848, OPS: 401.0, Avg(us): 31448, Min(us): 22320, Max(us): 76031, 50th(us): 31407, 90th(us): 32799, 95th(us): 33247, 99th(us): 34079, 99.9th(us): 46559, 99.99th(us): 74495
TXN_ERROR - Takes(s): 32.0, Count: 3808, OPS: 118.9, Avg(us): 28043, Min(us): 20288, Max(us): 67647, 50th(us): 28031, 90th(us): 29407, 95th(us): 29807, 99th(us): 30591, 99.9th(us): 37887, 99.99th(us): 67647
TxnGroup - Takes(s): 32.0, Count: 16656, OPS: 519.8, Avg(us): 30663, Min(us): 19408, Max(us): 76095, 50th(us): 31167, 90th(us): 32687, 95th(us): 33151, 99th(us): 34111, 99.9th(us): 40767, 99.99th(us): 73407
UPDATE - Takes(s): 32.1, Count: 99600, OPS: 3106.0, Avg(us): 3, Min(us): 1, Max(us): 545, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 11, 99.9th(us): 141, 99.99th(us): 334
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3685
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  75
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong state  46
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong txnId  2

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  322
rollback failed
  version mismatch  78
```

+ 32

```bash
----------------------------------
Run finished, takes 15.537523326s
COMMIT - Takes(s): 15.5, Count: 11074, OPS: 714.4, Avg(us): 8136, Min(us): 7072, Max(us): 16511, 50th(us): 7955, 90th(us): 9079, 95th(us): 9487, 99th(us): 10455, 99.9th(us): 13823, 99.99th(us): 15791
COMMIT_ERROR - Takes(s): 15.5, Count: 5566, OPS: 358.9, Avg(us): 4697, Min(us): 3830, Max(us): 13159, 50th(us): 4539, 90th(us): 5431, 95th(us): 5787, 99th(us): 6731, 99.9th(us): 9207, 99.99th(us): 10959
READ   - Takes(s): 15.5, Count: 99699, OPS: 6418.4, Avg(us): 3754, Min(us): 5, Max(us): 17743, 50th(us): 3667, 90th(us): 4223, 95th(us): 4471, 99th(us): 5079, 99.9th(us): 6187, 99.99th(us): 9399
READ_ERROR - Takes(s): 15.5, Count: 301, OPS: 19.4, Avg(us): 4783, Min(us): 3528, Max(us): 12559, 50th(us): 4607, 90th(us): 5763, 95th(us): 6183, 99th(us): 7783, 99.9th(us): 12559, 99.99th(us): 12559
Start  - Takes(s): 15.5, Count: 16672, OPS: 1072.9, Avg(us): 25, Min(us): 13, Max(us): 681, 50th(us): 21, 90th(us): 34, 95th(us): 41, 99th(us): 79, 99.9th(us): 378, 99.99th(us): 656
TOTAL  - Takes(s): 15.5, Count: 254858, OPS: 16402.3, Avg(us): 5100, Min(us): 1, Max(us): 45215, 50th(us): 3437, 90th(us): 26799, 95th(us): 30431, 99th(us): 32511, 99.9th(us): 34623, 99.99th(us): 38847
TXN    - Takes(s): 15.5, Count: 11074, OPS: 714.4, Avg(us): 30808, Min(us): 21920, Max(us): 44927, 50th(us): 30671, 90th(us): 32607, 95th(us): 33311, 99th(us): 34719, 99.9th(us): 38943, 99.99th(us): 40479
TXN_ERROR - Takes(s): 15.5, Count: 5566, OPS: 358.9, Avg(us): 27400, Min(us): 19248, Max(us): 36895, 50th(us): 27295, 90th(us): 29199, 95th(us): 29807, 99th(us): 31167, 99.9th(us): 34815, 99.99th(us): 35647
TxnGroup - Takes(s): 15.5, Count: 16640, OPS: 1072.6, Avg(us): 29655, Min(us): 18912, Max(us): 45215, 50th(us): 29887, 90th(us): 32271, 95th(us): 32927, 99th(us): 34335, 99.9th(us): 38207, 99.99th(us): 41951
UPDATE - Takes(s): 15.5, Count: 99699, OPS: 6418.0, Avg(us): 3, Min(us): 1, Max(us): 1105, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 11, 99.9th(us): 132, 99.99th(us): 288
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  5388
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  116
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong state  59
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong txnId  3

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  212
rollback failed
  version mismatch  89
```

+ 64

```bash
----------------------------------
Run finished, takes 8.659833279s
COMMIT - Takes(s): 8.6, Count: 9632, OPS: 1117.0, Avg(us): 9448, Min(us): 7176, Max(us): 28479, 50th(us): 9191, 90th(us): 11127, 95th(us): 11799, 99th(us): 13679, 99.9th(us): 17743, 99.99th(us): 19599
COMMIT_ERROR - Takes(s): 8.6, Count: 7008, OPS: 811.8, Avg(us): 5575, Min(us): 3786, Max(us): 35711, 50th(us): 5323, 90th(us): 6983, 95th(us): 7643, 99th(us): 9239, 99.9th(us): 14767, 99.99th(us): 16071
READ   - Takes(s): 8.7, Count: 99434, OPS: 11487.4, Avg(us): 4154, Min(us): 5, Max(us): 14351, 50th(us): 3903, 90th(us): 5135, 95th(us): 5683, 99th(us): 7043, 99.9th(us): 9295, 99.99th(us): 11615
READ_ERROR - Takes(s): 8.6, Count: 566, OPS: 65.8, Avg(us): 6265, Min(us): 3722, Max(us): 12775, 50th(us): 5871, 90th(us): 8703, 95th(us): 9823, 99th(us): 12143, 99.9th(us): 12663, 99.99th(us): 12775
Start  - Takes(s): 8.7, Count: 16704, OPS: 1928.7, Avg(us): 30, Min(us): 13, Max(us): 2673, 50th(us): 26, 90th(us): 39, 95th(us): 44, 99th(us): 198, 99.9th(us): 760, 99.99th(us): 1174
TOTAL  - Takes(s): 8.7, Count: 251476, OPS: 29038.4, Avg(us): 5512, Min(us): 1, Max(us): 58559, 50th(us): 3549, 90th(us): 28687, 95th(us): 33727, 99th(us): 37023, 99.9th(us): 40927, 99.99th(us): 45375
TXN    - Takes(s): 8.6, Count: 9632, OPS: 1117.0, Avg(us): 34542, Min(us): 26112, Max(us): 51423, 50th(us): 34271, 90th(us): 37407, 95th(us): 38527, 99th(us): 41407, 99.9th(us): 45087, 99.99th(us): 48063
TXN_ERROR - Takes(s): 8.6, Count: 7008, OPS: 811.9, Avg(us): 30867, Min(us): 19520, Max(us): 59039, 50th(us): 30623, 90th(us): 33695, 95th(us): 34815, 99th(us): 37727, 99.9th(us): 41887, 99.99th(us): 43007
TxnGroup - Takes(s): 8.6, Count: 16640, OPS: 1926.6, Avg(us): 32966, Min(us): 20944, Max(us): 58559, 50th(us): 33023, 90th(us): 36639, 95th(us): 37823, 99th(us): 40543, 99.9th(us): 45343, 99.99th(us): 52319
UPDATE - Takes(s): 8.7, Count: 99434, OPS: 11487.9, Avg(us): 3, Min(us): 1, Max(us): 1586, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 14, 99.9th(us): 210, 99.99th(us): 720
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  351
rollback failed
  version mismatch  215

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  6731
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  175
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong state  89
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong txnId  13
```

+ 96

```bash
----------------------------------
Run finished, takes 7.650768989s
COMMIT - Takes(s): 7.6, Count: 9077, OPS: 1192.8, Avg(us): 13114, Min(us): 7232, Max(us): 38879, 50th(us): 12959, 90th(us): 15655, 95th(us): 16511, 99th(us): 18655, 99.9th(us): 33215, 99.99th(us): 38815
COMMIT_ERROR - Takes(s): 7.6, Count: 7531, OPS: 988.6, Avg(us): 7535, Min(us): 3996, Max(us): 22911, 50th(us): 7279, 90th(us): 9639, 95th(us): 10479, 99th(us): 12791, 99.9th(us): 18191, 99.99th(us): 22815
READ   - Takes(s): 7.6, Count: 98722, OPS: 12911.8, Avg(us): 5403, Min(us): 5, Max(us): 22047, 50th(us): 4943, 90th(us): 7711, 95th(us): 8551, 99th(us): 11239, 99.9th(us): 15167, 99.99th(us): 17935
READ_ERROR - Takes(s): 7.6, Count: 1278, OPS: 168.1, Avg(us): 10192, Min(us): 4038, Max(us): 22655, 50th(us): 9807, 90th(us): 14159, 95th(us): 15495, 99th(us): 17759, 99.9th(us): 21727, 99.99th(us): 22655
Start  - Takes(s): 7.7, Count: 16704, OPS: 2182.9, Avg(us): 34, Min(us): 13, Max(us): 2025, 50th(us): 27, 90th(us): 40, 95th(us): 46, 99th(us): 257, 99.9th(us): 1211, 99.99th(us): 1992
TOTAL  - Takes(s): 7.7, Count: 248910, OPS: 32529.6, Avg(us): 7201, Min(us): 1, Max(us): 79807, 50th(us): 3691, 90th(us): 34367, 95th(us): 44383, 99th(us): 51615, 99.9th(us): 58687, 99.99th(us): 65407
TXN    - Takes(s): 7.6, Count: 9077, OPS: 1192.6, Avg(us): 45742, Min(us): 28880, Max(us): 67839, 50th(us): 45503, 90th(us): 52255, 95th(us): 54559, 99th(us): 59039, 99.9th(us): 64511, 99.99th(us): 67583
TXN_ERROR - Takes(s): 7.6, Count: 7531, OPS: 988.5, Avg(us): 41078, Min(us): 24544, Max(us): 64927, 50th(us): 40799, 90th(us): 47583, 95th(us): 49759, 99th(us): 54111, 99.9th(us): 58655, 99.99th(us): 64735
TxnGroup - Takes(s): 7.6, Count: 16608, OPS: 2177.3, Avg(us): 43582, Min(us): 24176, Max(us): 79807, 50th(us): 43327, 90th(us): 50975, 95th(us): 53247, 99th(us): 58239, 99.9th(us): 65791, 99.99th(us): 73599
UPDATE - Takes(s): 7.6, Count: 98722, OPS: 12910.2, Avg(us): 4, Min(us): 1, Max(us): 2069, 50th(us): 3, 90th(us): 5, 95th(us): 5, 99th(us): 16, 99.9th(us): 278, 99.99th(us): 1267
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  7217
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong state  133
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  130
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong txnId  51

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  769
rollback failed
  version mismatch  508
     key not found    1
```

+ 128

```bash
----------------------------------
Run finished, takes 7.723607254s
COMMIT - Takes(s): 7.7, Count: 8912, OPS: 1156.7, Avg(us): 17902, Min(us): 7088, Max(us): 45727, 50th(us): 17759, 90th(us): 21119, 95th(us): 22143, 99th(us): 24911, 99.9th(us): 37471, 99.99th(us): 43135
COMMIT_ERROR - Takes(s): 7.7, Count: 7728, OPS: 1007.8, Avg(us): 10558, Min(us): 4058, Max(us): 28559, 50th(us): 10287, 90th(us): 13399, 95th(us): 14503, 99th(us): 17391, 99.9th(us): 23247, 99.99th(us): 27279
READ   - Takes(s): 7.7, Count: 97820, OPS: 12669.6, Avg(us): 7065, Min(us): 5, Max(us): 34399, 50th(us): 6727, 90th(us): 10951, 95th(us): 12247, 99th(us): 16767, 99.9th(us): 24223, 99.99th(us): 29503
READ_ERROR - Takes(s): 7.7, Count: 2180, OPS: 283.9, Avg(us): 16604, Min(us): 4124, Max(us): 40383, 50th(us): 16191, 90th(us): 24095, 95th(us): 26319, 99th(us): 29679, 99.9th(us): 33343, 99.99th(us): 40383
Start  - Takes(s): 7.7, Count: 16768, OPS: 2170.8, Avg(us): 40, Min(us): 12, Max(us): 2771, 50th(us): 28, 90th(us): 43, 95th(us): 53, 99th(us): 424, 99.9th(us): 1600, 99.99th(us): 2463
TOTAL  - Takes(s): 7.7, Count: 246832, OPS: 31956.3, Avg(us): 9569, Min(us): 1, Max(us): 113471, 50th(us): 3757, 90th(us): 41727, 95th(us): 59263, 99th(us): 71999, 99.9th(us): 85375, 99.99th(us): 97151
TXN    - Takes(s): 7.7, Count: 8892, OPS: 1156.6, Avg(us): 60704, Min(us): 28736, Max(us): 113471, 50th(us): 60319, 90th(us): 72511, 95th(us): 76607, 99th(us): 85375, 99.9th(us): 95551, 99.99th(us): 106111
TXN_ERROR - Takes(s): 7.7, Count: 7728, OPS: 1007.9, Avg(us): 55786, Min(us): 25792, Max(us): 105919, 50th(us): 55231, 90th(us): 67903, 95th(us): 72127, 99th(us): 80767, 99.9th(us): 92351, 99.99th(us): 104383
TxnGroup - Takes(s): 7.7, Count: 16640, OPS: 2160.1, Avg(us): 58338, Min(us): 21152, Max(us): 107647, 50th(us): 58015, 90th(us): 71231, 95th(us): 75519, 99th(us): 84735, 99.9th(us): 98303, 99.99th(us): 105599
UPDATE - Takes(s): 7.7, Count: 97820, OPS: 12669.5, Avg(us): 5, Min(us): 1, Max(us): 2307, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 22, 99.9th(us): 349, 99.99th(us): 1919
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  1285
rollback failed
  version mismatch  882
     key not found   13

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  7321
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong state  245
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong txnId  128
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  34
```





#### Read:RMW = 1:1

##### Cherry Garcia

> We can directly use Workload F ‘s result

+ 8

```bash

```

+ 16

```bash

```

+ 32

```

```

+ 64

```bash

```

+ 96

```bash

```

+ 128

```bash

```

##### Oreo-P using remote

+ 8

```bash
----------------------------------
Run finished, takes 1m4.948373825s
COMMIT - Takes(s): 64.9, Count: 15945, OPS: 245.6, Avg(us): 7464, Min(us): 0, Max(us): 49663, 50th(us): 7571, 90th(us): 7919, 95th(us): 8051, 99th(us): 8367, 99.9th(us): 9191, 99.99th(us): 9831
COMMIT_ERROR - Takes(s): 64.9, Count: 719, OPS: 11.1, Avg(us): 4283, Min(us): 3682, Max(us): 5555, 50th(us): 4263, 90th(us): 4603, 95th(us): 4763, 99th(us): 5063, 99.9th(us): 5235, 99.99th(us): 5555
READ   - Takes(s): 64.9, Count: 99759, OPS: 1536.1, Avg(us): 3953, Min(us): 5, Max(us): 47103, 50th(us): 3945, 90th(us): 4359, 95th(us): 4475, 99th(us): 4691, 99.9th(us): 5031, 99.99th(us): 5475
READ_ERROR - Takes(s): 63.5, Count: 241, OPS: 3.8, Avg(us): 4342, Min(us): 3374, Max(us): 5531, 50th(us): 4359, 90th(us): 4787, 95th(us): 4947, 99th(us): 5071, 99.9th(us): 5531, 99.99th(us): 5531
Start  - Takes(s): 64.9, Count: 16672, OPS: 256.7, Avg(us): 22, Min(us): 13, Max(us): 349, 50th(us): 19, 90th(us): 28, 95th(us): 31, 99th(us): 44, 99.9th(us): 236, 99.99th(us): 335
TOTAL  - Takes(s): 64.9, Count: 214772, OPS: 3306.8, Avg(us): 7129, Min(us): 0, Max(us): 74815, 50th(us): 3875, 90th(us): 31231, 95th(us): 31711, 99th(us): 32271, 99.9th(us): 33087, 99.99th(us): 38911
TXN    - Takes(s): 64.9, Count: 15945, OPS: 245.6, Avg(us): 31271, Min(us): 19136, Max(us): 74815, 50th(us): 31471, 90th(us): 32127, 95th(us): 32319, 99th(us): 32895, 99.9th(us): 34111, 99.99th(us): 74111
TXN_ERROR - Takes(s): 64.9, Count: 719, OPS: 11.1, Avg(us): 28065, Min(us): 22384, Max(us): 70911, 50th(us): 28175, 90th(us): 28895, 95th(us): 29119, 99th(us): 29615, 99.9th(us): 30607, 99.99th(us): 70911
TxnGroup - Takes(s): 64.9, Count: 16664, OPS: 256.6, Avg(us): 31130, Min(us): 19056, Max(us): 74751, 50th(us): 31471, 90th(us): 32143, 95th(us): 32351, 99th(us): 32895, 99.9th(us): 34143, 99.99th(us): 74303
UPDATE - Takes(s): 64.9, Count: 49787, OPS: 766.6, Avg(us): 3, Min(us): 1, Max(us): 340, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 10, 99.9th(us): 74, 99.99th(us): 230
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  719

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  223
  read failed due to unknown txn status   11
rollback failed
  version mismatch  7
```

+ 16

```bash
----------------------------------
Run finished, takes 32.160088487s
COMMIT - Takes(s): 32.1, Count: 15305, OPS: 476.4, Avg(us): 7515, Min(us): 0, Max(us): 49343, 50th(us): 7599, 90th(us): 8131, 95th(us): 8287, 99th(us): 8615, 99.9th(us): 9535, 99.99th(us): 49055
COMMIT_ERROR - Takes(s): 32.1, Count: 1351, OPS: 42.1, Avg(us): 4352, Min(us): 3478, Max(us): 45471, 50th(us): 4263, 90th(us): 4835, 95th(us): 5023, 99th(us): 5319, 99.9th(us): 6111, 99.99th(us): 45471
READ   - Takes(s): 32.2, Count: 99662, OPS: 3099.3, Avg(us): 3907, Min(us): 5, Max(us): 45951, 50th(us): 3833, 90th(us): 4455, 95th(us): 4615, 99th(us): 4927, 99.9th(us): 5391, 99.99th(us): 6423
READ_ERROR - Takes(s): 31.8, Count: 338, OPS: 10.6, Avg(us): 4397, Min(us): 3374, Max(us): 5595, 50th(us): 4411, 90th(us): 5007, 95th(us): 5199, 99th(us): 5455, 99.9th(us): 5595, 99.99th(us): 5595
Start  - Takes(s): 32.2, Count: 16672, OPS: 518.4, Avg(us): 24, Min(us): 13, Max(us): 546, 50th(us): 21, 90th(us): 31, 95th(us): 38, 99th(us): 66, 99.9th(us): 297, 99.99th(us): 475
TOTAL  - Takes(s): 32.2, Count: 213239, OPS: 6630.5, Avg(us): 7004, Min(us): 0, Max(us): 74495, 50th(us): 3731, 90th(us): 30719, 95th(us): 31631, 99th(us): 32543, 99.9th(us): 33407, 99.99th(us): 68863
TXN    - Takes(s): 32.1, Count: 15305, OPS: 476.4, Avg(us): 31069, Min(us): 18480, Max(us): 73663, 50th(us): 31231, 90th(us): 32351, 95th(us): 32639, 99th(us): 33183, 99.9th(us): 36287, 99.99th(us): 73343
TXN_ERROR - Takes(s): 32.1, Count: 1351, OPS: 42.1, Avg(us): 27832, Min(us): 22080, Max(us): 70975, 50th(us): 27903, 90th(us): 29023, 95th(us): 29279, 99th(us): 29855, 99.9th(us): 70271, 99.99th(us): 70975
TxnGroup - Takes(s): 32.1, Count: 16656, OPS: 518.3, Avg(us): 30800, Min(us): 19040, Max(us): 74495, 50th(us): 31167, 90th(us): 32367, 95th(us): 32655, 99th(us): 33279, 99.9th(us): 34655, 99.99th(us): 73855
UPDATE - Takes(s): 32.2, Count: 49639, OPS: 1543.7, Avg(us): 3, Min(us): 1, Max(us): 1270, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 12, 99.9th(us): 167, 99.99th(us): 343
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  1351

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  278
  read failed due to unknown txn status   40
rollback failed
  version mismatch  20
```

+ 32

```bash
----------------------------------
Run finished, takes 15.942663231s
COMMIT - Takes(s): 15.9, Count: 14616, OPS: 918.0, Avg(us): 7723, Min(us): 0, Max(us): 38271, 50th(us): 7723, 90th(us): 8503, 95th(us): 8775, 99th(us): 9511, 99.9th(us): 26447, 99.99th(us): 34431
COMMIT_ERROR - Takes(s): 15.9, Count: 2024, OPS: 127.2, Avg(us): 4444, Min(us): 3454, Max(us): 13975, 50th(us): 4351, 90th(us): 5043, 95th(us): 5251, 99th(us): 5827, 99.9th(us): 12175, 99.99th(us): 13975
READ   - Takes(s): 15.9, Count: 99591, OPS: 6248.1, Avg(us): 3842, Min(us): 5, Max(us): 38591, 50th(us): 3739, 90th(us): 4395, 95th(us): 4627, 99th(us): 5115, 99.9th(us): 5755, 99.99th(us): 6399
READ_ERROR - Takes(s): 15.8, Count: 409, OPS: 25.8, Avg(us): 4618, Min(us): 3324, Max(us): 17599, 50th(us): 4463, 90th(us): 5343, 95th(us): 5583, 99th(us): 8623, 99.9th(us): 17599, 99.99th(us): 17599
Start  - Takes(s): 15.9, Count: 16672, OPS: 1045.7, Avg(us): 27, Min(us): 13, Max(us): 1233, 50th(us): 24, 90th(us): 36, 95th(us): 42, 99th(us): 87, 99.9th(us): 545, 99.99th(us): 1039
TOTAL  - Takes(s): 15.9, Count: 212178, OPS: 13308.6, Avg(us): 6858, Min(us): 0, Max(us): 66175, 50th(us): 3667, 90th(us): 30303, 95th(us): 31439, 99th(us): 32687, 99.9th(us): 34175, 99.99th(us): 52895
TXN    - Takes(s): 15.9, Count: 14616, OPS: 918.0, Avg(us): 30903, Min(us): 19888, Max(us): 65343, 50th(us): 30991, 90th(us): 32495, 95th(us): 32895, 99th(us): 33919, 99.9th(us): 50047, 99.99th(us): 62335
TXN_ERROR - Takes(s): 15.9, Count: 2024, OPS: 127.2, Avg(us): 27619, Min(us): 18592, Max(us): 36127, 50th(us): 27727, 90th(us): 29167, 95th(us): 29487, 99th(us): 30607, 99.9th(us): 35135, 99.99th(us): 36127
TxnGroup - Takes(s): 15.9, Count: 16640, OPS: 1045.1, Avg(us): 30490, Min(us): 18048, Max(us): 66175, 50th(us): 30863, 90th(us): 32399, 95th(us): 32831, 99th(us): 33695, 99.9th(us): 48831, 99.99th(us): 60607
UPDATE - Takes(s): 15.9, Count: 50043, OPS: 3140.0, Avg(us): 3, Min(us): 1, Max(us): 1084, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 14, 99.9th(us): 168, 99.99th(us): 638
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2024

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  287
  read failed due to unknown txn status  102
rollback failed
  version mismatch  20
```

+ 64

```bash
----------------------------------
Run finished, takes 7.629737313s
COMMIT - Takes(s): 7.6, Count: 13651, OPS: 1794.8, Avg(us): 7439, Min(us): 0, Max(us): 16047, 50th(us): 7467, 90th(us): 8223, 95th(us): 8543, 99th(us): 9279, 99.9th(us): 11407, 99.99th(us): 15943
COMMIT_ERROR - Takes(s): 7.6, Count: 2989, OPS: 393.2, Avg(us): 4352, Min(us): 3406, Max(us): 8895, 50th(us): 4239, 90th(us): 4927, 95th(us): 5235, 99th(us): 5991, 99.9th(us): 8383, 99.99th(us): 8895
READ   - Takes(s): 7.6, Count: 99618, OPS: 13064.1, Avg(us): 3677, Min(us): 5, Max(us): 13087, 50th(us): 3617, 90th(us): 4049, 95th(us): 4275, 99th(us): 4831, 99.9th(us): 5747, 99.99th(us): 6759
READ_ERROR - Takes(s): 7.6, Count: 382, OPS: 50.3, Avg(us): 4112, Min(us): 3266, Max(us): 6851, 50th(us): 4005, 90th(us): 4895, 95th(us): 5331, 99th(us): 5855, 99.9th(us): 6851, 99.99th(us): 6851
Start  - Takes(s): 7.6, Count: 16704, OPS: 2189.3, Avg(us): 28, Min(us): 13, Max(us): 1174, 50th(us): 25, 90th(us): 38, 95th(us): 43, 99th(us): 177, 99.9th(us): 654, 99.99th(us): 1113
TOTAL  - Takes(s): 7.6, Count: 210075, OPS: 27531.6, Avg(us): 6454, Min(us): 0, Max(us): 40671, 50th(us): 3571, 90th(us): 29071, 95th(us): 29887, 99th(us): 31183, 99.9th(us): 32959, 99.99th(us): 36927
TXN    - Takes(s): 7.6, Count: 13651, OPS: 1794.6, Avg(us): 29617, Min(us): 17760, Max(us): 40671, 50th(us): 29615, 90th(us): 31007, 95th(us): 31535, 99th(us): 32895, 99.9th(us): 37791, 99.99th(us): 40383
TXN_ERROR - Takes(s): 7.6, Count: 2989, OPS: 393.2, Avg(us): 26515, Min(us): 18928, Max(us): 33663, 50th(us): 26463, 90th(us): 27871, 95th(us): 28351, 99th(us): 29599, 99.9th(us): 32751, 99.99th(us): 33663
TxnGroup - Takes(s): 7.6, Count: 16640, OPS: 2187.6, Avg(us): 29035, Min(us): 18672, Max(us): 39711, 50th(us): 29423, 90th(us): 30831, 95th(us): 31343, 99th(us): 32479, 99.9th(us): 34463, 99.99th(us): 39263
UPDATE - Takes(s): 7.6, Count: 49811, OPS: 6532.1, Avg(us): 3, Min(us): 1, Max(us): 1968, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 13, 99.9th(us): 136, 99.99th(us): 467
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2989

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status    190
rollForward failed
  version mismatch  164
rollback failed
  version mismatch  28
```

+ 96

```bash
----------------------------------
Run finished, takes 5.687046914s
COMMIT - Takes(s): 5.7, Count: 13039, OPS: 2304.5, Avg(us): 8449, Min(us): 0, Max(us): 27471, 50th(us): 8407, 90th(us): 9903, 95th(us): 10479, 99th(us): 11943, 99.9th(us): 17807, 99.99th(us): 27439
COMMIT_ERROR - Takes(s): 5.7, Count: 3569, OPS: 631.3, Avg(us): 5137, Min(us): 3468, Max(us): 13463, 50th(us): 4907, 90th(us): 6407, 95th(us): 7027, 99th(us): 8239, 99.9th(us): 12615, 99.99th(us): 13463
READ   - Takes(s): 5.7, Count: 99385, OPS: 17492.7, Avg(us): 4076, Min(us): 5, Max(us): 14687, 50th(us): 3883, 90th(us): 4919, 95th(us): 5363, 99th(us): 6395, 99.9th(us): 7803, 99.99th(us): 9639
READ_ERROR - Takes(s): 5.6, Count: 615, OPS: 109.0, Avg(us): 5120, Min(us): 3304, Max(us): 10463, 50th(us): 4871, 90th(us): 6715, 95th(us): 7511, 99th(us): 8775, 99.9th(us): 9695, 99.99th(us): 10463
Start  - Takes(s): 5.7, Count: 16704, OPS: 2937.3, Avg(us): 40, Min(us): 13, Max(us): 2143, 50th(us): 28, 90th(us): 41, 95th(us): 50, 99th(us): 508, 99.9th(us): 1293, 99.99th(us): 1943
TOTAL  - Takes(s): 5.7, Count: 208310, OPS: 36622.7, Avg(us): 7126, Min(us): 0, Max(us): 57567, 50th(us): 3757, 90th(us): 31711, 95th(us): 33471, 99th(us): 36031, 99.9th(us): 39391, 99.99th(us): 46527
TXN    - Takes(s): 5.7, Count: 13039, OPS: 2304.3, Avg(us): 33080, Min(us): 20640, Max(us): 54143, 50th(us): 32991, 90th(us): 35647, 95th(us): 36703, 99th(us): 39135, 99.9th(us): 45311, 99.99th(us): 53599
TXN_ERROR - Takes(s): 5.7, Count: 3569, OPS: 631.3, Avg(us): 29804, Min(us): 20144, Max(us): 41983, 50th(us): 29647, 90th(us): 32335, 95th(us): 33343, 99th(us): 35775, 99.9th(us): 40671, 99.99th(us): 41983
TxnGroup - Takes(s): 5.7, Count: 16608, OPS: 2932.8, Avg(us): 32335, Min(us): 18864, Max(us): 57567, 50th(us): 32479, 90th(us): 35359, 95th(us): 36415, 99th(us): 38687, 99.9th(us): 46079, 99.99th(us): 54111
UPDATE - Takes(s): 5.7, Count: 49535, OPS: 8718.6, Avg(us): 4, Min(us): 1, Max(us): 1208, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 17, 99.9th(us): 342, 99.99th(us): 1065
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3569

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  297
  read failed due to unknown txn status  260
rollback failed
  version mismatch  58
```

+ 128

```bash
----------------------------------
Run finished, takes 5.04283073s
COMMIT - Takes(s): 5.0, Count: 12837, OPS: 2558.9, Avg(us): 10252, Min(us): 0, Max(us): 46431, 50th(us): 10159, 90th(us): 12623, 95th(us): 13479, 99th(us): 15847, 99.9th(us): 24287, 99.99th(us): 37791
COMMIT_ERROR - Takes(s): 5.0, Count: 3803, OPS: 759.4, Avg(us): 6176, Min(us): 3662, Max(us): 34463, 50th(us): 5895, 90th(us): 7979, 95th(us): 8831, 99th(us): 10791, 99.9th(us): 13519, 99.99th(us): 34463
READ   - Takes(s): 5.0, Count: 98990, OPS: 19646.6, Avg(us): 4755, Min(us): 5, Max(us): 21007, 50th(us): 4427, 90th(us): 6339, 95th(us): 6999, 99th(us): 8495, 99.9th(us): 11023, 99.99th(us): 13463
READ_ERROR - Takes(s): 5.0, Count: 1010, OPS: 201.6, Avg(us): 7121, Min(us): 3330, Max(us): 12871, 50th(us): 6867, 90th(us): 9895, 95th(us): 10671, 99th(us): 11999, 99.9th(us): 12775, 99.99th(us): 12871
Start  - Takes(s): 5.0, Count: 16768, OPS: 3325.1, Avg(us): 55, Min(us): 13, Max(us): 3493, 50th(us): 29, 90th(us): 48, 95th(us): 63, 99th(us): 1004, 99.9th(us): 2257, 99.99th(us): 2853
TOTAL  - Takes(s): 5.0, Count: 207661, OPS: 41165.7, Avg(us): 8376, Min(us): 0, Max(us): 77887, 50th(us): 4111, 90th(us): 36639, 95th(us): 39807, 99th(us): 44095, 99.9th(us): 49727, 99.99th(us): 58463
TXN    - Takes(s): 5.0, Count: 12837, OPS: 2559.2, Avg(us): 39042, Min(us): 22848, Max(us): 70015, 50th(us): 38911, 90th(us): 43551, 95th(us): 45215, 99th(us): 49087, 99.9th(us): 58431, 99.99th(us): 66047
TXN_ERROR - Takes(s): 5.0, Count: 3803, OPS: 759.5, Avg(us): 35350, Min(us): 21904, Max(us): 59871, 50th(us): 35167, 90th(us): 39903, 95th(us): 41311, 99th(us): 44927, 99.9th(us): 47871, 99.99th(us): 59871
TxnGroup - Takes(s): 5.0, Count: 16640, OPS: 3317.1, Avg(us): 38142, Min(us): 20160, Max(us): 77887, 50th(us): 38111, 90th(us): 43103, 95th(us): 44607, 99th(us): 48479, 99.9th(us): 55839, 99.99th(us): 62463
UPDATE - Takes(s): 5.0, Count: 49589, OPS: 9844.6, Avg(us): 6, Min(us): 1, Max(us): 2555, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 24, 99.9th(us): 677, 99.99th(us): 2415
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  590
  read failed due to unknown txn status  294
rollback failed
  version mismatch  126

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3803
```

##### Oreo-AA using remote

+ 8

```bash
----------------------------------
Run finished, takes 1m5.005028941s
COMMIT - Takes(s): 65.0, Count: 15918, OPS: 245.0, Avg(us): 7480, Min(us): 0, Max(us): 42815, 50th(us): 7575, 90th(us): 7919, 95th(us): 8039, 99th(us): 8375, 99.9th(us): 9047, 99.99th(us): 9463
COMMIT_ERROR - Takes(s): 65.0, Count: 746, OPS: 11.5, Avg(us): 4277, Min(us): 3534, Max(us): 5667, 50th(us): 4263, 90th(us): 4615, 95th(us): 4735, 99th(us): 4991, 99.9th(us): 5267, 99.99th(us): 5667
READ   - Takes(s): 65.0, Count: 99783, OPS: 1535.1, Avg(us): 3958, Min(us): 5, Max(us): 43839, 50th(us): 3951, 90th(us): 4359, 95th(us): 4475, 99th(us): 4691, 99.9th(us): 5047, 99.99th(us): 5475
READ_ERROR - Takes(s): 65.0, Count: 217, OPS: 3.3, Avg(us): 4391, Min(us): 3432, Max(us): 5219, 50th(us): 4383, 90th(us): 4835, 95th(us): 4891, 99th(us): 5031, 99.9th(us): 5219, 99.99th(us): 5219
Start  - Takes(s): 65.0, Count: 16672, OPS: 256.5, Avg(us): 21, Min(us): 13, Max(us): 413, 50th(us): 19, 90th(us): 28, 95th(us): 31, 99th(us): 42, 99.9th(us): 223, 99.99th(us): 285
TOTAL  - Takes(s): 65.0, Count: 214842, OPS: 3304.9, Avg(us): 7133, Min(us): 0, Max(us): 71871, 50th(us): 3881, 90th(us): 31263, 95th(us): 31743, 99th(us): 32287, 99.9th(us): 33055, 99.99th(us): 35871
TXN    - Takes(s): 65.0, Count: 15918, OPS: 245.0, Avg(us): 31324, Min(us): 19680, Max(us): 71615, 50th(us): 31503, 90th(us): 32143, 95th(us): 32351, 99th(us): 32863, 99.9th(us): 34015, 99.99th(us): 71487
TXN_ERROR - Takes(s): 65.0, Count: 746, OPS: 11.5, Avg(us): 28008, Min(us): 20048, Max(us): 64447, 50th(us): 28175, 90th(us): 28879, 95th(us): 29119, 99th(us): 29503, 99.9th(us): 30383, 99.99th(us): 64447
TxnGroup - Takes(s): 65.0, Count: 16664, OPS: 256.4, Avg(us): 31172, Min(us): 19312, Max(us): 71871, 50th(us): 31487, 90th(us): 32175, 95th(us): 32367, 99th(us): 32863, 99.9th(us): 34175, 99.99th(us): 70975
UPDATE - Takes(s): 65.0, Count: 49887, OPS: 767.5, Avg(us): 3, Min(us): 1, Max(us): 254, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 10, 99.9th(us): 100, 99.99th(us): 220
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  736
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  6
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong state  4

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  206
rollback failed
  version mismatch  11
```

+ 16

```bash
----------------------------------
Run finished, takes 32.149441973s
COMMIT - Takes(s): 32.1, Count: 15330, OPS: 477.0, Avg(us): 7508, Min(us): 0, Max(us): 52735, 50th(us): 7591, 90th(us): 8087, 95th(us): 8247, 99th(us): 8559, 99.9th(us): 9231, 99.99th(us): 52159
COMMIT_ERROR - Takes(s): 32.1, Count: 1326, OPS: 41.7, Avg(us): 4322, Min(us): 3480, Max(us): 5775, 50th(us): 4275, 90th(us): 4815, 95th(us): 4979, 99th(us): 5251, 99.9th(us): 5759, 99.99th(us): 5775
READ   - Takes(s): 32.1, Count: 99704, OPS: 3101.6, Avg(us): 3901, Min(us): 5, Max(us): 49439, 50th(us): 3825, 90th(us): 4447, 95th(us): 4607, 99th(us): 4911, 99.9th(us): 5391, 99.99th(us): 6083
READ_ERROR - Takes(s): 32.1, Count: 296, OPS: 9.2, Avg(us): 4546, Min(us): 3496, Max(us): 5627, 50th(us): 4539, 90th(us): 5127, 95th(us): 5223, 99th(us): 5451, 99.9th(us): 5627, 99.99th(us): 5627
Start  - Takes(s): 32.2, Count: 16672, OPS: 518.6, Avg(us): 24, Min(us): 13, Max(us): 512, 50th(us): 20, 90th(us): 31, 95th(us): 38, 99th(us): 63, 99.9th(us): 300, 99.99th(us): 507
TOTAL  - Takes(s): 32.1, Count: 213622, OPS: 6644.7, Avg(us): 6985, Min(us): 0, Max(us): 76735, 50th(us): 3719, 90th(us): 30671, 95th(us): 31567, 99th(us): 32495, 99.9th(us): 33311, 99.99th(us): 72639
TXN    - Takes(s): 32.1, Count: 15320, OPS: 477.0, Avg(us): 31027, Min(us): 18368, Max(us): 76735, 50th(us): 31167, 90th(us): 32303, 95th(us): 32607, 99th(us): 33151, 99.9th(us): 35135, 99.99th(us): 76543
TXN_ERROR - Takes(s): 32.1, Count: 1336, OPS: 41.7, Avg(us): 27790, Min(us): 22448, Max(us): 73343, 50th(us): 27855, 90th(us): 28991, 95th(us): 29231, 99th(us): 29743, 99.9th(us): 69119, 99.99th(us): 73343
TxnGroup - Takes(s): 32.1, Count: 16656, OPS: 518.5, Avg(us): 30760, Min(us): 18976, Max(us): 76735, 50th(us): 31103, 90th(us): 32335, 95th(us): 32623, 99th(us): 33183, 99.9th(us): 34591, 99.99th(us): 76543
UPDATE - Takes(s): 32.1, Count: 49950, OPS: 1553.9, Avg(us): 3, Min(us): 1, Max(us): 602, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 13, 99.9th(us): 147, 99.99th(us): 391
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  278
rollback failed
  version mismatch  18

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  1287
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong state  17
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  22
```

+ 32

```bash
----------------------------------
Run finished, takes 15.83794386s
COMMIT - Takes(s): 15.8, Count: 14399, OPS: 910.7, Avg(us): 7615, Min(us): 0, Max(us): 10527, 50th(us): 7651, 90th(us): 8375, 95th(us): 8623, 99th(us): 9135, 99.9th(us): 9743, 99.99th(us): 10407
COMMIT_ERROR - Takes(s): 15.8, Count: 2241, OPS: 141.7, Avg(us): 4381, Min(us): 3430, Max(us): 6135, 50th(us): 4311, 90th(us): 4947, 95th(us): 5119, 99th(us): 5567, 99.9th(us): 5943, 99.99th(us): 6135
READ   - Takes(s): 15.8, Count: 99643, OPS: 6293.2, Avg(us): 3832, Min(us): 5, Max(us): 9471, 50th(us): 3733, 90th(us): 4375, 95th(us): 4603, 99th(us): 5067, 99.9th(us): 5663, 99.99th(us): 6283
READ_ERROR - Takes(s): 15.8, Count: 357, OPS: 22.6, Avg(us): 4571, Min(us): 3450, Max(us): 6187, 50th(us): 4483, 90th(us): 5267, 95th(us): 5487, 99th(us): 5867, 99.9th(us): 6187, 99.99th(us): 6187
Start  - Takes(s): 15.8, Count: 16672, OPS: 1052.7, Avg(us): 27, Min(us): 13, Max(us): 1230, 50th(us): 25, 90th(us): 36, 95th(us): 42, 99th(us): 88, 99.9th(us): 644, 99.99th(us): 987
TOTAL  - Takes(s): 15.8, Count: 211490, OPS: 13352.3, Avg(us): 6802, Min(us): 0, Max(us): 37279, 50th(us): 3663, 90th(us): 30207, 95th(us): 31311, 99th(us): 32383, 99.9th(us): 33439, 99.99th(us): 34783
TXN    - Takes(s): 15.8, Count: 14399, OPS: 910.7, Avg(us): 30742, Min(us): 18464, Max(us): 36959, 50th(us): 30911, 90th(us): 32223, 95th(us): 32543, 99th(us): 33279, 99.9th(us): 34431, 99.99th(us): 36479
TXN_ERROR - Takes(s): 15.8, Count: 2241, OPS: 141.7, Avg(us): 27423, Min(us): 19472, Max(us): 31519, 50th(us): 27567, 90th(us): 28895, 95th(us): 29199, 99th(us): 29839, 99.9th(us): 30799, 99.99th(us): 31519
TxnGroup - Takes(s): 15.8, Count: 16640, OPS: 1051.9, Avg(us): 30281, Min(us): 18512, Max(us): 37279, 50th(us): 30751, 90th(us): 32175, 95th(us): 32527, 99th(us): 33247, 99.9th(us): 34495, 99.99th(us): 35999
UPDATE - Takes(s): 15.8, Count: 49737, OPS: 3141.2, Avg(us): 3, Min(us): 1, Max(us): 1047, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 14, 99.9th(us): 222, 99.99th(us): 817
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2134
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  80
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong state  26
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong txnId  1

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  326
rollback failed
  version mismatch  31
```

+ 64

```bash
----------------------------------
Run finished, takes 7.65563012s
COMMIT - Takes(s): 7.6, Count: 13398, OPS: 1756.3, Avg(us): 7489, Min(us): 0, Max(us): 14551, 50th(us): 7487, 90th(us): 8295, 95th(us): 8695, 99th(us): 9975, 99.9th(us): 12303, 99.99th(us): 14551
COMMIT_ERROR - Takes(s): 7.6, Count: 3242, OPS: 425.2, Avg(us): 4342, Min(us): 3432, Max(us): 9087, 50th(us): 4231, 90th(us): 4903, 95th(us): 5203, 99th(us): 6143, 99.9th(us): 8543, 99.99th(us): 9087
READ   - Takes(s): 7.7, Count: 99768, OPS: 13040.4, Avg(us): 3695, Min(us): 5, Max(us): 11871, 50th(us): 3627, 90th(us): 4087, 95th(us): 4327, 99th(us): 4943, 99.9th(us): 5971, 99.99th(us): 7707
READ_ERROR - Takes(s): 7.6, Count: 232, OPS: 30.6, Avg(us): 4433, Min(us): 3470, Max(us): 8463, 50th(us): 4275, 90th(us): 5155, 95th(us): 5587, 99th(us): 6627, 99.9th(us): 8463, 99.99th(us): 8463
Start  - Takes(s): 7.7, Count: 16704, OPS: 2181.7, Avg(us): 29, Min(us): 13, Max(us): 1186, 50th(us): 25, 90th(us): 38, 95th(us): 43, 99th(us): 198, 99.9th(us): 758, 99.99th(us): 1151
TOTAL  - Takes(s): 7.7, Count: 209613, OPS: 27380.9, Avg(us): 6457, Min(us): 0, Max(us): 39295, 50th(us): 3579, 90th(us): 29135, 95th(us): 29983, 99th(us): 31503, 99.9th(us): 34303, 99.99th(us): 37919
TXN    - Takes(s): 7.6, Count: 13398, OPS: 1756.2, Avg(us): 29781, Min(us): 17552, Max(us): 39295, 50th(us): 29727, 90th(us): 31279, 95th(us): 31951, 99th(us): 34175, 99.9th(us): 38335, 99.99th(us): 39295
TXN_ERROR - Takes(s): 7.6, Count: 3242, OPS: 425.2, Avg(us): 26615, Min(us): 18672, Max(us): 35711, 50th(us): 26527, 90th(us): 28015, 95th(us): 28671, 99th(us): 30703, 99.9th(us): 34879, 99.99th(us): 35711
TxnGroup - Takes(s): 7.6, Count: 16640, OPS: 2179.6, Avg(us): 29138, Min(us): 18080, Max(us): 38847, 50th(us): 29503, 90th(us): 31055, 95th(us): 31727, 99th(us): 33695, 99.9th(us): 36063, 99.99th(us): 38079
UPDATE - Takes(s): 7.7, Count: 49705, OPS: 6496.6, Avg(us): 3, Min(us): 1, Max(us): 1514, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 14, 99.9th(us): 185, 99.99th(us): 731
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3063
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  134
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong state  42
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong txnId  3

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  199
rollback failed
  version mismatch  33
```

+ 96

```bash
----------------------------------
Run finished, takes 5.80121793s
COMMIT - Takes(s): 5.8, Count: 12937, OPS: 2240.5, Avg(us): 8728, Min(us): 0, Max(us): 30991, 50th(us): 8623, 90th(us): 10303, 95th(us): 10919, 99th(us): 13415, 99.9th(us): 19823, 99.99th(us): 30911
COMMIT_ERROR - Takes(s): 5.8, Count: 3671, OPS: 636.2, Avg(us): 5283, Min(us): 3436, Max(us): 15679, 50th(us): 5039, 90th(us): 6619, 95th(us): 7239, 99th(us): 8959, 99.9th(us): 14863, 99.99th(us): 15679
READ   - Takes(s): 5.8, Count: 99579, OPS: 17173.7, Avg(us): 4137, Min(us): 5, Max(us): 14431, 50th(us): 3919, 90th(us): 5059, 95th(us): 5543, 99th(us): 6667, 99.9th(us): 8575, 99.99th(us): 11063
READ_ERROR - Takes(s): 5.8, Count: 421, OPS: 73.1, Avg(us): 6016, Min(us): 3622, Max(us): 12015, 50th(us): 5711, 90th(us): 8295, 95th(us): 9423, 99th(us): 10647, 99.9th(us): 12015, 99.99th(us): 12015
Start  - Takes(s): 5.8, Count: 16704, OPS: 2878.6, Avg(us): 41, Min(us): 13, Max(us): 2347, 50th(us): 28, 90th(us): 41, 95th(us): 51, 99th(us): 535, 99.9th(us): 1360, 99.99th(us): 2201
TOTAL  - Takes(s): 5.8, Count: 208518, OPS: 35935.6, Avg(us): 7238, Min(us): 0, Max(us): 61599, 50th(us): 3777, 90th(us): 32191, 95th(us): 34111, 99th(us): 36991, 99.9th(us): 41663, 99.99th(us): 49279
TXN    - Takes(s): 5.8, Count: 12937, OPS: 2240.3, Avg(us): 33722, Min(us): 18736, Max(us): 56767, 50th(us): 33567, 90th(us): 36575, 95th(us): 37823, 99th(us): 41183, 99.9th(us): 46783, 99.99th(us): 55423
TXN_ERROR - Takes(s): 5.8, Count: 3671, OPS: 636.3, Avg(us): 30411, Min(us): 19184, Max(us): 42495, 50th(us): 30239, 90th(us): 33279, 95th(us): 34431, 99th(us): 37279, 99.9th(us): 41535, 99.99th(us): 42495
TxnGroup - Takes(s): 5.8, Count: 16608, OPS: 2871.4, Avg(us): 32949, Min(us): 18656, Max(us): 61599, 50th(us): 33055, 90th(us): 36191, 95th(us): 37407, 99th(us): 40543, 99.9th(us): 47615, 99.99th(us): 59647
UPDATE - Takes(s): 5.8, Count: 49753, OPS: 8581.3, Avg(us): 4, Min(us): 1, Max(us): 1659, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 16, 99.9th(us): 352, 99.99th(us): 1174
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3421
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  185
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong state  60
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong txnId  5

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  354
rollback failed
  version mismatch  67
```

+ 128

```bash
----------------------------------
Run finished, takes 4.943070782s
COMMIT - Takes(s): 4.9, Count: 12593, OPS: 2565.1, Avg(us): 9950, Min(us): 0, Max(us): 38367, 50th(us): 9863, 90th(us): 12111, 95th(us): 12959, 99th(us): 15511, 99.9th(us): 22127, 99.99th(us): 35999
COMMIT_ERROR - Takes(s): 4.9, Count: 4047, OPS: 824.6, Avg(us): 6093, Min(us): 3452, Max(us): 20719, 50th(us): 5791, 90th(us): 7851, 95th(us): 8647, 99th(us): 11127, 99.9th(us): 17903, 99.99th(us): 20719
READ   - Takes(s): 4.9, Count: 99329, OPS: 20125.1, Avg(us): 4673, Min(us): 6, Max(us): 22351, 50th(us): 4367, 90th(us): 6159, 95th(us): 6811, 99th(us): 8471, 99.9th(us): 11423, 99.99th(us): 14639
READ_ERROR - Takes(s): 4.9, Count: 671, OPS: 137.4, Avg(us): 7349, Min(us): 3606, Max(us): 16223, 50th(us): 7023, 90th(us): 10031, 95th(us): 10999, 99th(us): 13951, 99.9th(us): 15943, 99.99th(us): 16223
Start  - Takes(s): 4.9, Count: 16768, OPS: 3391.4, Avg(us): 52, Min(us): 13, Max(us): 3461, 50th(us): 29, 90th(us): 43, 95th(us): 55, 99th(us): 846, 99.9th(us): 2499, 99.99th(us): 3245
TOTAL  - Takes(s): 4.9, Count: 207747, OPS: 42026.6, Avg(us): 8148, Min(us): 0, Max(us): 69887, 50th(us): 4067, 90th(us): 35871, 95th(us): 38879, 99th(us): 43103, 99.9th(us): 48927, 99.99th(us): 59007
TXN    - Takes(s): 4.9, Count: 12593, OPS: 2565.3, Avg(us): 38233, Min(us): 20976, Max(us): 65727, 50th(us): 38079, 90th(us): 42591, 95th(us): 44095, 99th(us): 47871, 99.9th(us): 52831, 99.99th(us): 61631
TXN_ERROR - Takes(s): 4.9, Count: 4047, OPS: 824.8, Avg(us): 34625, Min(us): 22176, Max(us): 55871, 50th(us): 34367, 90th(us): 38911, 95th(us): 40447, 99th(us): 43903, 99.9th(us): 48703, 99.99th(us): 55871
TxnGroup - Takes(s): 4.9, Count: 16640, OPS: 3381.3, Avg(us): 37298, Min(us): 21152, Max(us): 69887, 50th(us): 37247, 90th(us): 42111, 95th(us): 43711, 99th(us): 47903, 99.9th(us): 59007, 99.99th(us): 69375
UPDATE - Takes(s): 4.9, Count: 49824, OPS: 10092.9, Avg(us): 5, Min(us): 1, Max(us): 2785, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 17, 99.9th(us): 381, 99.99th(us): 1869
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3794
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  172
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong state  63
prepare phase failed: Remote prepare failed
validation failed in AA mode
  rollback failed due to wrong txnId  18

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  529
rollback failed
  version mismatch  141
     key not found    1
```

##### Oreo-AC using remote

+ 8

```bash
----------------------------------
Run finished, takes 1m4.883242821s
COMMIT - Takes(s): 64.8, Count: 15910, OPS: 245.3, Avg(us): 7473, Min(us): 0, Max(us): 48735, 50th(us): 7567, 90th(us): 7895, 95th(us): 8015, 99th(us): 8383, 99.9th(us): 9231, 99.99th(us): 41151
COMMIT_ERROR - Takes(s): 64.8, Count: 754, OPS: 11.6, Avg(us): 4258, Min(us): 3486, Max(us): 5247, 50th(us): 4247, 90th(us): 4575, 95th(us): 4739, 99th(us): 4947, 99.9th(us): 5159, 99.99th(us): 5247
READ   - Takes(s): 64.9, Count: 99772, OPS: 1537.8, Avg(us): 3950, Min(us): 5, Max(us): 45471, 50th(us): 3941, 90th(us): 4347, 95th(us): 4463, 99th(us): 4663, 99.9th(us): 4999, 99.99th(us): 5423
READ_ERROR - Takes(s): 64.4, Count: 228, OPS: 3.5, Avg(us): 4401, Min(us): 3452, Max(us): 5471, 50th(us): 4367, 90th(us): 4815, 95th(us): 4919, 99th(us): 5123, 99.9th(us): 5471, 99.99th(us): 5471
Start  - Takes(s): 64.9, Count: 16672, OPS: 257.0, Avg(us): 21, Min(us): 13, Max(us): 344, 50th(us): 19, 90th(us): 28, 95th(us): 31, 99th(us): 42, 99.9th(us): 213, 99.99th(us): 305
TOTAL  - Takes(s): 64.9, Count: 214908, OPS: 3312.2, Avg(us): 7116, Min(us): 0, Max(us): 73471, 50th(us): 3875, 90th(us): 31231, 95th(us): 31663, 99th(us): 32191, 99.9th(us): 32895, 99.99th(us): 41503
TXN    - Takes(s): 64.8, Count: 15910, OPS: 245.3, Avg(us): 31267, Min(us): 19392, Max(us): 73343, 50th(us): 31455, 90th(us): 32063, 95th(us): 32271, 99th(us): 32719, 99.9th(us): 33855, 99.99th(us): 72831
TXN_ERROR - Takes(s): 64.8, Count: 754, OPS: 11.6, Avg(us): 27946, Min(us): 22560, Max(us): 31087, 50th(us): 28143, 90th(us): 28815, 95th(us): 29023, 99th(us): 29407, 99.9th(us): 30495, 99.99th(us): 31087
TxnGroup - Takes(s): 64.9, Count: 16664, OPS: 256.9, Avg(us): 31113, Min(us): 18384, Max(us): 73471, 50th(us): 31439, 90th(us): 32095, 95th(us): 32287, 99th(us): 32735, 99.9th(us): 34175, 99.99th(us): 73343
UPDATE - Takes(s): 64.9, Count: 49980, OPS: 770.3, Avg(us): 3, Min(us): 1, Max(us): 661, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 10, 99.9th(us): 129, 99.99th(us): 268
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  745
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  9

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  227
rollback failed
  version mismatch  1
```

+ 16

```bash
----------------------------------
Run finished, takes 32.180994474s
COMMIT - Takes(s): 32.1, Count: 15348, OPS: 477.4, Avg(us): 7485, Min(us): 0, Max(us): 49311, 50th(us): 7583, 90th(us): 8091, 95th(us): 8247, 99th(us): 8551, 99.9th(us): 9295, 99.99th(us): 48863
COMMIT_ERROR - Takes(s): 32.2, Count: 1308, OPS: 40.7, Avg(us): 4312, Min(us): 3500, Max(us): 5851, 50th(us): 4267, 90th(us): 4803, 95th(us): 4931, 99th(us): 5239, 99.9th(us): 5695, 99.99th(us): 5851
READ   - Takes(s): 32.2, Count: 99677, OPS: 3097.8, Avg(us): 3911, Min(us): 5, Max(us): 46111, 50th(us): 3843, 90th(us): 4459, 95th(us): 4615, 99th(us): 4899, 99.9th(us): 5379, 99.99th(us): 5979
READ_ERROR - Takes(s): 31.9, Count: 323, OPS: 10.1, Avg(us): 4767, Min(us): 3494, Max(us): 46303, 50th(us): 4503, 90th(us): 5095, 95th(us): 5219, 99th(us): 5599, 99.9th(us): 46303, 99.99th(us): 46303
Start  - Takes(s): 32.2, Count: 16672, OPS: 518.1, Avg(us): 24, Min(us): 13, Max(us): 615, 50th(us): 21, 90th(us): 31, 95th(us): 38, 99th(us): 61, 99.9th(us): 288, 99.99th(us): 437
TOTAL  - Takes(s): 32.2, Count: 213156, OPS: 6623.6, Avg(us): 7014, Min(us): 0, Max(us): 74943, 50th(us): 3731, 90th(us): 30735, 95th(us): 31647, 99th(us): 32575, 99.9th(us): 33471, 99.99th(us): 69823
TXN    - Takes(s): 32.1, Count: 15348, OPS: 477.4, Avg(us): 31068, Min(us): 18560, Max(us): 73983, 50th(us): 31247, 90th(us): 32367, 95th(us): 32671, 99th(us): 33279, 99.9th(us): 35615, 99.99th(us): 73471
TXN_ERROR - Takes(s): 32.2, Count: 1308, OPS: 40.7, Avg(us): 27828, Min(us): 18960, Max(us): 71039, 50th(us): 27999, 90th(us): 29135, 95th(us): 29439, 99th(us): 29935, 99.9th(us): 31679, 99.99th(us): 71039
TxnGroup - Takes(s): 32.2, Count: 16656, OPS: 518.0, Avg(us): 30807, Min(us): 18512, Max(us): 74943, 50th(us): 31183, 90th(us): 32415, 95th(us): 32735, 99th(us): 33311, 99.9th(us): 36063, 99.99th(us): 73215
UPDATE - Takes(s): 32.2, Count: 49455, OPS: 1537.0, Avg(us): 3, Min(us): 1, Max(us): 707, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 12, 99.9th(us): 172, 99.99th(us): 397
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  1270
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  33
prepare phase failed: Remote prepare failed
  validation failed due to false assumption  5

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  303
rollback failed
  version mismatch  20
```

+ 32

```bash
----------------------------------
Run finished, takes 15.954589858s
COMMIT - Takes(s): 15.9, Count: 14427, OPS: 906.4, Avg(us): 7636, Min(us): 0, Max(us): 14207, 50th(us): 7687, 90th(us): 8423, 95th(us): 8671, 99th(us): 9215, 99.9th(us): 11847, 99.99th(us): 13095
COMMIT_ERROR - Takes(s): 15.9, Count: 2213, OPS: 139.0, Avg(us): 4398, Min(us): 3414, Max(us): 8919, 50th(us): 4335, 90th(us): 4979, 95th(us): 5171, 99th(us): 5567, 99.9th(us): 6195, 99.99th(us): 8919
READ   - Takes(s): 16.0, Count: 99635, OPS: 6246.6, Avg(us): 3852, Min(us): 5, Max(us): 6667, 50th(us): 3755, 90th(us): 4407, 95th(us): 4635, 99th(us): 5123, 99.9th(us): 5771, 99.99th(us): 6223
READ_ERROR - Takes(s): 15.8, Count: 365, OPS: 23.1, Avg(us): 4603, Min(us): 3594, Max(us): 6359, 50th(us): 4547, 90th(us): 5263, 95th(us): 5403, 99th(us): 5711, 99.9th(us): 6359, 99.99th(us): 6359
Start  - Takes(s): 16.0, Count: 16672, OPS: 1044.9, Avg(us): 27, Min(us): 13, Max(us): 1377, 50th(us): 25, 90th(us): 36, 95th(us): 42, 99th(us): 97, 99.9th(us): 457, 99.99th(us): 917
TOTAL  - Takes(s): 16.0, Count: 211801, OPS: 13274.7, Avg(us): 6830, Min(us): 0, Max(us): 41695, 50th(us): 3681, 90th(us): 30399, 95th(us): 31439, 99th(us): 32511, 99.9th(us): 33759, 99.99th(us): 39935
TXN    - Takes(s): 15.9, Count: 14427, OPS: 906.4, Avg(us): 30890, Min(us): 17824, Max(us): 41695, 50th(us): 31055, 90th(us): 32319, 95th(us): 32703, 99th(us): 33631, 99.9th(us): 39967, 99.99th(us): 41535
TXN_ERROR - Takes(s): 15.9, Count: 2213, OPS: 139.0, Avg(us): 27560, Min(us): 19104, Max(us): 36895, 50th(us): 27727, 90th(us): 29007, 95th(us): 29311, 99th(us): 30143, 99.9th(us): 33343, 99.99th(us): 36895
TxnGroup - Takes(s): 15.9, Count: 16640, OPS: 1044.7, Avg(us): 30433, Min(us): 15688, Max(us): 40479, 50th(us): 30911, 90th(us): 32287, 95th(us): 32655, 99th(us): 33439, 99.9th(us): 38463, 99.99th(us): 40383
UPDATE - Takes(s): 16.0, Count: 50000, OPS: 3134.7, Avg(us): 3, Min(us): 1, Max(us): 969, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 15, 99.9th(us): 182, 99.99th(us): 562
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  2115
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  87
prepare phase failed: Remote prepare failed
  validation failed due to false assumption  11

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  332
rollback failed
  version mismatch  33
```

+ 64

```bash
----------------------------------
Run finished, takes 7.955957107s
COMMIT - Takes(s): 7.9, Count: 13453, OPS: 1697.5, Avg(us): 7724, Min(us): 0, Max(us): 15079, 50th(us): 7611, 90th(us): 9023, 95th(us): 9519, 99th(us): 10807, 99.9th(us): 12119, 99.99th(us): 13439
COMMIT_ERROR - Takes(s): 7.9, Count: 3187, OPS: 402.5, Avg(us): 4579, Min(us): 3408, Max(us): 10215, 50th(us): 4351, 90th(us): 5627, 95th(us): 6115, 99th(us): 7147, 99.9th(us): 8527, 99.99th(us): 10215
READ   - Takes(s): 7.9, Count: 99825, OPS: 12564.5, Avg(us): 3824, Min(us): 6, Max(us): 13999, 50th(us): 3629, 90th(us): 4511, 95th(us): 4943, 99th(us): 6051, 99.9th(us): 8199, 99.99th(us): 13287
READ_ERROR - Takes(s): 7.9, Count: 175, OPS: 22.2, Avg(us): 4460, Min(us): 3606, Max(us): 7775, 50th(us): 4307, 90th(us): 5395, 95th(us): 5763, 99th(us): 6675, 99.9th(us): 7775, 99.99th(us): 7775
Start  - Takes(s): 8.0, Count: 16704, OPS: 2099.6, Avg(us): 41, Min(us): 14, Max(us): 2687, 50th(us): 30, 90th(us): 46, 95th(us): 54, 99th(us): 309, 99.9th(us): 1309, 99.99th(us): 2057
TOTAL  - Takes(s): 8.0, Count: 210080, OPS: 26404.7, Avg(us): 6685, Min(us): 0, Max(us): 44191, 50th(us): 3547, 90th(us): 29631, 95th(us): 31151, 99th(us): 33663, 99.9th(us): 36895, 99.99th(us): 42143
TXN    - Takes(s): 7.9, Count: 13453, OPS: 1697.5, Avg(us): 30849, Min(us): 17568, Max(us): 44191, 50th(us): 30655, 90th(us): 33311, 95th(us): 34335, 99th(us): 36671, 99.9th(us): 42463, 99.99th(us): 43935
TXN_ERROR - Takes(s): 7.9, Count: 3187, OPS: 402.5, Avg(us): 27672, Min(us): 18112, Max(us): 39999, 50th(us): 27439, 90th(us): 30079, 95th(us): 31119, 99th(us): 33407, 99.9th(us): 38591, 99.99th(us): 39999
TxnGroup - Takes(s): 7.9, Count: 16640, OPS: 2099.1, Avg(us): 30213, Min(us): 18624, Max(us): 43455, 50th(us): 30271, 90th(us): 32991, 95th(us): 33951, 99th(us): 36255, 99.9th(us): 39359, 99.99th(us): 40991
UPDATE - Takes(s): 7.9, Count: 50005, OPS: 6293.5, Avg(us): 6, Min(us): 1, Max(us): 2171, 50th(us): 5, 90th(us): 6, 95th(us): 7, 99th(us): 21, 99.9th(us): 397, 99.99th(us): 1276
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  141
rollback failed
  version mismatch  34

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3009
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  154
prepare phase failed: Remote prepare failed
  validation failed due to false assumption  24
```

+ 96

```bash
----------------------------------
Run finished, takes 5.759408909s
COMMIT - Takes(s): 5.7, Count: 12896, OPS: 2249.8, Avg(us): 8585, Min(us): 0, Max(us): 58751, 50th(us): 8423, 90th(us): 10015, 95th(us): 10631, 99th(us): 12711, 99.9th(us): 54527, 99.99th(us): 58623
COMMIT_ERROR - Takes(s): 5.7, Count: 3712, OPS: 648.2, Avg(us): 5142, Min(us): 3506, Max(us): 53599, 50th(us): 4883, 90th(us): 6331, 95th(us): 6835, 99th(us): 8727, 99.9th(us): 14839, 99.99th(us): 53599
READ   - Takes(s): 5.8, Count: 99608, OPS: 17311.6, Avg(us): 4121, Min(us): 6, Max(us): 54751, 50th(us): 3883, 90th(us): 4975, 95th(us): 5435, 99th(us): 6571, 99.9th(us): 8967, 99.99th(us): 52415
READ_ERROR - Takes(s): 5.7, Count: 392, OPS: 68.7, Avg(us): 5976, Min(us): 3788, Max(us): 55775, 50th(us): 5499, 90th(us): 7595, 95th(us): 8295, 99th(us): 10455, 99.9th(us): 55775, 99.99th(us): 55775
Start  - Takes(s): 5.8, Count: 16704, OPS: 2900.0, Avg(us): 37, Min(us): 14, Max(us): 2011, 50th(us): 28, 90th(us): 41, 95th(us): 47, 99th(us): 370, 99.9th(us): 1248, 99.99th(us): 1978
TOTAL  - Takes(s): 5.8, Count: 208616, OPS: 36210.9, Avg(us): 7175, Min(us): 0, Max(us): 87359, 50th(us): 3753, 90th(us): 31791, 95th(us): 33631, 99th(us): 36351, 99.9th(us): 53471, 99.99th(us): 84991
TXN    - Takes(s): 5.7, Count: 12896, OPS: 2249.9, Avg(us): 33470, Min(us): 19184, Max(us): 87359, 50th(us): 33087, 90th(us): 35903, 95th(us): 37023, 99th(us): 42303, 99.9th(us): 84287, 99.99th(us): 87103
TXN_ERROR - Takes(s): 5.7, Count: 3712, OPS: 648.2, Avg(us): 30200, Min(us): 21296, Max(us): 83135, 50th(us): 29743, 90th(us): 32607, 95th(us): 33791, 99th(us): 38719, 99.9th(us): 78655, 99.99th(us): 83135
TxnGroup - Takes(s): 5.7, Count: 16608, OPS: 2894.0, Avg(us): 32698, Min(us): 19440, Max(us): 86975, 50th(us): 32607, 90th(us): 35615, 95th(us): 36767, 99th(us): 40991, 99.9th(us): 83903, 99.99th(us): 86399
UPDATE - Takes(s): 5.8, Count: 49904, OPS: 8671.8, Avg(us): 4, Min(us): 1, Max(us): 1635, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 16, 99.9th(us): 375, 99.99th(us): 1261
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  327
rollback failed
  version mismatch  65

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3489
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  194
prepare phase failed: Remote prepare failed
  validation failed due to false assumption  29
```

+ 128

```bash
----------------------------------
Run finished, takes 5.350061959s
COMMIT - Takes(s): 5.3, Count: 12667, OPS: 2378.8, Avg(us): 11216, Min(us): 0, Max(us): 77119, 50th(us): 10927, 90th(us): 13495, 95th(us): 14391, 99th(us): 18479, 99.9th(us): 69823, 99.99th(us): 76031
COMMIT_ERROR - Takes(s): 5.3, Count: 3973, OPS: 746.1, Avg(us): 6768, Min(us): 3718, Max(us): 63519, 50th(us): 6243, 90th(us): 8319, 95th(us): 9087, 99th(us): 11983, 99.9th(us): 60575, 99.99th(us): 63519
READ   - Takes(s): 5.3, Count: 99173, OPS: 18548.8, Avg(us): 4990, Min(us): 5, Max(us): 65471, 50th(us): 4619, 90th(us): 6775, 95th(us): 7483, 99th(us): 9263, 99.9th(us): 13439, 99.99th(us): 54751
READ_ERROR - Takes(s): 5.3, Count: 827, OPS: 157.3, Avg(us): 8634, Min(us): 4128, Max(us): 66943, 50th(us): 8007, 90th(us): 11959, 95th(us): 13367, 99th(us): 15471, 99.9th(us): 21471, 99.99th(us): 66943
Start  - Takes(s): 5.4, Count: 16768, OPS: 3134.2, Avg(us): 53, Min(us): 14, Max(us): 3019, 50th(us): 29, 90th(us): 48, 95th(us): 63, 99th(us): 883, 99.9th(us): 2183, 99.99th(us): 2897
TOTAL  - Takes(s): 5.4, Count: 207221, OPS: 38724.5, Avg(us): 8855, Min(us): 0, Max(us): 123519, 50th(us): 4219, 90th(us): 38271, 95th(us): 42015, 99th(us): 47423, 99.9th(us): 81279, 99.99th(us): 104063
TXN    - Takes(s): 5.3, Count: 12667, OPS: 2378.6, Avg(us): 41405, Min(us): 22800, Max(us): 102719, 50th(us): 40959, 90th(us): 46463, 95th(us): 48543, 99th(us): 57695, 99.9th(us): 95679, 99.99th(us): 102015
TXN_ERROR - Takes(s): 5.3, Count: 3973, OPS: 746.1, Avg(us): 37502, Min(us): 22256, Max(us): 100543, 50th(us): 36863, 90th(us): 42399, 95th(us): 44607, 99th(us): 54431, 99.9th(us): 90303, 99.99th(us): 100543
TxnGroup - Takes(s): 5.3, Count: 16640, OPS: 3120.7, Avg(us): 40413, Min(us): 19344, Max(us): 123519, 50th(us): 39967, 90th(us): 46047, 95th(us): 47967, 99th(us): 56191, 99.9th(us): 104767, 99.99th(us): 123007
UPDATE - Takes(s): 5.3, Count: 49306, OPS: 9221.5, Avg(us): 5, Min(us): 1, Max(us): 2022, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 24, 99.9th(us): 558, 99.99th(us): 1531
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  3706
prepare phase failed: Remote prepare failed
  validation failed due to unknown status  216
prepare phase failed: Remote prepare failed
  validation failed due to false assumption  51

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  670
rollback failed
  version mismatch  156
     key not found    1
```

## High Latency

+ Latency = 50ms
+ operation count = 10000
+ TxnGroup = 6
+ zipfian constant = 0.9

### Workload A

#### Cherry Garcia

+ 8

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 8
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 2m0.222037698s
COMMIT - Takes(s): 119.9, Count: 1596, OPS: 13.3, Avg(us): 409264, Min(us): 0, Max(us): 817151, 50th(us): 406271, 90th(us): 608767, 95th(us): 609791, 99th(us): 710655, 99.9th(us): 761855, 99.99th(us): 817151
COMMIT_ERROR - Takes(s): 119.5, Count: 68, OPS: 0.6, Avg(us): 441730, Min(us): 202752, Max(us): 761343, 50th(us): 406527, 90th(us): 609279, 95th(us): 609279, 99th(us): 610303, 99.9th(us): 761343, 99.99th(us): 761343
READ   - Takes(s): 120.2, Count: 4926, OPS: 41.0, Avg(us): 51944, Min(us): 5, Max(us): 203263, 50th(us): 50687, 90th(us): 50975, 95th(us): 51103, 99th(us): 152063, 99.9th(us): 203135, 99.99th(us): 203263
READ_ERROR - Takes(s): 113.5, Count: 16, OPS: 0.1, Avg(us): 180592, Min(us): 151680, Max(us): 203135, 50th(us): 202495, 90th(us): 203007, 95th(us): 203135, 99th(us): 203135, 99.9th(us): 203135, 99.99th(us): 203135
Start  - Takes(s): 120.2, Count: 1672, OPS: 13.9, Avg(us): 33, Min(us): 14, Max(us): 428, 50th(us): 25, 90th(us): 41, 95th(us): 56, 99th(us): 220, 99.9th(us): 337, 99.99th(us): 428
TOTAL  - Takes(s): 120.2, Count: 16512, OPS: 137.3, Avg(us): 166657, Min(us): 0, Max(us): 1016319, 50th(us): 50559, 90th(us): 559103, 95th(us): 659455, 99th(us): 761343, 99.9th(us): 913919, 99.99th(us): 968703
TXN    - Takes(s): 119.9, Count: 1596, OPS: 13.3, Avg(us): 566783, Min(us): 253184, Max(us): 862719, 50th(us): 558591, 90th(us): 659967, 95th(us): 710655, 99th(us): 761855, 99.9th(us): 813055, 99.99th(us): 862719
TXN_ERROR - Takes(s): 119.5, Count: 68, OPS: 0.6, Avg(us): 548377, Min(us): 253440, Max(us): 812031, 50th(us): 558591, 90th(us): 659967, 95th(us): 660479, 99th(us): 712191, 99.9th(us): 812031, 99.99th(us): 812031
TxnGroup - Takes(s): 120.1, Count: 1664, OPS: 13.9, Avg(us): 563774, Min(us): 76, Max(us): 1016319, 50th(us): 558591, 90th(us): 761343, 95th(us): 812031, 99th(us): 913919, 99.9th(us): 968703, 99.99th(us): 1016319
UPDATE - Takes(s): 120.2, Count: 5058, OPS: 42.1, Avg(us): 4, Min(us): 1, Max(us): 274, 50th(us): 3, 90th(us): 6, 95th(us): 7, 99th(us): 24, 99.9th(us): 187, 99.99th(us): 225
Error Summary:

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
  transaction is aborted by other transaction      32
       prepare phase failed: version mismatch      12
prepare phase failed: rollForward failed
  version mismatch  12
prepare phase failed: rollback failed
                                                                           version mismatch  6
  prepare phase failed: rollback failed because the corresponding transaction has committed  4
                                                 prepare phase failed: get old state failed  2

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  6
rollback failed
                                                     version mismatch  5
  rollback failed because the corresponding transaction has committed  3
                                                 get old state failed  2
```

+ 16

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 16
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 1m0.448811923s
COMMIT - Takes(s): 60.0, Count: 1528, OPS: 25.5, Avg(us): 408663, Min(us): 0, Max(us): 815103, 50th(us): 407039, 90th(us): 608767, 95th(us): 611327, 99th(us): 712703, 99.9th(us): 764415, 99.99th(us): 815103
COMMIT_ERROR - Takes(s): 59.8, Count: 136, OPS: 2.3, Avg(us): 414132, Min(us): 101824, Max(us): 764415, 50th(us): 407295, 90th(us): 610303, 95th(us): 661503, 99th(us): 763391, 99.9th(us): 764415, 99.99th(us): 764415
READ   - Takes(s): 60.4, Count: 4889, OPS: 80.9, Avg(us): 53383, Min(us): 5, Max(us): 204415, 50th(us): 50783, 90th(us): 51199, 95th(us): 51359, 99th(us): 152959, 99.9th(us): 203903, 99.99th(us): 204415
READ_ERROR - Takes(s): 59.8, Count: 46, OPS: 0.8, Avg(us): 175613, Min(us): 150784, Max(us): 254719, 50th(us): 152831, 90th(us): 204031, 95th(us): 253951, 99th(us): 254719, 99.9th(us): 254719, 99.99th(us): 254719
Start  - Takes(s): 60.4, Count: 1680, OPS: 27.8, Avg(us): 37, Min(us): 15, Max(us): 466, 50th(us): 28, 90th(us): 48, 95th(us): 66, 99th(us): 249, 99.9th(us): 372, 99.99th(us): 466
TOTAL  - Takes(s): 60.5, Count: 16354, OPS: 270.5, Avg(us): 165416, Min(us): 0, Max(us): 1220607, 50th(us): 50559, 90th(us): 560639, 95th(us): 660479, 99th(us): 763391, 99.9th(us): 915967, 99.99th(us): 1016831
TXN    - Takes(s): 60.0, Count: 1528, OPS: 25.5, Avg(us): 573646, Min(us): 303360, Max(us): 913407, 50th(us): 560127, 90th(us): 661503, 95th(us): 712191, 99th(us): 813567, 99.9th(us): 864255, 99.99th(us): 913407
TXN_ERROR - Takes(s): 59.8, Count: 136, OPS: 2.3, Avg(us): 537797, Min(us): 253952, Max(us): 815103, 50th(us): 559103, 90th(us): 662527, 95th(us): 761855, 99th(us): 814079, 99.9th(us): 815103, 99.99th(us): 815103
TxnGroup - Takes(s): 60.4, Count: 1664, OPS: 27.5, Avg(us): 566805, Min(us): 86, Max(us): 1220607, 50th(us): 560127, 90th(us): 762879, 95th(us): 814079, 99th(us): 915455, 99.9th(us): 1016831, 99.99th(us): 1220607
UPDATE - Takes(s): 60.5, Count: 5065, OPS: 83.8, Avg(us): 5, Min(us): 1, Max(us): 480, 50th(us): 3, 90th(us): 6, 95th(us): 9, 99th(us): 38, 99.9th(us): 316, 99.99th(us): 391
Error Summary:

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
  transaction is aborted by other transaction      66
       prepare phase failed: version mismatch      33
prepare phase failed: rollForward failed
  version mismatch  16
prepare phase failed: rollback failed
                                                                           version mismatch  12
  prepare phase failed: rollback failed because the corresponding transaction has committed   6
                                                 prepare phase failed: get old state failed   3

  Operation:   READ
       Error  Count
       -----  -----
rollback failed
  version mismatch  19
rollForward failed
                                                     version mismatch  18
  rollback failed because the corresponding transaction has committed   6
                                                 get old state failed   3
```

+ 32

```
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 32
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 31.336843837s
COMMIT - Takes(s): 31.0, Count: 1462, OPS: 47.1, Avg(us): 398374, Min(us): 0, Max(us): 911871, 50th(us): 405759, 90th(us): 606207, 95th(us): 609279, 99th(us): 710655, 99.9th(us): 761855, 99.99th(us): 911871
COMMIT_ERROR - Takes(s): 31.0, Count: 202, OPS: 6.5, Avg(us): 433690, Min(us): 101312, Max(us): 1012735, 50th(us): 406015, 90th(us): 609791, 95th(us): 709631, 99th(us): 860671, 99.9th(us): 1012735, 99.99th(us): 1012735
READ   - Takes(s): 31.3, Count: 4961, OPS: 158.6, Avg(us): 53377, Min(us): 5, Max(us): 205183, 50th(us): 50527, 90th(us): 51199, 95th(us): 51519, 99th(us): 153471, 99.9th(us): 203391, 99.99th(us): 205183
READ_ERROR - Takes(s): 31.0, Count: 88, OPS: 2.8, Avg(us): 169082, Min(us): 150656, Max(us): 254719, 50th(us): 151935, 90th(us): 204287, 95th(us): 253055, 99th(us): 254079, 99.9th(us): 254719, 99.99th(us): 254719
Start  - Takes(s): 31.3, Count: 1680, OPS: 53.6, Avg(us): 46, Min(us): 13, Max(us): 1085, 50th(us): 29, 90th(us): 55, 95th(us): 94, 99th(us): 453, 99.9th(us): 1039, 99.99th(us): 1085
TOTAL  - Takes(s): 31.3, Count: 16180, OPS: 516.3, Avg(us): 162064, Min(us): 0, Max(us): 1216511, 50th(us): 50399, 90th(us): 559615, 95th(us): 658431, 99th(us): 764927, 99.9th(us): 962559, 99.99th(us): 1067007
TXN    - Takes(s): 31.0, Count: 1462, OPS: 47.1, Avg(us): 572780, Min(us): 302080, Max(us): 1012735, 50th(us): 558591, 90th(us): 660991, 95th(us): 711679, 99th(us): 808959, 99.9th(us): 914943, 99.99th(us): 1012735
TXN_ERROR - Takes(s): 31.0, Count: 202, OPS: 6.5, Avg(us): 555090, Min(us): 204800, Max(us): 1012735, 50th(us): 558079, 90th(us): 709631, 95th(us): 808959, 99th(us): 862719, 99.9th(us): 1012735, 99.99th(us): 1012735
TxnGroup - Takes(s): 31.3, Count: 1664, OPS: 53.1, Avg(us): 563373, Min(us): 1205, Max(us): 1216511, 50th(us): 558591, 90th(us): 761855, 95th(us): 814591, 99th(us): 962559, 99.9th(us): 1067007, 99.99th(us): 1216511
UPDATE - Takes(s): 31.3, Count: 4951, OPS: 158.0, Avg(us): 6, Min(us): 1, Max(us): 507, 50th(us): 3, 90th(us): 6, 95th(us): 9, 99th(us): 45, 99.9th(us): 384, 99.99th(us): 507
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  43
rollback failed
                                                     version mismatch  32
  rollback failed because the corresponding transaction has committed  13

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
  transaction is aborted by other transaction      93
       prepare phase failed: version mismatch      41
prepare phase failed: rollForward failed
  version mismatch  34
prepare phase failed: rollback failed
                                                                           version mismatch  23
  prepare phase failed: rollback failed because the corresponding transaction has committed  11
```

+ 64

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 16.238195359s
COMMIT - Takes(s): 15.9, Count: 1341, OPS: 84.1, Avg(us): 400626, Min(us): 0, Max(us): 962047, 50th(us): 406015, 90th(us): 607743, 95th(us): 610303, 99th(us): 760319, 99.9th(us): 864255, 99.99th(us): 962047
COMMIT_ERROR - Takes(s): 15.8, Count: 323, OPS: 20.4, Avg(us): 413792, Min(us): 101248, Max(us): 862719, 50th(us): 406015, 90th(us): 610303, 95th(us): 707583, 99th(us): 762367, 99.9th(us): 862719, 99.99th(us): 862719
READ   - Takes(s): 16.2, Count: 4855, OPS: 299.9, Avg(us): 54200, Min(us): 5, Max(us): 205055, 50th(us): 50591, 90th(us): 51295, 95th(us): 51711, 99th(us): 201983, 99.9th(us): 203775, 99.99th(us): 205055
READ_ERROR - Takes(s): 15.5, Count: 139, OPS: 9.0, Avg(us): 168034, Min(us): 150400, Max(us): 254207, 50th(us): 152575, 90th(us): 203391, 95th(us): 252543, 99th(us): 254079, 99.9th(us): 254207, 99.99th(us): 254207
Start  - Takes(s): 16.2, Count: 1680, OPS: 103.4, Avg(us): 44, Min(us): 14, Max(us): 1000, 50th(us): 29, 90th(us): 52, 95th(us): 87, 99th(us): 413, 99.9th(us): 668, 99.99th(us): 1000
TOTAL  - Takes(s): 16.2, Count: 15887, OPS: 978.3, Avg(us): 158022, Min(us): 0, Max(us): 1369087, 50th(us): 50399, 90th(us): 559615, 95th(us): 659455, 99th(us): 812031, 99.9th(us): 964607, 99.99th(us): 1115135
TXN    - Takes(s): 15.9, Count: 1341, OPS: 84.2, Avg(us): 580601, Min(us): 302592, Max(us): 965119, 50th(us): 559615, 90th(us): 709631, 95th(us): 759807, 99th(us): 814079, 99.9th(us): 962047, 99.99th(us): 965119
TXN_ERROR - Takes(s): 15.8, Count: 323, OPS: 20.4, Avg(us): 553571, Min(us): 153088, Max(us): 863231, 50th(us): 558079, 90th(us): 711679, 95th(us): 762367, 99th(us): 861695, 99.9th(us): 863231, 99.99th(us): 863231
TxnGroup - Takes(s): 16.2, Count: 1664, OPS: 102.5, Avg(us): 559754, Min(us): 77, Max(us): 1369087, 50th(us): 558591, 90th(us): 809983, 95th(us): 861695, 99th(us): 963583, 99.9th(us): 1115135, 99.99th(us): 1369087
UPDATE - Takes(s): 16.2, Count: 5006, OPS: 308.3, Avg(us): 6, Min(us): 1, Max(us): 739, 50th(us): 4, 90th(us): 6, 95th(us): 8, 99th(us): 32, 99.9th(us): 490, 99.99th(us): 637
Error Summary:

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
  transaction is aborted by other transaction     130
prepare phase failed: rollback failed
                        version mismatch  64
  prepare phase failed: version mismatch  59
prepare phase failed: rollForward failed
                                                                           version mismatch  58
  prepare phase failed: rollback failed because the corresponding transaction has committed  12

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  62
rollback failed
                                                     version mismatch  60
  rollback failed because the corresponding transaction has committed  17
```

+ 96

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 96
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 10.35036097s
COMMIT - Takes(s): 9.9, Count: 1306, OPS: 132.0, Avg(us): 393591, Min(us): 0, Max(us): 913919, 50th(us): 405759, 90th(us): 510719, 95th(us): 609791, 99th(us): 711167, 99.9th(us): 859647, 99.99th(us): 913919
COMMIT_ERROR - Takes(s): 10.0, Count: 326, OPS: 32.5, Avg(us): 402803, Min(us): 101184, Max(us): 861183, 50th(us): 405503, 90th(us): 607231, 95th(us): 657407, 99th(us): 761855, 99.9th(us): 861183, 99.99th(us): 861183
READ   - Takes(s): 10.3, Count: 4884, OPS: 472.0, Avg(us): 54065, Min(us): 5, Max(us): 253311, 50th(us): 50559, 90th(us): 51295, 95th(us): 51679, 99th(us): 201855, 99.9th(us): 203903, 99.99th(us): 253311
READ_ERROR - Takes(s): 9.8, Count: 150, OPS: 15.3, Avg(us): 169816, Min(us): 150528, Max(us): 254847, 50th(us): 152319, 90th(us): 252287, 95th(us): 253311, 99th(us): 253951, 99.9th(us): 254847, 99.99th(us): 254847
Start  - Takes(s): 10.4, Count: 1728, OPS: 166.9, Avg(us): 62, Min(us): 14, Max(us): 1324, 50th(us): 30, 90th(us): 52, 95th(us): 120, 99th(us): 963, 99.9th(us): 1270, 99.99th(us): 1324
TOTAL  - Takes(s): 10.4, Count: 15822, OPS: 1528.3, Avg(us): 153162, Min(us): 0, Max(us): 1219583, 50th(us): 50367, 90th(us): 558591, 95th(us): 656895, 99th(us): 807935, 99.9th(us): 968703, 99.99th(us): 1065983
TXN    - Takes(s): 9.9, Count: 1306, OPS: 132.0, Avg(us): 577154, Min(us): 302080, Max(us): 1013759, 50th(us): 559103, 90th(us): 663039, 95th(us): 712191, 99th(us): 809471, 99.9th(us): 965119, 99.99th(us): 1013759
TXN_ERROR - Takes(s): 10.0, Count: 326, OPS: 32.5, Avg(us): 538080, Min(us): 203648, Max(us): 964607, 50th(us): 556543, 90th(us): 662527, 95th(us): 758783, 99th(us): 860159, 99.9th(us): 964607, 99.99th(us): 964607
TxnGroup - Takes(s): 10.3, Count: 1632, OPS: 157.7, Avg(us): 546172, Min(us): 450, Max(us): 1219583, 50th(us): 558079, 90th(us): 762367, 95th(us): 815103, 99th(us): 967167, 99.9th(us): 1065983, 99.99th(us): 1219583
UPDATE - Takes(s): 10.4, Count: 4966, OPS: 479.7, Avg(us): 5, Min(us): 1, Max(us): 656, 50th(us): 4, 90th(us): 6, 95th(us): 9, 99th(us): 29, 99.9th(us): 462, 99.99th(us): 656
Error Summary:

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
  transaction is aborted by other transaction     103
       prepare phase failed: version mismatch      84
prepare phase failed: rollback failed
  version mismatch  67
prepare phase failed: rollForward failed
                                                                           version mismatch  61
  prepare phase failed: rollback failed because the corresponding transaction has committed  11

  Operation:   READ
       Error  Count
       -----  -----
rollback failed
  version mismatch  76
rollForward failed
                                                     version mismatch  61
  rollback failed because the corresponding transaction has committed  13
```

+ 128

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 8.264569518s
COMMIT - Takes(s): 8.0, Count: 1227, OPS: 154.2, Avg(us): 397700, Min(us): 0, Max(us): 812543, 50th(us): 405247, 90th(us): 559103, 95th(us): 609279, 99th(us): 711167, 99.9th(us): 809471, 99.99th(us): 812543
COMMIT_ERROR - Takes(s): 8.0, Count: 437, OPS: 54.6, Avg(us): 407226, Min(us): 101056, Max(us): 860671, 50th(us): 404223, 90th(us): 607743, 95th(us): 658431, 99th(us): 761855, 99.9th(us): 860671, 99.99th(us): 860671
READ   - Takes(s): 8.2, Count: 4762, OPS: 579.7, Avg(us): 54509, Min(us): 5, Max(us): 204927, 50th(us): 50495, 90th(us): 51295, 95th(us): 51839, 99th(us): 201983, 99.9th(us): 204287, 99.99th(us): 204927
READ_ERROR - Takes(s): 7.8, Count: 168, OPS: 21.6, Avg(us): 172630, Min(us): 150528, Max(us): 256127, 50th(us): 152063, 90th(us): 252415, 95th(us): 253439, 99th(us): 254591, 99.9th(us): 256127, 99.99th(us): 256127
Start  - Takes(s): 8.3, Count: 1680, OPS: 203.2, Avg(us): 43, Min(us): 14, Max(us): 1131, 50th(us): 29, 90th(us): 48, 95th(us): 62, 99th(us): 553, 99.9th(us): 961, 99.99th(us): 1131
TOTAL  - Takes(s): 8.3, Count: 15630, OPS: 1890.8, Avg(us): 151253, Min(us): 0, Max(us): 1265663, 50th(us): 50335, 90th(us): 558591, 95th(us): 657919, 99th(us): 809983, 99.9th(us): 963071, 99.99th(us): 1114111
TXN    - Takes(s): 8.0, Count: 1227, OPS: 154.2, Avg(us): 582397, Min(us): 302336, Max(us): 1061887, 50th(us): 559615, 90th(us): 708095, 95th(us): 758271, 99th(us): 860159, 99.9th(us): 914431, 99.99th(us): 1061887
TXN_ERROR - Takes(s): 8.0, Count: 437, OPS: 54.6, Avg(us): 548509, Min(us): 252288, Max(us): 961023, 50th(us): 557055, 90th(us): 709631, 95th(us): 760831, 99th(us): 861695, 99.9th(us): 961023, 99.99th(us): 961023
TxnGroup - Takes(s): 8.2, Count: 1664, OPS: 202.5, Avg(us): 541969, Min(us): 50304, Max(us): 1265663, 50th(us): 557055, 90th(us): 761855, 95th(us): 858623, 99th(us): 962559, 99.9th(us): 1114111, 99.99th(us): 1265663
UPDATE - Takes(s): 8.3, Count: 5070, OPS: 613.3, Avg(us): 5, Min(us): 1, Max(us): 1037, 50th(us): 3, 90th(us): 6, 95th(us): 8, 99th(us): 19, 99.9th(us): 503, 99.99th(us): 809
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollback failed
  version mismatch  95
rollForward failed
                                                     version mismatch  60
  rollback failed because the corresponding transaction has committed  13

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
  transaction is aborted by other transaction     133
prepare phase failed: rollback failed
                        version mismatch  131
  prepare phase failed: version mismatch   94
prepare phase failed: rollForward failed
                                                                           version mismatch  58
  prepare phase failed: rollback failed because the corresponding transaction has committed  21
```

#### Oreo

+ 8

```bash
----------------------------------
Run finished, takes 53.952913941s
COMMIT - Takes(s): 53.8, Count: 1608, OPS: 29.9, Avg(us): 100313, Min(us): 0, Max(us): 128639, 50th(us): 102207, 90th(us): 102719, 95th(us): 102911, 99th(us): 103295, 99.9th(us): 103999, 99.99th(us): 128639
COMMIT_ERROR - Takes(s): 53.4, Count: 56, OPS: 1.0, Avg(us): 52878, Min(us): 50848, Max(us): 102783, 50th(us): 52031, 90th(us): 52735, 95th(us): 52863, 99th(us): 53247, 99.9th(us): 102783, 99.99th(us): 102783
READ   - Takes(s): 53.9, Count: 4969, OPS: 92.2, Avg(us): 51080, Min(us): 5, Max(us): 77567, 50th(us): 51167, 90th(us): 51711, 95th(us): 51839, 99th(us): 52095, 99.9th(us): 52415, 99.99th(us): 77567
READ_ERROR - Takes(s): 49.7, Count: 19, OPS: 0.4, Avg(us): 51647, Min(us): 51072, Max(us): 52543, 50th(us): 51615, 90th(us): 51935, 95th(us): 52191, 99th(us): 52543, 99.9th(us): 52543, 99.99th(us): 52543
Start  - Takes(s): 54.0, Count: 1672, OPS: 31.0, Avg(us): 29, Min(us): 13, Max(us): 442, 50th(us): 26, 90th(us): 36, 95th(us): 42, 99th(us): 123, 99.9th(us): 396, 99.99th(us): 442
TOTAL  - Takes(s): 54.0, Count: 16533, OPS: 306.4, Avg(us): 75050, Min(us): 0, Max(us): 410879, 50th(us): 50975, 90th(us): 255743, 95th(us): 306687, 99th(us): 358143, 99.9th(us): 409087, 99.99th(us): 410623
TXN    - Takes(s): 53.9, Count: 1608, OPS: 29.9, Avg(us): 253496, Min(us): 101760, Max(us): 383231, 50th(us): 255743, 90th(us): 308479, 95th(us): 358143, 99th(us): 358911, 99.9th(us): 359935, 99.99th(us): 383231
TXN_ERROR - Takes(s): 53.4, Count: 56, OPS: 1.0, Avg(us): 199979, Min(us): 51840, Max(us): 307455, 50th(us): 205311, 90th(us): 257279, 95th(us): 257663, 99th(us): 306687, 99.9th(us): 307455, 99.99th(us): 307455
TxnGroup - Takes(s): 54.0, Count: 1664, OPS: 30.8, Avg(us): 251203, Min(us): 118, Max(us): 410879, 50th(us): 255743, 90th(us): 357119, 95th(us): 358399, 99th(us): 409087, 99.9th(us): 410623, 99.99th(us): 410879
UPDATE - Takes(s): 54.0, Count: 5012, OPS: 92.9, Avg(us): 3, Min(us): 1, Max(us): 393, 50th(us): 3, 90th(us): 5, 95th(us): 6, 99th(us): 12, 99.9th(us): 122, 99.99th(us): 207
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  49
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  5
prepare phase failed: Remote prepare failed
        read failed due to unknown txn status  1
  transaction is aborted by other transaction  1

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  14
  read failed due to unknown txn status   5
```

+ 16

```bash
----------------------------------
Run finished, takes 27.769333222s
COMMIT - Takes(s): 27.7, Count: 1582, OPS: 57.2, Avg(us): 100478, Min(us): 0, Max(us): 104703, 50th(us): 102527, 90th(us): 103295, 95th(us): 103487, 99th(us): 103871, 99.9th(us): 104191, 99.99th(us): 104703
COMMIT_ERROR - Takes(s): 27.3, Count: 82, OPS: 3.0, Avg(us): 52273, Min(us): 50880, Max(us): 53535, 50th(us): 52255, 90th(us): 53023, 95th(us): 53119, 99th(us): 53311, 99.9th(us): 53535, 99.99th(us): 53535
READ   - Takes(s): 27.7, Count: 4908, OPS: 177.1, Avg(us): 51027, Min(us): 6, Max(us): 53119, 50th(us): 51103, 90th(us): 51711, 95th(us): 51935, 99th(us): 52319, 99.9th(us): 52831, 99.99th(us): 53119
READ_ERROR - Takes(s): 27.3, Count: 36, OPS: 1.3, Avg(us): 51832, Min(us): 50720, Max(us): 53183, 50th(us): 51839, 90th(us): 52383, 95th(us): 52447, 99th(us): 53183, 99.9th(us): 53183, 99.99th(us): 53183
Start  - Takes(s): 27.8, Count: 1680, OPS: 60.5, Avg(us): 29, Min(us): 13, Max(us): 599, 50th(us): 25, 90th(us): 38, 95th(us): 45, 99th(us): 171, 99.9th(us): 288, 99.99th(us): 599
TOTAL  - Takes(s): 27.8, Count: 16472, OPS: 593.1, Avg(us): 74201, Min(us): 0, Max(us): 411647, 50th(us): 50911, 90th(us): 255871, 95th(us): 306943, 99th(us): 358655, 99.9th(us): 409343, 99.99th(us): 411391
TXN    - Takes(s): 27.7, Count: 1582, OPS: 57.2, Avg(us): 252278, Min(us): 102016, Max(us): 360703, 50th(us): 256127, 90th(us): 308735, 95th(us): 358399, 99th(us): 359679, 99.9th(us): 359935, 99.99th(us): 360703
TXN_ERROR - Takes(s): 27.3, Count: 82, OPS: 3.0, Avg(us): 197768, Min(us): 53120, Max(us): 309247, 50th(us): 205695, 90th(us): 257535, 95th(us): 258303, 99th(us): 308223, 99.9th(us): 309247, 99.99th(us): 309247
TxnGroup - Takes(s): 27.8, Count: 1664, OPS: 59.9, Avg(us): 248607, Min(us): 87, Max(us): 411647, 50th(us): 255871, 90th(us): 356863, 95th(us): 358911, 99th(us): 409343, 99.9th(us): 411391, 99.99th(us): 411647
UPDATE - Takes(s): 27.8, Count: 5056, OPS: 182.1, Avg(us): 3, Min(us): 1, Max(us): 209, 50th(us): 3, 90th(us): 5, 95th(us): 5, 99th(us): 11, 99.9th(us): 152, 99.99th(us): 205
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  21
  read failed due to unknown txn status  13
                          key not found   1
rollback failed
  version mismatch  1

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  68
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  8
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  4
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  2
```

+ 32

```bash
----------------------------------
Run finished, takes 13.769684147s
COMMIT - Takes(s): 13.7, Count: 1498, OPS: 109.6, Avg(us): 100850, Min(us): 0, Max(us): 105151, 50th(us): 102527, 90th(us): 103679, 95th(us): 103999, 99th(us): 104575, 99.9th(us): 105087, 99.99th(us): 105151
COMMIT_ERROR - Takes(s): 13.7, Count: 166, OPS: 12.1, Avg(us): 52667, Min(us): 50592, Max(us): 102591, 50th(us): 52287, 90th(us): 53503, 95th(us): 53919, 99th(us): 54655, 99.9th(us): 102591, 99.99th(us): 102591
READ   - Takes(s): 13.7, Count: 4948, OPS: 360.7, Avg(us): 50925, Min(us): 6, Max(us): 54047, 50th(us): 50943, 90th(us): 51647, 95th(us): 51871, 99th(us): 52383, 99.9th(us): 52959, 99.99th(us): 54047
READ_ERROR - Takes(s): 13.2, Count: 40, OPS: 3.0, Avg(us): 51920, Min(us): 50624, Max(us): 53791, 50th(us): 51839, 90th(us): 52799, 95th(us): 53119, 99th(us): 53791, 99.9th(us): 53791, 99.99th(us): 53791
Start  - Takes(s): 13.8, Count: 1680, OPS: 122.0, Avg(us): 29, Min(us): 13, Max(us): 398, 50th(us): 24, 90th(us): 38, 95th(us): 46, 99th(us): 195, 99.9th(us): 367, 99.99th(us): 398
TOTAL  - Takes(s): 13.8, Count: 16300, OPS: 1183.7, Avg(us): 73294, Min(us): 0, Max(us): 411647, 50th(us): 50751, 90th(us): 255487, 95th(us): 306431, 99th(us): 357887, 99.9th(us): 407807, 99.99th(us): 410367
TXN    - Takes(s): 13.7, Count: 1498, OPS: 109.6, Avg(us): 254306, Min(us): 101824, Max(us): 360703, 50th(us): 255743, 90th(us): 355583, 95th(us): 357631, 99th(us): 359423, 99.9th(us): 360447, 99.99th(us): 360703
TXN_ERROR - Takes(s): 13.7, Count: 166, OPS: 12.1, Avg(us): 197615, Min(us): 51968, Max(us): 307967, 50th(us): 205439, 90th(us): 258303, 95th(us): 306431, 99th(us): 307455, 99.9th(us): 307967, 99.99th(us): 307967
TxnGroup - Takes(s): 13.8, Count: 1664, OPS: 120.8, Avg(us): 246769, Min(us): 323, Max(us): 411647, 50th(us): 255487, 90th(us): 356351, 95th(us): 358143, 99th(us): 407295, 99.9th(us): 410367, 99.99th(us): 411647
UPDATE - Takes(s): 13.8, Count: 5012, OPS: 364.0, Avg(us): 3, Min(us): 1, Max(us): 436, 50th(us): 3, 90th(us): 5, 95th(us): 5, 99th(us): 14, 99.9th(us): 185, 99.99th(us): 419
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  136
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  27
prepare phase failed: Remote prepare failed
        read failed due to unknown txn status  2
  transaction is aborted by other transaction  1

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  23
  read failed due to unknown txn status  17
```

+ 64

```bash
----------------------------------
Run finished, takes 7.108643539s
COMMIT - Takes(s): 7.0, Count: 1417, OPS: 203.9, Avg(us): 101428, Min(us): 0, Max(us): 107711, 50th(us): 102399, 90th(us): 103551, 95th(us): 104063, 99th(us): 105087, 99.9th(us): 106815, 99.99th(us): 107711
COMMIT_ERROR - Takes(s): 7.0, Count: 247, OPS: 35.5, Avg(us): 52242, Min(us): 50944, Max(us): 55391, 50th(us): 52159, 90th(us): 53343, 95th(us): 53727, 99th(us): 55231, 99.9th(us): 55391, 99.99th(us): 55391
READ   - Takes(s): 7.1, Count: 4910, OPS: 695.8, Avg(us): 50826, Min(us): 5, Max(us): 53791, 50th(us): 50847, 90th(us): 51583, 95th(us): 51903, 99th(us): 52831, 99.9th(us): 53503, 99.99th(us): 53791
READ_ERROR - Takes(s): 6.8, Count: 53, OPS: 7.7, Avg(us): 51517, Min(us): 50368, Max(us): 54111, 50th(us): 51487, 90th(us): 52351, 95th(us): 52575, 99th(us): 53375, 99.9th(us): 54111, 99.99th(us): 54111
Start  - Takes(s): 7.1, Count: 1680, OPS: 236.3, Avg(us): 34, Min(us): 13, Max(us): 846, 50th(us): 24, 90th(us): 37, 95th(us): 49, 99th(us): 375, 99.9th(us): 842, 99.99th(us): 846
TOTAL  - Takes(s): 7.1, Count: 16125, OPS: 2268.6, Avg(us): 71713, Min(us): 0, Max(us): 410367, 50th(us): 50687, 90th(us): 254847, 95th(us): 305663, 99th(us): 357119, 99.9th(us): 360703, 99.99th(us): 409343
TXN    - Takes(s): 7.0, Count: 1417, OPS: 203.8, Avg(us): 254370, Min(us): 102080, Max(us): 363007, 50th(us): 255231, 90th(us): 355327, 95th(us): 357119, 99th(us): 359423, 99.9th(us): 362751, 99.99th(us): 363007
TXN_ERROR - Takes(s): 7.0, Count: 247, OPS: 35.5, Avg(us): 195730, Min(us): 51904, Max(us): 310015, 50th(us): 204671, 90th(us): 256255, 95th(us): 257663, 99th(us): 308479, 99.9th(us): 310015, 99.99th(us): 310015
TxnGroup - Takes(s): 7.1, Count: 1664, OPS: 235.8, Avg(us): 241932, Min(us): 51008, Max(us): 410367, 50th(us): 254847, 90th(us): 308223, 95th(us): 357119, 99th(us): 359167, 99.9th(us): 409343, 99.99th(us): 410367
UPDATE - Takes(s): 7.1, Count: 5037, OPS: 708.6, Avg(us): 3, Min(us): 1, Max(us): 529, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 13, 99.9th(us): 73, 99.99th(us): 290
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  230
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  11
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  4
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  2

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status     30
rollForward failed
  version mismatch  20
rollback failed
  version mismatch  3
```

+ 96

```bash
----------------------------------
Run finished, takes 4.761532323s
COMMIT - Takes(s): 4.7, Count: 1321, OPS: 283.7, Avg(us): 100498, Min(us): 0, Max(us): 107839, 50th(us): 102527, 90th(us): 103935, 95th(us): 104511, 99th(us): 105855, 99.9th(us): 107583, 99.99th(us): 107839
COMMIT_ERROR - Takes(s): 4.7, Count: 311, OPS: 66.8, Avg(us): 52453, Min(us): 50592, Max(us): 57343, 50th(us): 52255, 90th(us): 54047, 95th(us): 54527, 99th(us): 55071, 99.9th(us): 57343, 99.99th(us): 57343
READ   - Takes(s): 4.7, Count: 4853, OPS: 1030.4, Avg(us): 50853, Min(us): 6, Max(us): 55775, 50th(us): 50847, 90th(us): 51679, 95th(us): 52127, 99th(us): 53791, 99.9th(us): 55359, 99.99th(us): 55775
READ_ERROR - Takes(s): 4.6, Count: 74, OPS: 16.1, Avg(us): 51689, Min(us): 50400, Max(us): 54111, 50th(us): 51391, 90th(us): 53215, 95th(us): 53439, 99th(us): 54079, 99.9th(us): 54111, 99.99th(us): 54111
Start  - Takes(s): 4.8, Count: 1728, OPS: 362.9, Avg(us): 28, Min(us): 13, Max(us): 730, 50th(us): 25, 90th(us): 36, 95th(us): 44, 99th(us): 66, 99.9th(us): 288, 99.99th(us): 730
TOTAL  - Takes(s): 4.8, Count: 15928, OPS: 3343.8, Avg(us): 68941, Min(us): 0, Max(us): 410111, 50th(us): 50655, 90th(us): 254591, 95th(us): 305407, 99th(us): 356863, 99.9th(us): 406783, 99.99th(us): 409343
TXN    - Takes(s): 4.7, Count: 1321, OPS: 283.6, Avg(us): 251974, Min(us): 101696, Max(us): 364031, 50th(us): 255359, 90th(us): 310015, 95th(us): 356863, 99th(us): 359935, 99.9th(us): 363775, 99.99th(us): 364031
TXN_ERROR - Takes(s): 4.7, Count: 311, OPS: 66.8, Avg(us): 197647, Min(us): 51680, Max(us): 315135, 50th(us): 204927, 90th(us): 257663, 95th(us): 304895, 99th(us): 308223, 99.9th(us): 315135, 99.99th(us): 315135
TxnGroup - Takes(s): 4.8, Count: 1632, OPS: 342.7, Avg(us): 236288, Min(us): 812, Max(us): 410111, 50th(us): 254719, 90th(us): 308479, 95th(us): 356863, 99th(us): 406783, 99.9th(us): 409343, 99.99th(us): 410111
UPDATE - Takes(s): 4.8, Count: 5073, OPS: 1065.0, Avg(us): 3, Min(us): 1, Max(us): 237, 50th(us): 3, 90th(us): 5, 95th(us): 5, 99th(us): 14, 99.9th(us): 191, 99.99th(us): 230
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  288
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  19
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  3
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  1

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status     47
rollForward failed
  version mismatch  21
     key not found   4
rollback failed
  version mismatch  2
```

+ 128

```bash
----------------------------------
Run finished, takes 3.936102568s
COMMIT - Takes(s): 3.8, Count: 1319, OPS: 348.3, Avg(us): 100692, Min(us): 0, Max(us): 150271, 50th(us): 102335, 90th(us): 104191, 95th(us): 104959, 99th(us): 144255, 99.9th(us): 148991, 99.99th(us): 150271
COMMIT_ERROR - Takes(s): 3.8, Count: 345, OPS: 91.1, Avg(us): 52758, Min(us): 50624, Max(us): 96767, 50th(us): 51999, 90th(us): 54271, 95th(us): 55327, 99th(us): 59903, 99.9th(us): 96767, 99.99th(us): 96767
READ   - Takes(s): 3.9, Count: 4860, OPS: 1250.6, Avg(us): 51392, Min(us): 5, Max(us): 97599, 50th(us): 50815, 90th(us): 51807, 95th(us): 52511, 99th(us): 92479, 99.9th(us): 96767, 99.99th(us): 97599
READ_ERROR - Takes(s): 3.9, Count: 99, OPS: 25.5, Avg(us): 52527, Min(us): 50272, Max(us): 96511, 50th(us): 51359, 90th(us): 53247, 95th(us): 54111, 99th(us): 95167, 99.9th(us): 96511, 99.99th(us): 96511
Start  - Takes(s): 3.9, Count: 1680, OPS: 426.6, Avg(us): 26, Min(us): 13, Max(us): 921, 50th(us): 23, 90th(us): 34, 95th(us): 42, 99th(us): 66, 99.9th(us): 278, 99.99th(us): 921
TOTAL  - Takes(s): 3.9, Count: 15883, OPS: 4034.3, Avg(us): 70163, Min(us): 0, Max(us): 448767, 50th(us): 50655, 90th(us): 254463, 95th(us): 305407, 99th(us): 357119, 99.9th(us): 407551, 99.99th(us): 410367
TXN    - Takes(s): 3.8, Count: 1319, OPS: 348.2, Avg(us): 255963, Min(us): 101760, Max(us): 407807, 50th(us): 255615, 90th(us): 355327, 95th(us): 357119, 99th(us): 359935, 99.9th(us): 406783, 99.99th(us): 407807
TXN_ERROR - Takes(s): 3.8, Count: 345, OPS: 91.1, Avg(us): 197554, Min(us): 51936, Max(us): 353791, 50th(us): 204415, 90th(us): 257919, 95th(us): 305151, 99th(us): 310271, 99.9th(us): 353791, 99.99th(us): 353791
TxnGroup - Takes(s): 3.9, Count: 1664, OPS: 422.8, Avg(us): 236869, Min(us): 69, Max(us): 448767, 50th(us): 254335, 90th(us): 309247, 95th(us): 357631, 99th(us): 407295, 99.9th(us): 410367, 99.99th(us): 448767
UPDATE - Takes(s): 3.9, Count: 5041, OPS: 1280.2, Avg(us): 3, Min(us): 1, Max(us): 265, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 8, 99.9th(us): 164, 99.99th(us): 257
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  323
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  12
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  7
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  3

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status     70
rollForward failed
  version mismatch  18
     key not found   6
rollback failed
  version mismatch  5
```

### Workload F

#### Cherry Garcia

+ 8

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 8
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 2m0.342789999s
COMMIT - Takes(s): 119.9, Count: 1620, OPS: 13.5, Avg(us): 249817, Min(us): 0, Max(us): 408063, 50th(us): 253951, 90th(us): 306431, 95th(us): 355583, 99th(us): 406271, 99.9th(us): 407039, 99.99th(us): 408063
COMMIT_ERROR - Takes(s): 118.8, Count: 44, OPS: 0.4, Avg(us): 210185, Min(us): 101632, Max(us): 355327, 50th(us): 203135, 90th(us): 306687, 95th(us): 355327, 99th(us): 355327, 99.9th(us): 355327, 99.99th(us): 355327
READ   - Takes(s): 120.3, Count: 9961, OPS: 82.8, Avg(us): 52075, Min(us): 5, Max(us): 203263, 50th(us): 50719, 90th(us): 51007, 95th(us): 51135, 99th(us): 152191, 99.9th(us): 202879, 99.99th(us): 203263
READ_ERROR - Takes(s): 118.0, Count: 39, OPS: 0.3, Avg(us): 171725, Min(us): 151680, Max(us): 203519, 50th(us): 152447, 90th(us): 203263, 95th(us): 203263, 99th(us): 203519, 99.9th(us): 203519, 99.99th(us): 203519
Start  - Takes(s): 120.3, Count: 1672, OPS: 13.9, Avg(us): 31, Min(us): 14, Max(us): 309, 50th(us): 25, 90th(us): 39, 95th(us): 52, 99th(us): 213, 99.9th(us): 281, 99.99th(us): 309
TOTAL  - Takes(s): 120.3, Count: 21516, OPS: 178.8, Avg(us): 129005, Min(us): 0, Max(us): 913407, 50th(us): 50655, 90th(us): 508927, 95th(us): 609279, 99th(us): 660991, 99.9th(us): 761855, 99.99th(us): 862719
TXN    - Takes(s): 119.9, Count: 1620, OPS: 13.5, Avg(us): 564886, Min(us): 303616, Max(us): 862719, 50th(us): 558591, 90th(us): 660479, 95th(us): 710655, 99th(us): 761343, 99.9th(us): 812543, 99.99th(us): 862719
TXN_ERROR - Takes(s): 118.8, Count: 44, OPS: 0.4, Avg(us): 537637, Min(us): 406016, Max(us): 1065983, 50th(us): 508159, 90th(us): 659455, 95th(us): 659967, 99th(us): 1065983, 99.9th(us): 1065983, 99.99th(us): 1065983
TxnGroup - Takes(s): 120.0, Count: 1664, OPS: 13.9, Avg(us): 563129, Min(us): 303872, Max(us): 913407, 50th(us): 558591, 90th(us): 659967, 95th(us): 710143, 99th(us): 761855, 99.9th(us): 814591, 99.99th(us): 913407
UPDATE - Takes(s): 120.3, Count: 4979, OPS: 41.4, Avg(us): 5, Min(us): 1, Max(us): 249, 50th(us): 4, 90th(us): 7, 95th(us): 8, 99th(us): 22, 99.9th(us): 185, 99.99th(us): 249
Error Summary:

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
       prepare phase failed: version mismatch      31
  transaction is aborted by other transaction      13

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  22
rollback failed
                                                     version mismatch  11
  rollback failed because the corresponding transaction has committed   4
                                                 get old state failed   2
```

+ 16

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 16
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 59.992554688s
COMMIT - Takes(s): 59.5, Count: 1559, OPS: 26.2, Avg(us): 247924, Min(us): 0, Max(us): 408319, 50th(us): 253951, 90th(us): 306431, 95th(us): 356095, 99th(us): 406783, 99.9th(us): 408063, 99.99th(us): 408319
COMMIT_ERROR - Takes(s): 59.4, Count: 105, OPS: 1.8, Avg(us): 178238, Min(us): 50560, Max(us): 406783, 50th(us): 153215, 90th(us): 304895, 95th(us): 305919, 99th(us): 356351, 99.9th(us): 406783, 99.99th(us): 406783
READ   - Takes(s): 59.9, Count: 9911, OPS: 165.3, Avg(us): 52432, Min(us): 5, Max(us): 204287, 50th(us): 50687, 90th(us): 51231, 95th(us): 51359, 99th(us): 152319, 99.9th(us): 202879, 99.99th(us): 204031
READ_ERROR - Takes(s): 58.7, Count: 89, OPS: 1.5, Avg(us): 173328, Min(us): 150528, Max(us): 253439, 50th(us): 153087, 90th(us): 203263, 95th(us): 203647, 99th(us): 203775, 99.9th(us): 253439, 99.99th(us): 253439
Start  - Takes(s): 60.0, Count: 1680, OPS: 28.0, Avg(us): 34, Min(us): 14, Max(us): 560, 50th(us): 26, 90th(us): 43, 95th(us): 62, 99th(us): 254, 99.9th(us): 386, 99.99th(us): 560
TOTAL  - Takes(s): 60.0, Count: 21313, OPS: 355.3, Avg(us): 128104, Min(us): 0, Max(us): 916479, 50th(us): 50591, 90th(us): 556031, 95th(us): 608255, 99th(us): 708607, 99.9th(us): 809983, 99.99th(us): 866303
TXN    - Takes(s): 59.5, Count: 1559, OPS: 26.2, Avg(us): 569631, Min(us): 302592, Max(us): 916479, 50th(us): 559615, 90th(us): 661503, 95th(us): 710655, 99th(us): 763391, 99.9th(us): 861183, 99.99th(us): 916479
TXN_ERROR - Takes(s): 59.4, Count: 105, OPS: 1.8, Avg(us): 492280, Min(us): 354304, Max(us): 711167, 50th(us): 505599, 90th(us): 609791, 95th(us): 658943, 99th(us): 662527, 99.9th(us): 711167, 99.99th(us): 711167
TxnGroup - Takes(s): 59.7, Count: 1664, OPS: 27.9, Avg(us): 562485, Min(us): 302592, Max(us): 913407, 50th(us): 559103, 90th(us): 661503, 95th(us): 711167, 99th(us): 763903, 99.9th(us): 865279, 99.99th(us): 913407
UPDATE - Takes(s): 59.9, Count: 4940, OPS: 82.4, Avg(us): 5, Min(us): 1, Max(us): 425, 50th(us): 4, 90th(us): 7, 95th(us): 9, 99th(us): 21, 99.9th(us): 270, 99.99th(us): 425
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                                                     version mismatch  50
  rollback failed because the corresponding transaction has committed  22
rollback failed
      version mismatch  11
  get old state failed   6

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
       prepare phase failed: version mismatch      87
  transaction is aborted by other transaction      18
```

+ 32

```
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 32
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 30.214288464s
COMMIT - Takes(s): 29.8, Count: 1497, OPS: 50.3, Avg(us): 247768, Min(us): 0, Max(us): 408319, 50th(us): 253567, 90th(us): 306175, 95th(us): 355327, 99th(us): 405503, 99.9th(us): 408063, 99.99th(us): 408319
COMMIT_ERROR - Takes(s): 29.8, Count: 167, OPS: 5.6, Avg(us): 179176, Min(us): 50464, Max(us): 406271, 50th(us): 152959, 90th(us): 304383, 95th(us): 353535, 99th(us): 355839, 99.9th(us): 406271, 99.99th(us): 406271
READ   - Takes(s): 30.2, Count: 9895, OPS: 328.0, Avg(us): 52606, Min(us): 5, Max(us): 204031, 50th(us): 50527, 90th(us): 51231, 95th(us): 51519, 99th(us): 151935, 99.9th(us): 203135, 99.99th(us): 203903
READ_ERROR - Takes(s): 29.5, Count: 105, OPS: 3.6, Avg(us): 163332, Min(us): 150528, Max(us): 253695, 50th(us): 152063, 90th(us): 202623, 95th(us): 202879, 99th(us): 252543, 99.9th(us): 253695, 99.99th(us): 253695
Start  - Takes(s): 30.2, Count: 1680, OPS: 55.6, Avg(us): 45, Min(us): 15, Max(us): 1217, 50th(us): 29, 90th(us): 52, 95th(us): 81, 99th(us): 489, 99.9th(us): 975, 99.99th(us): 1217
TOTAL  - Takes(s): 30.2, Count: 21204, OPS: 701.7, Avg(us): 126208, Min(us): 0, Max(us): 864255, 50th(us): 50463, 90th(us): 508927, 95th(us): 607743, 99th(us): 708607, 99.9th(us): 809471, 99.99th(us): 861695
TXN    - Takes(s): 29.8, Count: 1497, OPS: 50.3, Avg(us): 571183, Min(us): 302336, Max(us): 862207, 50th(us): 558079, 90th(us): 660479, 95th(us): 710143, 99th(us): 762367, 99.9th(us): 861695, 99.99th(us): 862207
TXN_ERROR - Takes(s): 29.8, Count: 167, OPS: 5.6, Avg(us): 496754, Min(us): 303104, Max(us): 809983, 50th(us): 458495, 90th(us): 657407, 95th(us): 660479, 99th(us): 759807, 99.9th(us): 809983, 99.99th(us): 809983
TxnGroup - Takes(s): 29.9, Count: 1664, OPS: 55.6, Avg(us): 558588, Min(us): 253312, Max(us): 864255, 50th(us): 558079, 90th(us): 659967, 95th(us): 709631, 99th(us): 762879, 99.9th(us): 859647, 99.99th(us): 864255
UPDATE - Takes(s): 30.2, Count: 4971, OPS: 164.8, Avg(us): 8, Min(us): 1, Max(us): 792, 50th(us): 4, 90th(us): 8, 95th(us): 11, 99th(us): 159, 99.9th(us): 415, 99.99th(us): 792
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  65
rollback failed
                                                     version mismatch  22
  rollback failed because the corresponding transaction has committed  18

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
       prepare phase failed: version mismatch     132
  transaction is aborted by other transaction      35
```

+ 64

```bash
----------------------------------
Run finished, takes 15.867326345s
COMMIT - Takes(s): 15.6, Count: 1445, OPS: 92.9, Avg(us): 245789, Min(us): 0, Max(us): 409599, 50th(us): 253439, 90th(us): 307199, 95th(us): 355839, 99th(us): 405503, 99.9th(us): 407807, 99.99th(us): 409599
COMMIT_ERROR - Takes(s): 15.5, Count: 219, OPS: 14.2, Avg(us): 167373, Min(us): 50560, Max(us): 406271, 50th(us): 152447, 90th(us): 303359, 95th(us): 305151, 99th(us): 356351, 99.9th(us): 406271, 99.99th(us): 406271
READ   - Takes(s): 15.8, Count: 9809, OPS: 620.2, Avg(us): 52927, Min(us): 5, Max(us): 253951, 50th(us): 50495, 90th(us): 51199, 95th(us): 51647, 99th(us): 151935, 99.9th(us): 203007, 99.99th(us): 253055
READ_ERROR - Takes(s): 15.3, Count: 191, OPS: 12.5, Avg(us): 167098, Min(us): 150528, Max(us): 254207, 50th(us): 152063, 90th(us): 202495, 95th(us): 203391, 99th(us): 253695, 99.9th(us): 254207, 99.99th(us): 254207
Start  - Takes(s): 15.9, Count: 1680, OPS: 105.9, Avg(us): 42, Min(us): 14, Max(us): 1310, 50th(us): 30, 90th(us): 52, 95th(us): 68, 99th(us): 381, 99.9th(us): 700, 99.99th(us): 1310
TOTAL  - Takes(s): 15.9, Count: 20978, OPS: 1321.9, Avg(us): 125535, Min(us): 0, Max(us): 911871, 50th(us): 50431, 90th(us): 509183, 95th(us): 607743, 99th(us): 709631, 99.9th(us): 812031, 99.99th(us): 865279
TXN    - Takes(s): 15.6, Count: 1445, OPS: 92.9, Avg(us): 575894, Min(us): 301312, Max(us): 864255, 50th(us): 559103, 90th(us): 662527, 95th(us): 711679, 99th(us): 809983, 99.9th(us): 861695, 99.99th(us): 864255
TXN_ERROR - Takes(s): 15.5, Count: 219, OPS: 14.2, Avg(us): 503710, Min(us): 352768, Max(us): 909823, 50th(us): 506367, 90th(us): 610815, 95th(us): 659967, 99th(us): 763391, 99.9th(us): 909823, 99.99th(us): 909823
TxnGroup - Takes(s): 15.6, Count: 1664, OPS: 106.9, Avg(us): 557020, Min(us): 302080, Max(us): 911871, 50th(us): 558079, 90th(us): 662015, 95th(us): 711167, 99th(us): 812031, 99.9th(us): 865279, 99.99th(us): 911871
UPDATE - Takes(s): 15.8, Count: 4935, OPS: 312.0, Avg(us): 7, Min(us): 1, Max(us): 684, 50th(us): 4, 90th(us): 7, 95th(us): 9, 99th(us): 30, 99.9th(us): 587, 99.99th(us): 684
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                                                     version mismatch  113
  rollback failed because the corresponding transaction has committed   45
rollback failed
  version mismatch  33

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
       prepare phase failed: version mismatch     184
  transaction is aborted by other transaction      35
```

+ 96

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 96
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 10.898556117s
COMMIT - Takes(s): 10.6, Count: 1351, OPS: 127.5, Avg(us): 241429, Min(us): 0, Max(us): 408319, 50th(us): 253055, 90th(us): 306175, 95th(us): 354559, 99th(us): 357375, 99.9th(us): 407295, 99.99th(us): 408319
COMMIT_ERROR - Takes(s): 10.5, Count: 281, OPS: 26.6, Avg(us): 171811, Min(us): 50528, Max(us): 406527, 50th(us): 153087, 90th(us): 303359, 95th(us): 305151, 99th(us): 355839, 99.9th(us): 406527, 99.99th(us): 406527
READ   - Takes(s): 10.8, Count: 9750, OPS: 898.8, Avg(us): 52850, Min(us): 5, Max(us): 254463, 50th(us): 50495, 90th(us): 51231, 95th(us): 51615, 99th(us): 152063, 99.9th(us): 203519, 99.99th(us): 253567
READ_ERROR - Takes(s): 10.4, Count: 250, OPS: 24.1, Avg(us): 166024, Min(us): 150528, Max(us): 254591, 50th(us): 151935, 90th(us): 202623, 95th(us): 252159, 99th(us): 253567, 99.9th(us): 254591, 99.99th(us): 254591
Start  - Takes(s): 10.9, Count: 1728, OPS: 158.5, Avg(us): 51, Min(us): 13, Max(us): 1566, 50th(us): 29, 90th(us): 53, 95th(us): 90, 99th(us): 620, 99.9th(us): 1010, 99.99th(us): 1566
TOTAL  - Takes(s): 10.9, Count: 20693, OPS: 1898.5, Avg(us): 121718, Min(us): 0, Max(us): 1011711, 50th(us): 50399, 90th(us): 507647, 95th(us): 607743, 99th(us): 709631, 99.9th(us): 812031, 99.99th(us): 909823
TXN    - Takes(s): 10.6, Count: 1351, OPS: 127.5, Avg(us): 575849, Min(us): 302080, Max(us): 862719, 50th(us): 559103, 90th(us): 707071, 95th(us): 711167, 99th(us): 764927, 99.9th(us): 862719, 99.99th(us): 862719
TXN_ERROR - Takes(s): 10.5, Count: 281, OPS: 26.6, Avg(us): 508099, Min(us): 353024, Max(us): 862719, 50th(us): 505855, 90th(us): 657407, 95th(us): 708607, 99th(us): 811007, 99.9th(us): 862719, 99.99th(us): 862719
TxnGroup - Takes(s): 10.6, Count: 1632, OPS: 153.3, Avg(us): 550953, Min(us): 252928, Max(us): 1011711, 50th(us): 557055, 90th(us): 707071, 95th(us): 712191, 99th(us): 811519, 99.9th(us): 909823, 99.99th(us): 1011711
UPDATE - Takes(s): 10.8, Count: 4881, OPS: 450.0, Avg(us): 8, Min(us): 1, Max(us): 1341, 50th(us): 4, 90th(us): 7, 95th(us): 10, 99th(us): 45, 99.9th(us): 871, 99.99th(us): 1341
Error Summary:

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  140
rollback failed
                                                     version mismatch  70
  rollback failed because the corresponding transaction has committed  39
                                                 get old state failed   1

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
       prepare phase failed: version mismatch     239
  transaction is aborted by other transaction      42
```

+ 128

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 8.265175714s
COMMIT - Takes(s): 7.9, Count: 1326, OPS: 168.7, Avg(us): 243505, Min(us): 0, Max(us): 414975, 50th(us): 253055, 90th(us): 306175, 95th(us): 354559, 99th(us): 403711, 99.9th(us): 413951, 99.99th(us): 414975
COMMIT_ERROR - Takes(s): 7.9, Count: 338, OPS: 42.8, Avg(us): 159607, Min(us): 50496, Max(us): 405503, 50th(us): 152319, 90th(us): 254847, 95th(us): 304383, 99th(us): 356095, 99.9th(us): 405503, 99.99th(us): 405503
READ   - Takes(s): 8.2, Count: 9733, OPS: 1184.9, Avg(us): 52594, Min(us): 5, Max(us): 253823, 50th(us): 50495, 90th(us): 51231, 95th(us): 51583, 99th(us): 151807, 99.9th(us): 202879, 99.99th(us): 252287
READ_ERROR - Takes(s): 7.8, Count: 267, OPS: 34.4, Avg(us): 167817, Min(us): 150528, Max(us): 254975, 50th(us): 152063, 90th(us): 203007, 95th(us): 252287, 99th(us): 254079, 99.9th(us): 254975, 99.99th(us): 254975
Start  - Takes(s): 8.3, Count: 1680, OPS: 203.2, Avg(us): 42, Min(us): 13, Max(us): 1108, 50th(us): 29, 90th(us): 44, 95th(us): 55, 99th(us): 561, 99.9th(us): 1008, 99.99th(us): 1108
TOTAL  - Takes(s): 8.3, Count: 20710, OPS: 2505.7, Avg(us): 120881, Min(us): 0, Max(us): 959999, 50th(us): 50399, 90th(us): 507647, 95th(us): 606719, 99th(us): 708607, 99.9th(us): 812543, 99.99th(us): 913919
TXN    - Takes(s): 7.9, Count: 1326, OPS: 168.7, Avg(us): 577275, Min(us): 301568, Max(us): 910847, 50th(us): 559103, 90th(us): 663551, 95th(us): 710143, 99th(us): 808959, 99.9th(us): 909311, 99.99th(us): 910847
TXN_ERROR - Takes(s): 7.9, Count: 338, OPS: 42.8, Avg(us): 495908, Min(us): 303104, Max(us): 811519, 50th(us): 459519, 90th(us): 610815, 95th(us): 659455, 99th(us): 760319, 99.9th(us): 811519, 99.99th(us): 811519
TxnGroup - Takes(s): 8.0, Count: 1664, OPS: 207.7, Avg(us): 542721, Min(us): 253824, Max(us): 959999, 50th(us): 556543, 90th(us): 661503, 95th(us): 712191, 99th(us): 812543, 99.9th(us): 913919, 99.99th(us): 959999
UPDATE - Takes(s): 8.2, Count: 4981, OPS: 606.4, Avg(us): 7, Min(us): 1, Max(us): 1877, 50th(us): 4, 90th(us): 6, 95th(us): 9, 99th(us): 28, 99.9th(us): 721, 99.99th(us): 1877
Error Summary:

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
       prepare phase failed: version mismatch     304
  transaction is aborted by other transaction      34

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  138
rollback failed
                                                     version mismatch  82
  rollback failed because the corresponding transaction has committed  45
                                                 get old state failed   2
```

#### Oreo

+ 8

```bash
----------------------------------
Run finished, takes 1m24.894928176s
COMMIT - Takes(s): 84.5, Count: 1608, OPS: 19.0, Avg(us): 100329, Min(us): 0, Max(us): 129471, 50th(us): 101951, 90th(us): 102399, 95th(us): 102527, 99th(us): 102783, 99.9th(us): 103615, 99.99th(us): 129471
COMMIT_ERROR - Takes(s): 83.3, Count: 56, OPS: 0.7, Avg(us): 53308, Min(us): 51008, Max(us): 101951, 50th(us): 51519, 90th(us): 51903, 95th(us): 52063, 99th(us): 101887, 99.9th(us): 101951, 99.99th(us): 101951
READ   - Takes(s): 84.8, Count: 9977, OPS: 117.6, Avg(us): 51112, Min(us): 6, Max(us): 79487, 50th(us): 51199, 90th(us): 51647, 95th(us): 51807, 99th(us): 52031, 99.9th(us): 52479, 99.99th(us): 79231
READ_ERROR - Takes(s): 70.7, Count: 23, OPS: 0.3, Avg(us): 51605, Min(us): 50784, Max(us): 52799, 50th(us): 51679, 90th(us): 52031, 95th(us): 52063, 99th(us): 52799, 99.9th(us): 52799, 99.99th(us): 52799
Start  - Takes(s): 84.9, Count: 1672, OPS: 19.7, Avg(us): 28, Min(us): 14, Max(us): 365, 50th(us): 27, 90th(us): 36, 95th(us): 40, 99th(us): 64, 99.9th(us): 274, 99.99th(us): 365
TOTAL  - Takes(s): 84.9, Count: 21489, OPS: 253.1, Avg(us): 93070, Min(us): 0, Max(us): 436735, 50th(us): 51135, 90th(us): 409087, 95th(us): 409599, 99th(us): 410367, 99.9th(us): 410879, 99.99th(us): 436479
TXN    - Takes(s): 84.5, Count: 1608, OPS: 19.0, Avg(us): 407118, Min(us): 306176, Max(us): 436735, 50th(us): 409343, 90th(us): 410111, 95th(us): 410367, 99th(us): 410879, 99.9th(us): 435711, 99.99th(us): 436735
TXN_ERROR - Takes(s): 83.3, Count: 56, OPS: 0.7, Avg(us): 360082, Min(us): 308736, Max(us): 410367, 50th(us): 358911, 90th(us): 359679, 95th(us): 359679, 99th(us): 409343, 99.9th(us): 410367, 99.99th(us): 410367
TxnGroup - Takes(s): 84.6, Count: 1664, OPS: 19.7, Avg(us): 405046, Min(us): 305664, Max(us): 436735, 50th(us): 409343, 90th(us): 410111, 95th(us): 410367, 99th(us): 410879, 99.9th(us): 435967, 99.99th(us): 436735
UPDATE - Takes(s): 84.8, Count: 4960, OPS: 58.5, Avg(us): 4, Min(us): 2, Max(us): 231, 50th(us): 4, 90th(us): 5, 95th(us): 6, 99th(us): 14, 99.9th(us): 141, 99.99th(us): 231
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
                             version mismatch  54
  transaction is aborted by other transaction   2

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  20
rollback failed
  version mismatch  3
```

+ 16

```bash
----------------------------------
Run finished, takes 42.21261555s
COMMIT - Takes(s): 41.9, Count: 1544, OPS: 36.9, Avg(us): 100452, Min(us): 0, Max(us): 104703, 50th(us): 102143, 90th(us): 102847, 95th(us): 103103, 99th(us): 103615, 99.9th(us): 104319, 99.99th(us): 104703
COMMIT_ERROR - Takes(s): 41.9, Count: 120, OPS: 2.9, Avg(us): 52540, Min(us): 50656, Max(us): 103551, 50th(us): 51647, 90th(us): 52319, 95th(us): 52671, 99th(us): 101887, 99.9th(us): 103551, 99.99th(us): 103551
READ   - Takes(s): 42.2, Count: 9979, OPS: 236.7, Avg(us): 51035, Min(us): 6, Max(us): 89087, 50th(us): 51135, 90th(us): 51743, 95th(us): 51935, 99th(us): 52287, 99.9th(us): 52767, 99.99th(us): 53119
READ_ERROR - Takes(s): 41.7, Count: 21, OPS: 0.5, Avg(us): 51865, Min(us): 50976, Max(us): 52991, 50th(us): 51871, 90th(us): 52351, 95th(us): 52671, 99th(us): 52991, 99.9th(us): 52991, 99.99th(us): 52991
Start  - Takes(s): 42.2, Count: 1680, OPS: 39.8, Avg(us): 29, Min(us): 14, Max(us): 354, 50th(us): 26, 90th(us): 39, 95th(us): 45, 99th(us): 137, 99.9th(us): 333, 99.99th(us): 354
TOTAL  - Takes(s): 42.2, Count: 21342, OPS: 505.6, Avg(us): 91939, Min(us): 0, Max(us): 447487, 50th(us): 51007, 90th(us): 408831, 95th(us): 409855, 99th(us): 410879, 99.9th(us): 412159, 99.99th(us): 412671
TXN    - Takes(s): 41.9, Count: 1544, OPS: 36.9, Avg(us): 406859, Min(us): 254976, Max(us): 447487, 50th(us): 409343, 90th(us): 410623, 95th(us): 410879, 99th(us): 412159, 99.9th(us): 412671, 99.99th(us): 447487
TXN_ERROR - Takes(s): 41.9, Count: 120, OPS: 2.9, Avg(us): 358067, Min(us): 306432, Max(us): 410111, 50th(us): 358911, 90th(us): 360191, 95th(us): 360959, 99th(us): 408319, 99.9th(us): 410111, 99.99th(us): 410111
TxnGroup - Takes(s): 42.0, Count: 1664, OPS: 39.7, Avg(us): 402357, Min(us): 254976, Max(us): 446975, 50th(us): 409343, 90th(us): 410623, 95th(us): 410879, 99th(us): 411391, 99.9th(us): 412415, 99.99th(us): 446975
UPDATE - Takes(s): 42.2, Count: 4931, OPS: 117.0, Avg(us): 4, Min(us): 1, Max(us): 612, 50th(us): 4, 90th(us): 5, 95th(us): 6, 99th(us): 21, 99.9th(us): 249, 99.99th(us): 612
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
                             version mismatch  118
  transaction is aborted by other transaction    2

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  19
rollback failed
  version mismatch  2
```

+ 32

```bash
----------------------------------
Run finished, takes 21.18831699s
COMMIT - Takes(s): 20.9, Count: 1499, OPS: 71.8, Avg(us): 101070, Min(us): 0, Max(us): 105855, 50th(us): 102527, 90th(us): 103871, 95th(us): 104191, 99th(us): 104831, 99.9th(us): 105727, 99.99th(us): 105855
COMMIT_ERROR - Takes(s): 20.8, Count: 165, OPS: 7.9, Avg(us): 53940, Min(us): 50720, Max(us): 103743, 50th(us): 51999, 90th(us): 53247, 95th(us): 54271, 99th(us): 103167, 99.9th(us): 103743, 99.99th(us): 103743
READ   - Takes(s): 21.1, Count: 9969, OPS: 471.6, Avg(us): 50980, Min(us): 6, Max(us): 53983, 50th(us): 51007, 90th(us): 51775, 95th(us): 52031, 99th(us): 52543, 99.9th(us): 53247, 99.99th(us): 53823
READ_ERROR - Takes(s): 19.2, Count: 31, OPS: 1.6, Avg(us): 52141, Min(us): 50656, Max(us): 53823, 50th(us): 52031, 90th(us): 53279, 95th(us): 53503, 99th(us): 53823, 99.9th(us): 53823, 99.99th(us): 53823
Start  - Takes(s): 21.2, Count: 1680, OPS: 79.3, Avg(us): 33, Min(us): 13, Max(us): 609, 50th(us): 25, 90th(us): 39, 95th(us): 51, 99th(us): 311, 99.9th(us): 580, 99.99th(us): 609
TOTAL  - Takes(s): 21.2, Count: 21255, OPS: 1003.0, Avg(us): 91116, Min(us): 0, Max(us): 415487, 50th(us): 50911, 90th(us): 408575, 95th(us): 409855, 99th(us): 411391, 99.9th(us): 413695, 99.99th(us): 415231
TXN    - Takes(s): 20.9, Count: 1499, OPS: 71.8, Avg(us): 407129, Min(us): 304384, Max(us): 415487, 50th(us): 409343, 90th(us): 411135, 95th(us): 411903, 99th(us): 414719, 99.9th(us): 415487, 99.99th(us): 415487
TXN_ERROR - Takes(s): 20.8, Count: 165, OPS: 7.9, Avg(us): 359755, Min(us): 307968, Max(us): 410367, 50th(us): 358911, 90th(us): 360959, 95th(us): 362495, 99th(us): 409343, 99.9th(us): 410367, 99.99th(us): 410367
TxnGroup - Takes(s): 20.9, Count: 1664, OPS: 79.5, Avg(us): 400589, Min(us): 258944, Max(us): 413183, 50th(us): 409343, 90th(us): 410879, 95th(us): 411391, 99th(us): 412159, 99.9th(us): 412927, 99.99th(us): 413183
UPDATE - Takes(s): 21.1, Count: 4944, OPS: 233.9, Avg(us): 5, Min(us): 2, Max(us): 1230, 50th(us): 4, 90th(us): 5, 95th(us): 6, 99th(us): 20, 99.9th(us): 426, 99.99th(us): 1230
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
                             version mismatch  159
  transaction is aborted by other transaction    6

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  24
  read failed due to unknown txn status   5
rollback failed
  version mismatch  2
```

+ 64

```bash
----------------------------------
Run finished, takes 10.584252583s
COMMIT - Takes(s): 10.3, Count: 1385, OPS: 134.8, Avg(us): 100599, Min(us): 0, Max(us): 110207, 50th(us): 102399, 90th(us): 104383, 95th(us): 105215, 99th(us): 107903, 99.9th(us): 108991, 99.99th(us): 110207
COMMIT_ERROR - Takes(s): 10.2, Count: 279, OPS: 27.3, Avg(us): 54458, Min(us): 50816, Max(us): 106559, 50th(us): 51967, 90th(us): 54271, 95th(us): 57855, 99th(us): 104511, 99.9th(us): 106559, 99.99th(us): 106559
READ   - Takes(s): 10.5, Count: 9977, OPS: 947.1, Avg(us): 50891, Min(us): 5, Max(us): 56351, 50th(us): 50847, 90th(us): 51775, 95th(us): 52191, 99th(us): 53151, 99.9th(us): 54687, 99.99th(us): 55743
READ_ERROR - Takes(s): 7.9, Count: 23, OPS: 2.9, Avg(us): 52198, Min(us): 50432, Max(us): 53919, 50th(us): 52031, 90th(us): 53343, 95th(us): 53407, 99th(us): 53919, 99.9th(us): 53919, 99.99th(us): 53919
Start  - Takes(s): 10.6, Count: 1680, OPS: 158.7, Avg(us): 31, Min(us): 14, Max(us): 1005, 50th(us): 26, 90th(us): 40, 95th(us): 48, 99th(us): 106, 99.9th(us): 767, 99.99th(us): 1005
TOTAL  - Takes(s): 10.6, Count: 21071, OPS: 1990.6, Avg(us): 88601, Min(us): 0, Max(us): 422143, 50th(us): 50719, 90th(us): 407295, 95th(us): 409087, 99th(us): 411903, 99.9th(us): 419839, 99.99th(us): 421631
TXN    - Takes(s): 10.3, Count: 1385, OPS: 134.8, Avg(us): 406282, Min(us): 303360, Max(us): 422143, 50th(us): 408575, 90th(us): 411647, 95th(us): 413183, 99th(us): 420095, 99.9th(us): 421887, 99.99th(us): 422143
TXN_ERROR - Takes(s): 10.2, Count: 279, OPS: 27.3, Avg(us): 359283, Min(us): 305152, Max(us): 412415, 50th(us): 358143, 90th(us): 361727, 95th(us): 370943, 99th(us): 411135, 99.9th(us): 412415, 99.99th(us): 412415
TxnGroup - Takes(s): 10.3, Count: 1664, OPS: 161.2, Avg(us): 394881, Min(us): 261120, Max(us): 415743, 50th(us): 408063, 90th(us): 410879, 95th(us): 411903, 99th(us): 413183, 99.9th(us): 415487, 99.99th(us): 415743
UPDATE - Takes(s): 10.5, Count: 4980, OPS: 472.9, Avg(us): 4, Min(us): 1, Max(us): 867, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 16, 99.9th(us): 249, 99.99th(us): 867
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
                             version mismatch  267
  transaction is aborted by other transaction   12

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
  version mismatch  11
rollback failed
                       version mismatch  6
  read failed due to unknown txn status  6
```

+ 96

```bash
----------------------------------
Run finished, takes 7.066818208s
COMMIT - Takes(s): 6.7, Count: 1297, OPS: 194.9, Avg(us): 101546, Min(us): 0, Max(us): 108671, 50th(us): 103103, 90th(us): 105471, 95th(us): 106175, 99th(us): 107391, 99.9th(us): 108607, 99.99th(us): 108671
COMMIT_ERROR - Takes(s): 6.7, Count: 335, OPS: 50.0, Avg(us): 55347, Min(us): 50784, Max(us): 108031, 50th(us): 52191, 90th(us): 55103, 95th(us): 101951, 99th(us): 105855, 99.9th(us): 108031, 99.99th(us): 108031
READ   - Takes(s): 7.0, Count: 9977, OPS: 1422.3, Avg(us): 50925, Min(us): 6, Max(us): 56639, 50th(us): 50815, 90th(us): 51807, 95th(us): 52351, 99th(us): 53695, 99.9th(us): 55071, 99.99th(us): 56607
READ_ERROR - Takes(s): 6.6, Count: 23, OPS: 3.5, Avg(us): 52799, Min(us): 51744, Max(us): 54975, 50th(us): 52607, 90th(us): 54143, 95th(us): 54495, 99th(us): 54975, 99.9th(us): 54975, 99.99th(us): 54975
Start  - Takes(s): 7.1, Count: 1728, OPS: 244.5, Avg(us): 30, Min(us): 13, Max(us): 761, 50th(us): 26, 90th(us): 40, 95th(us): 49, 99th(us): 201, 99.9th(us): 548, 99.99th(us): 761
TOTAL  - Takes(s): 7.1, Count: 20922, OPS: 2960.1, Avg(us): 86435, Min(us): 0, Max(us): 422399, 50th(us): 50687, 90th(us): 407551, 95th(us): 409599, 99th(us): 412671, 99.9th(us): 419071, 99.99th(us): 421887
TXN    - Takes(s): 6.7, Count: 1297, OPS: 194.8, Avg(us): 407390, Min(us): 303616, Max(us): 422399, 50th(us): 409343, 90th(us): 412671, 95th(us): 414719, 99th(us): 420863, 99.9th(us): 422143, 99.99th(us): 422399
TXN_ERROR - Takes(s): 6.7, Count: 335, OPS: 50.0, Avg(us): 360663, Min(us): 305408, Max(us): 417535, 50th(us): 358655, 90th(us): 365055, 95th(us): 369919, 99th(us): 412927, 99.9th(us): 417535, 99.99th(us): 417535
TxnGroup - Takes(s): 6.8, Count: 1632, OPS: 241.5, Avg(us): 392253, Min(us): 303616, Max(us): 417791, 50th(us): 408319, 90th(us): 411391, 95th(us): 412415, 99th(us): 414207, 99.9th(us): 415999, 99.99th(us): 417791
UPDATE - Takes(s): 7.0, Count: 4991, OPS: 711.5, Avg(us): 4, Min(us): 1, Max(us): 1292, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 16, 99.9th(us): 186, 99.99th(us): 1292
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
                             version mismatch  317
  transaction is aborted by other transaction   18

  Operation:   READ
       Error  Count
       -----  -----
rollForward failed
                       version mismatch  13
  read failed due to unknown txn status   5
rollback failed
  version mismatch  5
```

+ 128

```bash
----------------------------------
Run finished, takes 5.36257347s
COMMIT - Takes(s): 5.1, Count: 1294, OPS: 256.1, Avg(us): 101400, Min(us): 0, Max(us): 113983, 50th(us): 102655, 90th(us): 104895, 95th(us): 106303, 99th(us): 110527, 99.9th(us): 113471, 99.99th(us): 113983
COMMIT_ERROR - Takes(s): 5.0, Count: 370, OPS: 74.1, Avg(us): 53805, Min(us): 50528, Max(us): 105343, 50th(us): 51967, 90th(us): 54655, 95th(us): 59391, 99th(us): 104127, 99.9th(us): 105343, 99.99th(us): 105343
READ   - Takes(s): 5.3, Count: 9969, OPS: 1876.8, Avg(us): 51071, Min(us): 6, Max(us): 81471, 50th(us): 50815, 90th(us): 51775, 95th(us): 52319, 99th(us): 54815, 99.9th(us): 80127, 99.99th(us): 81215
READ_ERROR - Takes(s): 5.0, Count: 31, OPS: 6.3, Avg(us): 51849, Min(us): 50496, Max(us): 57823, 50th(us): 51519, 90th(us): 53407, 95th(us): 53631, 99th(us): 57823, 99.9th(us): 57823, 99.99th(us): 57823
Start  - Takes(s): 5.4, Count: 1680, OPS: 313.3, Avg(us): 30, Min(us): 13, Max(us): 1479, 50th(us): 24, 90th(us): 37, 95th(us): 47, 99th(us): 228, 99.9th(us): 728, 99.99th(us): 1479
TOTAL  - Takes(s): 5.4, Count: 20878, OPS: 3891.9, Avg(us): 87076, Min(us): 0, Max(us): 451583, 50th(us): 50719, 90th(us): 407039, 95th(us): 409087, 99th(us): 412671, 99.9th(us): 447231, 99.99th(us): 450559
TXN    - Takes(s): 5.1, Count: 1294, OPS: 256.0, Avg(us): 407602, Min(us): 303360, Max(us): 451583, 50th(us): 408575, 90th(us): 412415, 95th(us): 417279, 99th(us): 448767, 99.9th(us): 451071, 99.99th(us): 451583
TXN_ERROR - Takes(s): 5.0, Count: 370, OPS: 74.1, Avg(us): 361656, Min(us): 304384, Max(us): 414463, 50th(us): 358399, 90th(us): 367103, 95th(us): 396031, 99th(us): 408319, 99.9th(us): 414463, 99.99th(us): 414463
TxnGroup - Takes(s): 5.1, Count: 1664, OPS: 326.0, Avg(us): 390702, Min(us): 257152, Max(us): 420351, 50th(us): 407807, 90th(us): 411391, 95th(us): 412927, 99th(us): 415743, 99.9th(us): 418559, 99.99th(us): 420351
UPDATE - Takes(s): 5.3, Count: 4977, OPS: 937.0, Avg(us): 4, Min(us): 1, Max(us): 898, 50th(us): 4, 90th(us): 5, 95th(us): 5, 99th(us): 16, 99.9th(us): 247, 99.99th(us): 898
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
                             version mismatch  361
  transaction is aborted by other transaction    9

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status     18
rollForward failed
  version mismatch  8
rollback failed
  version mismatch  5
```

### Pure Write

#### Cherry Garcia

+ 8

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 8
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 2m30.706074233s
COMMIT - Takes(s): 150.0, Count: 1491, OPS: 9.9, Avg(us): 727216, Min(us): 608256, Max(us): 1015295, 50th(us): 711679, 90th(us): 813055, 95th(us): 862207, 99th(us): 865279, 99.9th(us): 1014783, 99.99th(us): 1015295
COMMIT_ERROR - Takes(s): 150.0, Count: 173, OPS: 1.2, Avg(us): 624410, Min(us): 252800, Max(us): 867327, 50th(us): 708095, 90th(us): 811519, 95th(us): 863231, 99th(us): 864767, 99.9th(us): 867327, 99.99th(us): 867327
Start  - Takes(s): 150.7, Count: 1672, OPS: 11.1, Avg(us): 34, Min(us): 15, Max(us): 340, 50th(us): 27, 90th(us): 45, 95th(us): 68, 99th(us): 207, 99.9th(us): 295, 99.99th(us): 340
TOTAL  - Takes(s): 150.7, Count: 16318, OPS: 108.3, Avg(us): 205659, Min(us): 1, Max(us): 1015295, 50th(us): 3, 90th(us): 712191, 95th(us): 713215, 99th(us): 863743, 99.9th(us): 914943, 99.99th(us): 1015295
TXN    - Takes(s): 150.0, Count: 1491, OPS: 9.9, Avg(us): 727369, Min(us): 608256, Max(us): 1015295, 50th(us): 711679, 90th(us): 813055, 95th(us): 862207, 99th(us): 865791, 99.9th(us): 1015295, 99.99th(us): 1015295
TXN_ERROR - Takes(s): 150.0, Count: 173, OPS: 1.2, Avg(us): 624583, Min(us): 252928, Max(us): 867327, 50th(us): 708607, 90th(us): 811519, 95th(us): 863231, 99th(us): 864767, 99.9th(us): 867327, 99.99th(us): 867327
TxnGroup - Takes(s): 150.7, Count: 1664, OPS: 11.0, Avg(us): 713385, Min(us): 134, Max(us): 1015295, 50th(us): 711679, 90th(us): 813055, 95th(us): 862719, 99th(us): 865279, 99.9th(us): 965631, 99.99th(us): 1015295
UPDATE - Takes(s): 150.7, Count: 10000, OPS: 66.4, Avg(us): 3, Min(us): 1, Max(us): 411, 50th(us): 2, 90th(us): 4, 95th(us): 6, 99th(us): 20, 99.9th(us): 169, 99.99th(us): 316
Error Summary:

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
  transaction is aborted by other transaction      88
prepare phase failed: rollback failed
  version mismatch  25
prepare phase failed: rollForward failed
                                                                           version mismatch  22
                                                     prepare phase failed: version mismatch  21
  prepare phase failed: rollback failed because the corresponding transaction has committed  13
                                                 prepare phase failed: get old state failed   4
```

+ 16

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 16
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 1m17.114135565s
COMMIT - Takes(s): 76.4, Count: 1301, OPS: 17.0, Avg(us): 736267, Min(us): 606720, Max(us): 1016831, 50th(us): 711679, 90th(us): 858111, 95th(us): 864255, 99th(us): 917503, 99.9th(us): 1014783, 99.99th(us): 1016831
COMMIT_ERROR - Takes(s): 76.6, Count: 363, OPS: 4.7, Avg(us): 635319, Min(us): 201856, Max(us): 1014271, 50th(us): 663551, 90th(us): 862719, 95th(us): 865279, 99th(us): 916479, 99.9th(us): 1014271, 99.99th(us): 1014271
Start  - Takes(s): 77.1, Count: 1680, OPS: 21.8, Avg(us): 37, Min(us): 14, Max(us): 437, 50th(us): 29, 90th(us): 49, 95th(us): 67, 99th(us): 247, 99.9th(us): 363, 99.99th(us): 437
TOTAL  - Takes(s): 77.1, Count: 15946, OPS: 206.8, Avg(us): 193986, Min(us): 1, Max(us): 1016831, 50th(us): 3, 90th(us): 712703, 95th(us): 810495, 99th(us): 865279, 99.9th(us): 966655, 99.99th(us): 1016831
TXN    - Takes(s): 76.4, Count: 1301, OPS: 17.0, Avg(us): 736461, Min(us): 607232, Max(us): 1016831, 50th(us): 712191, 90th(us): 858111, 95th(us): 864255, 99th(us): 917503, 99.9th(us): 1014783, 99.99th(us): 1016831
TXN_ERROR - Takes(s): 76.6, Count: 363, OPS: 4.7, Avg(us): 635499, Min(us): 201984, Max(us): 1014271, 50th(us): 663551, 90th(us): 862719, 95th(us): 865279, 99th(us): 916479, 99.9th(us): 1014271, 99.99th(us): 1014271
TxnGroup - Takes(s): 77.1, Count: 1664, OPS: 21.6, Avg(us): 707445, Min(us): 550, Max(us): 1016831, 50th(us): 711679, 90th(us): 861183, 95th(us): 864767, 99th(us): 917503, 99.9th(us): 1014783, 99.99th(us): 1016831
UPDATE - Takes(s): 77.1, Count: 10000, OPS: 129.7, Avg(us): 3, Min(us): 1, Max(us): 316, 50th(us): 2, 90th(us): 5, 95th(us): 6, 99th(us): 22, 99.9th(us): 180, 99.99th(us): 266
Error Summary:

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
  transaction is aborted by other transaction     165
       prepare phase failed: version mismatch      60
prepare phase failed: rollForward failed
  version mismatch  60
prepare phase failed: rollback failed
                                                                           version mismatch  44
  prepare phase failed: rollback failed because the corresponding transaction has committed  28
                                                 prepare phase failed: get old state failed   6
```

+ 32

```
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 32
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 38.615460492s
COMMIT - Takes(s): 37.9, Count: 1076, OPS: 28.4, Avg(us): 742572, Min(us): 607232, Max(us): 1166335, 50th(us): 711167, 90th(us): 861183, 95th(us): 863231, 99th(us): 964607, 99.9th(us): 1068031, 99.99th(us): 1166335
COMMIT_ERROR - Takes(s): 38.2, Count: 588, OPS: 15.4, Avg(us): 643121, Min(us): 150912, Max(us): 1069055, 50th(us): 707071, 90th(us): 862207, 95th(us): 864767, 99th(us): 1014783, 99.9th(us): 1068031, 99.99th(us): 1069055
Start  - Takes(s): 38.6, Count: 1680, OPS: 43.5, Avg(us): 44, Min(us): 14, Max(us): 1163, 50th(us): 30, 90th(us): 51, 95th(us): 73, 99th(us): 341, 99.9th(us): 1086, 99.99th(us): 1163
TOTAL  - Takes(s): 38.6, Count: 15496, OPS: 401.3, Avg(us): 177632, Min(us): 1, Max(us): 1166335, 50th(us): 3, 90th(us): 711679, 95th(us): 811007, 99th(us): 864255, 99.9th(us): 1015295, 99.99th(us): 1166335
TXN    - Takes(s): 37.9, Count: 1076, OPS: 28.4, Avg(us): 742808, Min(us): 607232, Max(us): 1166335, 50th(us): 711167, 90th(us): 861695, 95th(us): 863743, 99th(us): 964607, 99.9th(us): 1069055, 99.99th(us): 1166335
TXN_ERROR - Takes(s): 38.2, Count: 588, OPS: 15.4, Avg(us): 643341, Min(us): 151040, Max(us): 1070079, 50th(us): 707583, 90th(us): 862207, 95th(us): 864767, 99th(us): 1014783, 99.9th(us): 1068031, 99.99th(us): 1070079
TxnGroup - Takes(s): 38.6, Count: 1664, OPS: 43.1, Avg(us): 693635, Min(us): 1132, Max(us): 1166335, 50th(us): 710655, 90th(us): 861695, 95th(us): 864255, 99th(us): 1012223, 99.9th(us): 1068031, 99.99th(us): 1166335
UPDATE - Takes(s): 38.6, Count: 10000, OPS: 259.0, Avg(us): 3, Min(us): 1, Max(us): 559, 50th(us): 3, 90th(us): 5, 95th(us): 6, 99th(us): 20, 99.9th(us): 241, 99.99th(us): 519
Error Summary:

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
  transaction is aborted by other transaction     262
prepare phase failed: rollback failed
                        version mismatch  122
  prepare phase failed: version mismatch  110
prepare phase failed: rollForward failed
                                                                           version mismatch  62
  prepare phase failed: rollback failed because the corresponding transaction has committed  31
                                                 prepare phase failed: get old state failed   1
```

+ 64

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 19.372295512s
COMMIT - Takes(s): 18.7, Count: 941, OPS: 50.4, Avg(us): 747209, Min(us): 606208, Max(us): 1115135, 50th(us): 710655, 90th(us): 861695, 95th(us): 863743, 99th(us): 964607, 99.9th(us): 1064959, 99.99th(us): 1115135
COMMIT_ERROR - Takes(s): 19.1, Count: 723, OPS: 37.8, Avg(us): 636001, Min(us): 100864, Max(us): 1267711, 50th(us): 658943, 90th(us): 861695, 95th(us): 910847, 99th(us): 1014271, 99.9th(us): 1165311, 99.99th(us): 1267711
Start  - Takes(s): 19.4, Count: 1680, OPS: 86.7, Avg(us): 42, Min(us): 14, Max(us): 1139, 50th(us): 30, 90th(us): 48, 95th(us): 58, 99th(us): 371, 99.9th(us): 1124, 99.99th(us): 1139
TOTAL  - Takes(s): 19.4, Count: 15226, OPS: 785.9, Avg(us): 165848, Min(us): 1, Max(us): 1267711, 50th(us): 3, 90th(us): 711167, 95th(us): 811007, 99th(us): 865791, 99.9th(us): 1015295, 99.99th(us): 1115135
TXN    - Takes(s): 18.7, Count: 941, OPS: 50.4, Avg(us): 747475, Min(us): 606208, Max(us): 1115135, 50th(us): 710655, 90th(us): 861695, 95th(us): 864767, 99th(us): 964607, 99.9th(us): 1064959, 99.99th(us): 1115135
TXN_ERROR - Takes(s): 19.1, Count: 723, OPS: 37.8, Avg(us): 636311, Min(us): 101248, Max(us): 1267711, 50th(us): 658943, 90th(us): 862207, 95th(us): 910847, 99th(us): 1014783, 99.9th(us): 1166335, 99.99th(us): 1267711
TxnGroup - Takes(s): 19.4, Count: 1664, OPS: 85.9, Avg(us): 672233, Min(us): 70, Max(us): 1267711, 50th(us): 710143, 90th(us): 861695, 95th(us): 864255, 99th(us): 1013247, 99.9th(us): 1115135, 99.99th(us): 1267711
UPDATE - Takes(s): 19.4, Count: 10000, OPS: 516.2, Avg(us): 4, Min(us): 1, Max(us): 882, 50th(us): 3, 90th(us): 4, 95th(us): 6, 99th(us): 18, 99.9th(us): 322, 99.99th(us): 595
Error Summary:

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
  transaction is aborted by other transaction     263
prepare phase failed: rollback failed
                        version mismatch  224
  prepare phase failed: version mismatch  127
prepare phase failed: rollForward failed
                                                                           version mismatch  84
  prepare phase failed: rollback failed because the corresponding transaction has committed  25
```

+ 96

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 96
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 12.828479549s
COMMIT - Takes(s): 12.1, Count: 820, OPS: 67.7, Avg(us): 746179, Min(us): 607744, Max(us): 1061887, 50th(us): 710655, 90th(us): 861183, 95th(us): 862719, 99th(us): 963583, 99.9th(us): 1014271, 99.99th(us): 1061887
COMMIT_ERROR - Takes(s): 12.5, Count: 812, OPS: 64.8, Avg(us): 629398, Min(us): 101504, Max(us): 1167359, 50th(us): 609791, 90th(us): 861183, 95th(us): 911359, 99th(us): 1012735, 99.9th(us): 1165311, 99.99th(us): 1167359
Start  - Takes(s): 12.8, Count: 1728, OPS: 134.7, Avg(us): 53, Min(us): 14, Max(us): 1415, 50th(us): 30, 90th(us): 47, 95th(us): 64, 99th(us): 735, 99.9th(us): 1363, 99.99th(us): 1415
TOTAL  - Takes(s): 12.8, Count: 15000, OPS: 1169.2, Avg(us): 152093, Min(us): 1, Max(us): 1167359, 50th(us): 3, 90th(us): 710655, 95th(us): 809983, 99th(us): 863231, 99.9th(us): 1013247, 99.99th(us): 1161215
TXN    - Takes(s): 12.1, Count: 820, OPS: 67.7, Avg(us): 746597, Min(us): 608256, Max(us): 1061887, 50th(us): 710655, 90th(us): 861695, 95th(us): 863231, 99th(us): 963583, 99.9th(us): 1014271, 99.99th(us): 1061887
TXN_ERROR - Takes(s): 12.5, Count: 812, OPS: 64.8, Avg(us): 629781, Min(us): 101632, Max(us): 1167359, 50th(us): 610303, 90th(us): 861695, 95th(us): 911359, 99th(us): 1012735, 99.9th(us): 1166335, 99.99th(us): 1167359
TxnGroup - Takes(s): 12.8, Count: 1632, OPS: 127.2, Avg(us): 647791, Min(us): 69, Max(us): 1167359, 50th(us): 709631, 90th(us): 861183, 95th(us): 863231, 99th(us): 1012735, 99.9th(us): 1161215, 99.99th(us): 1167359
UPDATE - Takes(s): 12.8, Count: 10000, OPS: 779.5, Avg(us): 3, Min(us): 1, Max(us): 1205, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 16, 99.9th(us): 379, 99.99th(us): 755
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: rollback failed
                             version mismatch  286
  transaction is aborted by other transaction  258
       prepare phase failed: version mismatch  152
prepare phase failed: rollForward failed
                                                                           version mismatch  79
  prepare phase failed: rollback failed because the corresponding transaction has committed  37
```

+ 128

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 10.29207057s
COMMIT - Takes(s): 9.6, Count: 803, OPS: 83.8, Avg(us): 748260, Min(us): 606720, Max(us): 1114111, 50th(us): 709631, 90th(us): 860671, 95th(us): 863231, 99th(us): 963583, 99.9th(us): 1112063, 99.99th(us): 1114111
COMMIT_ERROR - Takes(s): 10.1, Count: 861, OPS: 85.4, Avg(us): 623001, Min(us): 101568, Max(us): 1062911, 50th(us): 609791, 90th(us): 861183, 95th(us): 911871, 99th(us): 1012735, 99.9th(us): 1061887, 99.99th(us): 1062911
Start  - Takes(s): 10.3, Count: 1680, OPS: 163.2, Avg(us): 51, Min(us): 12, Max(us): 1289, 50th(us): 31, 90th(us): 47, 95th(us): 61, 99th(us): 883, 99.9th(us): 1161, 99.99th(us): 1289
TOTAL  - Takes(s): 10.3, Count: 14950, OPS: 1452.5, Avg(us): 150829, Min(us): 1, Max(us): 1114111, 50th(us): 3, 90th(us): 709631, 95th(us): 809471, 99th(us): 870399, 99.9th(us): 1014783, 99.99th(us): 1114111
TXN    - Takes(s): 9.6, Count: 803, OPS: 83.8, Avg(us): 748677, Min(us): 606720, Max(us): 1114111, 50th(us): 710143, 90th(us): 861183, 95th(us): 866303, 99th(us): 964095, 99.9th(us): 1112063, 99.99th(us): 1114111
TXN_ERROR - Takes(s): 10.1, Count: 861, OPS: 85.4, Avg(us): 623624, Min(us): 101760, Max(us): 1062911, 50th(us): 611327, 90th(us): 861695, 95th(us): 911871, 99th(us): 1012735, 99.9th(us): 1061887, 99.99th(us): 1062911
TxnGroup - Takes(s): 10.3, Count: 1664, OPS: 161.7, Avg(us): 632650, Min(us): 68, Max(us): 1114111, 50th(us): 708607, 90th(us): 861183, 95th(us): 868351, 99th(us): 1011711, 99.9th(us): 1062911, 99.99th(us): 1114111
UPDATE - Takes(s): 10.3, Count: 10000, OPS: 971.5, Avg(us): 4, Min(us): 1, Max(us): 1207, 50th(us): 3, 90th(us): 4, 95th(us): 5, 99th(us): 16, 99.9th(us): 408, 99.99th(us): 1155
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: rollback failed
                             version mismatch  331
  transaction is aborted by other transaction  251
       prepare phase failed: version mismatch  174
prepare phase failed: rollForward failed
                                                                           version mismatch  74
  prepare phase failed: rollback failed because the corresponding transaction has committed  31
```

#### Oreo

+ 8

```bash
----------------------------------
Run finished, takes 20.405475881s
COMMIT - Takes(s): 20.3, Count: 1479, OPS: 72.9, Avg(us): 102810, Min(us): 101312, Max(us): 106495, 50th(us): 102783, 90th(us): 103807, 95th(us): 104127, 99th(us): 104895, 99.9th(us): 106367, 99.99th(us): 106495
COMMIT_ERROR - Takes(s): 20.4, Count: 185, OPS: 9.1, Avg(us): 52266, Min(us): 51168, Max(us): 55519, 50th(us): 52159, 90th(us): 53247, 95th(us): 53503, 99th(us): 54175, 99.9th(us): 55519, 99.99th(us): 55519
Start  - Takes(s): 20.4, Count: 1672, OPS: 81.9, Avg(us): 26, Min(us): 14, Max(us): 310, 50th(us): 24, 90th(us): 33, 95th(us): 37, 99th(us): 49, 99.9th(us): 240, 99.99th(us): 310
TOTAL  - Takes(s): 20.4, Count: 16294, OPS: 798.5, Avg(us): 28569, Min(us): 1, Max(us): 106751, 50th(us): 3, 90th(us): 103103, 95th(us): 103551, 99th(us): 104383, 99.9th(us): 105599, 99.99th(us): 106623
TXN    - Takes(s): 20.3, Count: 1479, OPS: 72.9, Avg(us): 102948, Min(us): 101440, Max(us): 106623, 50th(us): 102911, 90th(us): 103935, 95th(us): 104319, 99th(us): 105279, 99.9th(us): 106623, 99.99th(us): 106623
TXN_ERROR - Takes(s): 20.4, Count: 185, OPS: 9.1, Avg(us): 52405, Min(us): 51264, Max(us): 56063, 50th(us): 52287, 90th(us): 53343, 95th(us): 53663, 99th(us): 54975, 99.9th(us): 56063, 99.99th(us): 56063
TxnGroup - Takes(s): 20.4, Count: 1664, OPS: 81.5, Avg(us): 96834, Min(us): 410, Max(us): 106751, 50th(us): 102783, 90th(us): 103871, 95th(us): 104255, 99th(us): 105023, 99.9th(us): 106559, 99.99th(us): 106751
UPDATE - Takes(s): 20.4, Count: 10000, OPS: 490.1, Avg(us): 2, Min(us): 1, Max(us): 234, 50th(us): 2, 90th(us): 3, 95th(us): 4, 99th(us): 7, 99.9th(us): 42, 99.99th(us): 140
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  118
prepare phase failed: Remote prepare failed
  version mismatch  32
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  32
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  3
```

+ 16

```bash
----------------------------------
Run finished, takes 9.941989687s
COMMIT - Takes(s): 9.8, Count: 1329, OPS: 135.1, Avg(us): 102783, Min(us): 101184, Max(us): 108799, 50th(us): 102655, 90th(us): 103807, 95th(us): 104255, 99th(us): 105855, 99.9th(us): 108543, 99.99th(us): 108799
COMMIT_ERROR - Takes(s): 9.9, Count: 335, OPS: 33.9, Avg(us): 52289, Min(us): 51072, Max(us): 57055, 50th(us): 52127, 90th(us): 53247, 95th(us): 53695, 99th(us): 55455, 99.9th(us): 57055, 99.99th(us): 57055
Start  - Takes(s): 9.9, Count: 1680, OPS: 169.0, Avg(us): 24, Min(us): 14, Max(us): 291, 50th(us): 22, 90th(us): 32, 95th(us): 36, 99th(us): 55, 99.9th(us): 259, 99.99th(us): 291
TOTAL  - Takes(s): 9.9, Count: 16002, OPS: 1609.3, Avg(us): 26643, Min(us): 1, Max(us): 110463, 50th(us): 3, 90th(us): 102975, 95th(us): 103423, 99th(us): 104575, 99.9th(us): 107967, 99.99th(us): 110463
TXN    - Takes(s): 9.8, Count: 1329, OPS: 135.1, Avg(us): 102921, Min(us): 101312, Max(us): 110463, 50th(us): 102783, 90th(us): 103999, 95th(us): 104447, 99th(us): 107199, 99.9th(us): 110463, 99.99th(us): 110463
TXN_ERROR - Takes(s): 9.9, Count: 335, OPS: 33.9, Avg(us): 52413, Min(us): 51200, Max(us): 59423, 50th(us): 52223, 90th(us): 53439, 95th(us): 53951, 99th(us): 55711, 99.9th(us): 59423, 99.99th(us): 59423
TxnGroup - Takes(s): 9.9, Count: 1664, OPS: 167.4, Avg(us): 91885, Min(us): 810, Max(us): 109119, 50th(us): 102527, 90th(us): 103807, 95th(us): 104255, 99th(us): 105599, 99.9th(us): 108671, 99.99th(us): 109119
UPDATE - Takes(s): 9.9, Count: 10000, OPS: 1005.9, Avg(us): 2, Min(us): 1, Max(us): 255, 50th(us): 2, 90th(us): 3, 95th(us): 4, 99th(us): 7, 99.9th(us): 71, 99.99th(us): 191
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  292
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  22
prepare phase failed: Remote prepare failed
  version mismatch  16
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  5
```

+ 32

```bash
----------------------------------
Run finished, takes 4.831936736s
COMMIT - Takes(s): 4.7, Count: 1178, OPS: 249.2, Avg(us): 103095, Min(us): 101056, Max(us): 114175, 50th(us): 102719, 90th(us): 104319, 95th(us): 106879, 99th(us): 110655, 99.9th(us): 113983, 99.99th(us): 114175
COMMIT_ERROR - Takes(s): 4.8, Count: 486, OPS: 101.8, Avg(us): 52540, Min(us): 51040, Max(us): 62271, 50th(us): 52095, 90th(us): 53983, 95th(us): 55135, 99th(us): 60991, 99.9th(us): 62271, 99.99th(us): 62271
Start  - Takes(s): 4.8, Count: 1680, OPS: 347.6, Avg(us): 27, Min(us): 14, Max(us): 605, 50th(us): 20, 90th(us): 31, 95th(us): 38, 99th(us): 222, 99.9th(us): 532, 99.99th(us): 605
TOTAL  - Takes(s): 4.8, Count: 15700, OPS: 3248.6, Avg(us): 24675, Min(us): 1, Max(us): 114879, 50th(us): 3, 90th(us): 102911, 95th(us): 103487, 99th(us): 108095, 99.9th(us): 112959, 99.99th(us): 114623
TXN    - Takes(s): 4.7, Count: 1178, OPS: 249.2, Avg(us): 103263, Min(us): 101120, Max(us): 114623, 50th(us): 102783, 90th(us): 104447, 95th(us): 107391, 99th(us): 111935, 99.9th(us): 114303, 99.99th(us): 114623
TXN_ERROR - Takes(s): 4.8, Count: 486, OPS: 101.8, Avg(us): 52717, Min(us): 51136, Max(us): 63199, 50th(us): 52191, 90th(us): 54143, 95th(us): 55327, 99th(us): 61983, 99.9th(us): 63199, 99.99th(us): 63199
TxnGroup - Takes(s): 4.8, Count: 1664, OPS: 344.4, Avg(us): 86683, Min(us): 438, Max(us): 114879, 50th(us): 102335, 90th(us): 103999, 95th(us): 105407, 99th(us): 110527, 99.9th(us): 113919, 99.99th(us): 114879
UPDATE - Takes(s): 4.8, Count: 10000, OPS: 2069.4, Avg(us): 2, Min(us): 1, Max(us): 540, 50th(us): 2, 90th(us): 3, 95th(us): 4, 99th(us): 6, 99.9th(us): 47, 99.99th(us): 188
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  434
prepare phase failed: Remote prepare failed
  version mismatch  26
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  17
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  9
```

+ 64

```bash
----------------------------------
Run finished, takes 2.456053844s
COMMIT - Takes(s): 2.3, Count: 978, OPS: 416.8, Avg(us): 104305, Min(us): 101184, Max(us): 120127, 50th(us): 102847, 90th(us): 109183, 95th(us): 112511, 99th(us): 116607, 99.9th(us): 119999, 99.99th(us): 120127
COMMIT_ERROR - Takes(s): 2.4, Count: 686, OPS: 286.2, Avg(us): 53609, Min(us): 51040, Max(us): 68095, 50th(us): 52223, 90th(us): 57695, 95th(us): 62079, 99th(us): 66815, 99.9th(us): 67903, 99.99th(us): 68095
Start  - Takes(s): 2.5, Count: 1680, OPS: 683.7, Avg(us): 27, Min(us): 13, Max(us): 682, 50th(us): 20, 90th(us): 31, 95th(us): 37, 99th(us): 218, 99.9th(us): 663, 99.99th(us): 682
TOTAL  - Takes(s): 2.5, Count: 15300, OPS: 6225.9, Avg(us): 22108, Min(us): 1, Max(us): 121023, 50th(us): 3, 90th(us): 102847, 95th(us): 104895, 99th(us): 112767, 99.9th(us): 119423, 99.99th(us): 121023
TXN    - Takes(s): 2.3, Count: 978, OPS: 416.8, Avg(us): 104540, Min(us): 101312, Max(us): 121023, 50th(us): 102975, 90th(us): 109311, 95th(us): 113983, 99th(us): 119423, 99.9th(us): 121023, 99.99th(us): 121023
TXN_ERROR - Takes(s): 2.4, Count: 686, OPS: 286.2, Avg(us): 53857, Min(us): 51104, Max(us): 70719, 50th(us): 52351, 90th(us): 57951, 95th(us): 63487, 99th(us): 69055, 99.9th(us): 70655, 99.99th(us): 70719
TxnGroup - Takes(s): 2.5, Count: 1664, OPS: 677.4, Avg(us): 80490, Min(us): 82, Max(us): 120383, 50th(us): 102015, 90th(us): 107071, 95th(us): 109887, 99th(us): 116479, 99.9th(us): 119551, 99.99th(us): 120383
UPDATE - Takes(s): 2.5, Count: 10000, OPS: 4070.8, Avg(us): 2, Min(us): 1, Max(us): 200, 50th(us): 2, 90th(us): 3, 95th(us): 4, 99th(us): 5, 99.9th(us): 33, 99.99th(us): 173
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  607
prepare phase failed: Remote prepare failed
  version mismatch  42
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  19
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  18
```

+ 96

```bash
----------------------------------
Run finished, takes 1.678863607s
COMMIT - Takes(s): 1.6, Count: 834, OPS: 530.6, Avg(us): 106673, Min(us): 101248, Max(us): 132223, 50th(us): 103743, 90th(us): 115903, 95th(us): 120767, 99th(us): 127551, 99.9th(us): 130943, 99.99th(us): 132223
COMMIT_ERROR - Takes(s): 1.6, Count: 798, OPS: 493.3, Avg(us): 55786, Min(us): 51040, Max(us): 84479, 50th(us): 52927, 90th(us): 65439, 95th(us): 70719, 99th(us): 76863, 99.9th(us): 83967, 99.99th(us): 84479
Start  - Takes(s): 1.7, Count: 1728, OPS: 1029.0, Avg(us): 24, Min(us): 14, Max(us): 249, 50th(us): 21, 90th(us): 31, 95th(us): 37, 99th(us): 59, 99.9th(us): 226, 99.99th(us): 249
TOTAL  - Takes(s): 1.7, Count: 15028, OPS: 8947.7, Avg(us): 20310, Min(us): 1, Max(us): 132479, 50th(us): 3, 90th(us): 103231, 95th(us): 107071, 99th(us): 121151, 99.9th(us): 128831, 99.99th(us): 132223
TXN    - Takes(s): 1.6, Count: 834, OPS: 530.5, Avg(us): 107175, Min(us): 101376, Max(us): 132479, 50th(us): 103871, 90th(us): 118591, 95th(us): 125695, 99th(us): 128831, 99.9th(us): 131071, 99.99th(us): 132479
TXN_ERROR - Takes(s): 1.6, Count: 798, OPS: 493.5, Avg(us): 56219, Min(us): 51168, Max(us): 84607, 50th(us): 53055, 90th(us): 66687, 95th(us): 75007, 99th(us): 79551, 99.9th(us): 84031, 99.99th(us): 84607
TxnGroup - Takes(s): 1.7, Count: 1632, OPS: 972.3, Avg(us): 77705, Min(us): 354, Max(us): 132351, 50th(us): 72127, 90th(us): 111487, 95th(us): 116159, 99th(us): 125759, 99.9th(us): 129279, 99.99th(us): 132351
UPDATE - Takes(s): 1.7, Count: 10000, OPS: 5954.2, Avg(us): 2, Min(us): 1, Max(us): 335, 50th(us): 2, 90th(us): 3, 95th(us): 4, 99th(us): 6, 99.9th(us): 27, 99.99th(us): 290
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  675
prepare phase failed: Remote prepare failed
  version mismatch  58
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  35
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  30
```

+ 128

```bash
----------------------------------
Run finished, takes 1.370666457s
COMMIT - Takes(s): 1.3, Count: 769, OPS: 609.8, Avg(us): 109454, Min(us): 101312, Max(us): 159359, 50th(us): 107455, 90th(us): 117759, 95th(us): 124799, 99th(us): 139647, 99.9th(us): 157823, 99.99th(us): 159359
COMMIT_ERROR - Takes(s): 1.3, Count: 895, OPS: 687.4, Avg(us): 58533, Min(us): 51168, Max(us): 95999, 50th(us): 55519, 90th(us): 67455, 95th(us): 86335, 99th(us): 93951, 99.9th(us): 95743, 99.99th(us): 95999
Start  - Takes(s): 1.4, Count: 1680, OPS: 1225.3, Avg(us): 25, Min(us): 14, Max(us): 270, 50th(us): 22, 90th(us): 31, 95th(us): 36, 99th(us): 55, 99.9th(us): 266, 99.99th(us): 270
TOTAL  - Takes(s): 1.4, Count: 14882, OPS: 10851.6, Avg(us): 19898, Min(us): 1, Max(us): 166783, 50th(us): 3, 90th(us): 105535, 95th(us): 110079, 99th(us): 122623, 99.9th(us): 148223, 99.99th(us): 166655
TXN    - Takes(s): 1.3, Count: 769, OPS: 610.1, Avg(us): 110030, Min(us): 101440, Max(us): 166783, 50th(us): 107647, 90th(us): 118335, 95th(us): 128255, 99th(us): 148095, 99.9th(us): 166655, 99.99th(us): 166783
TXN_ERROR - Takes(s): 1.3, Count: 895, OPS: 688.7, Avg(us): 59209, Min(us): 51232, Max(us): 101183, 50th(us): 55615, 90th(us): 67583, 95th(us): 94271, 99th(us): 100799, 99.9th(us): 101119, 99.99th(us): 101183
TxnGroup - Takes(s): 1.4, Count: 1664, OPS: 1212.6, Avg(us): 76493, Min(us): 274, Max(us): 159487, 50th(us): 64063, 90th(us): 112319, 95th(us): 117311, 99th(us): 134271, 99.9th(us): 150527, 99.99th(us): 159487
UPDATE - Takes(s): 1.4, Count: 10000, OPS: 7289.4, Avg(us): 2, Min(us): 1, Max(us): 269, 50th(us): 2, 90th(us): 3, 95th(us): 4, 99th(us): 6, 99.9th(us): 109, 99.99th(us): 256
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  750
prepare phase failed: Remote prepare failed
  version mismatch  88
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  30
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  27
```



## High Write

#### Cherry Garcia - 5

+ 50ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 19.631827472s
COMMIT - Takes(s): 19.0, Count: 1284, OPS: 67.5, Avg(us): 632933, Min(us): 506368, Max(us): 911359, 50th(us): 608767, 90th(us): 757759, 95th(us): 760831, 99th(us): 808447, 99.9th(us): 910335, 99.99th(us): 911359
COMMIT_ERROR - Takes(s): 19.3, Count: 700, OPS: 36.3, Avg(us): 534636, Min(us): 102784, Max(us): 1012223, 50th(us): 557055, 90th(us): 759807, 95th(us): 761855, 99th(us): 910335, 99.9th(us): 964607, 99.99th(us): 1012223
Start  - Takes(s): 19.6, Count: 2048, OPS: 104.3, Avg(us): 47, Min(us): 13, Max(us): 2223, 50th(us): 30, 90th(us): 49, 95th(us): 70, 99th(us): 494, 99.9th(us): 1454, 99.99th(us): 2223
TOTAL  - Takes(s): 19.6, Count: 16600, OPS: 845.5, Avg(us): 167085, Min(us): 1, Max(us): 1012223, 50th(us): 4, 90th(us): 609791, 95th(us): 706559, 99th(us): 761855, 99.9th(us): 863231, 99.99th(us): 923135
TXN    - Takes(s): 19.0, Count: 1284, OPS: 67.5, Avg(us): 633206, Min(us): 506624, Max(us): 911359, 50th(us): 609279, 90th(us): 757759, 95th(us): 761343, 99th(us): 808447, 99.9th(us): 910335, 99.99th(us): 911359
TXN_ERROR - Takes(s): 19.3, Count: 700, OPS: 36.3, Avg(us): 534897, Min(us): 103360, Max(us): 1012223, 50th(us): 557055, 90th(us): 759807, 95th(us): 762367, 99th(us): 910335, 99.9th(us): 964607, 99.99th(us): 1012223
TxnGroup - Takes(s): 19.6, Count: 1984, OPS: 101.1, Avg(us): 578496, Min(us): 59, Max(us): 1012223, 50th(us): 608255, 90th(us): 758783, 95th(us): 761343, 99th(us): 859647, 99.9th(us): 923135, 99.99th(us): 1012223
UPDATE - Takes(s): 19.6, Count: 10000, OPS: 509.3, Avg(us): 5, Min(us): 1, Max(us): 1370, 50th(us): 3, 90th(us): 5, 95th(us): 6, 99th(us): 19, 99.9th(us): 1221, 99.99th(us): 1332
Error Summary:

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
  transaction is aborted by other transaction     228
prepare phase failed: rollback failed
                        version mismatch  207
  prepare phase failed: version mismatch  121
prepare phase failed: rollForward failed
                                                                           version mismatch  110
  prepare phase failed: rollback failed because the corresponding transaction has committed   34
```

+ 100ms

```bash

```

+ 150ms

```bash

```

+ 200ms

```bash

```

#### Cherry Garcia - 4

+ 50ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 16.742923732s
COMMIT - Takes(s): 16.2, Count: 1445, OPS: 89.0, Avg(us): 528938, Min(us): 404224, Max(us): 813055, 50th(us): 507391, 90th(us): 610303, 95th(us): 658943, 99th(us): 663551, 99.9th(us): 811007, 99.99th(us): 813055
COMMIT_ERROR - Takes(s): 16.6, Count: 539, OPS: 32.4, Avg(us): 430312, Min(us): 101120, Max(us): 762879, 50th(us): 454911, 90th(us): 610303, 95th(us): 659455, 99th(us): 758783, 99.9th(us): 759807, 99.99th(us): 762879
Start  - Takes(s): 16.7, Count: 2048, OPS: 122.3, Avg(us): 52, Min(us): 14, Max(us): 1347, 50th(us): 29, 90th(us): 50, 95th(us): 80, 99th(us): 722, 99.9th(us): 1217, 99.99th(us): 1347
TOTAL  - Takes(s): 16.7, Count: 14922, OPS: 891.2, Avg(us): 167121, Min(us): 1, Max(us): 813567, 50th(us): 6, 90th(us): 508927, 95th(us): 606719, 99th(us): 660479, 99.9th(us): 761855, 99.99th(us): 813055
TXN    - Takes(s): 16.2, Count: 1445, OPS: 89.0, Avg(us): 529171, Min(us): 404224, Max(us): 813567, 50th(us): 507647, 90th(us): 610303, 95th(us): 659455, 99th(us): 664063, 99.9th(us): 811007, 99.99th(us): 813567
TXN_ERROR - Takes(s): 16.6, Count: 539, OPS: 32.4, Avg(us): 430570, Min(us): 101504, Max(us): 762879, 50th(us): 455167, 90th(us): 610303, 95th(us): 659455, 99th(us): 758783, 99.9th(us): 760319, 99.99th(us): 762879
TxnGroup - Takes(s): 16.7, Count: 1984, OPS: 118.5, Avg(us): 486225, Min(us): 56, Max(us): 813055, 50th(us): 507135, 90th(us): 609791, 95th(us): 658943, 99th(us): 663551, 99.9th(us): 809983, 99.99th(us): 813055
UPDATE - Takes(s): 16.7, Count: 8000, OPS: 477.8, Avg(us): 5, Min(us): 1, Max(us): 1319, 50th(us): 3, 90th(us): 5, 95th(us): 6, 99th(us): 22, 99.9th(us): 517, 99.99th(us): 1261
Error Summary:

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
  transaction is aborted by other transaction     153
prepare phase failed: rollback failed
  version mismatch  145
prepare phase failed: rollForward failed
                                                                           version mismatch  105
                                                     prepare phase failed: version mismatch   93
  prepare phase failed: rollback failed because the corresponding transaction has committed   41
                                                 prepare phase failed: get old state failed    2
```

+ 100ms

```bash

```

+ 150ms

```bash

```

+ 200ms

```bash

```

#### Cherry Garcia - 3

+ 50ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: false
Read Strategy: p
ConcurrentOptimizationLevel: 0
AsyncLevel: 1
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 0s ConnAdditionalLatency: 50ms
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 13.286063038s
COMMIT - Takes(s): 12.9, Count: 1590, OPS: 123.4, Avg(us): 416888, Min(us): 303360, Max(us): 711167, 50th(us): 405759, 90th(us): 409087, 95th(us): 509183, 99th(us): 559615, 99.9th(us): 658943, 99.99th(us): 711167
COMMIT_ERROR - Takes(s): 13.1, Count: 394, OPS: 30.1, Avg(us): 345355, Min(us): 101056, Max(us): 656383, 50th(us): 353791, 90th(us): 505599, 95th(us): 555519, 99th(us): 558591, 99.9th(us): 656383, 99.99th(us): 656383
Start  - Takes(s): 13.3, Count: 2032, OPS: 152.9, Avg(us): 54, Min(us): 13, Max(us): 1301, 50th(us): 30, 90th(us): 59, 95th(us): 142, 99th(us): 606, 99.9th(us): 1284, 99.99th(us): 1301
TOTAL  - Takes(s): 13.3, Count: 13196, OPS: 993.2, Avg(us): 159079, Min(us): 1, Max(us): 711167, 50th(us): 27, 90th(us): 407039, 95th(us): 408575, 99th(us): 557055, 99.9th(us): 655871, 99.99th(us): 711167
TXN    - Takes(s): 12.9, Count: 1590, OPS: 123.5, Avg(us): 417113, Min(us): 303616, Max(us): 711167, 50th(us): 405759, 90th(us): 409599, 95th(us): 509439, 99th(us): 560127, 99.9th(us): 658943, 99.99th(us): 711167
TXN_ERROR - Takes(s): 13.1, Count: 394, OPS: 30.1, Avg(us): 345600, Min(us): 101184, Max(us): 656383, 50th(us): 354047, 90th(us): 505855, 95th(us): 556031, 99th(us): 559103, 99.9th(us): 656383, 99.99th(us): 656383
TxnGroup - Takes(s): 13.3, Count: 1984, OPS: 149.3, Avg(us): 389618, Min(us): 40, Max(us): 711167, 50th(us): 405503, 90th(us): 410623, 95th(us): 509951, 99th(us): 559615, 99.9th(us): 659455, 99.99th(us): 711167
UPDATE - Takes(s): 13.3, Count: 6000, OPS: 451.6, Avg(us): 5, Min(us): 1, Max(us): 1411, 50th(us): 3, 90th(us): 5, 95th(us): 7, 99th(us): 28, 99.9th(us): 561, 99.99th(us): 1276
Error Summary:

                                   Operation:  COMMIT
                                        Error   Count
                                        -----   -----
  transaction is aborted by other transaction     104
prepare phase failed: rollForward failed
  version mismatch  99
prepare phase failed: rollback failed
                                                                           version mismatch  86
                                                     prepare phase failed: version mismatch  81
  prepare phase failed: rollback failed because the corresponding transaction has committed  23
                                                 prepare phase failed: get old state failed   1
```

+ 100ms

```bash

```

+ 150ms

```bash

```

+ 200ms

```bash

```

#### Oreo - 5

+ 50ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 50ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 3.072550709s
COMMIT - Takes(s): 3.0, Count: 1324, OPS: 447.1, Avg(us): 104512, Min(us): 101056, Max(us): 146431, 50th(us): 102591, 90th(us): 105279, 95th(us): 117375, 99th(us): 143615, 99.9th(us): 145791, 99.99th(us): 146431
COMMIT_ERROR - Takes(s): 3.0, Count: 660, OPS: 219.4, Avg(us): 53146, Min(us): 50912, Max(us): 78207, 50th(us): 51935, 90th(us): 54207, 95th(us): 64447, 99th(us): 73151, 99.9th(us): 75839, 99.99th(us): 78207
Start  - Takes(s): 3.1, Count: 2048, OPS: 666.6, Avg(us): 28, Min(us): 13, Max(us): 1163, 50th(us): 20, 90th(us): 31, 95th(us): 42, 99th(us): 278, 99.9th(us): 713, 99.99th(us): 1163
TOTAL  - Takes(s): 3.1, Count: 16680, OPS: 5426.9, Avg(us): 26705, Min(us): 1, Max(us): 147967, 50th(us): 3, 90th(us): 102847, 95th(us): 103679, 99th(us): 119231, 99.9th(us): 146175, 99.99th(us): 147583
TXN    - Takes(s): 3.0, Count: 1324, OPS: 447.1, Avg(us): 104709, Min(us): 101120, Max(us): 147583, 50th(us): 102655, 90th(us): 105535, 95th(us): 117951, 99th(us): 146047, 99.9th(us): 147583, 99.99th(us): 147583
TXN_ERROR - Takes(s): 3.0, Count: 660, OPS: 219.3, Avg(us): 53342, Min(us): 50976, Max(us): 78271, 50th(us): 52031, 90th(us): 54335, 95th(us): 65343, 99th(us): 74623, 99.9th(us): 76927, 99.99th(us): 78271
TxnGroup - Takes(s): 3.1, Count: 1984, OPS: 645.5, Avg(us): 84854, Min(us): 65, Max(us): 147967, 50th(us): 102079, 90th(us): 104319, 95th(us): 109439, 99th(us): 143871, 99.9th(us): 146431, 99.99th(us): 147967
UPDATE - Takes(s): 3.1, Count: 10000, OPS: 3253.9, Avg(us): 2, Min(us): 1, Max(us): 241, 50th(us): 2, 90th(us): 3, 95th(us): 4, 99th(us): 6, 99.9th(us): 39, 99.99th(us): 201
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  599
prepare phase failed: Remote prepare failed
  version mismatch  34
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  21
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  6
```

+ 100ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 100ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 5.898737536s
COMMIT - Takes(s): 5.7, Count: 1292, OPS: 227.0, Avg(us): 203230, Min(us): 201088, Max(us): 219519, 50th(us): 202623, 90th(us): 204543, 95th(us): 206975, 99th(us): 215935, 99.9th(us): 218239, 99.99th(us): 219519
COMMIT_ERROR - Takes(s): 5.8, Count: 692, OPS: 119.5, Avg(us): 102790, Min(us): 100928, Max(us): 118719, 50th(us): 102079, 90th(us): 104127, 95th(us): 108287, 99th(us): 115839, 99.9th(us): 118399, 99.99th(us): 118719
Start  - Takes(s): 5.9, Count: 2048, OPS: 347.1, Avg(us): 33, Min(us): 14, Max(us): 1116, 50th(us): 21, 90th(us): 32, 95th(us): 38, 99th(us): 564, 99.9th(us): 1090, 99.99th(us): 1116
TOTAL  - Takes(s): 5.9, Count: 16616, OPS: 2817.2, Avg(us): 51047, Min(us): 1, Max(us): 221183, 50th(us): 3, 90th(us): 202879, 95th(us): 203647, 99th(us): 209279, 99.9th(us): 218367, 99.99th(us): 221055
TXN    - Takes(s): 5.7, Count: 1292, OPS: 227.0, Avg(us): 203412, Min(us): 201216, Max(us): 221183, 50th(us): 202751, 90th(us): 204671, 95th(us): 207359, 99th(us): 218623, 99.9th(us): 221183, 99.99th(us): 221183
TXN_ERROR - Takes(s): 5.8, Count: 692, OPS: 119.5, Avg(us): 102988, Min(us): 100992, Max(us): 121343, 50th(us): 102207, 90th(us): 104191, 95th(us): 108351, 99th(us): 119551, 99.9th(us): 121215, 99.99th(us): 121343
TxnGroup - Takes(s): 5.9, Count: 1984, OPS: 336.4, Avg(us): 162665, Min(us): 62, Max(us): 219903, 50th(us): 202111, 90th(us): 204031, 95th(us): 205183, 99th(us): 215807, 99.9th(us): 218367, 99.99th(us): 219903
UPDATE - Takes(s): 5.9, Count: 10000, OPS: 1695.6, Avg(us): 2, Min(us): 1, Max(us): 239, 50th(us): 2, 90th(us): 3, 95th(us): 4, 99th(us): 6, 99.9th(us): 32, 99.99th(us): 238
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  630
prepare phase failed: Remote prepare failed
  version mismatch  38
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  17
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  7
```

+ 150ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 150ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 8.675932056s
COMMIT - Takes(s): 8.4, Count: 1339, OPS: 160.0, Avg(us): 303231, Min(us): 301056, Max(us): 318463, 50th(us): 302847, 90th(us): 304639, 95th(us): 307199, 99th(us): 315135, 99.9th(us): 317439, 99.99th(us): 318463
COMMIT_ERROR - Takes(s): 8.5, Count: 645, OPS: 75.7, Avg(us): 152888, Min(us): 150784, Max(us): 167679, 50th(us): 152191, 90th(us): 154623, 95th(us): 159743, 99th(us): 164607, 99.9th(us): 165375, 99.99th(us): 167679
Start  - Takes(s): 8.7, Count: 2048, OPS: 236.0, Avg(us): 26, Min(us): 14, Max(us): 410, 50th(us): 22, 90th(us): 33, 95th(us): 41, 99th(us): 112, 99.9th(us): 254, 99.99th(us): 410
TOTAL  - Takes(s): 8.7, Count: 16710, OPS: 1925.8, Avg(us): 77795, Min(us): 1, Max(us): 320255, 50th(us): 3, 90th(us): 303103, 95th(us): 303871, 99th(us): 309247, 99.9th(us): 317439, 99.99th(us): 318975
TXN    - Takes(s): 8.4, Count: 1339, OPS: 160.0, Avg(us): 303448, Min(us): 301056, Max(us): 320255, 50th(us): 302847, 90th(us): 305151, 95th(us): 307455, 99th(us): 317439, 99.9th(us): 318975, 99.99th(us): 320255
TXN_ERROR - Takes(s): 8.5, Count: 645, OPS: 75.7, Avg(us): 153138, Min(us): 150912, Max(us): 168063, 50th(us): 152191, 90th(us): 154751, 95th(us): 160511, 99th(us): 167679, 99.9th(us): 168063, 99.99th(us): 168063
TxnGroup - Takes(s): 8.7, Count: 1984, OPS: 228.7, Avg(us): 245737, Min(us): 120, Max(us): 318719, 50th(us): 302335, 90th(us): 304383, 95th(us): 305663, 99th(us): 314879, 99.9th(us): 317439, 99.99th(us): 318719
UPDATE - Takes(s): 8.7, Count: 10000, OPS: 1152.4, Avg(us): 2, Min(us): 1, Max(us): 305, 50th(us): 2, 90th(us): 3, 95th(us): 4, 99th(us): 7, 99.9th(us): 169, 99.99th(us): 257
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  568
prepare phase failed: Remote prepare failed
  version mismatch  38
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  32
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  7
```

+ 200ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 200ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 11.291504403s
COMMIT - Takes(s): 10.9, Count: 1302, OPS: 119.6, Avg(us): 403137, Min(us): 401152, Max(us): 417023, 50th(us): 402687, 90th(us): 404991, 95th(us): 407551, 99th(us): 414207, 99.9th(us): 416255, 99.99th(us): 417023
COMMIT_ERROR - Takes(s): 11.1, Count: 682, OPS: 61.5, Avg(us): 202680, Min(us): 200576, Max(us): 217087, 50th(us): 202111, 90th(us): 204159, 95th(us): 206207, 99th(us): 214143, 99.9th(us): 216319, 99.99th(us): 217087
Start  - Takes(s): 11.3, Count: 2048, OPS: 181.4, Avg(us): 32, Min(us): 14, Max(us): 1417, 50th(us): 22, 90th(us): 33, 95th(us): 42, 99th(us): 269, 99.9th(us): 1309, 99.99th(us): 1417
TOTAL  - Takes(s): 11.3, Count: 16636, OPS: 1473.1, Avg(us): 101676, Min(us): 1, Max(us): 420095, 50th(us): 3, 90th(us): 402943, 95th(us): 403711, 99th(us): 408575, 99.9th(us): 416511, 99.99th(us): 419327
TXN    - Takes(s): 10.9, Count: 1302, OPS: 119.6, Avg(us): 403334, Min(us): 401152, Max(us): 420095, 50th(us): 402687, 90th(us): 405247, 95th(us): 408063, 99th(us): 417535, 99.9th(us): 419327, 99.99th(us): 420095
TXN_ERROR - Takes(s): 11.1, Count: 682, OPS: 61.5, Avg(us): 202879, Min(us): 200704, Max(us): 219263, 50th(us): 202111, 90th(us): 204287, 95th(us): 206335, 99th(us): 217727, 99.9th(us): 219263, 99.99th(us): 219263
TxnGroup - Takes(s): 11.3, Count: 1984, OPS: 175.7, Avg(us): 323271, Min(us): 113, Max(us): 417535, 50th(us): 402175, 90th(us): 404479, 95th(us): 405759, 99th(us): 412415, 99.9th(us): 416511, 99.99th(us): 417535
UPDATE - Takes(s): 11.3, Count: 10000, OPS: 885.5, Avg(us): 2, Min(us): 1, Max(us): 214, 50th(us): 2, 90th(us): 3, 95th(us): 4, 99th(us): 7, 99.9th(us): 67, 99.99th(us): 212
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  616
prepare phase failed: Remote prepare failed
  version mismatch  35
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  19
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  12
```

#### Oreo - 4

+ 50ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 50ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 3.095304381s
COMMIT - Takes(s): 3.0, Count: 1471, OPS: 491.5, Avg(us): 102768, Min(us): 101120, Max(us): 113535, 50th(us): 102399, 90th(us): 103999, 95th(us): 105023, 99th(us): 110783, 99.9th(us): 113087, 99.99th(us): 113535
COMMIT_ERROR - Takes(s): 3.0, Count: 513, OPS: 168.9, Avg(us): 52272, Min(us): 50752, Max(us): 61759, 50th(us): 51903, 90th(us): 53279, 95th(us): 54399, 99th(us): 60479, 99.9th(us): 61471, 99.99th(us): 61759
Start  - Takes(s): 3.1, Count: 2048, OPS: 661.4, Avg(us): 24, Min(us): 14, Max(us): 365, 50th(us): 21, 90th(us): 31, 95th(us): 38, 99th(us): 70, 99.9th(us): 289, 99.99th(us): 365
TOTAL  - Takes(s): 3.1, Count: 14974, OPS: 4835.6, Avg(us): 31758, Min(us): 1, Max(us): 116159, 50th(us): 4, 90th(us): 102847, 95th(us): 103551, 99th(us): 106879, 99.9th(us): 114495, 99.99th(us): 116095
TXN    - Takes(s): 3.0, Count: 1471, OPS: 491.6, Avg(us): 102967, Min(us): 101184, Max(us): 116159, 50th(us): 102527, 90th(us): 104127, 95th(us): 105343, 99th(us): 114495, 99.9th(us): 116095, 99.99th(us): 116159
TXN_ERROR - Takes(s): 3.0, Count: 513, OPS: 168.9, Avg(us): 52462, Min(us): 50816, Max(us): 65855, 50th(us): 52031, 90th(us): 53407, 95th(us): 54847, 99th(us): 64831, 99.9th(us): 65791, 99.99th(us): 65855
TxnGroup - Takes(s): 3.1, Count: 1984, OPS: 641.0, Avg(us): 87120, Min(us): 414, Max(us): 113791, 50th(us): 102143, 90th(us): 103871, 95th(us): 104639, 99th(us): 110399, 99.9th(us): 113215, 99.99th(us): 113791
UPDATE - Takes(s): 3.1, Count: 8000, OPS: 2584.1, Avg(us): 2, Min(us): 1, Max(us): 207, 50th(us): 2, 90th(us): 3, 95th(us): 4, 99th(us): 7, 99.9th(us): 33, 99.99th(us): 202
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  455
prepare phase failed: Remote prepare failed
  version mismatch  29
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  25
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  4
```

+ 100ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 100ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 6.004946156s
COMMIT - Takes(s): 5.8, Count: 1449, OPS: 249.9, Avg(us): 203256, Min(us): 201088, Max(us): 220031, 50th(us): 202623, 90th(us): 205311, 95th(us): 207615, 99th(us): 214655, 99.9th(us): 219263, 99.99th(us): 220031
COMMIT_ERROR - Takes(s): 5.9, Count: 535, OPS: 90.7, Avg(us): 102688, Min(us): 100864, Max(us): 116223, 50th(us): 102015, 90th(us): 104255, 95th(us): 106751, 99th(us): 115647, 99.9th(us): 115967, 99.99th(us): 116223
Start  - Takes(s): 6.0, Count: 2048, OPS: 340.9, Avg(us): 25, Min(us): 14, Max(us): 385, 50th(us): 23, 90th(us): 32, 95th(us): 38, 99th(us): 57, 99.9th(us): 220, 99.99th(us): 385
TOTAL  - Takes(s): 6.0, Count: 14930, OPS: 2485.2, Avg(us): 62112, Min(us): 1, Max(us): 221567, 50th(us): 4, 90th(us): 203263, 95th(us): 204159, 99th(us): 210047, 99.9th(us): 218367, 99.99th(us): 221311
TXN    - Takes(s): 5.8, Count: 1449, OPS: 249.9, Avg(us): 203438, Min(us): 201216, Max(us): 221567, 50th(us): 202751, 90th(us): 205567, 95th(us): 207871, 99th(us): 218367, 99.9th(us): 221311, 99.99th(us): 221567
TXN_ERROR - Takes(s): 5.9, Count: 535, OPS: 90.7, Avg(us): 102886, Min(us): 100928, Max(us): 118335, 50th(us): 102079, 90th(us): 104383, 95th(us): 107903, 99th(us): 117567, 99.9th(us): 118271, 99.99th(us): 118335
TxnGroup - Takes(s): 6.0, Count: 1984, OPS: 330.3, Avg(us): 170345, Min(us): 350, Max(us): 220159, 50th(us): 202239, 90th(us): 204543, 95th(us): 206719, 99th(us): 214655, 99.9th(us): 217727, 99.99th(us): 220159
UPDATE - Takes(s): 6.0, Count: 8000, OPS: 1331.7, Avg(us): 2, Min(us): 1, Max(us): 233, 50th(us): 2, 90th(us): 3, 95th(us): 4, 99th(us): 7, 99.9th(us): 82, 99.99th(us): 188
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  468
prepare phase failed: Remote prepare failed
  version mismatch  26
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  22
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  19
```

+ 150ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 150ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 8.650489963s
COMMIT - Takes(s): 8.3, Count: 1456, OPS: 174.5, Avg(us): 303128, Min(us): 301056, Max(us): 314623, 50th(us): 302591, 90th(us): 304895, 95th(us): 307199, 99th(us): 312575, 99.9th(us): 314367, 99.99th(us): 314623
COMMIT_ERROR - Takes(s): 8.5, Count: 528, OPS: 62.2, Avg(us): 152566, Min(us): 150784, Max(us): 164479, 50th(us): 152063, 90th(us): 154111, 95th(us): 156671, 99th(us): 162815, 99.9th(us): 164351, 99.99th(us): 164479
Start  - Takes(s): 8.7, Count: 2048, OPS: 236.8, Avg(us): 31, Min(us): 13, Max(us): 806, 50th(us): 23, 90th(us): 36, 95th(us): 48, 99th(us): 303, 99.9th(us): 770, 99.99th(us): 806
TOTAL  - Takes(s): 8.7, Count: 14944, OPS: 1727.3, Avg(us): 92929, Min(us): 1, Max(us): 317439, 50th(us): 4, 90th(us): 303103, 95th(us): 304127, 99th(us): 308735, 99.9th(us): 315135, 99.99th(us): 317439
TXN    - Takes(s): 8.3, Count: 1456, OPS: 174.5, Avg(us): 303322, Min(us): 301056, Max(us): 317439, 50th(us): 302847, 90th(us): 305151, 95th(us): 307455, 99th(us): 315135, 99.9th(us): 317439, 99.99th(us): 317439
TXN_ERROR - Takes(s): 8.5, Count: 528, OPS: 62.2, Avg(us): 152779, Min(us): 150912, Max(us): 166783, 50th(us): 152191, 90th(us): 154111, 95th(us): 156927, 99th(us): 166143, 99.9th(us): 166783, 99.99th(us): 166783
TxnGroup - Takes(s): 8.7, Count: 1984, OPS: 229.3, Avg(us): 254865, Min(us): 56, Max(us): 315135, 50th(us): 302335, 90th(us): 304639, 95th(us): 306431, 99th(us): 311807, 99.9th(us): 314879, 99.99th(us): 315135
UPDATE - Takes(s): 8.7, Count: 8000, OPS: 924.8, Avg(us): 2, Min(us): 1, Max(us): 856, 50th(us): 2, 90th(us): 4, 95th(us): 4, 99th(us): 9, 99.9th(us): 179, 99.99th(us): 796
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  468
prepare phase failed: Remote prepare failed
  version mismatch  30
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  19
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  11
```

+ 200ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 200ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 12.104731694s
COMMIT - Takes(s): 11.7, Count: 1522, OPS: 130.1, Avg(us): 403308, Min(us): 400896, Max(us): 413439, 50th(us): 402943, 90th(us): 405503, 95th(us): 407807, 99th(us): 411903, 99.9th(us): 413439, 99.99th(us): 413439
COMMIT_ERROR - Takes(s): 11.9, Count: 462, OPS: 38.8, Avg(us): 202738, Min(us): 200832, Max(us): 211839, 50th(us): 202239, 90th(us): 204543, 95th(us): 207743, 99th(us): 210047, 99.9th(us): 211839, 99.99th(us): 211839
Start  - Takes(s): 12.1, Count: 2048, OPS: 169.2, Avg(us): 28, Min(us): 14, Max(us): 766, 50th(us): 23, 90th(us): 35, 95th(us): 42, 99th(us): 155, 99.9th(us): 648, 99.99th(us): 766
TOTAL  - Takes(s): 12.1, Count: 15076, OPS: 1245.3, Avg(us): 126890, Min(us): 1, Max(us): 416511, 50th(us): 4, 90th(us): 403455, 95th(us): 404479, 99th(us): 409855, 99.9th(us): 414719, 99.99th(us): 416255
TXN    - Takes(s): 11.7, Count: 1522, OPS: 130.1, Avg(us): 403509, Min(us): 400896, Max(us): 416511, 50th(us): 402943, 90th(us): 405503, 95th(us): 408319, 99th(us): 414719, 99.9th(us): 416255, 99.99th(us): 416511
TXN_ERROR - Takes(s): 11.9, Count: 462, OPS: 38.8, Avg(us): 202945, Min(us): 200960, Max(us): 215039, 50th(us): 202367, 90th(us): 204799, 95th(us): 207871, 99th(us): 213631, 99.9th(us): 215039, 99.99th(us): 215039
TxnGroup - Takes(s): 12.1, Count: 1984, OPS: 163.9, Avg(us): 345234, Min(us): 75, Max(us): 413695, 50th(us): 402431, 90th(us): 404991, 95th(us): 406527, 99th(us): 411903, 99.9th(us): 413439, 99.99th(us): 413695
UPDATE - Takes(s): 12.1, Count: 8000, OPS: 660.9, Avg(us): 2, Min(us): 1, Max(us): 364, 50th(us): 2, 90th(us): 4, 95th(us): 4, 99th(us): 7, 99.9th(us): 178, 99.99th(us): 277
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  383
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  40
prepare phase failed: Remote prepare failed
  version mismatch  35
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  4
```

#### Oreo - 3

+ 50ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 50ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 3.142902569s
COMMIT - Takes(s): 3.0, Count: 1641, OPS: 539.7, Avg(us): 102728, Min(us): 101120, Max(us): 111423, 50th(us): 102463, 90th(us): 104127, 95th(us): 105471, 99th(us): 108671, 99.9th(us): 110399, 99.99th(us): 111423
COMMIT_ERROR - Takes(s): 3.1, Count: 343, OPS: 111.1, Avg(us): 52320, Min(us): 50688, Max(us): 59999, 50th(us): 51935, 90th(us): 53823, 95th(us): 55103, 99th(us): 58399, 99.9th(us): 59999, 99.99th(us): 59999
Start  - Takes(s): 3.1, Count: 2032, OPS: 646.5, Avg(us): 28, Min(us): 14, Max(us): 1250, 50th(us): 24, 90th(us): 36, 95th(us): 45, 99th(us): 185, 99.9th(us): 369, 99.99th(us): 1250
TOTAL  - Takes(s): 3.1, Count: 13298, OPS: 4230.0, Avg(us): 38969, Min(us): 1, Max(us): 111999, 50th(us): 20, 90th(us): 103167, 95th(us): 103935, 99th(us): 107135, 99.9th(us): 111295, 99.99th(us): 111871
TXN    - Takes(s): 3.0, Count: 1641, OPS: 539.9, Avg(us): 102901, Min(us): 101184, Max(us): 111999, 50th(us): 102527, 90th(us): 104319, 95th(us): 105919, 99th(us): 110975, 99.9th(us): 111679, 99.99th(us): 111999
TXN_ERROR - Takes(s): 3.1, Count: 343, OPS: 111.1, Avg(us): 52497, Min(us): 50752, Max(us): 61247, 50th(us): 52031, 90th(us): 53951, 95th(us): 55199, 99th(us): 61055, 99.9th(us): 61247, 99.99th(us): 61247
TxnGroup - Takes(s): 3.1, Count: 1984, OPS: 631.3, Avg(us): 91077, Min(us): 306, Max(us): 111615, 50th(us): 102271, 90th(us): 104063, 95th(us): 105087, 99th(us): 108479, 99.9th(us): 110463, 99.99th(us): 111615
UPDATE - Takes(s): 3.1, Count: 6000, OPS: 1909.2, Avg(us): 2, Min(us): 1, Max(us): 341, 50th(us): 2, 90th(us): 4, 95th(us): 5, 99th(us): 9, 99.9th(us): 47, 99.99th(us): 185
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  282
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  31
prepare phase failed: Remote prepare failed
  version mismatch  30
```

+ 100ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 100ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 6.194922077s
COMMIT - Takes(s): 6.0, Count: 1625, OPS: 271.2, Avg(us): 202881, Min(us): 200960, Max(us): 213503, 50th(us): 202623, 90th(us): 204543, 95th(us): 205951, 99th(us): 209919, 99.9th(us): 211583, 99.99th(us): 213503
COMMIT_ERROR - Takes(s): 6.1, Count: 359, OPS: 59.0, Avg(us): 102354, Min(us): 100800, Max(us): 111039, 50th(us): 101887, 90th(us): 103807, 95th(us): 106047, 99th(us): 109247, 99.9th(us): 111039, 99.99th(us): 111039
Start  - Takes(s): 6.2, Count: 2032, OPS: 328.0, Avg(us): 27, Min(us): 14, Max(us): 478, 50th(us): 25, 90th(us): 35, 95th(us): 42, 99th(us): 77, 99.9th(us): 257, 99.99th(us): 478
TOTAL  - Takes(s): 6.2, Count: 13266, OPS: 2141.1, Avg(us): 76464, Min(us): 1, Max(us): 214399, 50th(us): 20, 90th(us): 203263, 95th(us): 204159, 99th(us): 207999, 99.9th(us): 213503, 99.99th(us): 214399
TXN    - Takes(s): 6.0, Count: 1625, OPS: 271.2, Avg(us): 203061, Min(us): 200960, Max(us): 214399, 50th(us): 202623, 90th(us): 204671, 95th(us): 206591, 99th(us): 213375, 99.9th(us): 214143, 99.99th(us): 214399
TXN_ERROR - Takes(s): 6.1, Count: 359, OPS: 59.0, Avg(us): 102530, Min(us): 100864, Max(us): 113215, 50th(us): 102015, 90th(us): 103935, 95th(us): 106559, 99th(us): 111359, 99.9th(us): 113215, 99.99th(us): 113215
TxnGroup - Takes(s): 6.2, Count: 1984, OPS: 320.2, Avg(us): 178753, Min(us): 62, Max(us): 213631, 50th(us): 202367, 90th(us): 204287, 95th(us): 205439, 99th(us): 209663, 99.9th(us): 212095, 99.99th(us): 213631
UPDATE - Takes(s): 6.2, Count: 6000, OPS: 968.5, Avg(us): 3, Min(us): 1, Max(us): 442, 50th(us): 2, 90th(us): 4, 95th(us): 4, 99th(us): 9, 99.9th(us): 205, 99.99th(us): 263
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  308
prepare phase failed: Remote prepare failed
  version mismatch  27
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  19
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  5
```

+ 150ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 150ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 9.410548661s
COMMIT - Takes(s): 9.1, Count: 1678, OPS: 184.3, Avg(us): 303246, Min(us): 300800, Max(us): 314879, 50th(us): 302847, 90th(us): 305151, 95th(us): 307711, 99th(us): 311295, 99.9th(us): 313087, 99.99th(us): 314879
COMMIT_ERROR - Takes(s): 9.3, Count: 306, OPS: 33.1, Avg(us): 152872, Min(us): 150656, Max(us): 161919, 50th(us): 152319, 90th(us): 154751, 95th(us): 158207, 99th(us): 160895, 99.9th(us): 161919, 99.99th(us): 161919
Start  - Takes(s): 9.4, Count: 2032, OPS: 215.9, Avg(us): 30, Min(us): 14, Max(us): 601, 50th(us): 25, 90th(us): 37, 95th(us): 48, 99th(us): 240, 99.9th(us): 522, 99.99th(us): 601
TOTAL  - Takes(s): 9.4, Count: 13372, OPS: 1421.0, Avg(us): 116362, Min(us): 1, Max(us): 315903, 50th(us): 21, 90th(us): 303871, 95th(us): 304895, 99th(us): 309503, 99.9th(us): 313343, 99.99th(us): 314879
TXN    - Takes(s): 9.1, Count: 1678, OPS: 184.3, Avg(us): 303417, Min(us): 300800, Max(us): 315903, 50th(us): 303103, 90th(us): 305407, 95th(us): 308223, 99th(us): 312063, 99.9th(us): 314367, 99.99th(us): 315903
TXN_ERROR - Takes(s): 9.3, Count: 306, OPS: 33.1, Avg(us): 153010, Min(us): 150784, Max(us): 163199, 50th(us): 152447, 90th(us): 155007, 95th(us): 158335, 99th(us): 161791, 99.9th(us): 163199, 99.99th(us): 163199
TxnGroup - Takes(s): 9.4, Count: 1984, OPS: 210.8, Avg(us): 271136, Min(us): 45, Max(us): 314879, 50th(us): 302591, 90th(us): 305151, 95th(us): 307199, 99th(us): 311295, 99.9th(us): 313343, 99.99th(us): 314879
UPDATE - Takes(s): 9.4, Count: 6000, OPS: 637.6, Avg(us): 3, Min(us): 1, Max(us): 423, 50th(us): 2, 90th(us): 4, 95th(us): 5, 99th(us): 11, 99.9th(us): 198, 99.99th(us): 293
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  241
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  39
prepare phase failed: Remote prepare failed
  version mismatch  24
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  2
```

+ 200ms

```bash
-----------------
DBType: oreo-rm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 64
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 2
HTTPAdditionalLatency: 200ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 12.305011154s
COMMIT - Takes(s): 11.9, Count: 1676, OPS: 140.9, Avg(us): 403030, Min(us): 400896, Max(us): 413183, 50th(us): 402687, 90th(us): 404991, 95th(us): 407039, 99th(us): 410879, 99.9th(us): 412671, 99.99th(us): 413183
COMMIT_ERROR - Takes(s): 12.1, Count: 308, OPS: 25.5, Avg(us): 202610, Min(us): 200704, Max(us): 210943, 50th(us): 201983, 90th(us): 204799, 95th(us): 207487, 99th(us): 210303, 99.9th(us): 210943, 99.99th(us): 210943
Start  - Takes(s): 12.3, Count: 2032, OPS: 165.1, Avg(us): 36, Min(us): 14, Max(us): 1299, 50th(us): 24, 90th(us): 37, 95th(us): 49, 99th(us): 413, 99.9th(us): 1257, 99.99th(us): 1299
TOTAL  - Takes(s): 12.3, Count: 13368, OPS: 1086.3, Avg(us): 154463, Min(us): 1, Max(us): 414719, 50th(us): 21, 90th(us): 403455, 95th(us): 404735, 99th(us): 409599, 99.9th(us): 413439, 99.99th(us): 414719
TXN    - Takes(s): 11.9, Count: 1676, OPS: 140.9, Avg(us): 403217, Min(us): 400896, Max(us): 414719, 50th(us): 402687, 90th(us): 404991, 95th(us): 407551, 99th(us): 413183, 99.9th(us): 414719, 99.99th(us): 414719
TXN_ERROR - Takes(s): 12.1, Count: 308, OPS: 25.5, Avg(us): 202814, Min(us): 200704, Max(us): 214015, 50th(us): 202111, 90th(us): 204927, 95th(us): 208511, 99th(us): 212863, 99.9th(us): 214015, 99.99th(us): 214015
TxnGroup - Takes(s): 12.3, Count: 1984, OPS: 161.2, Avg(us): 359631, Min(us): 52, Max(us): 414207, 50th(us): 402431, 90th(us): 404991, 95th(us): 406527, 99th(us): 411135, 99.9th(us): 412927, 99.99th(us): 414207
UPDATE - Takes(s): 12.3, Count: 6000, OPS: 487.6, Avg(us): 2, Min(us): 1, Max(us): 495, 50th(us): 2, 90th(us): 4, 95th(us): 5, 99th(us): 8, 99.9th(us): 111, 99.99th(us): 265
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  245
prepare phase failed: Remote prepare failed
  version mismatch  31
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  28
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  4
```

## Scalability

### Oreo

+ core = 1

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 3
HTTPAdditionalLatency: 3ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 1m22.259237559s
COMMIT - Takes(s): 82.1, Count: 10172, OPS: 123.9, Avg(us): 269532, Min(us): 7928, Max(us): 1196031, 50th(us): 280575, 90th(us): 405759, 95th(us): 500991, 99th(us): 698879, 99.9th(us): 988159, 99.99th(us): 1091583
COMMIT_ERROR - Takes(s): 82.2, Count: 6468, OPS: 78.7, Avg(us): 202354, Min(us): 4552, Max(us): 1695743, 50th(us): 197375, 90th(us): 309759, 95th(us): 403199, 99th(us): 597503, 99.9th(us): 804351, 99.99th(us): 1387519
READ   - Takes(s): 82.3, Count: 96232, OPS: 1169.9, Avg(us): 59132, Min(us): 6, Max(us): 1299455, 50th(us): 7791, 90th(us): 106047, 95th(us): 199807, 99th(us): 305919, 99.9th(us): 595967, 99.99th(us): 900095
READ_ERROR - Takes(s): 82.0, Count: 3768, OPS: 46.0, Avg(us): 134380, Min(us): 3710, Max(us): 1303551, 50th(us): 99263, 90th(us): 292863, 95th(us): 303871, 99th(us): 493055, 99.9th(us): 700415, 99.99th(us): 1303551
Start  - Takes(s): 82.3, Count: 16768, OPS: 203.8, Avg(us): 45, Min(us): 19, Max(us): 4103, 50th(us): 33, 90th(us): 53, 95th(us): 67, 99th(us): 158, 99.9th(us): 1852, 99.99th(us): 2863
TOTAL  - Takes(s): 82.3, Count: 246216, OPS: 2993.1, Avg(us): 101073, Min(us): 1, Max(us): 2099199, 50th(us): 3819, 90th(us): 398079, 95th(us): 602111, 99th(us): 908287, 99.9th(us): 1301503, 99.99th(us): 1675263
TXN    - Takes(s): 82.1, Count: 10172, OPS: 123.9, Avg(us): 611845, Min(us): 32752, Max(us): 2006015, 50th(us): 598527, 90th(us): 902143, 95th(us): 1004031, 99th(us): 1298431, 99.9th(us): 1692671, 99.99th(us): 1891327
TXN_ERROR - Takes(s): 82.2, Count: 6468, OPS: 78.7, Avg(us): 621949, Min(us): 24784, Max(us): 2801663, 50th(us): 600575, 90th(us): 904703, 95th(us): 1007615, 99th(us): 1305599, 99.9th(us): 1702911, 99.99th(us): 2001919
TxnGroup - Takes(s): 82.2, Count: 16640, OPS: 202.3, Avg(us): 614706, Min(us): 23648, Max(us): 2099199, 50th(us): 598015, 90th(us): 905215, 95th(us): 1010687, 99th(us): 1298431, 99.9th(us): 1607679, 99.99th(us): 1905663
UPDATE - Takes(s): 82.3, Count: 96232, OPS: 1169.9, Avg(us): 5, Min(us): 1, Max(us): 3059, 50th(us): 4, 90th(us): 6, 95th(us): 7, 99th(us): 15, 99.9th(us): 443, 99.99th(us): 1772
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  5571
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  564
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  287
prepare phase failed: Remote prepare failed
        read failed due to unknown txn status  44
  transaction is aborted by other transaction   2

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   2121
rollForward failed
  version mismatch  1045
rollback failed
  version mismatch  597
     key not found    5
```

+ core = 2

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 3
HTTPAdditionalLatency: 3ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 39.507318108s
COMMIT - Takes(s): 39.4, Count: 9902, OPS: 251.3, Avg(us): 122377, Min(us): 8068, Max(us): 607231, 50th(us): 100991, 90th(us): 198911, 95th(us): 209279, 99th(us): 300799, 99.9th(us): 399615, 99.99th(us): 513023
COMMIT_ERROR - Takes(s): 39.5, Count: 6738, OPS: 170.8, Avg(us): 89495, Min(us): 4456, Max(us): 502271, 50th(us): 93247, 90th(us): 184959, 95th(us): 197119, 99th(us): 212863, 99.9th(us): 312319, 99.99th(us): 402687
READ   - Takes(s): 39.5, Count: 96706, OPS: 2448.0, Avg(us): 30089, Min(us): 6, Max(us): 400639, 50th(us): 6099, 90th(us): 90559, 95th(us): 97599, 99th(us): 182015, 99.9th(us): 211583, 99.99th(us): 309503
READ_ERROR - Takes(s): 39.4, Count: 3294, OPS: 83.6, Avg(us): 65233, Min(us): 3806, Max(us): 306431, 50th(us): 81023, 90th(us): 104767, 95th(us): 114815, 99th(us): 201087, 99.9th(us): 299263, 99.99th(us): 306431
Start  - Takes(s): 39.5, Count: 16768, OPS: 424.4, Avg(us): 51, Min(us): 22, Max(us): 4163, 50th(us): 33, 90th(us): 56, 95th(us): 77, 99th(us): 430, 99.9th(us): 2163, 99.99th(us): 3977
TOTAL  - Takes(s): 39.5, Count: 246624, OPS: 6242.3, Avg(us): 48764, Min(us): 1, Max(us): 910335, 50th(us): 3929, 90th(us): 196991, 95th(us): 299263, 99th(us): 412415, 99.9th(us): 599551, 99.99th(us): 709631
TXN    - Takes(s): 39.4, Count: 9902, OPS: 251.3, Avg(us): 299918, Min(us): 31872, Max(us): 910335, 50th(us): 296959, 90th(us): 409855, 95th(us): 496895, 99th(us): 600063, 99.9th(us): 778751, 99.99th(us): 900607
TXN_ERROR - Takes(s): 39.5, Count: 6738, OPS: 170.7, Avg(us): 292751, Min(us): 29760, Max(us): 804863, 50th(us): 296447, 90th(us): 407295, 95th(us): 494591, 99th(us): 596991, 99.9th(us): 703999, 99.99th(us): 803327
TxnGroup - Takes(s): 39.5, Count: 16640, OPS: 421.5, Avg(us): 296493, Min(us): 25024, Max(us): 801279, 50th(us): 295167, 90th(us): 409599, 95th(us): 493311, 99th(us): 595455, 99.9th(us): 703487, 99.99th(us): 798207
UPDATE - Takes(s): 39.5, Count: 96706, OPS: 2447.9, Avg(us): 6, Min(us): 1, Max(us): 4083, 50th(us): 4, 90th(us): 6, 95th(us): 8, 99th(us): 16, 99.9th(us): 669, 99.99th(us): 2739
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  5889
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  552
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  266
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  31

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1809
rollForward failed
  version mismatch  865
rollback failed
  version mismatch  618
     key not found    2
```

+ core = 4

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 3
HTTPAdditionalLatency: 3ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 19.749386109s
COMMIT - Takes(s): 19.7, Count: 9627, OPS: 488.2, Avg(us): 59847, Min(us): 8472, Max(us): 227071, 50th(us): 67647, 90th(us): 95807, 95th(us): 102911, 99th(us): 164863, 99.9th(us): 193663, 99.99th(us): 208895
COMMIT_ERROR - Takes(s): 19.7, Count: 7013, OPS: 355.8, Avg(us): 44673, Min(us): 4928, Max(us): 287231, 50th(us): 28351, 90th(us): 87039, 95th(us): 94015, 99th(us): 107135, 99.9th(us): 177023, 99.99th(us): 193919
READ   - Takes(s): 19.7, Count: 97125, OPS: 4918.7, Avg(us): 15378, Min(us): 6, Max(us): 271103, 50th(us): 6059, 90th(us): 57727, 95th(us): 67647, 99th(us): 84159, 99.9th(us): 103871, 99.99th(us): 172287
READ_ERROR - Takes(s): 19.7, Count: 2875, OPS: 145.9, Avg(us): 30777, Min(us): 3946, Max(us): 197375, 50th(us): 14079, 90th(us): 75583, 95th(us): 84351, 99th(us): 101311, 99.9th(us): 152319, 99.99th(us): 197375
Start  - Takes(s): 19.8, Count: 16768, OPS: 848.8, Avg(us): 54, Min(us): 21, Max(us): 5071, 50th(us): 34, 90th(us): 60, 95th(us): 81, 99th(us): 519, 99.9th(us): 2441, 99.99th(us): 3041
TOTAL  - Takes(s): 19.7, Count: 246912, OPS: 12502.0, Avg(us): 24270, Min(us): 1, Max(us): 502783, 50th(us): 4025, 90th(us): 97407, 95th(us): 157439, 99th(us): 205183, 99.9th(us): 291071, 99.99th(us): 367615
TXN    - Takes(s): 19.7, Count: 9627, OPS: 488.2, Avg(us): 150556, Min(us): 32448, Max(us): 502783, 50th(us): 134655, 90th(us): 205055, 95th(us): 218495, 99th(us): 291839, 99.9th(us): 365567, 99.99th(us): 423423
TXN_ERROR - Takes(s): 19.7, Count: 7013, OPS: 355.7, Avg(us): 146385, Min(us): 29232, Max(us): 409599, 50th(us): 127231, 90th(us): 203647, 95th(us): 214271, 99th(us): 288255, 99.9th(us): 366335, 99.99th(us): 396799
TxnGroup - Takes(s): 19.7, Count: 16640, OPS: 843.4, Avg(us): 148548, Min(us): 21680, Max(us): 502271, 50th(us): 137343, 90th(us): 203519, 95th(us): 216319, 99th(us): 287487, 99.9th(us): 364031, 99.99th(us): 407807
UPDATE - Takes(s): 19.7, Count: 97125, OPS: 4917.7, Avg(us): 7, Min(us): 1, Max(us): 5287, 50th(us): 4, 90th(us): 6, 95th(us): 8, 99th(us): 16, 99.9th(us): 993, 99.99th(us): 2615
Error Summary:

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1446
rollForward failed
  version mismatch  838
rollback failed
  version mismatch  590
     key not found    1

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  6155
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  535
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  300
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  23
```

+ core = 8

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 3
HTTPAdditionalLatency: 3ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 9.637757973s
COMMIT - Takes(s): 9.6, Count: 9362, OPS: 975.9, Avg(us): 26311, Min(us): 8088, Max(us): 100351, 50th(us): 24127, 90th(us): 39391, 95th(us): 46879, 99th(us): 64063, 99.9th(us): 86719, 99.99th(us): 96703
COMMIT_ERROR - Takes(s): 9.6, Count: 7278, OPS: 759.5, Avg(us): 20587, Min(us): 4612, Max(us): 90367, 50th(us): 18575, 90th(us): 32895, 95th(us): 38623, 99th(us): 53887, 99.9th(us): 71039, 99.99th(us): 79871
READ   - Takes(s): 9.6, Count: 97270, OPS: 10099.0, Avg(us): 7812, Min(us): 6, Max(us): 109695, 50th(us): 5959, 90th(us): 13831, 95th(us): 17839, 99th(us): 29071, 99.9th(us): 49919, 99.99th(us): 77631
READ_ERROR - Takes(s): 9.6, Count: 2730, OPS: 285.1, Avg(us): 12796, Min(us): 4070, Max(us): 103231, 50th(us): 10199, 90th(us): 23583, 95th(us): 29071, 99th(us): 42879, 99.9th(us): 65535, 99.99th(us): 103231
Start  - Takes(s): 9.6, Count: 16768, OPS: 1738.9, Avg(us): 57, Min(us): 20, Max(us): 5011, 50th(us): 33, 90th(us): 61, 95th(us): 82, 99th(us): 491, 99.9th(us): 2565, 99.99th(us): 4431
TOTAL  - Takes(s): 9.6, Count: 246672, OPS: 25586.0, Avg(us): 11704, Min(us): 1, Max(us): 182271, 50th(us): 4093, 90th(us): 50751, 95th(us): 70719, 99th(us): 95743, 99.9th(us): 126335, 99.99th(us): 153599
TXN    - Takes(s): 9.6, Count: 9362, OPS: 975.7, Avg(us): 73133, Min(us): 31808, Max(us): 173311, 50th(us): 70207, 90th(us): 95743, 95th(us): 105407, 99th(us): 127039, 99.9th(us): 151039, 99.99th(us): 172159
TXN_ERROR - Takes(s): 9.6, Count: 7278, OPS: 759.2, Avg(us): 70288, Min(us): 25568, Max(us): 170239, 50th(us): 67647, 90th(us): 92799, 95th(us): 101503, 99th(us): 122047, 99.9th(us): 153343, 99.99th(us): 166911
TxnGroup - Takes(s): 9.6, Count: 16640, OPS: 1730.6, Avg(us): 71781, Min(us): 27072, Max(us): 182271, 50th(us): 69119, 90th(us): 94591, 95th(us): 103615, 99th(us): 124607, 99.9th(us): 155775, 99.99th(us): 174847
UPDATE - Takes(s): 9.6, Count: 97270, OPS: 10095.3, Avg(us): 7, Min(us): 1, Max(us): 3481, 50th(us): 4, 90th(us): 6, 95th(us): 8, 99th(us): 17, 99.9th(us): 1197, 99.99th(us): 2389
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  6374
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  563
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  327
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  14

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1390
rollForward failed
  version mismatch  812
rollback failed
  version mismatch  528
```

+ core = 12

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 3
HTTPAdditionalLatency: 3ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 9.274272127s
COMMIT - Takes(s): 9.2, Count: 9534, OPS: 1033.1, Avg(us): 25465, Min(us): 8176, Max(us): 115711, 50th(us): 23839, 90th(us): 36927, 95th(us): 42015, 99th(us): 56351, 99.9th(us): 81087, 99.99th(us): 111231
COMMIT_ERROR - Takes(s): 9.2, Count: 7106, OPS: 769.7, Avg(us): 20069, Min(us): 4668, Max(us): 97023, 50th(us): 18607, 90th(us): 31039, 95th(us): 35327, 99th(us): 44991, 99.9th(us): 62015, 99.99th(us): 75327
READ   - Takes(s): 9.3, Count: 97255, OPS: 10489.6, Avg(us): 7568, Min(us): 6, Max(us): 84863, 50th(us): 5895, 90th(us): 13319, 95th(us): 17087, 99th(us): 26399, 99.9th(us): 40575, 99.99th(us): 59295
READ_ERROR - Takes(s): 9.2, Count: 2745, OPS: 298.1, Avg(us): 12257, Min(us): 3894, Max(us): 53439, 50th(us): 9703, 90th(us): 22495, 95th(us): 27423, 99th(us): 38335, 99.9th(us): 49407, 99.99th(us): 53439
Start  - Takes(s): 9.3, Count: 16768, OPS: 1807.7, Avg(us): 55, Min(us): 22, Max(us): 4683, 50th(us): 33, 90th(us): 62, 95th(us): 82, 99th(us): 518, 99.9th(us): 2585, 99.99th(us): 4343
TOTAL  - Takes(s): 9.3, Count: 246986, OPS: 26630.5, Avg(us): 11398, Min(us): 1, Max(us): 163711, 50th(us): 4111, 90th(us): 50399, 95th(us): 69375, 99th(us): 90623, 99.9th(us): 114495, 99.99th(us): 140031
TXN    - Takes(s): 9.2, Count: 9534, OPS: 1033.0, Avg(us): 70938, Min(us): 29280, Max(us): 153855, 50th(us): 68991, 90th(us): 90175, 95th(us): 98047, 99th(us): 114559, 99.9th(us): 137087, 99.99th(us): 153855
TXN_ERROR - Takes(s): 9.2, Count: 7106, OPS: 769.5, Avg(us): 68092, Min(us): 25920, Max(us): 141695, 50th(us): 66047, 90th(us): 88191, 95th(us): 95487, 99th(us): 111167, 99.9th(us): 132223, 99.99th(us): 140415
TxnGroup - Takes(s): 9.2, Count: 16640, OPS: 1799.2, Avg(us): 69614, Min(us): 24832, Max(us): 163711, 50th(us): 67967, 90th(us): 89855, 95th(us): 97599, 99th(us): 113663, 99.9th(us): 140031, 99.99th(us): 155007
UPDATE - Takes(s): 9.3, Count: 97255, OPS: 10489.5, Avg(us): 6, Min(us): 1, Max(us): 3833, 50th(us): 4, 90th(us): 6, 95th(us): 7, 99th(us): 16, 99.9th(us): 960, 99.99th(us): 2541
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  6276
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  521
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  305
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  4

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1415
rollForward failed
  version mismatch  830
rollback failed
  version mismatch  500
```

+ core = 16

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 3
HTTPAdditionalLatency: 3ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 9.403273025s
COMMIT - Takes(s): 9.4, Count: 9528, OPS: 1017.9, Avg(us): 25863, Min(us): 8296, Max(us): 100223, 50th(us): 24175, 90th(us): 37567, 95th(us): 42975, 99th(us): 56383, 99.9th(us): 77183, 99.99th(us): 99391
COMMIT_ERROR - Takes(s): 9.4, Count: 7112, OPS: 759.9, Avg(us): 20556, Min(us): 4516, Max(us): 91903, 50th(us): 19151, 90th(us): 31583, 95th(us): 35839, 99th(us): 47615, 99.9th(us): 65535, 99.99th(us): 87999
READ   - Takes(s): 9.4, Count: 97334, OPS: 10355.1, Avg(us): 7597, Min(us): 6, Max(us): 85631, 50th(us): 5951, 90th(us): 13351, 95th(us): 16975, 99th(us): 25935, 99.9th(us): 39359, 99.99th(us): 57951
READ_ERROR - Takes(s): 9.3, Count: 2666, OPS: 285.3, Avg(us): 12612, Min(us): 3968, Max(us): 79039, 50th(us): 9943, 90th(us): 23375, 95th(us): 29071, 99th(us): 40735, 99.9th(us): 60511, 99.99th(us): 79039
Start  - Takes(s): 9.4, Count: 16768, OPS: 1782.8, Avg(us): 58, Min(us): 20, Max(us): 5515, 50th(us): 32, 90th(us): 62, 95th(us): 82, 99th(us): 593, 99.9th(us): 2723, 99.99th(us): 4759
TOTAL  - Takes(s): 9.4, Count: 247132, OPS: 26280.6, Avg(us): 11487, Min(us): 1, Max(us): 186879, 50th(us): 4119, 90th(us): 50623, 95th(us): 70079, 99th(us): 91391, 99.9th(us): 115391, 99.99th(us): 139519
TXN    - Takes(s): 9.4, Count: 9528, OPS: 1017.8, Avg(us): 71551, Min(us): 31552, Max(us): 166399, 50th(us): 69695, 90th(us): 91135, 95th(us): 98687, 99th(us): 114687, 99.9th(us): 136447, 99.99th(us): 153343
TXN_ERROR - Takes(s): 9.4, Count: 7112, OPS: 760.0, Avg(us): 68862, Min(us): 27232, Max(us): 157567, 50th(us): 66815, 90th(us): 88959, 95th(us): 96831, 99th(us): 113663, 99.9th(us): 134527, 99.99th(us): 155775
TxnGroup - Takes(s): 9.4, Count: 16640, OPS: 1773.8, Avg(us): 70293, Min(us): 22240, Max(us): 186879, 50th(us): 68607, 90th(us): 90623, 95th(us): 98559, 99th(us): 115327, 99.9th(us): 140415, 99.99th(us): 162431
UPDATE - Takes(s): 9.4, Count: 97334, OPS: 10353.6, Avg(us): 7, Min(us): 1, Max(us): 3601, 50th(us): 4, 90th(us): 6, 95th(us): 7, 99th(us): 17, 99.9th(us): 1314, 99.99th(us): 2675
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  6281
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  515
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  315
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  1

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1333
rollForward failed
  version mismatch  826
rollback failed
  version mismatch  507
```
