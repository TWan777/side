# Omini Chain Bridge

## Introduction

This specification outlines a solution enabling users to bridge assets without having to trust any third parties.

## Architecture
![Component](./components.png)

## Definition

 - `Light Client` empowers clients (applications, devices, blockchains, etc.) to engage with blockchains and efficiently verify the state on that blockchain through cryptographic methods, without the need to process the entire blockchain state.
 - `Threshold Signature Scheme (TSS)` enables users to establish a flexible threshold policy. With TSS technology, signing commands are replaced by distributed computations, eliminating the private key as a single point of failure. For instance, if three users each receive a share of the private signing key, at least two out of the three users must collaborate to construct the signature for a transaction.
 - `Vault` is an external account on counterparty blockchains used to store escrowed assets. It is controlled by a Threshold Signature Scheme (TSS).
 - `Relayer` is a permissionless off-chain process with the ability to read the state of and submit transactions to a defined set of ledgers using the SIDE bridge protocol.
 - `TSS network` is a collaborative network consisting of validators, each possessing a share of the TSS private key. This collective ownership is crucial for controlling the vault account effectively.

## Technical Specification

Similar to many other bridge solutions, we wrap bridged assets into pegged assets with a 1:1 ratio. Anyone can mint pegged assets by initiating an `inbound transaction` or burn pegged assets by executing an `outbound transaction`.

### Transaction Flow 
![flow](./transaction%20flow.png)

### Light Client

The light client traces states on counterparty chains and can be implemented in two types: probabilistic finality for PoW consensus blockchains (e.g., Bitcoin) and deterministic finality for PoS blockchains (e.g., Cosmos and Ethereum). 

### TSS Network

Among the numerous Threshold Signature Schemes, the [Multi-Party Threshold Signature Scheme](https://github.com/bnb-chain/tss-lib) as the optimal choice due to its resharing feature. This feature enables the TSS network to reshare the shares of the private key when the validator set undergoes changes.

All validators are required to operate a TSS Node to be eligible for rewards. Similar to signing blocks, validators must sign a minimum of 80% of transactions within a slashing epoch. Failure to meet this criterion results in the loss of rewards, including their block rewards.

 - Setup
   
```go
// Set up the parameters
// Note: The `id` and `moniker` fields are for convenience to allow you to easily track participants.
// The `id` should be a unique string representing this party in the network and `moniker` can be anything (even left blank).
// The `uniqueKey` is a unique identifying key for this peer (such as its p2p public key) as a big.Int.
thisParty := tss.NewPartyID(id, moniker, validator_pubkey)
```

 - Coordination

The TSS network randomly selects a node as the leader in each epoch. This leader is responsible for handling tasks such as fetching signing requests and writing signatures back to the blockchain.

