# CCA Core

`CCA Core` is a Cosmos SDK module implementation empowering users to create and manage light clients, update clients, establish channels, update transaction signatures, and support diverse queries. Integration with this module facilitates seamless functionality within your application.


## State

### Client
   
Client is mainly composed of id, channelId, type, latestHeader, comfirmation, frozen, validators(optional), difficulty(optional),  where id is the unique identifier of the client.

- Client: 0x01 | classID | -> ProtocolBuffer(Client)

### Channel
   
Channel is mainly composed of id, clientId, appId, externalChainId, vaultAddress, where id is the unique identifier of the Channel.

- Channel: 0x02 | channelID | -> ProtocolBuffer(Channel)

### Request
   
Requests are created within specific modules to implement their unique business logic. There are two types of requests: `<Intent>Request` and `SigningRequest`.

`CCA Core` is responsible for monitoring the execution of `<Intent>Request`, facilitating transactions that interoperate with the vault on an counterparty chain, and notifying the states to the specific sub app.

`CCA Core` also handles the acceptance and verification of signatures in `SigningRequest`, tracing the execution result on the counterparty chain.

- Request: 0x03 | RequestId | -> ProtocolBuffer(Request)

To prevent replay attacks, the states of both `IntentRequest` and `SigningRequest` transactions must be stored on the controller chain.

```ts
interface IntentRequest {
   channelId: string
   action: string,
   expectedSender: string, // the expected sender of inboundTx on counterparty chain
   hash: string,
   status: Enum,
   inboundTx: bytes[],
   createAt: u64,
}

interface SigningRequest {
   channelId: string,
   action: string,  // can be defined in app, such as AtomicSwap, LSD
   hash: string,
   status: Enum,
   outboundTx: bytes[],
   createAt: u64,
}
```

To enhance security, `IntentRequests` have a time limit of 24 hours. Any assets deposited after this designated time frame will not be processed and, consequently, will be forfeited. Users must adhere to the 24-hour limit to ensure the successful completion of the asset deposit on the external chain. Additionally, only one open request is allowed per account.

There should be the following functions to facilitate obtaining transactions or iterating through transactions.

```ts
function getRequest(hash: string) {

}

function getRequests(status: Enum, start: int, limit: int) {

}
```


### Remote Transaction

Remote Transaction are mainly composed of channelId, direction, transaction, txResponse, height, proof. 

- Remote Transaction: 0x04 | ChannelId | TxHash | -> ProtocolBuffer(Transaction)

## Messages

In this section we describe the processing of messages for the CCA Core module.

### MsgCreateLightClient
```proto
message MsgCreateLightClient {
  option (cosmos.msg.v1.signer) = "proposer";
  option (amino.name)           = "cosmos-sdk/v1/MsgCreateLightClient";

  string id = 1;
  string channelId = 2;
  string client_type = 3;
  Header latest_header = 4;
  u64 comfirmation = 5;
  bool frozen = 6;
  repeated string validator = 7;
  u64 difficulty = 8;
}
```

> **NOTE**
> This message can only be executed when a proposal is passed.

### MsgCreateChannel
```proto
message MsgCreateChannel {
  option (cosmos.msg.v1.signer) = "proposer";
  option (amino.name)           = "cosmos-sdk/v1/MsgCreateChannel";

  string id = 1;
  string client_id = 2;
  string app_id = 4;
  string hd_path = 5;
}
```

> **NOTE**
> This message can only be executed when a proposal is passed.

The TSS network will generate a vault address based on the `chain_id` of the client and the `hd_path`. Once a vault address is assigned to the channel, the channel should be able to facilitate the bridging of messages between the connected blockchains.

The id is a hash of `client_id` and `app_id`, indicating that there should be only one channel available between an app and a vault. The one-to-one mapping design is implemented to isolate risks between apps.

Message handling should fail if:

- The provided `id` already exists.
- The provided `client_id` does not exist.
- The provided `hd_path` has been used in another channel.

### MsgUpdateClient

```proto
message MsgUpdateClient {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name)           = "cosmos-sdk/v1/MsgUpdateClient";

  string client_id = 1;
  Header header = 2;
  repeated string[] validators = 3;
}
```

Message handling should fail if:

- The provided `height` of the header is less than the `height` of `latestHeader` of the client.
- The provided `client_id` does not exist.
- `client.verify(header: Header)` returns `false`.

### MsgSubmitSignature

```proto
message MsgSubmitSignature {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name)           = "cosmos-sdk/v1/MsgSubmitSignature";

  string request_id = 1;
  string signature = 2;
}
```

Message handling should fail if:

- The provided `request_id` does not exist.
- The provided `signature` is not valid according to the request's outbound transaction and public key.

### MsgSubmitRemoteTranction

```proto
message MsgSubmitRemoteTranction {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name)           = "cosmos-sdk/v1/MsgSubmitRemoteTranction";

  string channel_id = 1;
  string direction = 2;  // In or Out
  u64: height;
  any transaction = 2;
  string proof = 3;
}
```

Message handling should fail if:

- The provided `request_id` does not exist.
- The provided `height` does not exist.
- The provided `proof` cannot prove that the `transaction` is included in the specified `height`.
- The provided `sender` of the transaction does not match the `expected_sender` of the Request.




  
