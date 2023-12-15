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

## Msgs

