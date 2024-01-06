---
title: sideHub Router
stage: draft
category: IBC/APP
kind: instantiation
author: Marian <marian@side.one>
created: 2024-01-05
modified: 2024-01-05
requires: 24, 25
---

## Synopsis

This module will forward packets to connected any chains to the sideHub.

### Motivation

sideHub Router enables chains connected to an intermediary (router) chain to communicate and perform actions such as token transfers and contract calls without direct connections. This can expand the reachability of chains and enhance interoperability within the Cosmos ecosystem.

### Definitions

`sideHub Router`: A middleware that forwards IBC packets from a source chain to a destination chain, passing through an intermediary chain.

`CargoPacket`: A custom packet structure used by the router to encapsulate the data and instructions for forwarding.

### Desired Properties

- `Transparency`: The router should operate in a way that is transparent to the source and destination chains.
- `Reliability`: Packets forwarded by the router should be reliably delivered or a clear failure should be reported back to the source chain.
- `Statelessness`: The router should be as stateless as possible, not holding the state of other chains.

## Technical Specification

### General Design

The router will consist of a CosmWasm smart contract deployed on an intermediary chain. It will receive IBC packets from a source chain, inspect them to determine the destination, and then use IBC to send them to the destination chain.

### Data Structures

#### Data packets

```ts
interface IbcEndpoint {
  port_id: string;
  channel_id: string;
}
interface IbcTimeout {
  timestamp: bigint;
  height: string;
}
```

```ts
interface CargoPacket {
    srcEndpoint: IbcEndpoint;
    packetTimeout: IbcTimeout;
    retriesRemaining: number;
    packetData: []byte[];
    timeout: number;
}

interface ForwardablePacket {
  data: []byte; // data will to interchain-swap packet data.
  memo: string;
}

interface ForwardMetadata {
    target: IbcEndpoint;
    timeout: IbcTimeout;
    retries: number
}

interface PacketMetadata {
    forward: ForwardMetadata;
}

```

Once the setup function has been called, channels can be created via the IBC routing module.

#### Define IBCMiddleware

```typescript
class SideHubRouter {
    private ics4Wrapper: ICS4Wrapper;
    private channel: ChannelKeeper;
    private scope: ScopedKeeper
    private store: Map<string, Any>;
    function SendForwardablePacket(
	    srcEndpoint: IbcEndpoint,
        timeout: IbcTimeout,
	    data: byte[]
    ): number  {

        abortTransactionUnless(this.channel.GetChannel(srcEndpoint.portId, srcEndpoint.channelId))
        abortTransactionUnless(this.channel.GetNextSequenceSend(srcEndpoint.portId, srcEndpoint.channelId))

	    const {channelCap, ok } = this.scope.GetCapability(host.ChannelCapabilityPat(srcEndpoint.portId, srcEndpoint.channelId))
        abortTransactionUnless(ok);
	    const sequence = kics4Wrapper.SendPacket(ctx, channelCap, srcEndpoint.portId,  srcEndpoint.channelId, timeout.height, timeout.timestamp, data)
	    return &sequence
    }

    // GetAndClearCargoPacket will fetch an CargoPacket from the store, remove it if it exists, and return it.
    function GetAndClearCargoPacket(
    	channel string,
    	port string,
    	sequence uint64,
    ): CargoPacket|undefined {
    	const key := returnCargoPacketKey(channel, port, sequence)
    	if !store.Has(key) {
    		// this is either not a forwarded packet, or it is the final destination for the refund.
    		return undefined;
    	}
    	const bz = store.Get(key)
    	// done with packet key now, delete.
    	store.Delete(key)
    	const cargoPacket:CargoPacket =   protobuf.MustUnmarshal(bz)
    	return cargoPacket
    }


    function returnCargoPacketKey(channelID, portID string, sequence uint64):string {
        return `${channelID}/${portID}/${sequence}`;
    }



    function parseMemo(memo: string): {isRouted: boolean, ibcEndpoint: IbcEndpoint} {
        const memoJson = json.parse(memo);
        if(memoJson !== undefined) {
            return  {isRouted: false, ibcEndpoint: undefined};
        }
        return {
            portId: memoJson.portId,
            channelId: memoJson.channelId
        }
    }

}
```

```typescript
class IBCMiddleware {
    app: IBCModule;
    router: SideHubRouter;
    retriesOnTimeout: number;

    constructor(
        app: IBCModule,
        router: SideHubRouter,
        retriesOnTimeout: number
    ) {
        this.app = app;
        this.router = router;
        this.retriesOnTimeout = retriesOnTimeout
    }

    function OnChainOpenInit() {
        order: Order, // channeltypes module
        connectionHops: string[],
        portId: string,
        channelId: string,
        chanCap: Capability,        // capability module
        counterParty: CounterParty, // channeltypes module
        version: string
    } {
        return this.app.OnChainOpenInit(order, connectionHops, portId, channelId,chanCap, counterParty, version);
    }


    // OnChanOpenTry implements the IBCModule interface.
    function  OnChanOpenTry(
	    order Order,
	    connectionHops []string,
	    portID, channelID string,
	    chanCap Capability,
	    counterparty Counterparty,
	    counterpartyVersion string,
    ) (version string, err error) {
	    return this.app.OnChanOpenTry(order, connectionHops, portID, channelID, chanCap, counterparty, counterpartyVersion)
    }

    // OnChanOpenAck implements the IBCModule interface.
    function OnChanOpenAck(
    	portID, channelID string,
    	counterpartyChannelID string,
    	counterpartyVersion string,
    ) {
    	return this.app.OnChanOpenAck( portID, channelID, counterpartyChannelID, counterpartyVersion)
    }

    // OnChanOpenConfirm implements the IBCModule interface.
    function OnChanOpenConfirm(portID, channelID string) error {
    	return this.app.OnChanOpenConfirm(portID, channelID)
    }

    // OnChanCloseInit implements the IBCModule interface.
    function OnChanCloseInit(portID, channelID string) error {
    	return this.app.OnChanCloseInit(portID, channelID)
    }

    // OnChanCloseConfirm implements the IBCModule interface.
    function OnChanCloseConfirm(portID, channelID string) error {
    	return this.app.OnChanCloseConfirm(ctx, portID, channelID)
    }

    // OnRecvPacket checks the memo field on this packet and if the metadata inside's root key indicates this packet
    function OnRecvPacket(
    	packet channeltypes.Packet,
    	relayer sdk.AccAddress,
    ) Acknowledgement {
    	let data: ForwardablePacket = protobuf.UnmarshalJSON(packet.GetData())
        abortTransactionUnless(json.parse(data.Memo))
    	const m:PacketMetadata = json.parse(data.Memo);
    	const metadata := m.forward;
    	const timeout := time.Duration(metadata.Timeout)
        const retries = metadata.retries
        const {isRouted, targetEndpoint} = this.router.paseMemo(data.memo)
        if(isRouted) {
             abortTransactionUnless(this.router.SendForwardablePacket(targetEndpoint,data,"",retries,timeout))
        }else{
            this.app.onReceivedPacket(packet,relayer)
        }
    	return nil
    }

    // OnAcknowledgementPacket implements the IBCModule interface.
    function OnAcknowledgementPacket(
    	packet Packet, // channel types
    	acknowledgement []byte,
    	relayer string,
    ) {
        const data:ForwardablePacket = abortTransactionUnless(protobuf.UnmarshalJSON(packet.GetData()))

        const ack:Acknowledgement = abortTransactionUnless(protobuf.UnmarshalJSON(acknowledgement))
    	return this.app.OnAcknowledgementPacket(packet, acknowledgement, relayer)
    }


    // OnTimeoutPacket implements the IBCModule interface.
    function OnTimeoutPacket(packet channeltypes.Packet, relayer sdk.AccAddress) {
        try {
            const data:ForwardablePacket = protobuf.UnmarshalJSON(packet.GetData())

        }catch (e) {
            return this.app.OnTimeoutPacket(packet, relayer)
        }

	    inFlightPacket, err := im.keeper.TimeoutShouldRetry(ctx, packet)
	    if inFlightPacket != nil {
	    	if err != nil {
	    		im.keeper.RemoveInFlightPacket(ctx, packet)
	    		// this is a forwarded packet, so override handling to avoid refund from being processed on this chain.
	    		// WriteAcknowledgement with proxied ack to return success/fail to previous chain.
	    		return im.keeper.WriteAcknowledgementForForwardedPacket(ctx, packet, data, inFlightPacket, newErrorAcknowledgement(err))
	    	}
	    	// timeout should be retried. In order to do that, we need to handle this timeout to refund on this chain first.
	    	if err := im.app.OnTimeoutPacket(ctx, packet, relayer); err != nil {
	    		return err
	    	}
	    	return im.keeper.RetryTimeout(ctx, packet.SourceChannel, packet.SourcePort, data, inFlightPacket)
	    }

	    return im.app.OnTimeoutPacket(ctx, packet, relayer)
    }
}

```

## Example Implementation

https://github.com/sideprotocol/side

## Other Implementations

Coming soon.

## History

Oct 9, 2023 - Draft written

Jan 01, 2024 - Draft revised

## References

https://github.com/cosmos/ibc-apps/tree/main/middleware/packet-forward-middleware
https://github.com/cosmos/ibc/issues/126

## Copyright

All content herein is licensed under [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).
