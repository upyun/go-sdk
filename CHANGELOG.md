# UPYUN Go SDK ChangeLog

## Version v3.0.3 (2022.8.1)

1. fix err when content-length == 0
2. list add content-type info

## Version v3.0.2 (2021.1.13)

1. add test 
2. add error struct

## Version v3.0.1 (2021.1.6)

### features

1. escape copy/move source key
2. add lint

## Version v3.0.0 (2020.7.24)

### features

1. support go mod

## Version 2.2.0 (2020.4.1)

### features

1. move file
2. copy file
3. output the folder to Json format

## Version 2.1.0 (2017.2.14)

### features

1. restruct go sdk
2. support new signature
3. more flexible

## Version 2.0.0 (2015.12.30)

### features

1. restruct go sdk
2. add multipart upload and media api
3. add GetLargeList in REST api

## Version 1.1.0

date: 2015.06.10

### features

1. add purge and form api
2. user can use io.Reader or io.Writer instead of os.File in REST Put or Get
3. add user agent in http request headers

### Change

1. change default chunk size to 32kb, and SetChunkSize will influenced entire program once used.


### Bugfix

1. fix a bug when using md5 in put
