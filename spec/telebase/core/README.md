# Telebase Core

`Telebase Core` is a Cosmos SDK module implementation empowering users to create and manage light clients, update clients, establish channels, update transaction signatures, and support diverse queries. Integration with this module facilitates seamless functionality within your application.


## State

### Client
   
Client is mainly composed of id, channelId, type, latestHeader, comfirmation, frozen, validators(optional), difficulty(optional),  where id is the unique identifier of the client.

- Client: 0x01 | classID | -> ProtocolBuffer(Client)

### Channel
   
Channel is mainly composed of id, clientId, appId, externalChainId, vaultAddress, where id is the unique identifier of the Channel.

- Channel: 0x02 | channelID | -> ProtocolBuffer(Channel)

### Request
   
Requests are created within specific modules to implement their unique business logic. There are two types of requests: `IntentRequest` and `SigningRequest`.

Telebase Core is responsible for monitoring the execution of `IntentRequest`, facilitating transactions that interoperate with the vault on an counterparty chain, and notifying the states to the specific sub app.

Telebase Core also handles the acceptance and verification of signatures in `SigningRequest`, tracing the execution result on the counterparty chain.

- Request: 0x03 | RequestId | -> ProtocolBuffer(Request)

## Messages

In this section we describe the processing of messages for the Telebase core module.

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
