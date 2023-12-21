# Channel

A channel functions as a conduit for interchain operations between a designated module on the controller blockchain and a CCA on a remote blockchain. It guarantees that the vault exclusively accepts transactions sent from the specified module. Additionally, it triggers a function of the CCA applications to update states by monitoring received transactions that interact with the CCA on the remote blockchain.

## Data Structure
```ts
interface Channel {
   id: string;
   // specific which client is used for verification of inbound transaction. 
   clientId: string;
   vaultAddress: string;
   appId: string;
   // verify if this transaction is included in the external chain.
   function verify(txHash: string, proof: byte[]);
}
```

