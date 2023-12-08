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
 - `TSS Communitee` is a group of validators, each owning a share of the TSS private key used to control the vault account.

## Technical Specification

Similar to many other bridge solutions, we wrap bridged assets into pegged assets with a 1:1 ratio. Anyone can mint pegged assets by initiating an `inbound transaction` or burn pegged assets by executing an `outbound transaction`.

### Transaction Flow 
![flow](./transaction%20flow.png)
