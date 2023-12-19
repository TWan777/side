# Routing

## Synopsis

The routing module is a default implementation of a secondary module responsible for accepting inbound transactions and invoking functions of the CCA core to ensure authenticity and transaction relay. This module maintains a lookup table of apps, allowing it to efficiently find and invoke the appropriate app when a transaction is received. External relayers only need to relay transactions to the routing module, simplifying the transaction processing workflow.

## Motivation

The default CCA handler adopts a receiver call pattern, requiring modules to individually call the CCA handler for tasks like binding to a channel, sending and receiving transactions, and more. While flexible and straightforward, this approach can be a bit intricate and may demand additional effort from relayer processes, which need to monitor the state of multiple modules. This standard introduces a CCA "routing module" designed to automate prevalent functionalities, route transactions, and streamline the responsibilities of relayers.

Additionally, the routing module can function as the module manager, incorporating logic to decide when modules are permitted to bind to channels.

## Definitions

All functions provided by the CCA handler interface are defined as in Baseapp.

## Desired Properties

 - Apps should seamlessly bind to channels and take ownership through the routing module.
 - The routing module should introduce no additional overhead for transaction sends and receives, except for the layer of call indirection.
 - The routing module must invoke specified handler functions on the app precisely when needed to act upon packets.

## Technical Specification

### Lookup Application

```ts
function lookupApp(channelId: Identifier) {
    return privateStore.get(callbackPath(channelId))
}
```

### Transaction Handler

**Inbound Transaction Lifescope Manager**


