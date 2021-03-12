# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [api/proto/v1/archer.proto](#api/proto/v1/archer.proto)
    - [CancelRequest](#v1.CancelRequest)
    - [CancelResponse](#v1.CancelResponse)
    - [ProcessRequest](#v1.ProcessRequest)
    - [ProcessResponse](#v1.ProcessResponse)
    - [SampleInfo](#v1.SampleInfo)
    - [WatchRequest](#v1.WatchRequest)
    - [WatchResponse](#v1.WatchResponse)
  
    - [State](#v1.State)
  
    - [Archer](#v1.Archer)
  
- [Scalar Value Types](#scalar-value-types)



<a name="api/proto/v1/archer.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## api/proto/v1/archer.proto



<a name="v1.CancelRequest"></a>

### CancelRequest
CancelRequest will cancel processing for a sample.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| apiVersion | [string](#string) |  | api version |
| id | [string](#string) |  | identifier for the sample |






<a name="v1.CancelResponse"></a>

### CancelResponse
CancelResponse.






<a name="v1.ProcessRequest"></a>

### ProcessRequest
ProcessRequest will request a sample to be processed by Archer.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| apiVersion | [string](#string) |  | api version |
| sampleID | [string](#string) |  | sampleID is the sample identifier - users job to assign this and make it unique |
| inputFASTQfiles | [string](#string) | repeated | inputFASTQfiles for this sample |
| scheme | [string](#string) |  | scheme denotes the amplicon scheme used for the sample |
| schemeVersion | [int32](#int32) |  | schemeVersion denotes the amplicon scheme version used |
| endpoint | [string](#string) |  | endpoint for processed data to be sent |






<a name="v1.ProcessResponse"></a>

### ProcessResponse
ProcessResponse


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| apiVersion | [string](#string) |  | api version |
| id | [string](#string) |  | identifier for the sample that processing was started for (used to monitor or cancel the sample) |






<a name="v1.SampleInfo"></a>

### SampleInfo
SampleInfo describes how a sample was
processed by Archer.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sampleID | [string](#string) |  | sampleID is the sample identifier - as returned by Process() |
| processRequest | [ProcessRequest](#v1.ProcessRequest) |  | the original message used to start the sample processing |
| state | [State](#v1.State) |  | state the sample is in |
| errors | [string](#string) | repeated | errors will contain encountered errors (if state is STATE_ERROR, otherwise this will be empty) |
| filesDiscovered | [int32](#int32) |  | filesDiscovered is the number of files found for this sample |
| startTime | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | startTime for processing |
| endTime | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | endTime for processing (unset if processing still running) |






<a name="v1.WatchRequest"></a>

### WatchRequest
WatchRequest to monitor sample processing.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| apiVersion | [string](#string) |  | api version |
| sendFinished | [bool](#bool) |  | sendFinished will tell Archer to also send information on samples that have completed procesing |






<a name="v1.WatchResponse"></a>

### WatchResponse
WatchResponse.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| apiVersion | [string](#string) |  | api version |
| samples | [SampleInfo](#v1.SampleInfo) | repeated | current state of samples |





 


<a name="v1.State"></a>

### State
State of a sample being handled by Archer.

| Name | Number | Description |
| ---- | ------ | ----------- |
| STATE_RUNNING | 0 | sample prep is running |
| STATE_SUCCESS | 1 | sample prep is complete with no errors |
| STATE_ERROR | 2 | sample prep has stopped due to errors |
| STATE_CANCELLED | 3 | sample prep was cancelled via a call to cancel() |


 

 


<a name="v1.Archer"></a>

### Archer
Archer manages sample processing prior to
running CLIMB pipelines.

This includes read qc, contamination filtering,
compression and endpoint upload.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Process | [ProcessRequest](#v1.ProcessRequest) | [ProcessResponse](#v1.ProcessResponse) | Process will begin processing for a sample. |
| Cancel | [CancelRequest](#v1.CancelRequest) | [CancelResponse](#v1.CancelResponse) | Cancel will cancel processing for a sample. |
| Watch | [WatchRequest](#v1.WatchRequest) | [WatchResponse](#v1.WatchResponse) stream | Watch sample processing, returning messages when sample processing starts, stops or updates The current state of all currently-processing samples will be returned in the initial set of messages, with the option of also including finished samples. |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

