# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [api/auction_service.proto](#api_auction_service-proto)
    - [BidProto](#auctionservice-BidProto)
    - [CreateLotRequest](#auctionservice-CreateLotRequest)
    - [CurrentPriceProto](#auctionservice-CurrentPriceProto)
    - [DeleteLotRequest](#auctionservice-DeleteLotRequest)
    - [DeleteLotResponse](#auctionservice-DeleteLotResponse)
    - [GetCurrentPriceRequest](#auctionservice-GetCurrentPriceRequest)
    - [GetLotRequest](#auctionservice-GetLotRequest)
    - [GetLotsRequest](#auctionservice-GetLotsRequest)
    - [GetLotsResponse](#auctionservice-GetLotsResponse)
    - [LotFull](#auctionservice-LotFull)
    - [LotShort](#auctionservice-LotShort)
    - [PlaceBidRequest](#auctionservice-PlaceBidRequest)
    - [UpdateLotRequest](#auctionservice-UpdateLotRequest)
  
    - [LotStatus](#auctionservice-LotStatus)
  
    - [BidGrpcService](#auctionservice-BidGrpcService)
    - [LotGrpcService](#auctionservice-LotGrpcService)
  
- [api/gateway_messages.proto](#api_gateway_messages-proto)
    - [AuctionBrief](#apigateway-AuctionBrief)
    - [AuctionDetail](#apigateway-AuctionDetail)
    - [AuctionListResponse](#apigateway-AuctionListResponse)
    - [AuctionUpdate](#apigateway-AuctionUpdate)
    - [BidHistory](#apigateway-BidHistory)
    - [CreateAuctionRequest](#apigateway-CreateAuctionRequest)
    - [CreateAuctionResponse](#apigateway-CreateAuctionResponse)
    - [LoginRequest](#apigateway-LoginRequest)
    - [LoginResponse](#apigateway-LoginResponse)
    - [LogoutRequest](#apigateway-LogoutRequest)
    - [LogoutResponse](#apigateway-LogoutResponse)
    - [PlaceBidRequest](#apigateway-PlaceBidRequest)
    - [PlaceBidResponse](#apigateway-PlaceBidResponse)
    - [RefreshRequest](#apigateway-RefreshRequest)
    - [RefreshResponse](#apigateway-RefreshResponse)
    - [RegisterRequest](#apigateway-RegisterRequest)
    - [RegisterResponse](#apigateway-RegisterResponse)
    - [SubscribeRequest](#apigateway-SubscribeRequest)
    - [SubscribeResponse](#apigateway-SubscribeResponse)
    - [UnsubscribeRequest](#apigateway-UnsubscribeRequest)
    - [UnsubscribeResponse](#apigateway-UnsubscribeResponse)
  
- [api/template_messages.proto](#api_template_messages-proto)
    - [Template](#templateservice-Template)
  
- [api/template_service.proto](#api_template_service-proto)
    - [ApigatewayService](#templateservice-ApigatewayService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="api_auction_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## api/auction_service.proto



<a name="auctionservice-BidProto"></a>

### BidProto



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int64](#int64) |  |  |
| lot_id | [int64](#int64) |  |  |
| bidder_id | [int64](#int64) |  |  |
| amount | [string](#string) |  |  |
| created_at | [string](#string) |  | ISO-8601 |






<a name="auctionservice-CreateLotRequest"></a>

### CreateLotRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| title | [string](#string) |  |  |
| description | [string](#string) |  |  |
| starting_price | [string](#string) |  |  |
| status | [LotStatus](#auctionservice-LotStatus) |  |  |
| image_url | [string](#string) |  |  |
| seller_id | [int64](#int64) |  |  |
| ends_at | [string](#string) |  | ISO-8601, empty = not set |






<a name="auctionservice-CurrentPriceProto"></a>

### CurrentPriceProto



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| lot_id | [int64](#int64) |  |  |
| current_price | [string](#string) |  |  |






<a name="auctionservice-DeleteLotRequest"></a>

### DeleteLotRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int64](#int64) |  |  |






<a name="auctionservice-DeleteLotResponse"></a>

### DeleteLotResponse







<a name="auctionservice-GetCurrentPriceRequest"></a>

### GetCurrentPriceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| lot_id | [int64](#int64) |  |  |






<a name="auctionservice-GetLotRequest"></a>

### GetLotRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int64](#int64) |  |  |






<a name="auctionservice-GetLotsRequest"></a>

### GetLotsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page | [int32](#int32) |  |  |
| size | [int32](#int32) |  |  |
| sort | [string](#string) |  | field name, e.g. &#34;createdAt&#34; |
| filter_by_status | [bool](#bool) |  |  |
| status | [LotStatus](#auctionservice-LotStatus) |  |  |






<a name="auctionservice-GetLotsResponse"></a>

### GetLotsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| content | [LotShort](#auctionservice-LotShort) | repeated |  |
| page | [int32](#int32) |  |  |
| size | [int32](#int32) |  |  |
| total_elements | [int64](#int64) |  |  |
| total_pages | [int32](#int32) |  |  |
| first | [bool](#bool) |  |  |
| last | [bool](#bool) |  |  |






<a name="auctionservice-LotFull"></a>

### LotFull
Full lot detail (returned by create / get-by-id / update)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int64](#int64) |  |  |
| title | [string](#string) |  |  |
| description | [string](#string) |  |  |
| starting_price | [string](#string) |  | decimal as string, e.g. &#34;1500.00&#34; |
| current_price | [string](#string) |  |  |
| status | [LotStatus](#auctionservice-LotStatus) |  |  |
| image_url | [string](#string) |  |  |
| seller_id | [int64](#int64) |  |  |
| created_at | [string](#string) |  | ISO-8601 |
| updated_at | [string](#string) |  |  |
| ends_at | [string](#string) |  |  |






<a name="auctionservice-LotShort"></a>

### LotShort
Brief lot info (returned inside paginated list)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int64](#int64) |  |  |
| title | [string](#string) |  |  |
| current_price | [string](#string) |  |  |
| status | [LotStatus](#auctionservice-LotStatus) |  |  |
| image_url | [string](#string) |  |  |
| ends_at | [string](#string) |  |  |






<a name="auctionservice-PlaceBidRequest"></a>

### PlaceBidRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| lot_id | [int64](#int64) |  |  |
| bidder_id | [int64](#int64) |  |  |
| amount | [string](#string) |  | decimal as string |






<a name="auctionservice-UpdateLotRequest"></a>

### UpdateLotRequest
optional fields: only sent fields are updated (partial update)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int64](#int64) |  |  |
| title | [string](#string) | optional |  |
| description | [string](#string) | optional |  |
| starting_price | [string](#string) | optional |  |
| status | [LotStatus](#auctionservice-LotStatus) | optional |  |
| image_url | [string](#string) | optional |  |
| ends_at | [string](#string) | optional |  |





 


<a name="auctionservice-LotStatus"></a>

### LotStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| DRAFT | 0 |  |
| ACTIVE | 1 |  |
| CLOSED | 2 |  |


 

 


<a name="auctionservice-BidGrpcService"></a>

### BidGrpcService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| PlaceBid | [PlaceBidRequest](#auctionservice-PlaceBidRequest) | [BidProto](#auctionservice-BidProto) |  |
| GetCurrentPrice | [GetCurrentPriceRequest](#auctionservice-GetCurrentPriceRequest) | [CurrentPriceProto](#auctionservice-CurrentPriceProto) |  |


<a name="auctionservice-LotGrpcService"></a>

### LotGrpcService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateLot | [CreateLotRequest](#auctionservice-CreateLotRequest) | [LotFull](#auctionservice-LotFull) |  |
| GetLot | [GetLotRequest](#auctionservice-GetLotRequest) | [LotFull](#auctionservice-LotFull) |  |
| GetLots | [GetLotsRequest](#auctionservice-GetLotsRequest) | [GetLotsResponse](#auctionservice-GetLotsResponse) |  |
| UpdateLot | [UpdateLotRequest](#auctionservice-UpdateLotRequest) | [LotFull](#auctionservice-LotFull) |  |
| DeleteLot | [DeleteLotRequest](#auctionservice-DeleteLotRequest) | [DeleteLotResponse](#auctionservice-DeleteLotResponse) |  |

 



<a name="api_gateway_messages-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## api/gateway_messages.proto



<a name="apigateway-AuctionBrief"></a>

### AuctionBrief
Auction brief information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| auction_id | [string](#string) |  |  |
| title | [string](#string) |  |  |
| description | [string](#string) |  |  |
| photo_urls | [string](#string) | repeated |  |
| start_price | [int64](#int64) |  |  |
| current_price | [int64](#int64) |  |  |
| end_time | [int64](#int64) |  |  |
| status | [string](#string) |  |  |
| creator_id | [string](#string) |  |  |






<a name="apigateway-AuctionDetail"></a>

### AuctionDetail
Auction full information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| auction_id | [string](#string) |  |  |
| title | [string](#string) |  |  |
| description | [string](#string) |  |  |
| photo_urls | [string](#string) | repeated |  |
| start_price | [int64](#int64) |  |  |
| current_price | [int64](#int64) |  |  |
| end_time | [int64](#int64) |  |  |
| status | [string](#string) |  |  |
| creator_id | [string](#string) |  |  |
| bids | [BidHistory](#apigateway-BidHistory) | repeated |  |






<a name="apigateway-AuctionListResponse"></a>

### AuctionListResponse
Paged auction list response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| auctions | [AuctionBrief](#apigateway-AuctionBrief) | repeated |  |
| page | [int32](#int32) |  |  |
| page_size | [int32](#int32) |  |  |
| total_pages | [int32](#int32) |  |  |
| total_items | [int32](#int32) |  |  |






<a name="apigateway-AuctionUpdate"></a>

### AuctionUpdate
Auction update notification


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| auction_id | [string](#string) |  |  |
| current_price | [int64](#int64) |  |  |
| last_bid_time | [int64](#int64) |  |  |
| status | [string](#string) |  |  |
| subscribers | [string](#string) | repeated |  |






<a name="apigateway-BidHistory"></a>

### BidHistory
Bid history entry


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bidder_id | [string](#string) |  |  |
| amount | [int64](#int64) |  |  |
| timestamp | [int64](#int64) |  |  |






<a name="apigateway-CreateAuctionRequest"></a>

### CreateAuctionRequest
Create auction request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| title | [string](#string) |  |  |
| description | [string](#string) |  |  |
| photo_urls | [string](#string) | repeated |  |
| start_price | [int64](#int64) |  |  |
| end_time | [int64](#int64) |  |  |






<a name="apigateway-CreateAuctionResponse"></a>

### CreateAuctionResponse
Create auction response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| auction_id | [string](#string) |  |  |






<a name="apigateway-LoginRequest"></a>

### LoginRequest
User login request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| login | [string](#string) |  |  |
| password | [string](#string) |  |  |






<a name="apigateway-LoginResponse"></a>

### LoginResponse
User login response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| access_token | [string](#string) |  |  |
| refresh_token | [string](#string) |  |  |
| access_token_expiry | [int64](#int64) |  |  |
| refresh_token_expiry | [int64](#int64) |  |  |






<a name="apigateway-LogoutRequest"></a>

### LogoutRequest
User logout request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| refresh_token | [string](#string) |  |  |






<a name="apigateway-LogoutResponse"></a>

### LogoutResponse
User logout response






<a name="apigateway-PlaceBidRequest"></a>

### PlaceBidRequest
Place bid request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| auction_id | [string](#string) |  |  |
| amount | [int64](#int64) |  |  |






<a name="apigateway-PlaceBidResponse"></a>

### PlaceBidResponse
Place bid response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| success | [bool](#bool) |  |  |
| message | [string](#string) |  |  |






<a name="apigateway-RefreshRequest"></a>

### RefreshRequest
Token refresh request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| refresh_token | [string](#string) |  |  |






<a name="apigateway-RefreshResponse"></a>

### RefreshResponse
Token refresh response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| access_token | [string](#string) |  |  |
| refresh_token | [string](#string) |  |  |
| access_token_expiry | [int64](#int64) |  |  |
| refresh_token_expiry | [int64](#int64) |  |  |






<a name="apigateway-RegisterRequest"></a>

### RegisterRequest
User registration request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| login | [string](#string) |  |  |
| password | [string](#string) |  |  |
| email | [string](#string) |  |  |






<a name="apigateway-RegisterResponse"></a>

### RegisterResponse
User registration response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |






<a name="apigateway-SubscribeRequest"></a>

### SubscribeRequest
Subscribe to auction request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| auction_id | [string](#string) |  |  |






<a name="apigateway-SubscribeResponse"></a>

### SubscribeResponse
Subscribe to auction response






<a name="apigateway-UnsubscribeRequest"></a>

### UnsubscribeRequest
Unsubscribe from auction request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| auction_id | [string](#string) |  |  |






<a name="apigateway-UnsubscribeResponse"></a>

### UnsubscribeResponse
Unsubscribe from auction response





 

 

 

 



<a name="api_template_messages-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## api/template_messages.proto



<a name="templateservice-Template"></a>

### Template
Комментарий к message





 

 

 

 



<a name="api_template_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## api/template_service.proto


 

 

 


<a name="templateservice-ApigatewayService"></a>

### ApigatewayService
Комментарий к сервису

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Templates | [Template](#templateservice-Template) | [Template](#templateservice-Template) | Комментарии к методам |

 



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

