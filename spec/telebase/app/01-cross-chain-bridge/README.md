### Cross chain Bridge

The Cross-Chain Bridge is a Telebase app that enables users to deposit assets from an external blockchain to mint Peggy assets on the SIDE blockchain. It also facilitates the burning of Peggy assets on the SIDE blockchain to withdraw native assets on the external blockchain.

Similar to many other bridge solutions, we wrap bridged assets into pegged assets with a 1:1 ratio. Users have the flexibility to mint pegged assets by initiating a `IntentRequest` or burn pegged assets through the execution of a `SigningRequest` to withdraw native assets.

### Definition

 - `Peggy Token`: A tokenized asset pegged on the SIDE blockchain, enabling smooth interoperability. This involves locking assets on a counterparty chain and minting equivalent tokens on the SIDE blockchain, facilitating seamless cross-chain asset transfers while maintaining a fixed value ratio. For consistency and clarity, all peggy tokens must adhere to the naming convention: they should commence with the prefix `side/` followed by a hash generated from `chain_id`, `vault address`, and `symbol`. This ensures a standardized and identifiable nomenclature for peggy tokens.

#### Transaction Flow 
![flow](./bridge_workflow.png)
