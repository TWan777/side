# Channel

A channel functions as a conduit for interchain operations between a designated module on the controller blockchain and a CCA on a remote blockchain. It guarantees that the vault exclusively accepts transactions sent from the specified module. Additionally, it triggers a function of the CCA applications to update states by monitoring received transactions that interact with the CCA on the remote blockchain.

## Desired Properties
 - Each channel should be exclusive to a single application.
 - There should only exist one channel between a client and an application, identifiable by the hash of `clientId` and `appId`.
 - A channel can be closed via governance.
 - The primary purpose of a channel is to authenticate transactions.

## Technical Specification

### Data Structure
```ts
interface Channel {
   id: string;
   // specific which client is used for verification of inbound transaction. 
   clientId: string;
   vaultAddress: string;
   appId: string;
   status: string; // open and close
   // verify if this transaction is included in the external chain.
   function verify(txHash: string, proof: byte[]);
}
```

### Messages

Messages are defined in `CCA Core`
