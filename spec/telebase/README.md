# Telebase

## Introduction

This specification outlines a solution enabling users to manage crosschain assets without having to trust any third parties.

## Architecture
![Component](../bridge/architecture.png)


## Definition

 - `Light Client` empowers clients (applications, devices, blockchains, etc.) to engage with blockchains and efficiently verify the state on that blockchain through cryptographic methods, without the need to process the entire blockchain state.
 - `Threshold Signature Scheme (TSS)` enables users to establish a flexible threshold policy. With TSS technology, signing commands are replaced by distributed computations, eliminating the private key as a single point of failure. For instance, if three users each receive a share of the private signing key, at least two out of the three users must collaborate to construct the signature for a transaction.
 - `Vault` is an external account or a contract on counterparty blockchains used to store escrowed assets. It is controlled by a Threshold Signature Scheme (TSS).
 - `Channel`, A channel functions as a conduit for interchain operations between a designated module on the SIDE blockchain and a vault on an external blockchain. It guarantees that the vault exclusively accepts transactions sent from the specified module. Additionally, it triggers an acknowledgment function of the module to update states by monitoring received transactions that interact with the vault on the external blockchain.
 - `Relayer` is a permissionless off-chain process with the ability to read the state of and submit transactions to a defined set of ledgers using the SIDE bridge protocol.
 - `TSS network` is a collaborative network consisting of validators, each possessing a share of the TSS private key. This collective ownership is crucial for controlling the vault account effectively.
